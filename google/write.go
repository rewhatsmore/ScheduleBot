package google

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	db "schedule.sqlc.dev/app/db/sqlc"
)

type TrainingsUsersList struct {
	TrainingDate string
	Users        []string
}

const spreadsheetId string = "108QDbpBF6HY2PvEuRnhDCQw3XSHiSq9QkyeFGTyJf10"

func AddSheet(spreadsheetId, title string) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		return err
	}

	addSheetRequest := sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title: title,
		},
	}
	request := sheets.Request{
		AddSheet: &addSheetRequest,
	}
	batchUpdateRequest := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&request},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetId, &batchUpdateRequest).Do()
	if err != nil {
		log.Fatalf("Unable to add new sheet: %v", err)
		return err
	}

	log.Printf("Sheet '%s' added successfully", title)
	return nil
}

func AddNewUserToTable(userName string) (int64, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return 0, err
	}

	readRange := "A2:A" // Диапазон для чтения ФИО

	// Получение текущего списка ФИО
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		log.Println(err)
		return 0, err
	}

	// Определение позиции для нового ФИО
	startRow := len(resp.Values) + 2
	writeRange := fmt.Sprintf("A%d", startRow)

	vr := &sheets.ValueRange{
		Values: [][]interface{}{{userName}},
	}

	// Запись нового ФИО в таблицу
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, vr).ValueInputOption("RAW").Do()
	if err != nil {
		err = fmt.Errorf("unable to update data in sheet: %v", err)
		log.Println(err)
		return 0, err
	}

	fmt.Println("New name added successfully.")
	return int64(startRow), nil
}

func AddTrainingsToTable(date time.Time, groupType db.GroupTypeEnum) (int64, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return 0, err
	}

	// Чтение первой строки для поиска свободных ячеек
	readRange := "1:1"

	// Получение текущих данных из первой строки
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		log.Println(err)
		return 0, err
	}
	fmt.Println(resp)

	// Определение следующей свободной ячейки
	startColumn := len(resp.Values[0]) + 1

	// Даты тренировок на следующую неделю
	dateString := date.Format("02.01. 15:04")
	if groupType == db.GroupTypeEnumChild {
		dateString += " (Д)"
	}

	// Преобразование даты в формат [][]interface{}
	dateValues := [][]interface{}{{dateString}}

	// Определение диапазона для записи даты тренировки
	writeRange := fmt.Sprintf("R1C%d", startColumn)

	vr := &sheets.ValueRange{
		Values: dateValues,
	}

	// Запись дат тренировок в таблицу
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, vr).ValueInputOption("RAW").Do()
	if err != nil {
		err = fmt.Errorf("unable to update data in sheet: %v", err)
		log.Println(err)
		return 0, err
	}

	fmt.Println("Dates added successfully.")
	return int64(startColumn), nil
}

func AddAppointmentToTable(rowNum, colNum int64) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return err
	}

	// Символ для добавления
	checkmark := "✔"

	// Преобразование символа в формат [][]interface{}
	values := [][]interface{}{{checkmark}}

	// Определение диапазона для записи символа
	writeRange := fmt.Sprintf("R%dC%d", rowNum, colNum)

	vr := &sheets.ValueRange{
		Values: values,
	}

	// Запись символа в таблицу
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, vr).ValueInputOption("RAW").Do()
	if err != nil {
		err = fmt.Errorf("unable to update data in sheet: %v", err)
		log.Println(err)
		return err
	}

	fmt.Println("Checkmark added successfully.")
	return nil
}

func DeleteAppointment(rowNum, colNum int64) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return err
	}

	// Пустое значение для удаления содержимого ячейки
	emptyValue := ""

	// Преобразование пустого значения в формат [][]interface{}
	values := [][]interface{}{{emptyValue}}

	// Определение диапазона для записи пустого значения
	writeRange := fmt.Sprintf("R%dC%d", rowNum, colNum)

	vr := &sheets.ValueRange{
		Values: values,
	}

	// Запись пустого значения в таблицу
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, vr).ValueInputOption("RAW").Do()
	if err != nil {
		err = fmt.Errorf("unable to update data in sheet: %v", err)
		log.Println(err)
		return err
	}

	fmt.Println("Cell cleared successfully.")
	return nil
}
