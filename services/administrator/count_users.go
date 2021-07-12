package administrator

import (
	"database/sql"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"log"
)

const (
	getUsersCountQuery    = "SELECT COUNT(*) FROM users;"
	getDistinctUsersQuery = "SELECT COUNT(DISTINCT id) FROM subs;"
)

func countUsers(botLang string) int {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query(getUsersCountQuery)
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
		rows, err := handler.DataBase.Query(getUsersCountQuery)
		if err != nil {
			log.Println(err.Error())
		}
		sum += readRows(rows)
	}
	return sum
}

func countBlockedUsers(botLang string) int {
	//var count int
	//for _, value := range assets.AdminSettings.BlockedUsers {
	//	count += value
	//}
	//return count
	return assets.AdminSettings.BlockedUsers[botLang]
}

func countSubscribers(botLang string) int {
	rows, err := bots.Bots[botLang].DataBase.Query(getDistinctUsersQuery)
	if err != nil {
		log.Println(err.Error())
	}

	return readRows(rows)
}
