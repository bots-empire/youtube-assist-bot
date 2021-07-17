package auth

import (
	"database/sql"
	"fmt"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/db"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
	"time"
)

const (
	updateBalanceQuery       = "UPDATE users SET balance = ? WHERE id = ?;"
	updateLastVoiceQuery     = "UPDATE users SET last_voice = ? WHERE id = ?;"
	updateCompleteTodayQuery = "UPDATE users SET completed_today = ? WHERE id = ?;"
	updateAfterTaskQuery     = "UPDATE users SET balance = ?, completed = ?, completed_today = ? WHERE id = ?;"

	updateAfterBonusQuery = "UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;"

	getSubsUserQuery = "SELECT * FROM subs WHERE id = ?;"
	updateSubsQuery  = "INSERT INTO subs VALUES(?);"
)

func (u *User) MakeMoney(s bots.Situation, breakTime int64) {
	if time.Now().Unix()/86400 > u.LastView/86400 {
		u.resetWatchDayCounter(s.BotLang)
	}

	if u.CompletedToday >= assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDay {
		u.reachedMaxAmountPerDay(s.BotLang)
		return
	}

	if u.LastView+breakTime > time.Now().Unix() {
		u.breakTimeNotPassed(s.BotLang)
		return
	}

	s.Params.Link, s.Err = assets.GetTask(s)
	if s.Err != nil {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "task_not_found"))
		msgs2.SendMsgToUser(s.BotLang, msg)
		return
	}

	u.sendMoneyStatistic(s)
	if s.Params.Partition == "youtube" {
		u.sendInvitationToWatchLink(s)
	} else {
		u.sendInvitationToWatchVideo(s)
	}
	u.LastView = time.Now().Unix()

	dataBase := bots.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateLastVoiceQuery, u.LastView, u.ID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

	go transferMoney(s, breakTime)
}

func (u *User) resetWatchDayCounter(botLang string) {
	u.CompletedToday = 0

	dataBase := bots.GetDB(botLang)
	rows, err := dataBase.Query(updateCompleteTodayQuery, u.CompletedToday, u.ID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()
}

func (u *User) sendMoneyStatistic(s bots.Situation) {
	text := assets.LangText(u.Language, "make_money_statistic")
	text = fmt.Sprintf(text, u.CompletedToday, assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDay,
		u.Balance, assets.AdminSettings.Parameters[s.BotLang].WatchReward)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("back_to_make_money_button")),
	).Build(u.Language)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

func (u *User) sendInvitationToWatchLink(s bots.Situation) {
	text := assets.LangText(u.Language, "invitation_to_watch_"+s.Params.Partition)
	text = fmt.Sprintf(text, s.Params.Link.Url)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	makeMoneyLevel := db.RdbGetMakeMoneyLevel(s)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlDataButton("next_task_il_button", makeMoneyLevel)),
	).Build(u.Language)

	db.RdbSetUser(s.BotLang, u.ID, "/make_money_"+s.Params.Partition)
	db.RdbSetMakeMoneyLevel(s, strconv.Itoa(int(assets.AdminSettings.Parameters[s.BotLang].SecondBetweenViews)))
	msgs2.SendMsgToUser(s.BotLang, msg)
}

func (u *User) sendInvitationToWatchVideo(s bots.Situation) {
	videoCfg := tgbotapi.NewVideoShare(int64(s.UserID), s.Params.Link.FileID)

	makeMoneyLevel := db.RdbGetMakeMoneyLevel(s)
	videoCfg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlDataButton("next_task_il_button", makeMoneyLevel)),
	).Build(u.Language)

	db.RdbSetMakeMoneyLevel(s, strconv.Itoa(s.Params.Link.Duration))
	db.RdbSetUser(s.BotLang, u.ID, "/make_money_"+s.Params.Partition)
	msgs2.SendMsgToUser(s.BotLang, videoCfg)
}

func transferMoney(s bots.Situation, breakTime int64) {
	time.Sleep(time.Second * time.Duration(breakTime))

	u := GetUser(s.BotLang, s.UserID)
	u.Balance += assets.AdminSettings.Parameters[s.BotLang].WatchReward
	u.Completed++
	u.CompletedToday++

	dataBase := bots.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateAfterTaskQuery, u.Balance, u.Completed, u.CompletedToday, u.ID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()
}

func (u *User) reachedMaxAmountPerDay(botLang string) {
	text := assets.LangText(u.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDay, assets.AdminSettings.Parameters[botLang].MaxOfVideoPerDay)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[u.Language].Url)),
	).Build(u.Language)

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) breakTimeNotPassed(botLang string) {
	text := assets.LangText(u.Language, "break_time_not_passed")
	msg := tgbotapi.NewMessage(int64(u.ID), text)

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) WithdrawMoneyFromBalance(s bots.Situation, amount string) {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "incorrect_amount"))
		msgs2.SendMsgToUser(s.BotLang, msg)
		return
	}

	if amountInt < assets.AdminSettings.Parameters[s.BotLang].MinWithdrawalAmount {
		u.minAmountNotReached(s.BotLang)
		return
	}

	if u.Balance < amountInt {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "lack_of_funds"))
		msgs2.SendMsgToUser(s.BotLang, msg)
		return
	}

	sendInvitationToSubs(s, amount)
}

func (u *User) minAmountNotReached(botLang string) {
	text := assets.LangText(u.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount)

	msgs2.NewParseMessage(botLang, int64(u.ID), text)
}

func sendInvitationToSubs(s bots.Situation, amount string) {
	text := msgs2.GetFormatText(s.UserLang, "withdrawal_not_subs_text")

	msg := tgbotapi.NewMessage(int64(s.UserID), text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.UserLang].Url)),
		msgs2.NewIlRow(msgs2.NewIlDataButton("im_subscribe_button", "/withdrawal_money?"+amount)),
	).Build(s.UserLang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

func (u *User) CheckSubscribeToWithdrawal(s bots.Situation, amount int) bool {
	if u.Balance < amount {
		return false
	}

	if !u.CheckSubscribe(s) {
		sendInvitationToSubs(s, strconv.Itoa(amount))
		return false
	}

	u.Balance -= amount
	dataBase := bots.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateBalanceQuery, u.Balance, u.ID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

	msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "successfully_withdrawn"))
	msgs2.SendMsgToUser(s.BotLang, msg)
	return true
}

func (u *User) GetABonus(s bots.Situation) {
	if !u.CheckSubscribe(s) {
		text := assets.LangText(u.Language, "user_dont_subscribe_the_channel")
		msgs2.SendSimpleMsg(s.BotLang, int64(u.ID), text)
		return
	}

	if u.TakeBonus {
		text := assets.LangText(u.Language, "bonus_already_have")
		msgs2.SendSimpleMsg(s.BotLang, int64(u.ID), text)
		return
	}

	u.Balance += assets.AdminSettings.Parameters[s.BotLang].BonusAmount
	dataBase := bots.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateAfterBonusQuery, u.Balance, true, u.ID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

	text := assets.LangText(u.Language, "bonus_have_received")
	msgs2.SendSimpleMsg(s.BotLang, int64(u.ID), text)
}

func (u *User) CheckSubscribe(s bots.Situation) bool {
	fmt.Println(assets.AdminSettings.AdvertisingChan[s.BotLang].ChannelID)
	member, err := bots.Bots[s.BotLang].Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: assets.AdminSettings.AdvertisingChan[s.BotLang].ChannelID,
		UserID: s.UserID,
	})

	if err == nil {
		addMemberToSubsBase(s)
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
	if member.IsMember() {
		return true
	}
	return false
}

func addMemberToSubsBase(s bots.Situation) {
	dataBase := bots.GetDB(s.BotLang)
	rows, err := dataBase.Query(getSubsUserQuery, s.UserID)
	if err != nil {
		panic(err.Error())
	}

	user := readUser(rows)
	if user.ID != 0 {
		return
	}
	rows, err = dataBase.Query(updateSubsQuery, s.UserID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()
}

func readUser(rows *sql.Rows) User {
	defer rows.Close()

	var users []User

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, User{
			ID: id,
		})
	}
	if len(users) == 0 {
		users = append(users, User{
			ID: 0,
		})
	}
	return users[0]
}
