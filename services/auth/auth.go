package auth

import (
	"database/sql"
	"fmt"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

const (
	getUsersUserQuery        = "SELECT * FROM users WHERE id = ?;"
	newUserQuery             = "INSERT INTO users VALUES(?, 0, 0, 0, 0, 0, 0, 0, 0, 0, FALSE, ?);"
	updateAfterReferralQuery = "UPDATE users SET balance = ?, referral_count = ? WHERE id = ?;"
	getLangQuery             = "SELECT lang FROM users WHERE id = ?;"

	errFoundTwoUsers = "The number if users fond is not equal to one"
)

func CheckingTheUser(botLang string, message *tgbotapi.Message) {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query(getUsersUserQuery, message.From.ID)
	if err != nil {
		text := "Fatal Err with DB - auth.18 //" + err.Error()
		msgs.NewParseMessage("it", 1418862576, text)
		return
	}

	users := ReadUsers(rows)

	switch len(users) {
	case 0:
		user := createSimpleUser(botLang, message)
		referralID := pullReferralID(message)
		user.AddNewUser(botLang, referralID)
	case 1:
	default:
		text := "There were two identical users where id = " + strconv.Itoa(message.From.ID) + " in " + botLang + " bot"
		msgs.NewParseMessage("it", 1418862576, text)
		log.Println(text)
		return
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

func createSimpleUser(botLang string, message *tgbotapi.Message) User {
	//lang := message.From.LanguageCode
	//if !strings.Contains("en,de,it,pt,es", lang) || lang == "" {
	//	lang = "en"
	//}

	return User{
		ID:       message.From.ID,
		Language: bots.Bots[botLang].LanguageInBot,
	}
}

func (u *User) AddNewUser(botLang string, referralID int) {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query(newUserQuery, u.ID, u.Language)
	if err != nil {
		text := "Fatal Err with DB - auth.81 //" + err.Error()
		//msgs.NewParseMessage("it", 1418862576, text)
		log.Println(text)
		return
	}
	rows.Close()

	if referralID == u.ID || referralID == 0 {
		return
	}

	baseUser, err := GetUser(botLang, referralID)
	if err != nil {
		return
	}

	baseUser.Balance += assets.AdminSettings.Parameters[botLang].ReferralAmount
	rows, err = dataBase.Query(updateAfterReferralQuery, baseUser.Balance, baseUser.ReferralCount+1, baseUser.ID)
	if err != nil {
		text := "Fatal Err with DB - auth.96 //" + err.Error()
		msgs.NewParseMessage("it", 1418862576, text)
		return
	}
	rows.Close()
}

func GetUser(botLang string, id int) (*User, error) {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query(getUsersUserQuery, id)
	if err != nil {
		return nil, err
	}

	users := ReadUsers(rows)
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return users[0], nil
}

func ReadUsers(rows *sql.Rows) []*User {
	defer rows.Close()

	var users []*User

	for rows.Next() {
		var (
			id, balance,
			completed, completedT, completedY, completedA,
			referralCount int
			lastVoiceT, lastVoiceY, lastVoiceA int64
			takeBonus                          bool
			lang                               string
		)

		if err := rows.Scan(&id, &balance, &completed, &completedT, &completedY, &completedA,
			&lastVoiceT, &lastVoiceY, &lastVoiceA, &referralCount, &takeBonus, &lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, &User{
			ID:            id,
			Balance:       balance,
			Completed:     completed,
			CompletedT:    completedT,
			CompletedY:    completedY,
			CompletedA:    completedA,
			LastViewT:     lastVoiceT,
			LastViewY:     lastVoiceY,
			LastViewA:     lastVoiceA,
			ReferralCount: referralCount,
			TakeBonus:     takeBonus,
			Language:      lang,
		})
	}

	return users
}

func GetLang(botLang string, id int) string {
	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query(getLangQuery, id)
	if err != nil {
		text := "Fatal Err with DB - auth.158 //" + err.Error()
		msgs.NewParseMessage("it", 1418862576, text)
		return "it"
	}

	return GetLangFromRow(rows, botLang)
}

func GetLangFromRow(rows *sql.Rows, botLang string) string {
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
		log.Println(errFoundTwoUsers)
		return botLang
	}
	return users[0].Language
}
