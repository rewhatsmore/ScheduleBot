package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
	"schedule.sqlc.dev/app/google"
)

// todo naming
const makeApp = "ma"
const cancelApp = "ca"
const backMenu = "bc"
const listTrainings = "lt"
const listChildrenTrainings = "lc"
const myTrainings = "mt"
const trainUsersList = "tu"
const refreshList = "rl"
const refreshChildrenList = "rc"
const adultListTrainingUsers = "tl"
const childListTrainingUsers = "ut"
const maxAppointments = 15

// const childApointmentFlag = "ct"
const backMenuText = "⬅ назад в меню"
const refreshListText = "🔄 обновить список"

func HandleCallback(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	data := callBack.Data[:2]
	switch data {
	case makeApp:
		msg, err := handleTrainingAppointment(callBack, queries)
		if err != nil {
			return err
		}
		return msg.UpdateMsg(bot, callBack.Message)
	case cancelApp:
		return handleDeleteAppointment(callBack, bot, queries)
	case listTrainings:
		fmt.Println("1. сейчас будем формировать трени для юзера")
		msg, err := listTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return msg.UpdateMsg(bot, callBack.Message)
	case refreshList:
		fmt.Println("1. сейчас будем формировать трени для юзера")
		msg, err := listTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return sendNewMessageAndDeleteOld(bot, msg, callBack.Message)
	case refreshChildrenList:
		msg, err := listChildrenTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return sendNewMessageAndDeleteOld(bot, msg, callBack.Message)
	case refreshGuestList:
		internalUserIdString := callBack.Data[2:]
		internalUserId, err := strconv.Atoi(internalUserIdString)
		if err != nil {
			return err
		}
		fmt.Println("1. сейчас будем формировать трени для юзера")
		msg, err := listTrainingsForGuest(queries, int32(internalUserId), callBack)
		if err != nil {
			return err
		}
		return sendNewMessageAndDeleteOld(bot, msg, callBack.Message)
	case refreshChildrenGuestList:
		internalUserIdString := callBack.Data[2:]
		internalUserId, err := strconv.Atoi(internalUserIdString)
		if err != nil {
			return err
		}
		msg, err := listChildrenTrainingsForGuest(queries, int32(internalUserId), callBack)
		if err != nil {
			return err
		}
		return sendNewMessageAndDeleteOld(bot, msg, callBack.Message)
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
		return typeTrainingListUsersRequest(bot, callBack)
	case adultListTrainingUsers:
		return listTrainingUsers(bot, queries, callBack.Message)
	case childListTrainingUsers:
		return listChildrenTrainingUsers(bot, queries, callBack.Message)
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
	case sendMessageToAll:
		return adminMessageToAllRequest(bot, callBack)
	case writeUserManually:
		return adminNewUserNameRequest(bot, callBack)
	case adminManagingGuests:
		return adminManageGuests(bot, queries, callBack)
	case manageGuest:
		return adminManageGuest(bot, callBack)
	case adminDeleteGuests:
		return adminDeleteGuest(bot, queries, callBack)
	case guestTypeTrainingRequest:
		return adminTypeGuestTrainingRequest(bot, callBack.Data[2:], callBack.From.ID)
	case adultGuestListTraining:
		internalUserIdString := callBack.Data[2:]
		internalUserId, err := strconv.Atoi(internalUserIdString)
		if err != nil {
			return err
		}
		msg, err := listTrainingsForGuest(queries, int32(internalUserId), callBack)
		if err != nil {
			return err
		}
		return msg.UpdateMsg(bot, callBack.Message)
	case childGuestListTraining:
		internalUserIdString := callBack.Data[2:]
		internalUserId, err := strconv.Atoi(internalUserIdString)
		if err != nil {
			return err
		}
		msg, err := listChildrenTrainingsForGuest(queries, int32(internalUserId), callBack)
		if err != nil {
			return err
		}
		return msg.UpdateMsg(bot, callBack.Message)
	case adminMakeGuestAppointment:
		return handleAdminAppointment(bot, queries, callBack)
	case adminDeleteGuestAppointment:
		return handleAdminDeleteAppointment(callBack, bot, queries)
	default:
		return nil
	}
}

func sendNewMessageAndDeleteOld(bot *tgbotapi.BotAPI, newMsg *Msg, oldMsg *tgbotapi.Message) error {
	// Отправляем новое сообщение
	err := newMsg.SendMsg(bot)
	if err != nil {
		return err
	}

	// Удаляем старое сообщение
	deleteMsg := tgbotapi.NewDeleteMessage(oldMsg.Chat.ID, oldMsg.MessageID)
	_, err = bot.Request(deleteMsg)
	if err != nil {
		return err
	}

	return nil
}

func handleTrainingAppointment(callBack *tgbotapi.CallbackQuery, queries *db.Queries) (*Msg, error) {
	var msg *Msg
	callbackText := callBack.Data[2:]

	callBackData := strings.Split(callbackText, ",")

	trainingId, err := strconv.Atoi(callBackData[0])
	if err != nil {
		return nil, err
	}

	additionalChildNumber, err := strconv.Atoi(callBackData[1])
	if err != nil {
		return nil, err
	}

	columnNumber, err := strconv.Atoi(callBackData[2])
	if err != nil {
		return nil, err
	}

	user, err := queries.GetUser(context.Background(), callBack.From.ID)
	if err != nil {
		log.Println(err)
	}

	var usersCount int
	if additionalChildNumber != -1 {
		count, err := queries.GetAppointmentCount(context.Background(), int64(trainingId))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		usersCount = int(count)
	} else {
		count, err := childrenAppointmentCount(queries, int64(trainingId))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		usersCount = count + 1
	}

	if usersCount < maxAppointments {
		arg := db.CreateAppointmentParams{
			TrainingID:            int64(trainingId),
			InternalUserID:        int64(user.InternalUserID),
			AdditionalChildNumber: int64(additionalChildNumber),
		}

		_, err = queries.CreateAppointment(context.Background(), arg)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = google.AddAppointmentToTable(user.RowNumber, int64(columnNumber), additionalChildNumber)
		if err != nil {
			log.Println(err)
		}
	}

	if additionalChildNumber != -1 {
		msg, err = listChildrenTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return nil, err
		}
	} else {
		msg, err = listTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return nil, err
		}
	}

	return msg, nil
}

// СДЕЛАТЬ!!!
func handleDeleteAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
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

	err = queries.DeleteAppointment(context.Background(), int64(appointmentId))
	if err != nil {
		log.Println(err)
		return err
	}

	user, err := queries.GetUser(context.Background(), callBack.From.ID)
	if err != nil {
		log.Println(err)
	}

	err = google.DeleteAppointment(user.RowNumber, int64(columnNumber), sheetName)
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
func listTrainingsForUser(queries *db.Queries, telegramUserID int64) (*Msg, error) {
	fmt.Println("Мы в методе формирования трень")
	msg := &Msg{
		UserID: telegramUserID,
		Text:   "Расписание на неделю. Выбери тренировки для записи. Повторное нажатие для отмены.",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
	refreshRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(refreshListText, refreshList)}

	user, err := queries.GetUser(context.Background(), telegramUserID)
	if err != nil {
		return nil, err
	}

	arg := db.ListTrainingsForSendParams{
		InternalUserID: int64(user.InternalUserID),
		GroupType:      db.GroupTypeEnumAdult,
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
		} else if trainingForSend.AppointmentCount < maxAppointments {
			text = "☐  " + text
		} else {
			text = "🚫  " + text + " (мест нет)"
			data = refreshList + callBackData
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		fmt.Println(text)
		fmt.Println(data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, refreshRow)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)
	msg.ReplyMarkup = keyboard

	return msg, nil
}

// создание отправка детских тренировок для записи и отмены
func listChildrenTrainingsForUser(queries *db.Queries, telegramUserID int64) (*Msg, error) {
	fmt.Println("Начало создания детских трень")
	msg := &Msg{
		UserID: telegramUserID,
		Text:   "Расписание дети!!! Поставь галочку для записи. Повторное нажатие для отмены.",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
	refreshRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(refreshListText, refreshChildrenList)}

	user, err := queries.GetUser(context.Background(), telegramUserID)
	if err != nil {
		return nil, err
	}

	arg := db.ListTrainingsForSendParams{
		InternalUserID: int64(user.InternalUserID),
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
		childCount, err := childrenAppointmentCount(queries, trainingForSend.TrainingID)
		if err != nil {
			return msg, err
		}
		if trainingForSend.AppointmentID == 0 && childCount >= (maxAppointments-1) {
			text := "🚫  " + textOfTraining + " (мест нет)"
			data := refreshChildrenList
			btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
			fmt.Println("text:", text, "data:", data)
			row := []tgbotapi.InlineKeyboardButton{btn}
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
			continue
		}

		textSlice := []string{
			"☐  " + textOfTraining + " взр + реб",
			"☐  " + textOfTraining + " 1 реб",
			"☐  " + textOfTraining + " 2 реб",
			"☐  " + textOfTraining + " взр + 2 реб",
		}

		for i, text := range textSlice {
			if (i == 0 || i == 3) && trainingForSend.DateAndTime.Weekday() == time.Sunday && trainingForSend.DateAndTime.Hour() == 13 {
				continue
			}

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
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, refreshRow)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg.ReplyMarkup = keyboard

	return msg, nil
}

func typeTrainingListUsersRequest(bot *tgbotapi.BotAPI, callBack *tgbotapi.CallbackQuery) error {
	msg := &Msg{
		UserID: callBack.From.ID,
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
	adultRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Взрослые", adultListTrainingUsers)}
	childRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Дети", childListTrainingUsers)}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, adultRow, childRow, backRow)
	msg.Text = "Посмотреть, кто записан:"
	msg.ReplyMarkup = keyboard

	return msg.UpdateMsg(bot, callBack.Message)
}

// Кто уже записан взрослые
func listTrainingUsers(bot *tgbotapi.BotAPI, queries *db.Queries, message *tgbotapi.Message) error {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
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

	if len(adultTrainings) == 0 {
		msg.Text = "Пока расписания нет, но скоро обязательно появится!"
		return msg.UpdateMsg(bot, message)
	}

	//взрослые
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

	return msg.UpdateMsg(bot, message)
}

// Кто уже записан дети
func listChildrenTrainingUsers(bot *tgbotapi.BotAPI, queries *db.Queries, message *tgbotapi.Message) error {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg := &Msg{
		ReplyMarkup: keyboard,
	}

	fmt.Println("запрашиваем трени")

	childTrainings, err := queries.ListChildrenTrainings(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("детские есть")

	if len(childTrainings) == 0 {
		msg.Text = "Пока расписания нет, но скоро обязательно появится!"
		return msg.UpdateMsg(bot, message)
	}

	//дети
	for _, training := range childTrainings {
		text := fmt.Sprintf("<ins>🏅 <strong>%s</strong></ins>\n", CreateTextOfTraining(training.DateAndTime))

		msg.Text += text
		users, err := queries.ListTrainingUsers(context.Background(), training.TrainingID)
		if err != nil {
			log.Panicln(err)
		}
		for i, user := range users {

			textSlice := []string{"взр + реб", "1 реб", "2 реб", "взр + 2 реб"}
			userText := fmt.Sprintf("        <em>%d. %s (%s)</em>\n", i+1, user.FullName, textSlice[user.AdditionalChildNumber])
			msg.Text += userText
		}
		msg.Text += "\n"
	}
	return msg.UpdateMsg(bot, message)
}

func childrenAppointmentCount(queries *db.Queries, trainingID int64) (int, error) {
	appointments, err := queries.ListAppointments(context.Background(), trainingID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, appointment := range appointments {
		if appointment.AdditionalChildNumber == 0 {
			count += 2
		} else {
			count += int(appointment.AdditionalChildNumber)
		}
	}

	fmt.Println(count)
	return count, nil
}
