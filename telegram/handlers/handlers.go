package telegram

import (
	"context"
	"database/sql"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
)

type Msg struct {
	UserID      int64
	Text        string
	ReplyMarkup interface{}
}

func HandleDeleteUser(userID int64, queries *db.Queries) error {
	err := queries.DeleteUser(context.Background(), userID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (msg *Msg) SendMsg(bot *tgbotapi.BotAPI) error {
	message := tgbotapi.NewMessage(msg.UserID, msg.Text)
	if msg.ReplyMarkup != nil {
		message.ReplyMarkup = msg.ReplyMarkup
	}
	_, err := bot.Send(message)
	return err
}

func (msg *Msg) UpdateMsg(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, msg.Text, msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup))
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(editMsg)
	return err
}

// CreateTextOfTraining creates text for button
func CreateTextOfTraining(dateAndTime time.Time) string {
	engTime := dateAndTime.Format("Mon 02.01 в 15:04")
	dateTime := translateWeekDay(engTime)
	return dateTime
}

func translateWeekDay(s string) string {
	dict := map[string]string{"Mon": "пн", "Tue": "вт", "Wed": "ср", "Thu": "чт", "Fri": "пт", "Sat": "сб", "Sun": "вс"}
	oldWD := s[:3]
	newWD := dict[oldWD]
	return strings.Replace(s, oldWD, newWD, 1)
}

//func emptyKeyboard() *tgbotapi.InlineKeyboardMarkup {
//	keyboard := tgbotapi.InlineKeyboardMarkup{}
//	row := []tgbotapi.InlineKeyboardButton{}
//	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
//	return &keyboard
//}

func backMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	return &keyboard
}
