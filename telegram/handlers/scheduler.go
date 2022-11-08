package telegram

import (
	"context"
	"fmt"
	"log"
	"schedule.sqlc.dev/app/conf"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	db "schedule.sqlc.dev/app/db/sqlc"
)

func Scheduler(queries *db.Queries, bot *tgbotapi.BotAPI, config conf.Config) {
	c := cron.New()
	_, err := c.AddFunc("0 14 * * 0", func() {
		// every Sunday at 14-00
		err := createSchedule(queries)
		if err != nil {
			HandleError(config.AdminID, err)
			return
		}
		err = ScheduleNotification(queries, bot)
		if err != nil {
			HandleError(config.AdminID, err)
		}
	})
	if err != nil {
		HandleError(config.AdminID, err)
	}
	_, err = c.AddFunc("0 17 * * *", func() {
		// every day at 17-00
		err := trainingNotification(queries, bot)
		if err != nil {
			HandleError(config.AdminID, err)
		}
	})
	if err != nil {
		HandleError(config.AdminID, err)
	}
	c.Start()
}

func trainingNotification(queries *db.Queries, bot *tgbotapi.BotAPI) error {
	usersForAlert, err := queries.ListUsersForAlert(context.Background())
	if err != nil {
		return errNotificationDb //no urgent
	}
	for _, userForAlert := range usersForAlert {
		text := fmt.Sprintf("Напоминалка! Завтра у тебя тренировка: 🥷 %s. Если у тебя изменились планы, пожалуйста, отмени свою запись.", CreateTextOfTraining(userForAlert.DateAndTime))
		msg := &Msg{
			UserID:      userForAlert.UserID,
			Text:        text,
			ReplyMarkup: backMenuKeyboard(),
		}

		err := msg.SendMsg(bot)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func createSchedule(queries *db.Queries) error {
	var haveErrors error
	trainings, err := queries.ListLastWeekTrainings(context.Background())
	if err != nil || len(trainings) == 0 {
		return errCreateSchedule
	}
	for _, training := range trainings {
		arg := db.CreateTrainingParams{
			Place:       training.Place,
			DateAndTime: training.DateAndTime.Add(time.Hour * 24 * 7),
		}
		trainingNew, err := queries.CreateTraining(context.Background(), arg)
		log.Println("inserted:", trainingNew.TrainingID, trainingNew.DateAndTime)
		if err != nil {
			log.Println(err)
			haveErrors = errCreateSchedule
		}
	}
	return haveErrors
}

// ScheduleNotification sens schedule to user, works only if createSchedule completed successfully
func ScheduleNotification(queries *db.Queries, bot *tgbotapi.BotAPI) error {
	msg, err := listTrainingsForUser(queries, 0)
	if err != nil {
		return errNotificationDb
	}
	users, err := queries.ListUsers(context.Background())
	if err != nil {
		return errNotificationDb
	}

	for _, user := range users {
		msg.UserID = user.UserID
		err := msg.SendMsg(bot)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
