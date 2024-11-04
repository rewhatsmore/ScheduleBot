package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
	"schedule.sqlc.dev/app/google"
)

func HandleMessage(message *tgbotapi.Message, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	switch message.ReplyToMessage.Text {
	case insertFullName:
		return handleName(message, bot, queries)
	case insertDateAndTime:
		return adminTypeRequest(message, bot)
	case insertDateAndTimeAgain:
		return adminTypeRequest(message, bot)
	case insertMessageToAll:
		return adminSendMessageToAll(bot, message)
	case insertNewUserName:
		return adminNewUserTypeRequest(bot, message, queries)
	default:
		return nil
	}

}

func handleName(message *tgbotapi.Message, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	fullName := message.Text

	rowNumber, err := google.AddNewUserToTable(fullName)
	if err != nil {
		return errAddUserToSheet
	}

	google.AddUserToChildTable(fullName, rowNumber)

	arg := db.CreateUserParams{
		TelegramUserID: message.From.ID,
		FullName:       fullName,
		RowNumber:      rowNumber,
	}
	_, err = queries.CreateUser(context.Background(), arg)
	if err != nil {
		return err
	}
	msg := listFunctions(queries, message.From.ID)
	return msg.SendMsg(bot)
}

func HandleNewTraining(callback *tgbotapi.CallbackQuery, queries *db.Queries, bot *tgbotapi.BotAPI) error {
	msg := &Msg{
		UserID: callback.From.ID,
	}

	groupType := callback.Data[:2]
	text := callback.Data[2:]

	dateAndTime, err := time.Parse("02.01.2006 15:04", text)
	_ = err

	arg := db.CreateTrainingParams{
		DateAndTime: dateAndTime,
		GroupType:   db.GroupTypeEnumAdult,
	}

	if groupType == newChildTraining {
		arg.GroupType = db.GroupTypeEnumChild
	}

	columnNumber, err := google.AddTrainingToTable(arg.DateAndTime, arg.GroupType)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
	}

	arg.ColumnNumber = int64(columnNumber)

	_, err = queries.CreateTraining(context.Background(), arg)
	if err != nil {
		return err
	}

	msg.Text = "Тренеровка успешно добавлена"
	msg.ReplyMarkup = *backMenuKeyboard()

	return msg.UpdateMsg(bot, callback.Message)

}
