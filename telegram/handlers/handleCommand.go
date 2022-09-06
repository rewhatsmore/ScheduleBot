package telegram

import (
	"context"
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
)

const commandStart = "start"
const commandMenu = "menu"
const insertFullName = "Для записи на тренировки введи свое имя и фамилию."

func checkUser(userID int64, queries *db.Queries) error {
	_, err := queries.GetUser(context.Background(), userID)
	return err
}

func HandleCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	err := checkUser(message.Chat.ID, queries)
	if err == nil {
		switch message.Command() {
		case commandStart:
			return handleStart(message.Chat.ID, bot, queries)
		case commandMenu:
			msg := listFunctions(queries, message.From.ID)
			return msg.SendMsg(bot)
		default:
			return handleUncnowCommand(message, bot)
		}
	} else if err == sql.ErrNoRows {
		return askUserName(bot, message.Chat.ID)
	} else {
		return err
	}
}

func handleStart(userID int64, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	msg := listFunctions(queries, userID)
	return msg.SendMsg(bot)
}

func listFunctions(queries *db.Queries, userID int64) *Msg {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Запись/отмена записи", listTrainings)},
		{tgbotapi.NewInlineKeyboardButtonData("Список моих тренировок", myTrainings)},
		{tgbotapi.NewInlineKeyboardButtonData("Узнать кто уже записан?", trainUserslList)},
	}

	user, err := queries.GetUser(context.Background(), userID)
	if err != nil {
		log.Println(err)
	}
	if user.IsAdmin {
		newRow := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Меню Админа", adminMenu),
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, newRow)
	}
	return &Msg{
		UserID:      userID,
		Text:        "Добро пожаловать в нашу дружную команду! Выбери действие:",
		ReplyMarkup: keyboard,
	}
}

//TODO: Unknown
func handleUncnowCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Извини, я пока не знаю эту команду.")
	_, err := bot.Send(msg)
	return err
}

func askUserName(bot *tgbotapi.BotAPI, userID int64) error {
	msg := &Msg{
		UserID: userID,
		Text:   insertFullName,
		ReplyMarkup: tgbotapi.ForceReply{
			ForceReply: true,
		},
	}
	return msg.SendMsg(bot)
}
