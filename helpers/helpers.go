package helpers

import "time"

func GetNextMonthString() string {
	nextMonthes := map[time.Month]string{
		time.December:  "Январь",
		time.January:   "Февраль",
		time.February:  "Март",
		time.March:     "Апрель",
		time.April:     "Май",
		time.May:       "Июнь",
		time.June:      "Июль",
		time.July:      "Август",
		time.August:    "Сентябрь",
		time.September: "Октябрь",
		time.October:   "Ноябрь",
		time.November:  "Декабрь",
	}

	currentMonth := time.Now().Month()
	return nextMonthes[currentMonth]
}

func TranslateWeekDay(weekday string) string {
	dict := map[string]string{"Mon": "пн", "Tue": "вт", "Wed": "ср", "Thu": "чт", "Fri": "пт", "Sat": "сб", "Sun": "вс"}
	return dict[weekday]
}
