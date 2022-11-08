package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
)

const adminMenu = "am"
const cancelTraining = "ct"
const adminListTr = "al"
const cancelCheck = "ch"
const adminDaT = "ad"
const insertDateAndTime = "Для создания тренировки введи с клавиатуры дату, время и место проведения тренировки как в примере:\n 02.01.2006 15:04/ зал Ninja Way"
const insertDateAndTimeAgain = "Данные введены в неверном формате. Попробуй еще раз. образец: 02.01.2006 15:04/ зал Ninja way"

//создание и отправка меню админа
func listAdminFunctions(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Отменить тренировку", adminListTr)},
		{tgbotapi.NewInlineKeyboardButtonData("Добавить тренировку", adminDaT)},
		{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)},
	}

	msg := &Msg{
		Text:        "Меню Администратора:",
		ReplyMarkup: keyboard,
	}

	return msg.UpdateMsg(bot, message)
}

//создание и отправка списка тренировок, для отмены
func adminListTrainings(bot *tgbotapi.BotAPI, queries *db.Queries, message *tgbotapi.Message) error {
	msg := &Msg{
		Text: "Выбери тренировку, чтобы отменить.",
	}
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)}

	trainings, err := queries.ListTrainings(context.Background())
	if err != nil {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)
		msg.ReplyMarkup = keyboard
		return msg.UpdateMsg(bot, message)
	}

	for _, training := range trainings {

		var row []tgbotapi.InlineKeyboardButton
		text := CreateTextOfTraining(training.DateAndTime)
		data := cancelCheck + training.DateAndTime.Format("/02.01 в 15:04/") + fmt.Sprintf("%d", training.TrainingID)

		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)
	msg.ReplyMarkup = keyboard

	return msg.UpdateMsg(bot, message)
}

// запрос времени и даты новой тренировки у админа
func adminDateAntTimeRequest(userID int64, bot *tgbotapi.BotAPI) error {
	msg := &Msg{
		UserID: userID,
		Text:   insertDateAndTime,
		ReplyMarkup: tgbotapi.ForceReply{
			ForceReply: true,
		},
	}
	return msg.SendMsg(bot)
}

//создание и отправка уточнения отмены тренировки
func adminCancelCheck(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	callBackData := strings.Split(callBack.Data, "/")
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Да", cancelTraining+callBack.Data[2:]),
		tgbotapi.NewInlineKeyboardButtonData("Нет", adminMenu),
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	msg := &Msg{
		UserID:      callBack.From.ID,
		Text:        fmt.Sprintf("Удалить тренировку %s ?", callBackData[1]),
		ReplyMarkup: keyboard,
	}

	return msg.UpdateMsg(bot, callBack.Message)
}

// Удаление тренировки
func adminCancelTraining(bot *tgbotapi.BotAPI, queries *db.Queries, callBack *tgbotapi.CallbackQuery) error {
	callBackData := strings.Split(callBack.Data, "/")
	dateTimeString := callBackData[1]
	trainingId, err := strconv.Atoi(callBackData[2])
	if err != nil {
		return err
	}

	trainingUsers, err := queries.ListTrainingUsers(context.Background(), int64(trainingId))
	if err != nil {
		return err
	}

	err = queries.DeleteTraining(context.Background(), int64(trainingId))
	if err != nil {
		return err
	}

	for _, trainingUser := range trainingUsers {
		text := fmt.Sprintf("Внимание!!! Отмена тренировки %s. Посмотри изменения в расписании и выбери другую тренировку, при необходимости.", dateTimeString)
		msg := tgbotapi.NewMessage(trainingUser.UserID, text)
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
	}

	return adminListTrainings(bot, queries, callBack.Message)
}

// todo: вынести оповещение в отдельную функцию
// func cancelTrainingAlert(bot *tgbotapi.BotAPI, trainingUsers []db.ListTrainingUsersRow, dateTimeString string) {
// 	for _, trainingUser := range trainingUsers {
// 		text := fmt.Sprintf("Внимание!!! Отмена тренировки %s. Посмотри изменения в расписании и выбери другую тренировку, при необходимости.", dateTimeString)
// 		msg := tgbotapi.NewMessage(trainingUser.UserID, text)
// 		_, err := bot.Send(msg)
// 		if err != nil {
// 			return err
// 		}
// 	}
// }
