package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "schedule.sqlc.dev/app/db/sqlc"
)

const makeApp = "ma"
const cancelApp = "ca"
const backMenu = "bc"
const listTrainings = "lt"
const myTrainings = "mt"
const trainUsersList = "tu"
const backMenuText = "⬅ назад в меню"

func HandleCallback(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	switch callBack.Data[:2] {
	case makeApp:
		return handleTrainingAppointment(callBack, bot, queries)
	case cancelApp:
		return handleDeleteAppointment(callBack, bot, queries)
	case listTrainings:
		msg, err := listTrainingsForUser(queries, callBack.From.ID)
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
		return adminDateAntTimeRequest(callBack.From.ID, bot)
	default:
		return nil
	}
}

func handleTrainingAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	trainingId, err := strconv.Atoi(callBack.Data[2:])
	if err != nil {
		return err
	}
	//TODO: separate method and move below to new file AppointmentService
	// it should response with boolean, err and does not work with TG events
	arg := db.CreateAppointmentParams{
		TrainingID: int64(trainingId),
		UserID:     callBack.From.ID,
	}
	_, err = queries.CreateAppointment(context.Background(), arg)
	if err != nil {
		log.Println(err)
		return err
	}

	msg, err := listTrainingsForUser(queries, callBack.From.ID)
	if err != nil {
		return err
	}
	//end TODO
	return msg.UpdateMsg(bot, callBack.Message)
}

func handleDeleteAppointment(callBack *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, queries *db.Queries) error {
	appointmentId, err := strconv.Atoi(callBack.Data[2:])
	if err != nil {
		return err
	}

	err = queries.DeleteAppointment(context.Background(), int64(appointmentId))
	if err != nil {
		log.Println(err)
		return err
	}

	msg, err := listTrainingsForUser(queries, callBack.From.ID)
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
	msg := &Msg{
		UserID: userID,
		Text:   "Расписание на неделю. Выбери тренировки для записи. Повторное нажатие для отмены.",
	}

	keyboard := tgbotapi.InlineKeyboardMarkup{}
	backRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(backMenuText, backMenu)}

	trainingsForSend, err := queries.ListTrainingsForSend(context.Background(), userID)
	if err != nil {
		return msg, err
	}

	for _, trainingForSend := range trainingsForSend {
		var row []tgbotapi.InlineKeyboardButton
		text := CreateTextOfTraining(trainingForSend.DateAndTime)
		data := makeApp + strconv.Itoa(int(trainingForSend.TrainingID))
		if trainingForSend.AppointmentID != 0 {
			text = "✅  " + text
			data = cancelApp + fmt.Sprintf("%d", trainingForSend.AppointmentID)
			fmt.Println(data)
		} else {
			text = "☐  " + text
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, backRow)

	msg.ReplyMarkup = keyboard

	return msg, nil
}

func listTrainingUsers(bot *tgbotapi.BotAPI, queries *db.Queries, message *tgbotapi.Message) error {

	msg := &Msg{
		ReplyMarkup: *backMenuKeyboard(),
	}

	trainings, err := queries.ListTrainings(context.Background())
	if err != nil {
		return err
	}

	if len(trainings) == 0 {
		msg.Text = "Пока расписания нет, но скоро обязательно появится!"
	}

	for _, training := range trainings {
		text := fmt.Sprintf("<ins>🏅 <strong>%s</strong></ins>\n", CreateTextOfTraining(training.DateAndTime))
		msg.Text += text
		users, err := queries.ListTrainingUsers(context.Background(), training.TrainingID)
		if err != nil {
			log.Panicln(err)
		}
		for i, user := range users {
			msg.Text += fmt.Sprintf("        <em>%d. %s</em>\n", i+1, user.FullName)
		}
		msg.Text += "\n"
	}
	return msg.UpdateMsg(bot, message)
}
