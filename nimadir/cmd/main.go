package main

import (
	"nimadir/internal/db"
	"nimadir/internal/telegram"
)

func main() {
	db.InitDB("data.db")
	defer db.GetConn().Close()

	telegram.StartBot()
}
