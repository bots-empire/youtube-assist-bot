package auth

import (
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

func (u *User) MakeMoney(s bots.Situation, breakTime int64) {
	if time.Now().Unix()/86400 > u.LastView/86400 {
		u.resetWatchDayCounter(s.BotLang)
	}

	if u.CompletedToday >= assets.AdminSettings.MaxOfVideoPerDay {
		u.reachedMaxAmountPerDay(s.BotLang)
		return
	}

	if u.LastView+breakTime > time.Now().Unix() {
		u.breakTimeNotPassed(s.BotLang)
		return
	}

	s.Params.AdLink, s.Err = assets.GetTask(s)
	if s.Err != nil {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "task_not_found"))
		msgs2.SendMsgToUser(s.BotLang, msg)
		return
	}

	db.RdbSetUser(s.BotLang, u.ID, s.Params.Partition+"?"+strconv.Itoa(int(breakTime)))
	u.sendMoneyStatistic(s)
	u.sendInvitationToWatch(s)
	u.LastView = time.Now().Unix()

	dataBase := bots.GetDB(s.BotLang)
	_, err := dataBase.Query("UPDATE users SET last_voice = ? WHERE id = ?;", u.LastView, u.ID)
	if err != nil {
		panic(err.Error())
	}

	go transferMoney(s, breakTime)
}

func (u *User) resetWatchDayCounter(botLang string) {
	u.CompletedToday = 0

	dataBase := bots.GetDB(botLang)
	_, err := dataBase.Query("UPDATE users SET completed_today = ? WHERE id = ?;",
		u.CompletedToday, u.ID)
	if err != nil {
		panic(err.Error())
	}
}

func (u *User) sendMoneyStatistic(s bots.Situation) {
	text := assets.LangText(u.Language, "make_money_statistic")
	text = fmt.Sprintf(text, u.CompletedToday, assets.AdminSettings.MaxOfVideoPerDay,
		u.Balance, assets.AdminSettings.WatchReward)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("back_to_make_money_button")),
	).Build(u.Language)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

func (u *User) sendInvitationToWatch(s bots.Situation) {
	text := assets.LangText(u.Language, "invitation_to_watch_"+s.Params.Partition)
	text = fmt.Sprintf(text, s.Params.AdLink)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("invitation_to_watch_il_button", s.Params.AdLink)),
	).Build(u.Language)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

func transferMoney(s bots.Situation, breakTime int64) {
	time.Sleep(time.Second * time.Duration(breakTime))

	u := GetUser(s.BotLang, s.UserID)
	u.Balance += assets.AdminSettings.WatchReward
	u.Completed++
	u.CompletedToday++

	dataBase := bots.GetDB(s.BotLang)
	_, err := dataBase.Query("UPDATE users SET balance = ?, completed = ?, completed_today = ? WHERE id = ?;",
		u.Balance, u.Completed, u.CompletedToday, u.ID)
	if err != nil {
		panic(err.Error())
	}
}

func (u *User) reachedMaxAmountPerDay(botLang string) {
	text := assets.LangText(u.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.MaxOfVideoPerDay, assets.AdminSettings.MaxOfVideoPerDay)
	msg := tgbotapi.NewMessage(int64(u.ID), text)

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) breakTimeNotPassed(botLang string) {
	text := assets.LangText(u.Language, "break_time_not_passed")
	msg := tgbotapi.NewMessage(int64(u.ID), text)

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) WithdrawMoneyFromBalance(botLang string, amount string) bool {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "incorrect_amount"))
		msgs2.SendMsgToUser(botLang, msg)
		return false
	}

	if amountInt < assets.AdminSettings.MinWithdrawalAmount {
		u.minAmountNotReached(botLang)
		return false
	}

	if u.Balance < amountInt {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "lack_of_funds"))
		msgs2.SendMsgToUser(botLang, msg)
		return false
	}

	u.Balance -= amountInt
	dataBase := bots.GetDB(botLang)
	_, err = dataBase.Query("UPDATE users SET balance = ? WHERE id = ?;", u.Balance, u.ID)
	if err != nil {
		panic(err.Error())
	}

	msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "successfully_withdrawn"))
	msgs2.SendMsgToUser(botLang, msg)
	return true
}

func (u *User) minAmountNotReached(botLang string) {
	text := assets.LangText(u.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.MinWithdrawalAmount)

	msgs2.NewParseMessage(botLang, int64(u.ID), text)
}

func (u User) GetABonus(botLang string) {
	if u.TakeBonus {
		text := assets.LangText(u.Language, "bonus_already_have")
		msgs2.SendSimpleMsg(botLang, int64(u.ID), text)
		return
	}

	u.Balance += assets.AdminSettings.BonusAmount
	dataBase := bots.GetDB(botLang)
	_, err := dataBase.Query("UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;", u.Balance, true, u.ID)
	if err != nil {
		panic(err.Error())
	}

	text := assets.LangText(u.Language, "bonus_have_received")
	msgs2.SendSimpleMsg(botLang, int64(u.ID), text)
}
