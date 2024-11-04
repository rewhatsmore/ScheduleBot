package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
	"schedule.sqlc.dev/app/google"
)

const adminMenu = "am"
const newAdultTraining = "an"
const newChildTraining = "cn"
const sendMessageToAll = "sm"
const writeUserManually = "wm"
const cancelTraining = "ct"
const adminListTr = "al"
const cancelCheck = "cc"
const adminDaT = "at"
const manageGuest = "mo"
const adminManagingGuests = "mg"
const adminDeleteGuests = "dg"
const adminMakeGuestAppointment = "ag"
const adminDeleteGuestAppointment = "ad"
const guestTypeTrainingRequest = "tr"
const adultGuestListTraining = "lg"
const childGuestListTraining = "cl"
const refreshGuestList = "rg"
const refreshChildrenGuestList = "gl"
const insertDateAndTime = "Введи дату и время новой тренировки по шаблону:\n 02.01.2026 15:04"
const insertDateAndTimeAgain = "Данные введены в неверном формате. Попробуй еще раз. Образец: 02.01.2006 15:04"
const insertMessageToAll = "Введи сообщение, которое хочешь отправить всем:"
const insertNewUserName = "Введи имя гостя:"

// создание и отправка меню админа
func listAdminFunctions(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Отменить тренировку", adminListTr)},
		{tgbotapi.NewInlineKeyboardButtonData("Добавить тренировку", adminDaT)},
		{tgbotapi.NewInlineKeyboardButtonData("Отправить сообщение всем", sendMessageToAll)},
		{tgbotapi.NewInlineKeyboardButtonData("Добавить гостя вручную", writeUserManually)},
		{tgbotapi.NewInlineKeyboardButtonData("Управлять записью гостей", adminManagingGuests)},

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

// запрос типа тренировки для гостя
func adminTypeGuestTrainingRequest(bot *tgbotapi.BotAPI, internalUserIdString string, adminId int64) error {
	msg := &Msg{
		UserID: adminId,
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)}
	adultRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Взрослые", adultGuestListTraining+internalUserIdString)}
	childRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Дети", childGuestListTraining+internalUserIdString)}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, adultRow, childRow, backRow)
	msg.Text = "Выбери тип тренировки для запист гостя:"
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

	training, err := queries.GetTraining(context.Background(), int64(trainingId))
	if err != nil {
		log.Println(err)
	}

	trainingUsers, err := queries.ListTrainingUsers(context.Background(), int64(trainingId))
	if err != nil {
		return err
	}

	err = queries.DeleteTraining(context.Background(), int64(trainingId))
	if err != nil {
		return err
	}

	sheetName := "Adult"
	if training.GroupType == db.GroupTypeEnumChild {
		sheetName = "Child"
	}

	err = google.FillColumnWithColor(sheetName, training.ColumnNumber)
	if err != nil {
		return err
	}

	for _, trainingUser := range trainingUsers {
		text := fmt.Sprintf("Внимание!!! Отмена тренировки %s. Посмотри изменения в расписании и выбери другую тренировку, при необходимости.", dateTimeString)
		keyboard := tgbotapi.InlineKeyboardMarkup{}
		backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)
		msg := tgbotapi.NewMessage(trainingUser.TelegramUserID, text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
	}

	return adminListTrainings(bot, queries, callBack.Message)
}

func adminMessageToAllRequest(bot *tgbotapi.BotAPI, callBack *tgbotapi.CallbackQuery) error {
	msg := &Msg{
		UserID: callBack.From.ID,
		Text:   insertMessageToAll,
		ReplyMarkup: tgbotapi.ForceReply{
			ForceReply: true,
		},
	}
	return msg.SendMsg(bot)
}

func adminNewUserTypeRequest(bot *tgbotapi.BotAPI, message *tgbotapi.Message, queries *db.Queries) error {
	internalUserIdString, err := adminNewGuestUserAdd(message, queries)
	if err != nil {
		return err
	}

	return adminTypeGuestTrainingRequest(bot, fmt.Sprintf("%d", internalUserIdString), message.From.ID)
}

func adminNewGuestUserAdd(message *tgbotapi.Message, queries *db.Queries) (int32, error) {
	fullName := message.Text
	fmt.Println("Добавляем во взрослую таблицу")

	rowNumber, err := google.AddNewUserToTable(fullName)
	if err != nil {
		fmt.Println(err)
		return 0, errAddUserToSheet
	}
	fmt.Println("Добавляем в детскую таблицу")
	err = google.AddUserToChildTable(fullName, rowNumber)
	if err != nil {
		fmt.Println(err)
		return 0, errAddUserToSheet
	}

	arg := db.CreateUserParams{
		TelegramUserID: -1,
		FullName:       fullName,
		RowNumber:      rowNumber,
	}
	fmt.Println("Добавляем в базу")
	user, err := queries.CreateUser(context.Background(), arg)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return user.InternalUserID, nil
}

func adminSendMessageToAll(bot *tgbotapi.BotAPI, message *tgbotapi.Message, queries *db.Queries) error {
	users, err := queries.ListUsers(context.Background())
	if err != nil {
		return errNotificationDb
	}
	for _, user := range users {
		if !user.IsAdmin {
			continue
		}
		msg := &Msg{
			UserID: user.TelegramUserID,
			Text:   message.Text,
		}
		err := msg.SendMsg(bot)
		if err != nil {
			log.Println(err)
		}
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg := &Msg{
		UserID:      message.From.ID,
		Text:        "Сообщение отправлено всем активным пользователям",
		ReplyMarkup: keyboard,
	}
	return msg.SendMsg(bot)
}

func adminNewUserNameRequest(bot *tgbotapi.BotAPI, callBack *tgbotapi.CallbackQuery) error {
	msg := &Msg{
		UserID: callBack.From.ID,
		Text:   insertNewUserName,
		ReplyMarkup: tgbotapi.ForceReply{
			ForceReply: true,
		},
	}
	return msg.SendMsg(bot)
}

func adminManageGuests(bot *tgbotapi.BotAPI, queries *db.Queries, callBack *tgbotapi.CallbackQuery) error {
	guests, err := queries.ListGuests(context.Background())
	if err != nil {
		return err
	}
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)}

	for _, guest := range guests {
		callbackData := fmt.Sprintf("%s%d", manageGuest, guest.InternalUserID)
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(guest.FullName, callbackData),
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg := &Msg{
		UserID:      callBack.From.ID,
		Text:        "Список гостей",
		ReplyMarkup: keyboard,
	}
	return msg.UpdateMsg(bot, callBack.Message)
}

func adminManageGuest(bot *tgbotapi.BotAPI, callBack *tgbotapi.CallbackQuery) error {
	internalUserIdString := callBack.Data[2:]

	keyboard := tgbotapi.InlineKeyboardMarkup{}

	keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Удалить гостя", adminDeleteGuests+internalUserIdString)},
		{tgbotapi.NewInlineKeyboardButtonData("Записать/отменить запись", guestTypeTrainingRequest+internalUserIdString)},
		{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)},
	}
	msg := &Msg{
		UserID:      callBack.From.ID,
		Text:        "Выбери действие",
		ReplyMarkup: keyboard,
	}
	return msg.UpdateMsg(bot, callBack.Message)
}

func adminDeleteGuest(bot *tgbotapi.BotAPI, queries *db.Queries, callBack *tgbotapi.CallbackQuery) error {
	internalUserIdString := callBack.Data[2:]
	internalUserID, err := strconv.Atoi(internalUserIdString)
	if err != nil {
		return err
	}

	user, err := queries.GetUserByInternalID(context.Background(), int32(internalUserID))
	if err != nil {
		return err
	}

	err = queries.DeleteUser(context.Background(), int32(internalUserID))
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	err = google.MarkRowAsDeleted("Adult", user.RowNumber)
	if err != nil {
		fmt.Println(err)
	}

	err = google.MarkRowAsDeleted("Child", user.RowNumber)
	if err != nil {
		fmt.Println(err)
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData(backMenuText, adminMenu)},
	}

	msg := &Msg{
		UserID:      callBack.From.ID,
		Text:        "Гость успешно удален",
		ReplyMarkup: keyboard,
	}
	return msg.UpdateMsg(bot, callBack.Message)
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

func listTrainingsForGuest(queries *db.Queries, internalUserID int32, callBack *tgbotapi.CallbackQuery) (*Msg, error) {
	fmt.Println("Мы в методе формирования трень")
	msg := &Msg{
		UserID: callBack.From.ID,
		Text:   "Выбери тренировки для записи гостя. Повторное нажатие для отмены.",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	readyButton := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Готово", adminMenu)}

	arg := db.ListTrainingsForSendParams{
		InternalUserID: int64(internalUserID),
		GroupType:      db.GroupTypeEnumAdult,
	}

	fmt.Println("Сейчас запрошу трени для юзера")
	trainingsForSend, err := queries.ListTrainingsForSend(context.Background(), arg)
	if err != nil {
		return msg, err
	}
	fmt.Println("Запрошены. теперь клаву делаем")

	for _, trainingForSend := range trainingsForSend {
		callBackData := fmt.Sprintf("%d,%d,%d,%d,%d",
			trainingForSend.TrainingID,
			trainingForSend.AdditionalChildNumber,
			trainingForSend.ColumnNumber,
			trainingForSend.AppointmentID,
			internalUserID)

		fmt.Println("Строка: " + callBackData)

		var row []tgbotapi.InlineKeyboardButton
		text := CreateTextOfTraining(trainingForSend.DateAndTime)
		data := adminMakeGuestAppointment + callBackData
		if trainingForSend.AppointmentID != 0 {
			text = "✅  " + text + " (записан)"
			data = adminDeleteGuestAppointment + callBackData
			fmt.Println(data)
		} else if trainingForSend.AppointmentCount < maxAppointments {
			text = "☐  " + text
		} else {
			text = "🚫  " + text + " (мест нет)"
			data = refreshGuestList + fmt.Sprintf("%d", internalUserID)
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		fmt.Println(text)
		fmt.Println(data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, readyButton)
	msg.ReplyMarkup = keyboard

	return msg, nil
}

// создание отправка детских тренировок для записи и отмены
func listChildrenTrainingsForGuest(queries *db.Queries, internalUserID int32, callBack *tgbotapi.CallbackQuery) (*Msg, error) {
	fmt.Println("Начало создания детских трень")
	msg := &Msg{
		UserID: callBack.From.ID,
		Text:   "Выбери детские тренировки для записи гостя. Повторное нажатие для отмены.",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	readyButton := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Готово", adminMenu)}

	arg := db.ListTrainingsForSendParams{
		InternalUserID: int64(internalUserID),
		GroupType:      db.GroupTypeEnumChild,
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
		textOfTraining := CreateTextOfTraining(trainingForSend.DateAndTime)
		if trainingForSend.AppointmentID == 0 && trainingForSend.AppointmentCount >= 15 {
			text := "🚫  " + textOfTraining + " (мест нет)"
			data := refreshChildrenGuestList + fmt.Sprintf("%d", internalUserID)
			btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
			fmt.Println("text:", text, "data:", data)
			row := []tgbotapi.InlineKeyboardButton{btn}
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
			continue
		}

		textSlice := []string{
			"☐  " + textOfTraining + " взр + реб",
			"☐  " + textOfTraining + " 1 реб",
			"☐  " + textOfTraining + " 2 реб"}

		for i, text := range textSlice {

			var row []tgbotapi.InlineKeyboardButton
			fmt.Printf("%d.%d запись\n", j, i)

			data := adminMakeGuestAppointment
			fmt.Println("	номер записи:", trainingForSend.AppointmentID, "номер детей:", trainingForSend.AdditionalChildNumber, "i:", i)
			if trainingForSend.AppointmentID != 0 && trainingForSend.AdditionalChildNumber == int64(i) {
				fmt.Println("запись с галочкой")
				text = strings.ReplaceAll(text+" (вы записаны)", "☐  ", "✅  ")
				data = adminDeleteGuestAppointment
			}

			callBackData := fmt.Sprintf("%d,%d,%d,%d,%d", trainingForSend.TrainingID, i, trainingForSend.ColumnNumber, trainingForSend.AppointmentID, internalUserID)

			data += callBackData

			btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
			fmt.Println("text:", text, "data:", data)
			row = append(row, btn)
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		}

	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, readyButton)

	msg.ReplyMarkup = keyboard

	return msg, nil
}

func handleAdminAppointment(bot *tgbotapi.BotAPI, queries *db.Queries, callBack *tgbotapi.CallbackQuery) error {
	var msg *Msg
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

	sheetName := "Adult"
	if additionalChildNumber != -1 {
		sheetName = "Child"
	}

	columnNumber, err := strconv.Atoi(callBackData[2])
	if err != nil {
		return err
	}

	internalUserID, err := strconv.Atoi(callBackData[4])
	if err != nil {
		return err
	}

	user, err := queries.GetUserByInternalID(context.Background(), int32(internalUserID))
	if err != nil {
		log.Println(err)
		return err
	}

	usersCount, err := queries.GetAppointmentCount(context.Background(), int64(trainingId))
	if err != nil {
		log.Println(err)
		return err
	}

	if usersCount < maxAppointments {
		arg := db.CreateAppointmentParams{
			TrainingID:            int64(trainingId),
			InternalUserID:        int64(internalUserID),
			AdditionalChildNumber: int64(additionalChildNumber),
		}

		_, err = queries.CreateAppointment(context.Background(), arg)
		if err != nil {
			log.Println(err)
			return err
		}

		err = google.AddAppointmentToTable(user.RowNumber, int64(columnNumber), sheetName)
		if err != nil {
			log.Println(err)
		}
	}

	if additionalChildNumber != -1 {
		msg, err = listChildrenTrainingsForGuest(queries, int32(internalUserID), callBack)
		if err != nil {
			return err
		}
	} else {
		msg, err = listTrainingsForGuest(queries, user.InternalUserID, callBack)
		if err != nil {
			return err
		}
	}

	return msg.UpdateMsg(bot, callBack.Message)
}

func handleAdminDeleteAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	callbackText := callBack.Data[2:]

	callbackData := strings.Split(callbackText, ",")

	additionalChildNumber, err := strconv.Atoi(callbackData[1])
	if err != nil {
		return err
	}

	sheetName := "Adult"
	if additionalChildNumber != -1 {
		sheetName = "Child"
	}

	columnNumber, err := strconv.Atoi(callbackData[2])
	if err != nil {
		return err
	}

	appointmentId, err := strconv.Atoi(callbackData[3])
	if err != nil {
		return err
	}

	internalUserID, err := strconv.Atoi(callbackData[4])
	if err != nil {
		return err
	}

	err = queries.DeleteAppointment(context.Background(), int64(appointmentId))
	if err != nil {
		log.Println(err)
		return err
	}

	user, err := queries.GetUserByInternalID(context.Background(), int32(internalUserID))
	if err != nil {
		log.Println(err)
	}

	err = google.DeleteAppointment(user.RowNumber, int64(columnNumber), sheetName)
	if err != nil {
		log.Println(err)
	}

	msg, err := listTrainingsForGuest(queries, user.InternalUserID, callBack)
	if err != nil {
		return err
	}

	if additionalChildNumber == -1 {
		return msg.UpdateMsg(bot, callBack.Message)
	}

	msg, err = listChildrenTrainingsForGuest(queries, int32(internalUserID), callBack)
	if err != nil {
		return err
	}

	return msg.UpdateMsg(bot, callBack.Message)
}
