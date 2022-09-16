package telegram

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	errNotificationDb = errors.New("DB error in notifications")        //no urgent
	errCreateSchedule = errors.New("Unable to create proper schedule") //urgent
)

func handleError(err error, userID int64) *Msg {
	msg := &Msg{}
	switch err {
	case errNotificationDb, errCreateSchedule:
		//msg.UserID = myID
		msg.Text = "Ошибка приложения" + fmt.Sprint(err)
	default:
		msg.UserID = userID
		msg.Text = "Произошла ошибка. Попробуй еще раз позже."

	}
	return msg
}
