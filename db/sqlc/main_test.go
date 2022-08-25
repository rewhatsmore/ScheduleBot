package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries

func TestMain(m *testing.M) {

	testDB, err := sql.Open("postgres", "postgresql://root:reginapost@localhost:5432/shedule?sslmode=disable")
	if err != nil {
		log.Fatal("can not connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
