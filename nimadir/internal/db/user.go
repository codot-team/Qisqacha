package db

import "database/sql"

type User struct {
	ID       int
	ChatID   int64
	Username string
	Tariff   string
}

func GetUserByChatID(chatID int64) (*User, error) {
	row := GetConn().QueryRow("SELECT id, chat_id, username, tariff FROM users WHERE chat_id = ?", chatID)
	u := &User{}
	err := row.Scan(&u.ID, &u.ChatID, &u.Username, &u.Tariff)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func CreateUser(chatID int64, username, tariff string) error {
	_, err := GetConn().Exec("INSERT INTO users (chat_id, username, tariff) VALUES (?, ?, ?)", chatID, username, tariff)
	return err
}
