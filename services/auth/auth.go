package auth

import (
	"database/sql"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

func CheckingTheUser(botLang string, message *tgbotapi.Message) {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query("SELECT * FROM users WHERE id = ?;", message.From.ID)
	if err != nil {
		panic(err.Error())
	}

	users := ReadUsers(rows)

	switch len(users) {
	case 0:
		user := createSimpleUser(message)
		referralID := pullReferralID(message)
		user.AddNewUser(botLang, referralID)
	case 1:
	default:
		panic("There were two identical users")
	}
}

func pullReferralID(message *tgbotapi.Message) int {
	str := strings.Split(message.Text, " ")
	if len(str) < 2 {
		return 0
	}

	id, err := strconv.Atoi(str[1])
	if err != nil {
		log.Println(err)
		return 0
	}

	if id > 0 {
		return id
	}
	return 0
}

func createSimpleUser(message *tgbotapi.Message) User {
	lang := message.From.LanguageCode
	if !strings.Contains("en,de,it,pt,es", lang) || lang == "" {
		lang = "en"
	}

	return User{
		ID:       message.From.ID,
		Language: lang,
	}
}

func (u *User) AddNewUser(botLang string, referralID int) {
	dataBase := bots.GetDB(botLang)
	_, err := dataBase.Query("INSERT INTO users VALUES(?, 0, 0, 0, 0, 0, FALSE, ?);", u.ID, u.Language)
	if err != nil {
		panic(err.Error())
	}

	if referralID == u.ID || referralID == 0 {
		return
	}

	baseUser := GetUser(botLang, referralID)
	baseUser.Balance += assets.AdminSettings.ReferralAmount
	_, err = dataBase.Query("UPDATE users SET balance = ?, referral_count = ? WHERE id = ?;",
		baseUser.Balance, baseUser.ReferralCount+1, baseUser.ID)
	if err != nil {
		panic(err.Error())
	}
}

func GetUser(botLang string, id int) User {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query("SELECT * FROM users WHERE id = ?;", id)
	if err != nil {
		panic(err.Error())
	}

	users := ReadUsers(rows)

	return users[0]
}

func ReadUsers(rows *sql.Rows) []User {
	defer rows.Close()

	var users []User

	for rows.Next() {
		var (
			id, balance, completed, completedToday, referralCount int
			lastVoice                                             int64
			takeBonus                                             bool
			lang                                                  string
		)

		if err := rows.Scan(&id, &balance, &completed, &completedToday, &lastVoice, &referralCount, &takeBonus, &lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, User{
			ID:             id,
			Balance:        balance,
			Completed:      completed,
			CompletedToday: completedToday,
			LastView:       lastVoice,
			ReferralCount:  referralCount,
			TakeBonus:      takeBonus,
			Language:       lang,
		})
	}

	return users
}

func GetLang(botLang string, id int) string {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query("SELECT lang FROM users WHERE id = ?;", id)
	if err != nil {
		panic(err.Error())
	}

	return GetLangFromRow(rows)
}

func GetLangFromRow(rows *sql.Rows) string {
	defer rows.Close()

	var users []User

	for rows.Next() {
		var (
			lang string
		)

		if err := rows.Scan(&lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, User{
			Language: lang,
		})
	}

	if len(users) != 1 {
		log.Println("The number if users fond is not equal to one")
	}
	return users[0].Language
}