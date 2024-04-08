package common

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func GetDB() *sql.DB {
	if DB != nil {
		return DB
	}

	DB, err := sql.Open("sqlite3", "cache.db")
	if err != nil {
		log.Fatalln(err)
	}

	return DB
}
