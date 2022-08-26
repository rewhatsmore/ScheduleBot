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
)

const commandStart = "start"
const commandMenu = "menu" /////////////////////////////////////
const commandName = "n"
const commandTrain = "t"
const makeApp = "ma"
const cancelApp = "ca"
const backMenu = "bc"
const listTrainings = "lt"
const myTrainings = "mt"
const trainUserslList = "tu"
const adminMenu = "am"
const cancelTraining = "ct"
const adminListTr = "al"
const cancelCheck = "ch"
const adminDaT = "ad"

func handleCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	switch message.Command() {
	case commandStart:
		return handleStart(message, bot)
	case commandName:
		return handleName(message, bot, queries)
	case commandMenu:
		text, keyboard := listFunctions(queries, message.From.ID)
		return sendMsg(bot, message, text, keyboard)
	case commandTrain:
		return handleNewTraining(message, queries, bot)
	default:
		return handleUncnowCommand(message, bot)
	}
}

func handleNewTraining(message *tgbotapi.Message, queries *db.Queries, bot *tgbotapi.BotAPI) error {
	inputData := strings.Split(strings.TrimPrefix(message.Text, "/t "), ",")
	place := inputData[1]
	dateAndTime, err := time.Parse("02.01.2006 15:04", inputData[0])
	if err != nil {
		text := "Данные введены в неверном формате. Попробуй еще раз. образец: /t 02.01.2006 15:04"
		keyboard := tgbotapi.InlineKeyboardMarkup{}
		row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад в меню", adminMenu)}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		return sendMsg(bot, message, text, keyboard)
	}

	arg := db.CreateTrainingParams{
		Place:       place,
		DateAndTime: dateAndTime,
	}
	_, err = queries.CreateTraining(context.Background(), arg)
	if err != nil {
		return err
	}
	text := "Тренеровка успешно добавлена"
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад в меню", adminMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	return sendMsg(bot, message, text, keyboard)

}

func handleCallback(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	switch callBack.Data[:2] {
	case makeApp:
		return handleTrainingAppointment(callBack, bot, queries)
	case cancelApp:
		return handleDeleteAppointment(callBack, bot, queries)
	case listTrainings:
		text, keyboard, err := listTrainingsForUser(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return updateMsg(bot, callBack.Message, text, keyboard)
	case backMenu:
		text, keyboard := listFunctions(queries, callBack.From.ID)
		return updateMsg(bot, callBack.Message, text, keyboard)
	case myTrainings:
		text, keyboard, err := listMyTrainings(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return updateMsg(bot, callBack.Message, text, keyboard)
	case trainUserslList:
		text, keyboard, err := listTrainingUsers(queries, callBack.From.ID)
		if err != nil {
			return err
		}
		return updateMsg(bot, callBack.Message, text, keyboard)
	case adminMenu:
		text, keyboard := listAdminFunctions()
		return updateMsg(bot, callBack.Message, text, keyboard)
	case adminListTr:
		text, keyboard, err := adminListTrainings(queries)
		if err != nil {
			return err
		}
		return updateMsg(bot, callBack.Message, text, keyboard)
	case cancelCheck:
		text, keyboard := adminCancelCheck(callBack)
		return updateMsg(bot, callBack.Message, text, keyboard)
	case cancelTraining:
		return adminCancelTraining(bot, queries, callBack)
	case adminDaT:
		text, keyboard := adminDateAntTimeRequest()
		return updateMsg(bot, callBack.Message, text, keyboard)
	default:
		return nil
	}
}

func handleTrainingAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	trainingId, err := strconv.Atoi(callBack.Data[2:])
	if err != nil {
		return err
	}

	arg := db.CreateAppointmentParams{
		TrainingID: int64(trainingId),
		UserID:     callBack.From.ID,
	}
	_, err = queries.CreateAppointment(context.Background(), arg)
	if err != nil {
		return err
	}

	text, keyboard, err := listTrainingsForUser(queries, callBack.From.ID)
	if err != nil {
		return err
	}
	return updateMsg(bot, callBack.Message, text, keyboard)
}

func handleDeleteAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	appointmentId, err := strconv.Atoi(callBack.Data[2:])
	if err != nil {
		return err
	}

	err = queries.DeleteAppointment(context.Background(), int64(appointmentId))
	if err != nil {
		return err
	}

	text, keyboard, err := listTrainingsForUser(queries, callBack.From.ID)
	if err != nil {
		return err
	}
	return updateMsg(bot, callBack.Message, text, keyboard)
}

func handleStart(message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Для записи на тренировки представься как в примере: \n /n Иван Иванов")
	_, err := bot.Send(msg)
	return err
}

func handleName(message *tgbotapi.Message, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	fullName := strings.TrimPrefix(message.Text, "/n")
	arg := db.CreateUserParams{
		UserID:   message.From.ID,
		FullName: fullName,
	}
	_, err := queries.CreateUser(context.Background(), arg) // контекс повторяется/////////////////////////////////
	if err != nil {
		return err //Обработать ошибку/////////////////////////////
	} /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	msg := tgbotapi.NewMessage(message.From.ID, "Регистрация прошла успешно")
	_, err = bot.Send(msg)
	if err != nil {
		return err //Обработать ошибку/////////////////////////////
	}
	text, keyboard := listFunctions(queries, message.From.ID)
	err = sendMsg(bot, message, text, keyboard)
	return err
}

func listFunctions(queries *db.Queries, userID int64) (string, tgbotapi.InlineKeyboardMarkup) {

	msgText := "Выбери действие:"

	btnText := map[string]string{
		listTrainings:   "Запись/отмена записи",
		myTrainings:     "Список моих тренировок",
		trainUserslList: "Кто уже записан?",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	for data, text := range btnText {
		var row []tgbotapi.InlineKeyboardButton
		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	user, err := queries.GetUser(context.Background(), userID)
	if err != nil {
		log.Println(err)
	}
	if user.IsAdmin {
		newRow := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Меню Админа", adminMenu),
		}
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, newRow)
	}
	return msgText, keyboard
}

func listAdminFunctions() (string, tgbotapi.InlineKeyboardMarkup) {

	msgText := "Меню Администратора:"

	btnText := map[string]string{
		adminListTr: "Отменить тренировку",
		adminDaT:    "Добавить тренировку",
		backMenu:    "⬅ назад в меню",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	for data, text := range btnText {
		var row []tgbotapi.InlineKeyboardButton
		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	return msgText, keyboard
}

func listMyTrainings(queries *db.Queries, userID int64) (string, tgbotapi.InlineKeyboardMarkup, error) {
	text := "Твои тренировки:\n"
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад в меню", "bc")}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	userTrainings, err := queries.ListUserTrainings(context.Background(), userID)
	if err != nil {
		return text, keyboard, err //вывести сообщение если нет тренировок???/////////////
	}

	for _, userTraining := range userTrainings {
		training := db.Training{
			TrainingID:  userTraining.TrainingID,
			Place:       userTraining.Place,
			DateAndTime: userTraining.DateAndTime,
		}
		text += "🏅" + createTextOfTraining(training) + "\n"
	}

	return text, keyboard, nil
}

//// создание тренировок для записи и отмены
func listTrainingsForUser(queries *db.Queries, userID int64) (string, tgbotapi.InlineKeyboardMarkup, error) {
	text := "Расписание на неделю. Выбери тренировки для записи. Повторное нажатие для отмены."
	keyboard := tgbotapi.InlineKeyboardMarkup{}

	trainingsForSend, err := queries.ListTrainingsForSend(context.Background(), userID)
	if err != nil {
		return text, keyboard, err
	}

	for _, trainingForSend := range trainingsForSend {
		training := db.Training{
			TrainingID:  trainingForSend.TrainingID,
			Place:       trainingForSend.Place,
			DateAndTime: trainingForSend.DateAndTime,
		}
		var row []tgbotapi.InlineKeyboardButton
		text := createTextOfTraining(training)
		data := makeApp + strconv.Itoa(int(trainingForSend.TrainingID))
		if trainingForSend.AppointmentID != 0 {
			text = "✅ " + text
			data = cancelApp + fmt.Sprintf("%d", training.TrainingID)
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад в меню", "bc")}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	return text, keyboard, nil
}

func listTrainingUsers(queries *db.Queries, userID int64) (string, tgbotapi.InlineKeyboardMarkup, error) {
	var text string
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад в меню", "bc")}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	trainings, err := queries.ListTrainings(context.Background())
	if err != nil {
		return "Расписание пока недоступно, попробуй позже", keyboard, err
	}
	for _, training := range trainings {
		text += "🏅 " + createTextOfTraining(training) + "\n"
		users, err := queries.ListTrainingUsers(context.Background(), training.TrainingID)
		if err != nil {
			log.Panicln(err)
		}
		for i, user := range users {
			text += fmt.Sprintf("%d. %s\n", i+1, user.FullName)
		}
		text += "\n"
	}
	return text, keyboard, nil
}

func adminListTrainings(queries *db.Queries) (string, tgbotapi.InlineKeyboardMarkup, error) {
	text := "Выбери тренировку, чтобы отменить."
	keyboard := tgbotapi.InlineKeyboardMarkup{}

	trainings, err := queries.ListTrainings(context.Background())
	if err != nil {
		return text, keyboard, err
	}

	for _, training := range trainings {

		var row []tgbotapi.InlineKeyboardButton
		text := createTextOfTraining(training)
		data := cancelCheck + training.DateAndTime.Format("/02.01 в 15:04/") + fmt.Sprintf("%d", training.TrainingID)

		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад в меню", adminMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	return text, keyboard, nil
}

func adminCancelCheck(callBack *tgbotapi.CallbackQuery) (string, tgbotapi.InlineKeyboardMarkup) {
	callBackData := strings.Split(callBack.Data, "/")
	text := fmt.Sprintf("Удалить тренировку %s ?", callBackData[1])
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Да", cancelTraining+callBackData[2]),
		tgbotapi.NewInlineKeyboardButtonData("Нет", adminMenu),
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	return text, keyboard
}

func adminCancelTraining(bot *tgbotapi.BotAPI, queries *db.Queries, callBack *tgbotapi.CallbackQuery) error {
	trainingId, err := strconv.Atoi(callBack.Data[2:])
	if err != nil {
		return err
	}

	err = queries.DeleteTraining(context.Background(), int64(trainingId))
	if err != nil {
		return err
	}

	text, keyboard, err := adminListTrainings(queries)
	if err != nil {
		return err
	}
	return updateMsg(bot, callBack.Message, text, keyboard)
}

// func ()  {

// }

func adminDateAntTimeRequest() (string, tgbotapi.InlineKeyboardMarkup) {
	text := "Для создания тренировки введи  с клавиатуры дату, время и место проведения тренировки как в примере:\n/t 02.01.2006 15:04, зал Ninja Way"
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅ назад в меню", adminMenu)}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	return text, keyboard
}

func sendMsg(bot *tgbotapi.BotAPI, message *tgbotapi.Message, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	return err
}

func updateMsg(bot *tgbotapi.BotAPI, message *tgbotapi.Message, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, keyboard)
	_, err := bot.Send(editMsg)
	return err
}

func handleUncnowCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Извини, я пока не знаю эту команду.")
	_, err := bot.Send(msg)
	return err
}

// создание текста кнопки
func createTextOfTraining(training db.Training) string {
	engTime := training.DateAndTime.Format("Mon 02.01 в 15:04")
	time := translateWeekDay(engTime)
	place := training.Place
	return fmt.Sprintf("%s, %s", time, place)
}

func translateWeekDay(s string) string {
	dict := map[string]string{"Mon": "пн", "Tue": "вт", "Wed": "ср", "Thu": "чт", "Fri": "пт", "Sat": "сб", "Sun": "вс"}
	old := s[:3]
	new := dict[old]
	return strings.Replace(s, old, new, 1)
}
