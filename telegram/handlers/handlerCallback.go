package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
	"schedule.sqlc.dev/app/google"
)

const makeApp = "ma"
const cancelApp = "ca"
const backMenu = "bc"
const listTrainings = "lt"
const listChildrenTrainings = "lc"
const myTrainings = "mt"
const trainUsersList = "tu"

// const childApointmentFlag = "ct"
const backMenuText = "⬅ назад в меню"

func HandleCallback(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	data := callBack.Data[:2]
	switch data {
	case makeApp:
		return handleTrainingAppointment(callBack, bot, queries)
	case cancelApp:
		return handleDeleteAppointment(callBack, bot, queries)
	case listTrainings:
		fmt.Println("1. сейчас будем формировать трени для юзера")
		msg, err := listTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return msg.UpdateMsg(bot, callBack.Message)
	case listChildrenTrainings:
		msg, err := listChildrenTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return msg.UpdateMsg(bot, callBack.Message)
	case backMenu:
		msg := listFunctions(queries, callBack.From.ID)
		return msg.UpdateMsg(bot, callBack.Message)
	case myTrainings:
		return listMyTrainings(bot, queries, callBack.Message)
	case trainUsersList:
		return listTrainingUsers(bot, queries, callBack.Message)
	case adminMenu:
		return listAdminFunctions(bot, callBack.Message)
	case adminListTr:
		return adminListTrainings(bot, queries, callBack.Message)
	case cancelCheck:
		return adminCancelCheck(callBack, bot)
	case cancelTraining:
		return adminCancelTraining(bot, queries, callBack)
	case adminDaT:
		return adminDateAndTimeRequest(callBack.From.ID, bot)
	case newAdultTraining, newChildTraining:
		return HandleNewTraining(callBack, queries, bot)
	default:
		return nil
	}
}

func handleTrainingAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	callbackText := callBack.Data[2:]

	callBackData := strings.Split(callbackText, ",")

	trainingId, err := strconv.Atoi(callBackData[0])
	if err != nil {
		return err
	}

	additionalChildNumber, err := strconv.Atoi(callBackData[1])
	if err != nil {
		return err
	}

	columnNumber, err := strconv.Atoi(callBackData[2])
	if err != nil {
		return err
	}

	//TODO: separate method and move below to new file AppointmentService
	// it should response with boolean, err and does not work with TG events
	arg := db.CreateAppointmentParams{
		TrainingID:            int64(trainingId),
		UserID:                callBack.From.ID,
		AdditionalChildNumber: int64(additionalChildNumber),
	}

	_, err = queries.CreateAppointment(context.Background(), arg)
	if err != nil {
		log.Println(err)
		return err
	}

	user, err := queries.GetUser(context.Background(), arg.UserID)
	if err != nil {
		log.Println(err)
	}

	err = google.AddAppointmentToTable(user.RowNumber, int64(columnNumber))
	if err != nil {
		log.Println(err)
	}

	var msg *Msg

	if arg.AdditionalChildNumber != -1 {
		msg, err = listChildrenTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
	} else {
		msg, err = listTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
	}

	//end TODO
	return msg.UpdateMsg(bot, callBack.Message)
}

// СДЕЛАТЬ!!!
func handleDeleteAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	callbackText := callBack.Data[2:]

	callbackData := strings.Split(callbackText, ",")

	additionalChildNumber, err := strconv.Atoi(callbackData[1])
	if err != nil {
		return err
	}

	columnNumber, err := strconv.Atoi(callbackData[2])
	if err != nil {
		return err
	}

	appointmentId, err := strconv.Atoi(callbackData[3])
	if err != nil {
		return err
	}

	err = queries.DeleteAppointment(context.Background(), int64(appointmentId))
	if err != nil {
		log.Println(err)
		return err
	}

	user, err := queries.GetUser(context.Background(), callBack.From.ID)
	if err != nil {
		log.Println(err)
	}

	err = google.DeleteAppointment(user.RowNumber, int64(columnNumber))
	if err != nil {
		log.Println(err)
	}

	msg, err := listTrainingsForUser(queries, callBack.From.ID)
	if err != nil {
		return err
	}

	if additionalChildNumber == -1 {
		return msg.UpdateMsg(bot, callBack.Message)
	}

	msg, err = listChildrenTrainingsForUser(queries, callBack.From.ID)
	if err != nil {
		return err
	}

	return msg.UpdateMsg(bot, callBack.Message)
}

// создание отправка списка тренировок на которые записан пользователь
func listMyTrainings(bot *tgbotapi.BotAPI, queries *db.Queries, message *tgbotapi.Message) error {
	msg := &Msg{
		Text:        "Твои тренировки:\n\n",
		ReplyMarkup: *backMenuKeyboard(),
	}

	userTrainings, err := queries.ListUserTrainings(context.Background(), message.Chat.ID)
	if err != nil {
		return err
	}

	for _, userTraining := range userTrainings {
		msg.Text += "🏅 " + CreateTextOfTraining(userTraining.DateAndTime) + "\n\n"
	}

	return msg.UpdateMsg(bot, message)
}

// создание отправка тренировок для записи и отмены
func listTrainingsForUser(queries *db.Queries, userID int64) (*Msg, error) {
	fmt.Println("Мы в методе формирования трень")
	msg := &Msg{
		UserID: userID,
		Text:   "Расписание на неделю. Выбери тренировки для записи. Повторное нажатие для отмены.",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}

	arg := db.ListTrainingsForSendParams{
		UserID:    userID,
		GroupType: db.GroupTypeEnumAdult,
	}

	fmt.Println("Сейчас запрошу трени для юзера")
	trainingsForSend, err := queries.ListTrainingsForSend(context.Background(), arg)
	if err != nil {
		return msg, err
	}
	fmt.Println("Запрошены. теперь клаву делаем")

	for _, trainingForSend := range trainingsForSend {

		callBackData := fmt.Sprintf("%d,%d,%d,%d", trainingForSend.TrainingID, trainingForSend.AdditionalChildNumber, trainingForSend.ColumnNumber, trainingForSend.AppointmentID)

		fmt.Println("Строка: " + callBackData)

		var row []tgbotapi.InlineKeyboardButton
		text := CreateTextOfTraining(trainingForSend.DateAndTime)
		data := makeApp + callBackData
		if trainingForSend.AppointmentID != 0 {
			text = "✅  " + text + " (вы записаны)"
			data = cancelApp + callBackData
			fmt.Println(data)
		} else {
			text = "☐  " + text
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		fmt.Println(text)
		fmt.Println(data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg.ReplyMarkup = keyboard

	return msg, nil
}

// создание отправка детских тренировок для записи и отмены
func listChildrenTrainingsForUser(queries *db.Queries, userID int64) (*Msg, error) {
	fmt.Println("Начало создания детских трень")
	msg := &Msg{
		UserID: userID,
		Text:   "Расписание дети!!! Поставь галочку для записи. Повторное нажатие для отмены.",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}

	arg := db.ListTrainingsForSendParams{
		UserID:    userID,
		GroupType: db.GroupTypeEnumChild,
	}

	fmt.Println("Начало получения списка. должно быть 2")

	trainingsForSend, err := queries.ListTrainingsForSend(context.Background(), arg)
	if err != nil {
		return msg, err
	}

	fmt.Println("Списко получен")
	for i, tr := range trainingsForSend {
		fmt.Println(i, tr)
	}

	for j, trainingForSend := range trainingsForSend {

		fmt.Println(j, "-я тренировка, id:", trainingForSend.TrainingID)

		textSlice := []string{
			"☐  " + CreateTextOfTraining(trainingForSend.DateAndTime) + " взр + реб",
			"☐  " + CreateTextOfTraining(trainingForSend.DateAndTime) + " 1 реб",
			"☐  " + CreateTextOfTraining(trainingForSend.DateAndTime) + " 2 реб"}

		for i, text := range textSlice {

			var row []tgbotapi.InlineKeyboardButton
			fmt.Printf("%d.%d запись\n", j, i)

			data := makeApp
			fmt.Println("	номер записи:", trainingForSend.AppointmentID, "номер детей:", trainingForSend.AdditionalChildNumber, "i:", i)
			if trainingForSend.AppointmentID != 0 && trainingForSend.AdditionalChildNumber == int64(i) {
				fmt.Println("запись с галочкой")
				text = strings.ReplaceAll(text+" (вы записаны)", "☐  ", "✅  ")
				data = cancelApp
			}

			callBackData := fmt.Sprintf("%d,%d,%d,%d", trainingForSend.TrainingID, i, trainingForSend.ColumnNumber, trainingForSend.AppointmentID)

			data += callBackData

			btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
			fmt.Println("text:", text, "data:", data)
			row = append(row, btn)
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		}

	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg.ReplyMarkup = keyboard

	return msg, nil
}

// Кто уже записан
func listTrainingUsers(bot *tgbotapi.BotAPI, queries *db.Queries, message *tgbotapi.Message) error {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg := &Msg{
		ReplyMarkup: keyboard,
	}

	fmt.Println("запрашиваем трени")

	adultTrainings, err := queries.ListAdultTrainings(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("взрослые есть")

	childTrainings, err := queries.ListChildrenTrainings(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("детские есть")

	if len(adultTrainings) == 0 && len(childTrainings) == 0 {
		msg.Text = "Пока расписания нет, но скоро обязательно появится!"
		return msg.UpdateMsg(bot, message)
	}

	//взрослые
	msg.Text += "<ins><strong>ВЗРОСЛЫЕ:</strong></ins>\n\n"
	for _, training := range adultTrainings {
		text := fmt.Sprintf("<ins>🏅 <strong>%s</strong></ins>\n", CreateTextOfTraining(training.DateAndTime))

		msg.Text += text
		users, err := queries.ListTrainingUsers(context.Background(), training.TrainingID)
		if err != nil {
			log.Panicln(err)
		}
		for i, user := range users {
			userText := fmt.Sprintf("        <em>%d. %s</em>\n", i+1, user.FullName)
			msg.Text += userText
		}
		msg.Text += "\n"
	}

	fmt.Println("Сформирован список взрослых")

	//дети
	msg.Text += "<ins><strong>ДЕТИ:</strong></ins>\n\n"
	for _, training := range childTrainings {
		text := fmt.Sprintf("<ins>🏅 <strong>%s (дети)</strong></ins>\n", CreateTextOfTraining(training.DateAndTime))

		msg.Text += text
		users, err := queries.ListTrainingUsers(context.Background(), training.TrainingID)
		if err != nil {
			log.Panicln(err)
		}
		for i, user := range users {

			textSlice := []string{"взр + реб", "1 реб", "2 реб"}
			userText := fmt.Sprintf("        <em>%d. %s (%s)</em>\n", i+1, user.FullName, textSlice[user.AdditionalChildNumber])
			msg.Text += userText
		}
		msg.Text += "\n"
	}
	return msg.UpdateMsg(bot, message)
}
