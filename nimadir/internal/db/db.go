package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var conn *sql.DB

func InitDB(file string) {
	var err error
	conn, err = sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER UNIQUE,
		username TEXT,
		tariff TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func GetConn() *sql.DB {
	return conn
}
