package telegram

import (
	"context"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
)

func HandleMessage(message *tgbotapi.Message, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	switch message.ReplyToMessage.Text {
	case insertFullName:
		return handleName(message, bot, queries)
	case insertDateAndTime:
		return handleNewTraining(message, queries, bot)
	case insertDateAndTimeAgain:
		return handleNewTraining(message, queries, bot)
	default:
		return nil
	}

}

func handleName(message *tgbotapi.Message, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	fullName := message.Text
	arg := db.CreateUserParams{
		UserID:   message.From.ID,
		FullName: fullName,
	}
	_, err := queries.CreateUser(context.Background(), arg)
	if err != nil {
		return err
	}
	msg := listFunctions(queries, message.From.ID)
	return msg.SendMsg(bot)
}

func handleNewTraining(message *tgbotapi.Message, queries *db.Queries, bot *tgbotapi.BotAPI) error {
	msg := &Msg{
		UserID: message.From.ID,
	}
	inputData := strings.Split(message.Text, "/")
	dateAndTime, err := time.Parse("02.01.2006 15:04", inputData[0])
	if err != nil || dateAndTime.Before(time.Now()) {
		msg.Text = "Данные введены в неверном формате. Попробуй еще раз. образец: 02.01.2006 15:04/ зал Ninja way"
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
		}
		return msg.SendMsg(bot)
	}
	place := inputData[1]

	arg := db.CreateTrainingParams{
		Place:       place,
		DateAndTime: dateAndTime,
	}
	_, err = queries.CreateTraining(context.Background(), arg)
	if err != nil {
		return err
	}
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	msg.Text = "Тренеровка успешно добавлена"
	msg.ReplyMarkup = keyboard

	return msg.SendMsg(bot)

}
