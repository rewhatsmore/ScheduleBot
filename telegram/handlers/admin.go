package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
)

const adminMenu = "am"
const newAdultTraining = "an"
const newChildTraining = "cn"
const cancelTraining = "ct"
const adminListTr = "al"
const cancelCheck = "cc"
const adminDaT = "at"
const insertDateAndTime = "Введи дату и время новой тренировки по шаблону. Д(дети), В(взрослые):\n 02.01.2026 15:04 В"
const insertDateAndTimeAgain = "Данные введены в неверном формате. Попробуй еще раз. Образец: 02.01.2006 15:04 В"

// создание и отправка меню админа
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

// создание и отправка списка тренировок, для отмены
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
		if training.GroupType == db.GroupTypeEnumChild {
			text += " (дети)"
		}
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
func adminDateAndTimeRequest(userID int64, bot *tgbotapi.BotAPI) error {
	msg := &Msg{
		UserID: userID,
		Text:   insertDateAndTime,
		ReplyMarkup: tgbotapi.ForceReply{
			ForceReply: true,
		},
	}
	return msg.SendMsg(bot)
}

// запрос типа новой тренировки у админа
func adminTypeRequest(message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	msg := &Msg{
		UserID: message.From.ID,
	}

	dateAndTime, err := time.Parse("02.01.2006 15:04", message.Text)
	if err != nil || dateAndTime.Before(time.Now()) {
		msg.Text = insertDateAndTimeAgain
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
		}
		return msg.SendMsg(bot)
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)}
	adultRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Взрослые", newAdultTraining+message.Text)}
	childRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Дети", newChildTraining+message.Text)}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, adultRow, childRow, backRow)
	msg.Text = "В какое расписание добавить тренировку?"
	msg.ReplyMarkup = keyboard

	return msg.SendMsg(bot)
}

// создание и отправка уточнения отмены тренировки
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
