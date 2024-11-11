package google

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	db "schedule.sqlc.dev/app/db/sqlc"
	helpers "schedule.sqlc.dev/app/helpers"
)

type TrainingsUsersList struct {
	TrainingDate string
	Users        []string
}

const spreadsheetId string = "108QDbpBF6HY2PvEuRnhDCQw3XSHiSq9QkyeFGTyJf10"

func columnNumberToName(n int) string {
	var columnName string
	for n > 0 {
		n-- // Adjust for 0-based index
		columnName = string(rune('A'+n%26)) + columnName
		n /= 26
	}
	return columnName
}

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
	sheetName := "Adult"
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return 0, err
	}

	readRange := fmt.Sprintf("%s!A5:A", sheetName) // Диапазон для чтения ФИО

	// Получение текущего списка ФИО
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		log.Println(err)
		return 0, err
	}

	// Определение позиции для нового ФИО
	startRow := len(resp.Values) + 5
	writeRange := fmt.Sprintf("%s!A%d", sheetName, startRow)

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

func AddUserToChildTable(userName string, rowNum int64) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return err
	}

	vr := &sheets.ValueRange{
		Values: [][]interface{}{{userName}},
	}

	writeRange := fmt.Sprintf("%s!A%d", "Child", rowNum)

	// Запись нового ФИО в таблицу
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, vr).ValueInputOption("RAW").Do()
	if err != nil {
		err = fmt.Errorf("unable to update data in sheet: %v", err)
		log.Println(err)
		return err
	}

	fmt.Println("New name added successfully.")
	return nil
}

func AddTrainingToTable(date time.Time, groupType db.GroupTypeEnum) (int64, error) {
	sheetName := "Adult"
	if groupType == db.GroupTypeEnumChild {
		sheetName = "Child"
	}
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return 0, err
	}

	// Чтение первой строки для поиска свободных ячеек
	readRange := fmt.Sprintf("%s!1:1", sheetName)

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
	columnLetter := columnNumberToName(startColumn)

	// Форматирование даты, дня недели и времени
	dateString := date.Format("02.01.")
	dayOfWeek := helpers.TranslateWeekDay(date.Format("Mon"))
	timeString := date.Format("15:04")

	// Преобразование данных в формат [][]interface{}
	dateValues := [][]interface{}{
		{dateString},
		{dayOfWeek},
		{timeString},
	}

	// Определение диапазона для записи данных
	writeRange := fmt.Sprintf("%s!%s1:%s3", sheetName, columnLetter, columnLetter)
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

func AddAppointmentToTable(rowNum, colNum int64, additionalChildNumber int) error {
	if additionalChildNumber == -1 {
		return addAppointmentToAdultTable(rowNum, colNum)
	}
	return addAppointmentToChildTable(rowNum, colNum, additionalChildNumber)
}

func addAppointmentToAdultTable(rowNum, colNum int64) error {
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
	writeRange := fmt.Sprintf("%s!R%dC%d", "Adult", rowNum, colNum)

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

func addAppointmentToChildTable(rowNum, colNum int64, additionalChildNumber int) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return err
	}

	// Символ для добавления

	text := additionalChildNumber
	if additionalChildNumber == 0 {
		text = 2
	}

	// Преобразование символа в формат [][]interface{}
	values := [][]interface{}{{text}}

	// Определение диапазона для записи символа
	writeRange := fmt.Sprintf("%s!R%dC%d", "Child", rowNum, colNum)

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

	fmt.Println("Number added successfully.")
	return nil
}

func DeleteAppointment(rowNum, colNum int64, sheetName string) error {

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
	writeRange := fmt.Sprintf("%s!R%dC%d", sheetName, rowNum, colNum)

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

func HideFilledColumns(sheetName string) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return err
	}

	// Чтение первой строки для поиска заполненных ячеек
	readRange := fmt.Sprintf("%s!1:1", sheetName)

	// Получение текущих данных из первой строки
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		log.Println(err)
		return err
	}

	// Получение свойств листа для получения sheetId
	sheetResp, err := srv.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve sheet properties: %v", err)
		log.Println(err)
		return err
	}

	var sheetId int64
	for _, sheet := range sheetResp.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetId = sheet.Properties.SheetId
			break
		}
	}

	// Определение последней заполненной колонки
	lastFilledColumn := -1
	for i, cell := range resp.Values[0] {
		if cell != "" {
			lastFilledColumn = i
		}
	}

	if lastFilledColumn < 1 {
		fmt.Println("No columns to hide.")
		return nil
	}

	// Создание запроса на скрытие колонок от B до последней заполненной
	hideColumnRequest := &sheets.Request{
		UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetId,
				Dimension:  "COLUMNS",
				StartIndex: 1, // B соответствует индексу 1
				EndIndex:   int64(lastFilledColumn + 1),
			},
			Properties: &sheets.DimensionProperties{
				HiddenByUser: true,
			},
			Fields: "hiddenByUser",
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{hideColumnRequest},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequest).Do()
	if err != nil {
		err = fmt.Errorf("unable to hide columns in sheet: %v", err)
		log.Println(err)
		return err
	}

	fmt.Println("Filled columns hidden successfully.")
	return nil
}

func FillColumnWithColor(sheetName string, columnNumber int64) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return err
	}

	// Получение свойств листа для получения sheetId
	sheetResp, err := srv.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve sheet properties: %v", err)
		log.Println(err)
		return err
	}

	var sheetId int64
	for _, sheet := range sheetResp.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetId = sheet.Properties.SheetId
			break
		}
	}

	lightRed := &sheets.Color{
		Red:   1.0,
		Green: 0.8,
		Blue:  0.8,
	}

	// Создание запроса на изменение цвета колонки
	updateCellsRequest := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetId,
				StartColumnIndex: columnNumber - 1, // Индекс начинается с 0
				EndColumnIndex:   columnNumber,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: lightRed,
				},
			},
			Fields: "userEnteredFormat.backgroundColor",
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{updateCellsRequest},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequest).Do()
	if err != nil {
		err = fmt.Errorf("unable to fill column with color: %v", err)
		log.Println(err)
		return err
	}

	fmt.Println("Column filled with color successfully.")
	return nil
}

func MarkRowAsDeleted(sheetName string, rowNumber int64) error {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		err = fmt.Errorf("unable to retrieve Sheets client: %v", err)
		log.Println(err)
		return err
	}

	// Получение свойств листа для получения sheetId
	sheetResp, err := srv.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve sheet properties: %v", err)
		log.Println(err)
		return err
	}

	var sheetId int64
	for _, sheet := range sheetResp.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetId = sheet.Properties.SheetId
			break
		}
	}

	lightRed := &sheets.Color{
		Red:   1.0,
		Green: 0.8,
		Blue:  0.8,
	}

	// Создание запроса на изменение цвета строки
	updateCellsRequest := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:       sheetId,
				StartRowIndex: rowNumber - 1, // Индекс начинается с 0
				EndRowIndex:   rowNumber,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: lightRed,
				},
			},
			Fields: "userEnteredFormat.backgroundColor",
		},
	}

	// Чтение первой ячейки строки
	readRange := fmt.Sprintf("%s!A%d", sheetName, rowNumber)
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		log.Println(err)
		return err
	}

	// Добавление текста "(удален)" к существующему содержимому
	var existingText string
	if len(resp.Values) > 0 && len(resp.Values[0]) > 0 {
		existingText = resp.Values[0][0].(string)
	}
	newText := existingText + " (удален)"

	// Запись обновленного текста в первую ячейку строки
	writeRange := fmt.Sprintf("%s!A%d", sheetName, rowNumber)
	vr := &sheets.ValueRange{
		Values: [][]interface{}{{newText}},
	}

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, vr).ValueInputOption("RAW").Do()
	if err != nil {
		err = fmt.Errorf("unable to update data in sheet: %v", err)
		log.Println(err)
		return err
	}

	// Выполнение пакетного обновления
	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{updateCellsRequest},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequest).Do()
	if err != nil {
		err = fmt.Errorf("unable to fill row with color: %v", err)
		log.Println(err)
		return err
	}

	fmt.Println("Row marked as deleted successfully.")
	return nil
}
