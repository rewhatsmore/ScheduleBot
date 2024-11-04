package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"schedule.sqlc.dev/app/conf"
	db "schedule.sqlc.dev/app/db/sqlc"
	"schedule.sqlc.dev/app/telegram"
)

func main() {
	config, err := conf.LoadConfig(".")
	if err != nil {
		log.Fatal("can not load config", err)
	}
	dbSource := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		config.PostgresUser, config.PostgresPassword, config.Host, config.PostgresPort, config.StageDBName)
	dbConn, err := sql.Open(config.DBDriver, dbSource)
	if err != nil {
		log.Fatal("can not connect to db:", err)
	}

	queries := db.New(dbConn)

	telegram.StartBot(config, queries)
}
