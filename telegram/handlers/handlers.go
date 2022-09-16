package telegram

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Msg struct {
	UserID      int64
	Text        string
	ReplyMarkup interface{}
}

func (msg *Msg) SendMsg(bot *tgbotapi.BotAPI) error {
	message := tgbotapi.NewMessage(msg.UserID, msg.Text)
	message.ReplyMarkup = msg.ReplyMarkup
	_, err := bot.Send(message)
	return err
}

func (msg *Msg) UpdateMsg(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, msg.Text, msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup))
	_, err := bot.Send(editMsg)
	return err
}

// CreateTextOfTraining creates text for button
func CreateTextOfTraining(dateAndTime time.Time, place string) string {
	engTime := dateAndTime.Format("Mon 02.01 в 15:04")
	dateTime := translateWeekDay(engTime)
	return fmt.Sprintf("%s, %s", dateTime, place)
}

func translateWeekDay(s string) string {
	dict := map[string]string{"Mon": "пн", "Tue": "вт", "Wed": "ср", "Thu": "чт", "Fri": "пт", "Sat": "сб", "Sun": "вс"}
	oldWD := s[:3]
	newWD := dict[oldWD]
	return strings.Replace(s, oldWD, newWD, 1)
}
