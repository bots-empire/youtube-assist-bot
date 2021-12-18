package auth

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	updateBalanceQuery       = "UPDATE users SET balance = ? WHERE id = ?;"
	updateLastVoiceQuery     = "UPDATE users SET %s = ?, %s = ? WHERE id = ?;"
	updateCompleteTodayQuery = "UPDATE users SET completed_today = ?, completed_y = ?, completed_a = ? WHERE id = ?;"
	updateAfterTaskQuery     = "UPDATE users SET balance = ?, completed = ? WHERE id = ?;"

	updateAfterBonusQuery = "UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;"

	getSubsUserQuery = "SELECT * FROM subs WHERE id = ?;"
	updateSubsQuery  = "INSERT INTO subs VALUES(?);"
)

func MakeMoney(s model.Situation, breakTime int64) error {
	if getCountOfViewInPart(s.Params.Partition, s.User) != 0 {
		if time.Now().Unix()/86400 > getLastViewInPart(s.Params.Partition, s.User)/86400 {
			resetWatchDayCounter(s.BotLang, s.User.ID)
		}

		if checkCompleteTodayWithPart(s.BotLang, s.Params.Partition, s.User) {
			return reachedMaxAmountPerDay(s.BotLang, s.User, s.Params.Partition)
		}

		if getLastViewInPart(s.Params.Partition, s.User)+breakTime > time.Now().Unix() {
			return breakTimeNotPassed(s.BotLang, s.User)
		}
	}

	s.Params.Link, s.Err = assets.GetTask(s)
	if s.Err != nil {
		msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "task_not_found"))
		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	err := sendMoneyStatistic(s)
	if err != nil {
		return err
	}

	if s.Params.Partition == "youtube" {
		sendInvitationToWatchLink(s)
	} else {
		sendInvitationToWatchVideo(s)
	}

	firstSymbolInPart := string([]rune(s.Params.Partition)[0])
	completeToday := increasePartCounter(s.User, s.Params.Partition)
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(fmt.Sprintf(updateLastVoiceQuery,
		"last_voice_"+firstSymbolInPart,
		"completed_"+firstSymbolInPart),
		time.Now().Unix(), completeToday, s.User.ID)

	if err != nil {
		text := "Fatal Err with DB - methods.63 //" + err.Error()
		_ = msgs.NewParseMessage("it", 1418862576, text)
		return nil
	}
	rows.Close()

	go transferMoney(s, breakTime)
	return nil
}

func getCountOfViewInPart(partition string, user *model.User) int {
	switch partition {
	case "youtube":
		return user.CompletedY
	case "tiktok":
		return user.CompletedT
	case "advertisement":
		return user.CompletedA
	}
	return 0
}

func resetWatchDayCounter(botLang string, userID int64) {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query(updateCompleteTodayQuery, 0, 0, 0, userID)
	if err != nil {
		text := "Fatal Err with DB - methods.108 //" + err.Error()
		_ = msgs.NewParseMessage("it", 1418862576, text)
		return
	}
	rows.Close()
}

func checkCompleteTodayWithPart(botLang, partition string, user *model.User) bool {
	switch partition {
	case "youtube":
		return user.CompletedY >= assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayY
	case "tiktok":
		return user.CompletedT >= assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayT
	case "advertisement":
		return user.CompletedA >= assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayA
	}
	return true
}

func sendMoneyStatistic(s model.Situation) error {
	text := assets.LangText(s.User.Language, "make_money_statistic")
	countVideoToday := getMaxOfVideoPerDayWithPart(s.BotLang, s.Params.Partition)

	partKey := getPartKeyFromPart(s.Params.Partition)
	text = fmt.Sprintf(text, assets.LangText(s.User.Language, partKey), getCompleteTodayInPart(s.User, s.Params.Partition),
		countVideoToday, s.User.Balance, assets.AdminSettings.Parameters[s.BotLang].WatchReward)

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ParseMode = "HTML"

	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("back_to_make_money_button")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func getMaxOfVideoPerDayWithPart(botLang, partition string) int {
	switch partition {
	case "youtube":
		return assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayY
	case "tiktok":
		return assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayT
	case "advertisement":
		return assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayA
	}
	return 0
}

func getCompleteTodayInPart(user *model.User, partition string) int {
	switch partition {
	case "youtube":
		return user.CompletedY
	case "tiktok":
		return user.CompletedT
	case "advertisement":
		return user.CompletedA
	}
	return 0
}

func getPartKeyFromPart(partition string) string {
	switch partition {
	case "youtube":
		return "make_money_youtube"
	case "tiktok":
		return "make_money_tiktok"
	case "advertisement":
		return "make_money_advertisement"
	}
	return ""
}

func sendInvitationToWatchLink(s model.Situation) {
	text := assets.LangText(s.User.Language, "invitation_to_watch_"+s.Params.Partition)
	text = fmt.Sprintf(text, s.Params.Link.Url)
	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ParseMode = "HTML"

	makeMoneyLevel := db.RdbGetMakeMoneyLevel(s)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("next_task_il_button", makeMoneyLevel)),
	).Build(s.User.Language)

	db.RdbSetUser(s.BotLang, s.User.ID, "/make_money_"+s.Params.Partition)
	db.RdbSetMakeMoneyLevel(s, strconv.Itoa(int(assets.AdminSettings.Parameters[s.BotLang].SecondBetweenViews)))
	msgs.SendMsgToUser(s.BotLang, msg)
}

func sendInvitationToWatchVideo(s model.Situation) {
	videoCfg := tgbotapi.NewVideo(s.User.ID, tgbotapi.FileID(s.Params.Link.FileID))

	makeMoneyLevel := db.RdbGetMakeMoneyLevel(s)
	videoCfg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("next_task_il_button", makeMoneyLevel)),
	).Build(s.User.Language)

	db.RdbSetMakeMoneyLevel(s, strconv.Itoa(s.Params.Link.Duration))
	db.RdbSetUser(s.BotLang, s.User.ID, "/make_money_"+s.Params.Partition)
	msgs.SendMsgToUser(s.BotLang, videoCfg)
}

func getLastViewInPart(partition string, user *model.User) int64 {
	switch partition {
	case "youtube":
		return user.LastViewY
	case "tiktok":
		return user.LastViewT
	case "advertisement":
		return user.LastViewA
	}
	return 0
}

func transferMoney(s model.Situation, breakTime int64) {
	time.Sleep(time.Second * time.Duration(breakTime))

	s.User.Balance += assets.AdminSettings.Parameters[s.BotLang].WatchReward
	s.User.Completed++

	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateAfterTaskQuery, s.User.Balance, s.User.Completed, s.User.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.232 //" + err.Error()
		_ = msgs.NewParseMessage("it", 1418862576, text)
		return
	}
	rows.Close()
}

func increasePartCounter(user *model.User, partition string) int {
	switch partition {
	case "youtube":
		return user.CompletedY + 1
	case "tiktok":
		return user.CompletedT + 1
	case "advertisement":
		return user.CompletedA + 1
	}
	return 0
}

func reachedMaxAmountPerDay(botLang string, user *model.User, partition string) error {
	text := assets.LangText(user.Language, "reached_max_amount_per_day")
	var maxPerDayInPart int
	switch partition {
	case "youtube":
		maxPerDayInPart = assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayY
	case "tiktok":
		maxPerDayInPart = assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayT
	case "advertisement":
		maxPerDayInPart = assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDayA
	}

	text = fmt.Sprintf(text, maxPerDayInPart, maxPerDayInPart)
	msg := tgbotapi.NewMessage(user.ID, text)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[user.Language].Url)),
	).Build(user.Language)

	return msgs.SendMsgToUser(botLang, msg)
}

func breakTimeNotPassed(botLang string, user *model.User) error {
	text := assets.LangText(user.Language, "break_time_not_passed")
	msg := tgbotapi.NewMessage(user.ID, text)

	return msgs.SendMsgToUser(botLang, msg)
}

func WithdrawMoneyFromBalance(s model.Situation, amount string) error {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "incorrect_amount"))
		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	if amountInt < assets.AdminSettings.Parameters[s.BotLang].MinWithdrawalAmount {
		return minAmountNotReached(s.BotLang, s.User)
	}

	if s.User.Balance < amountInt {
		msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "lack_of_funds"))
		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	return sendInvitationToSubs(s, amount)
}

func minAmountNotReached(botLang string, user *model.User) error {
	text := assets.LangText(user.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount)

	return msgs.NewParseMessage(botLang, user.ID, text)
}

func sendInvitationToSubs(s model.Situation, amount string) error {
	text := msgs.GetFormatText(s.User.Language, "withdrawal_not_subs_text")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.User.Language].Url)),
		msgs.NewIlRow(msgs.NewIlDataButton("im_subscribe_button", "/withdrawal_money?"+amount)),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func CheckSubscribeToWithdrawal(s model.Situation, amount int) bool {
	if s.User.Balance < amount {
		return false
	}

	if !CheckSubscribe(s) {
		_ = sendInvitationToSubs(s, strconv.Itoa(amount))
		return false
	}

	s.User.Balance -= amount
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateBalanceQuery, s.User.Balance, s.User.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.335 //" + err.Error()
		_ = msgs.NewParseMessage("it", 1418862576, text)
		return false
	}
	rows.Close()

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "successfully_withdrawn"))
	_ = msgs.SendMsgToUser(s.BotLang, msg)
	return true
}

func GetABonus(s model.Situation) error {
	if !CheckSubscribe(s) {
		text := assets.LangText(s.User.Language, "user_dont_subscribe_the_channel")

		return msgs.SendSimpleMsg(s.BotLang, s.User.ID, text)
	}

	if s.User.TakeBonus {
		text := assets.LangText(s.User.Language, "bonus_already_have")

		return msgs.SendSimpleMsg(s.BotLang, s.User.ID, text)
	}

	s.User.Balance += assets.AdminSettings.Parameters[s.BotLang].BonusAmount
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateAfterBonusQuery, s.User.Balance, true, s.User.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.363 //" + err.Error()

		return msgs.NewParseMessage("it", 1418862576, text)
	}
	rows.Close()

	text := assets.LangText(s.User.Language, "bonus_have_received")
	return msgs.SendSimpleMsg(s.BotLang, s.User.ID, text)
}

func CheckSubscribe(s model.Situation) bool {
	member, err := model.Bots[s.BotLang].Bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: assets.AdminSettings.AdvertisingChan[s.BotLang].ChannelID,
			UserID: s.User.ID,
		},
	})

	if err == nil {
		if err := addMemberToSubsBase(s); err != nil {
			return false
		}
		return checkMemberStatus(member)
	}
	return false
}

func checkMemberStatus(member tgbotapi.ChatMember) bool {
	if member.IsAdministrator() {
		return true
	}
	if member.IsCreator() {
		return true
	}
	if member.Status == "member" {
		return true
	}
	return false
}

func addMemberToSubsBase(s model.Situation) error {
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(getSubsUserQuery, s.User.ID)
	if err != nil {
		return err
	}

	user, err := readUser(rows)
	if err != nil {
		return err
	}

	if user.ID != 0 {
		return nil
	}
	rows, err = dataBase.Query(updateSubsQuery, s.User.ID)
	if err != nil {
		return err
	}
	_ = rows.Close()
	return nil
}

func readUser(rows *sql.Rows) (*model.User, error) {
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return nil, model.ErrScanSqlRow
		}

		users = append(users, &model.User{
			ID: id,
		})
	}
	if len(users) == 0 {
		users = append(users, &model.User{
			ID: 0,
		})
	}
	return users[0], nil
}
