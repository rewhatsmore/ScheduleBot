package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	db "schedule.sqlc.dev/app/db/sqlc"
	"schedule.sqlc.dev/app/table"
	"schedule.sqlc.dev/app/telegram"
)

func main() {
	testDB, err := sql.Open(os.Getenv("DB_DRIVER"), os.Getenv("DB_SOURCE"))
	if err != nil {
		log.Fatal("can not connect to db:", err)
	}

	queries := db.New(testDB)

	table.StartcCreateTable(queries)

	telegram.StartBot(os.Getenv("TELEGRAM_BOT_TOKEN"), queries)
}
