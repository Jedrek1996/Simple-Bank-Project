package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// Package db contains all the unit tests only use testing.M in main
var testQueries *Queries

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource)

	if err != nil {
		log.Fatal("Error connecting to db:", err)
	}

	testQueries = New(conn)

	os.Exit(m.Run())

	fmt.Println(testQueries)
	fmt.Println(conn)
}
