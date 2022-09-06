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

// создание текста тренировки для кнопки
func CreateTextOfTraining(dateAndTime time.Time, place string) string {
	engTime := dateAndTime.Format("Mon 02.01 в 15:04")
	time := translateWeekDay(engTime)
	return fmt.Sprintf("%s, %s", time, place)
}

func translateWeekDay(s string) string {
	dict := map[string]string{"Mon": "пн", "Tue": "вт", "Wed": "ср", "Thu": "чт", "Fri": "пт", "Sat": "сб", "Sun": "вс"}
	old := s[:3]
	new := dict[old]
	return strings.Replace(s, old, new, 1)
}
