package auth

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	errFoundTwoUsers = "The number if users fond is not equal to one"
)

func CheckingTheUser(botLang string, message *tgbotapi.Message) (*model.User, error) {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query("SELECT * FROM users WHERE id = ?;", message.From.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get user")
	}

	users, err := ReadUsers(rows)
	if err != nil {
		return nil, errors.Wrap(err, "read user")
	}

	switch len(users) {
	case 0:
		user := createSimpleUser(botLang, message)
		parent := pullReferralID(botLang, message)
		if err := addNewUser(user, botLang, parent); err != nil {
			return nil, errors.Wrap(err, "add new user")
		}
		return user, nil
	case 1:
		return users[0], nil
	default:
		return nil, model.ErrFoundTwoUsers
	}
}

func pullReferralID(botLang string, message *tgbotapi.Message) int64 {
	readParams := strings.Split(message.Text, " ")
	if len(readParams) < 2 {
		return 0
	}

	linkInfo, err := model.DecodeLink(botLang, readParams[1])
	if err != nil || linkInfo == nil {
		if err != nil {
			msgs.SendNotificationToDeveloper("some err in decode link: " + err.Error())
		}

		model.IncomeBySource.WithLabelValues(
			model.GetGlobalBot(botLang).BotLink,
			botLang,
			"unknown",
		).Inc()

		return parseOldLink(readParams[1])
	}

	model.IncomeBySource.WithLabelValues(
		model.GetGlobalBot(botLang).BotLink,
		botLang,
		linkInfo.Source,
	).Inc()

	return linkInfo.ReferralID
}

func parseOldLink(str string) int64 {
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}

	if id > 0 {
		return id
	}

	return 0
}

func createSimpleUser(botLang string, message *tgbotapi.Message) *model.User {
	return &model.User{
		ID:       message.From.ID,
		Language: botLang,
	}
}

func addNewUser(u *model.User, botLang string, referralID int64) error {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query("INSERT INTO users VALUES(?, 0, 0, 0, 0, 0, 0, 0, 0, 0, FALSE, ?);", u.ID, u.Language)
	if err != nil {
		text := "Fatal Err with DB - auth.70 //" + err.Error()
		//msgs.NewParseMessage("it", 1418862576, text)
		log.Println(text)
		return errors.Wrap(err, "query failed")
	}
	_ = rows.Close()

	if referralID == u.ID || referralID == 0 {
		return nil
	}

	baseUser, err := GetUser(botLang, referralID)
	if err != nil {
		return errors.Wrap(err, "get user")
	}
	baseUser.Balance += assets.AdminSettings.Parameters[botLang].ReferralAmount
	rows, err = dataBase.Query("UPDATE users SET balance = ?, referral_count = ? WHERE id = ?;",
		baseUser.Balance, baseUser.ReferralCount+1, baseUser.ID)
	if err != nil {
		text := "Fatal Err with DB - auth.85 //" + err.Error()
		_ = msgs.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	_ = rows.Close()

	return nil
}

func GetUser(botLang string, id int64) (*model.User, error) {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query("SELECT * FROM users WHERE id = ?;", id)
	if err != nil {
		return nil, err
	}

	users, err := ReadUsers(rows)
	if err != nil || len(users) == 0 {
		return nil, model.ErrUserNotFound
	}

	return users[0], nil
}

func ReadUsers(rows *sql.Rows) ([]*model.User, error) {
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		user := model.User{}

		if err := rows.Scan(
			&user.ID,
			&user.Balance,
			&user.Completed,
			&user.CompletedT,
			&user.CompletedY,
			&user.CompletedA,
			&user.LastViewT,
			&user.LastViewY,
			&user.LastViewA,
			&user.ReferralCount,
			&user.TakeBonus,
			&user.Language); err != nil {
			return nil, errors.Wrap(err, "failed scan sql row")
		}

		users = append(users, &user)
	}

	return users, nil
}

func GetLang(botLang string, id int64) string {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query("SELECT lang FROM users WHERE id = ?;", id)
	if err != nil {
		text := "Fatal Err with DB - auth.158 //" + err.Error()
		_ = msgs.NewParseMessage("it", 1418862576, text)
		return "it"
	}

	return GetLangFromRow(rows, botLang)
}

func GetLangFromRow(rows *sql.Rows, botLang string) string {
	defer rows.Close()

	var users []model.User

	for rows.Next() {
		var (
			lang string
		)

		if err := rows.Scan(&lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, model.User{
			Language: lang,
		})
	}

	if len(users) != 1 {
		log.Println(errFoundTwoUsers)
		return botLang
	}
	return users[0].Language
}
