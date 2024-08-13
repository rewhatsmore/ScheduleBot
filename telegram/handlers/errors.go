package telegram

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	errNotificationDb = errors.New("DB error in notifications")        //no urgent
	errCreateSchedule = errors.New("Unable to create proper schedule") //urgent
	errAddUserToSheet = errors.New("Unable to add user name to google sheet")
	errDeleteUser     = errors.New("Unable to delete user")
)

func HandleError(userID int64, err error) *Msg {
	msg := &Msg{
		ReplyMarkup: nil,
		UserID:      userID,
	}
	switch err {
	case errNotificationDb, errCreateSchedule, errAddUserToSheet:
		fmt.Println(err)
		msg.Text = "Ошибка приложения" + fmt.Sprint(err)
	case errDeleteUser:
		fmt.Println(err)
		msg.Text = "Ошибка приложения" + fmt.Sprintf("%v %d", err, userID)
	default:
		fmt.Println(err)
		msg.ReplyMarkup = *backMenuKeyboard()
		msg.Text = "Произошла ошибка. Попробуй еще раз позже."
	}
	return msg
}
