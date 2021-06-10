package administrator

import (
	"database/sql"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"log"
)

func countUsers(botLang string) int {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query("SELECT COUNT(*) FROM users;")
	if err != nil {
		log.Println(err.Error())
	}
	return readRows(rows)
}

func readRows(rows *sql.Rows) int {
	defer rows.Close()

	var count int

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			panic("Failed to scan row: " + err.Error())
		}
	}

	return count
}

func countAllUsers() int {
	var sum int
	for _, handler := range bots.Bots {
		rows, err := handler.DataBase.Query("SELECT COUNT(*) FROM users;")
		if err != nil {
			log.Println(err.Error())
		}
		sum += readRows(rows)
	}
	return sum
}

func countBlockedUsers() int {
	var count int
	for _, value := range assets.AdminSettings.BlockedUsers {
		count += value
	}
	return count
}
