package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	db "schedule.sqlc.dev/app/db/sqlc"
)

func Scheduler(queries *db.Queries, bot *tgbotapi.BotAPI) {
	c := cron.New()
	c.AddFunc("0 14 * * 0", func() {
		// every Sunday at 14-00
		createTable(queries, bot)
	})
	c.AddFunc("0 17 * * *", func() {
		// every day at 17-00
		alertUsers(queries, bot)
	})
	c.Start()
}

func alertUsers(queries *db.Queries, bot *tgbotapi.BotAPI) {
	usersForAlert, err := queries.ListUsersForAlert(context.Background())
	if err != nil {
		log.Println(err)
	}
	for _, userForAlert := range usersForAlert {
		text := fmt.Sprintf("Напоминалка! Завтра у тебя тренировка: 🥷 %s. Если у тебя изменились планы, пожалуйста, отмени свою запись.", CreateTextOfTraining(userForAlert.DateAndTime, userForAlert.Place))
		keyboard := tgbotapi.InlineKeyboardMarkup{}
		backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)
		msg := &Msg{
			UserID:      userForAlert.UserID,
			Text:        text,
			ReplyMarkup: keyboard,
		}

		err := msg.SendMsg(bot)
		if err != nil {
			log.Println(err)
		}
	}
}

func createTable(queries *db.Queries, bot *tgbotapi.BotAPI) {
	trainings, err := queries.ListLastWeekTrainings(context.Background())
	if err != nil {
		//TODO: send error messago to Regina
		log.Println(err)
	}
	for _, training := range trainings {
		arg := db.CreateTrainingParams{
			Place:       training.Place,
			DateAndTime: training.DateAndTime.Add(time.Hour * 24 * 7),
		}
		trainingNew, err := queries.CreateTraining(context.Background(), arg)
		log.Println("inserted:", trainingNew.TrainingID, trainingNew.DateAndTime)
		if err != nil {
			//TODO: send error messago to Regina
			log.Println(err)
		}
	}

	//TODO: create package `notifioncation`
	msg, err := listTrainingsForUser(queries, 0)
	if err != nil {
		//TODO: send error messago to Regina
		log.Println(err)
	}
	users, err := queries.ListUsers(context.Background())
	if err != nil {
		//TODO: send error messago to Regina
		log.Println(err)
	}

	for _, user := range users {
		msg.UserID = user.UserID
		err := msg.SendMsg(bot)
		if err != nil {
			log.Println(err)
		}
	}
}
