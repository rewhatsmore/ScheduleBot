package telegram

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	errNotificationDb = errors.New("DB error in notifications")        //no urgent
	errCreateSchedule = errors.New("Unable to create proper schedule") //urgent
	errDeleteUser     = errors.New("Unable to delete user")
)

// TO DO: change id arguments to 1 argument and assign keybord outside
func HandleError(userID int64, AdminID int64, err error) *Msg {
	msg := &Msg{
		ReplyMarkup: emptyKeyboard(),
		UserID:      AdminID,
	}
	switch err {
	case errNotificationDb, errCreateSchedule:
		msg.Text = "Ошибка приложения" + fmt.Sprint(err)
	case errDeleteUser:
		msg.Text = "Ошибка приложения" + fmt.Sprintf("%v %d", err, userID)
	default:
		msg.UserID = userID
		msg.Text = "Произошла ошибка. Попробуй еще раз позже."
	}
	return msg
}
