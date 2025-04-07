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
const typeOfChildrenAppointment = "tc"

const maxAppointments = 15
const maxWednesdayAppointments = 12

const backMenuText = "⬅ назад в меню"
const refreshListText = "🔄 обновить список"

var childTexts = []string{
	" взр + реб",
	" 1 реб",
	" 2 реб",
	" взр + 2 реб",
}

func HandleCallback(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	data := callBack.Data[:2]
	switch data {
	case makeApp:
		msg, err := handleTrainingAppointment(callBack, queries)
		if err != nil {
			return err
		}
		msg.UserID = callBack.Message.Chat.ID
		return sendNewMessageAndDeleteOld(bot, msg, callBack.Message)
	case cancelApp:
		return handleDeleteAppointment(callBack, bot, queries)
	case listTrainings:
		msg, err := listTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return msg.UpdateMsg(bot, callBack.Message)
	case refreshList:
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
	case typeOfChildrenAppointment:
		return listChildrenAppointmentOptions(callBack, bot, queries)
	default:
		return nil
	}
}

func listChildrenAppointmentOptions(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {

	data := callBack.Data[2:]
	callbackData := strings.Split(data, ",")

	trainingId, err := strconv.Atoi(callbackData[0])
	if err != nil {
		return err
	}

	training, err := queries.GetTraining(context.Background(), int64(trainingId))
	if err != nil {
		return err
	}

	overallAppointmentCount, err := strconv.Atoi(callbackData[3])
	if err != nil {
		return err
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад", listChildrenTrainings)}

	keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("1 ребёнок", makeApp+fmt.Sprintf("%d,%s", 1, data))},
	}

	//less children on wednesday
	if overallAppointmentCount <= maxWednesdayAppointments-2 || (training.DateAndTime.Weekday() != time.Wednesday && overallAppointmentCount <= maxAppointments-2) {
		twoChildButton := tgbotapi.NewInlineKeyboardButtonData("2 ребёнка", makeApp+fmt.Sprintf("%d,%s", 2, data))
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{twoChildButton})
		if training.DateAndTime.Weekday() == time.Saturday || (training.DateAndTime.Weekday() == time.Sunday && training.DateAndTime.Hour() == 10) {
			childAndAdultButton := tgbotapi.NewInlineKeyboardButtonData("Взрослый + ребёнок", makeApp+fmt.Sprintf("%d,%s", 0, data))
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{childAndAdultButton})
		}
	}

	if overallAppointmentCount <= maxAppointments-3 && (training.DateAndTime.Weekday() == time.Saturday || (training.DateAndTime.Weekday() == time.Sunday && training.DateAndTime.Hour() == 10)) {
		twoChildAdAdultButton := tgbotapi.NewInlineKeyboardButtonData("Взрослый + 2 ребёнка", makeApp+fmt.Sprintf("%d,%s", 3, data))
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{twoChildAdAdultButton})
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg := &Msg{
		Text:        fmt.Sprintf("Кого записать на <ins><strong>%s</strong></ins>", CreateTextOfTraining(training.DateAndTime)),
		ReplyMarkup: keyboard,
	}

	return msg.UpdateMsg(bot, callBack.Message)
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

	additionalChildNumber, err := strconv.Atoi(callBackData[0])
	if err != nil {
		return nil, err
	}

	trainingId, err := strconv.Atoi(callBackData[1])
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
		count, err := childrenAppointmentCount(queries, int64(trainingId))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		usersCount = count + 1
	} else {
		count, err := queries.GetAppointmentCount(context.Background(), int64(trainingId))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if count < 0 {
			fmt.Println("count is less than 0")
		}
		usersCount = int(count)
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

	additionalChildNumber, err := strconv.Atoi(callbackData[0])
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

	user := db.User{
		TelegramUserID: telegramUserID,
		InternalUserID: 0,
	}
	var err error
	if telegramUserID != 0 {
		user, err = queries.GetUser(context.Background(), telegramUserID)
		if err != nil {
			return nil, err
		}
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
		callBackData := fmt.Sprintf("%d,%d,%d,%d", trainingForSend.AdditionalChildNumber, trainingForSend.TrainingID, trainingForSend.ColumnNumber, trainingForSend.AppointmentID)

		fmt.Println("Строка: " + callBackData)

		var row []tgbotapi.InlineKeyboardButton
		text := CreateTextOfTraining(trainingForSend.DateAndTime)
		data := makeApp + callBackData
		if trainingForSend.AppointmentID != 0 {
			if trainingForSend.DateAndTime.Before(time.Now().Add(4 * time.Hour)) {
				continue
			}
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

	trainingsForSend, err := queries.ListTrainingsForSend(context.Background(), arg)
	if err != nil {
		return msg, err
	}

	for i, tr := range trainingsForSend {
		fmt.Println(i, tr)
	}

	for _, trainingForSend := range trainingsForSend {

		childCount, err := childrenAppointmentCount(queries, trainingForSend.TrainingID)
		if err != nil {
			return msg, err
		}
		text := CreateTextOfTraining(trainingForSend.DateAndTime)
		data := fmt.Sprintf("%d,%d,%d", trainingForSend.TrainingID, trainingForSend.ColumnNumber, trainingForSend.AppointmentID)
		if trainingForSend.AppointmentID != 0 {
			if trainingForSend.DateAndTime.Before(time.Now().Add(4 * time.Hour)) {
				continue
			}
			text = "✅  " + text + " (" + childTexts[trainingForSend.AdditionalChildNumber] + ")"
			data = fmt.Sprintf("%d,%s", trainingForSend.AdditionalChildNumber, data)
			data = cancelApp + data
		} else if trainingForSend.AppointmentID == 0 && (childCount >= (maxAppointments) || (trainingForSend.DateAndTime.Weekday() == time.Wednesday && childCount >= (maxWednesdayAppointments))) {
			text = "🚫  " + text + " (мест нет)"
			data = refreshChildrenList
		} else {
			text = "☐  " + text
			data = fmt.Sprintf("%s,%d", data, childCount)

			data = typeOfChildrenAppointment + data
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		row := []tgbotapi.InlineKeyboardButton{btn}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
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
	refreshRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(refreshListText, adultListTrainingUsers)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, refreshRow, backRow)

	msg := &Msg{
		ReplyMarkup: keyboard,
		UserID:      message.Chat.ID,
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

	return sendNewMessageAndDeleteOld(bot, msg, message)
}

// Кто уже записан дети
func listChildrenTrainingUsers(bot *tgbotapi.BotAPI, queries *db.Queries, message *tgbotapi.Message) error {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}
	refreshRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(refreshListText, childListTrainingUsers)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, refreshRow, backRow)

	msg := &Msg{
		ReplyMarkup: keyboard,
		UserID:      message.Chat.ID,
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
			userText := fmt.Sprintf("        <em>%d. %s (%s)</em>\n", i+1, user.FullName, childTexts[user.AdditionalChildNumber])
			msg.Text += userText
		}
		msg.Text += "\n"
	}
	return sendNewMessageAndDeleteOld(bot, msg, message)
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
