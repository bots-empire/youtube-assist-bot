package services

import (
	"fmt"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/db"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/Stepan1328/youtube-assist-bot/services/administrator"
	"github.com/Stepan1328/youtube-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	updateCounterHeader = "Today Update's counter: %d"
	updatePrintHeader   = "update number: %d	// youtube-bot-update:	"
	extraneousUpdate    = "extraneous update"
	notificationChatID  = 1418862576

	updateBalanceQuery = "UPDATE users SET balance = ? WHERE id = ?;"
)

type MessagesHandlers struct {
	Handlers map[string]bots.Handler
}

func (h *MessagesHandlers) GetHandler(command string) bots.Handler {
	return h.Handlers[command]
}

func (h *MessagesHandlers) Init() {
	//Start command
	h.OnCommand("/start", NewStartCommand())
	h.OnCommand("/admin", administrator.NewAdminCommand())
	h.OnCommand("/getUpdate", administrator.NewGetUpdateCommand())

	//Main command
	h.OnCommand("/main_make_money", NewMakeMoneyCommand())
	//h.OnCommand("/main_spend_money", NewSpendMoneyCommand())
	h.OnCommand("/main_profile", NewSendProfileCommand())
	h.OnCommand("/main_money_for_a_friend", NewMoneyForAFriendCommand())
	h.OnCommand("/main_statistic", NewSendStatisticsCommand())
	h.OnCommand("/main_more_money", NewMoreMoneyCommand())

	//Make money command
	h.OnCommand("/make_money_youtube?", NewLinkTaskCommand())
	h.OnCommand("/make_money_tiktok?", NewVideoTaskCommand())
	h.OnCommand("/make_money_advertisement?", NewVideoTaskCommand())

	//Spend money command
	h.OnCommand("/spend_money_withdrawal", NewSpendMoneyWithdrawalCommand())
	h.OnCommand("/paypal_method", NewPaypalReqCommand())
	h.OnCommand("/credit_card_method", NewCreditCardReqCommand())
	h.OnCommand("/withdrawal_method", NewWithdrawalMethodCommand())
	h.OnCommand("/withdrawal_req_amount", NewReqWithdrawalAmountCommand())
	h.OnCommand("/withdrawal_exit", NewWithdrawalAmountCommand())
	//h.OnCommand("/promotion_choice", NewPromotionCommand())
	//h.OnCommand("/promotion_case", NewPromotionCaseAnswerCommand())

	//Log out command
	h.OnCommand("/admin_log_out", NewAdminLogOutCommand())

	log.Println("Messages Handlers Initialized")
}

func (h *MessagesHandlers) OnCommand(command string, handler bots.Handler) {
	h.Handlers[command] = handler
}

func ActionsWithUpdates(botLang string, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		checkUpdate(botLang, &update)
	}
}

func checkUpdate(botLang string, update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	PrintNewUpdate(botLang, update)
	if update.Message != nil {
		auth.CheckingTheUser(botLang, update.Message)
		situation := createSituationFromMsg(botLang, update.Message)

		//PrintNewSituation(situation)
		checkMessage(situation)
		return
	}

	if update.CallbackQuery != nil {
		situation := createSituationFromCallback(botLang, update.CallbackQuery)

		//PrintNewSituation(situation)
		checkCallbackQuery(situation)
		return
	}
}

func PrintNewUpdate(botLang string, update *tgbotapi.Update) {
	if (time.Now().Unix()+6500)/86400 > int64(assets.UpdateStatistic.Day) {
		text := fmt.Sprintf(updateCounterHeader, assets.UpdateStatistic.Counter)
		msgID := msgs2.NewIDParseMessage(administrator.DefaultNotificationBot, 1418862576, text)
		msgs2.SendMsgToUser(administrator.DefaultNotificationBot, tgbotapi.PinChatMessageConfig{
			ChatID:    notificationChatID,
			MessageID: msgID,
		})
		assets.UpdateStatistic.Counter = 0
		assets.UpdateStatistic.Day = int(time.Now().Unix()+6500) / 86400
	}
	assets.UpdateStatistic.Counter++
	assets.SaveUpdateStatistic()

	fmt.Printf(updatePrintHeader, assets.UpdateStatistic.Counter)
	if update.Message != nil {
		if update.Message.Text != "" {
			fmt.Println(botLang, update.Message.Text)
			return
		}
	}

	if update.CallbackQuery != nil {
		fmt.Println(botLang, update.CallbackQuery.Data)
		return
	}

	fmt.Println(botLang, extraneousUpdate)
}

//func PrintNewSituation(situation bots.Situation) {
//	bytes, err := json.MarshalIndent(situation, "", "   ")
//	if err != nil {
//		log.Println(err)
//		return
//	}
//	log.Println("New update:\n", string(bytes), "\n")
//}

func createSituationFromMsg(botLang string, message *tgbotapi.Message) bots.Situation {
	return bots.Situation{
		Message:  message,
		BotLang:  botLang,
		UserID:   message.From.ID,
		UserLang: auth.GetLang(botLang, message.From.ID),
		Params: bots.Parameters{
			Level: db.GetLevel(botLang, message.From.ID),
		},
	}
}

func createSituationFromCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery) bots.Situation {
	return bots.Situation{
		CallbackQuery: callbackQuery,
		BotLang:       botLang,
		UserID:        callbackQuery.From.ID,
		UserLang:      auth.GetLang(botLang, callbackQuery.From.ID),
		Command:       strings.Split(callbackQuery.Data, "?")[0],
		Params: bots.Parameters{
			Level: db.GetLevel(botLang, callbackQuery.From.ID),
		},
	}
}

func checkMessage(situation bots.Situation) {
	if situation.Command == "" {
		situation.Command, situation.Err = assets.GetCommandFromText(situation)
	}

	if situation.Err == nil {
		Handler := bots.Bots[situation.BotLang].MessageHandler.
			GetHandler(situation.Command)

		if Handler != nil {
			Handler.Serve(situation)
			return
		}
	}

	situation.Command = strings.Split(situation.Params.Level, "?")[0]

	Handler := bots.Bots[situation.BotLang].MessageHandler.
		GetHandler(situation.Command)

	if Handler != nil {
		Handler.Serve(situation)
		return
	}

	if administrator.CheckAdminMessage(situation) {
		return
	}

	emptyLevel(situation.BotLang, situation.Message, situation.UserLang)
	log.Println(situation.Err)
}

func emptyLevel(botLang string, message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_msg_dont_recognize"))
	msgs2.SendMsgToUser(botLang, msg)
}

type StartCommand struct {
}

func NewStartCommand() *StartCommand {
	return &StartCommand{}
}

func (c *StartCommand) Serve(s bots.Situation) {
	if strings.Contains(s.Message.Text, "new_admin") {
		s.Command = s.Message.Text
		administrator.CheckNewAdmin(s)
		return
	}

	text := assets.LangText(s.UserLang, "main_select_menu")
	db.RdbSetUser(s.BotLang, s.UserID, "main")

	msg := tgbotapi.NewMessage(int64(s.UserID), text)
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("main_make_money")),
		msgs2.NewRow(msgs2.NewDataButton("spend_money_withdrawal")),
		//msgs2.NewRow(msgs2.NewDataButton("main_spend_money")),
		msgs2.NewRow(msgs2.NewDataButton("main_money_for_a_friend"),
			msgs2.NewDataButton("main_more_money")),
		msgs2.NewRow(msgs2.NewDataButton("main_profile"),
			msgs2.NewDataButton("main_statistic")),
	).Build(s.UserLang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type MakeMoneyCommand struct {
}

func NewMakeMoneyCommand() *MakeMoneyCommand {
	return &MakeMoneyCommand{}
}

func (c *MakeMoneyCommand) Serve(s bots.Situation) {
	text := assets.LangText(s.UserLang, "make_money_text")
	db.RdbSetUser(s.BotLang, s.UserID, "main")

	msg := tgbotapi.NewMessage(int64(s.UserID), text)
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("make_money_youtube")),
		msgs2.NewRow(msgs2.NewDataButton("make_money_tiktok")),
		msgs2.NewRow(msgs2.NewDataButton("make_money_advertisement")),
		msgs2.NewRow(msgs2.NewDataButton("back_to_main_menu_button")),
	).Build(s.UserLang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type LinkTaskCommand struct {
}

func NewLinkTaskCommand() *LinkTaskCommand {
	return &LinkTaskCommand{}
}

func (c *LinkTaskCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)

	s.Params.Partition = strings.Split(strings.Replace(s.Command, "/make_money_", "", 1), "?")[0]
	user.MakeMoney(s, assets.AdminSettings.Parameters[s.BotLang].SecondBetweenViews)
}

type VideoTaskCommand struct {
}

func NewVideoTaskCommand() *VideoTaskCommand {
	return &VideoTaskCommand{}
}

func (c *VideoTaskCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)

	user.MakeMoney(s, getMakeMoneyDuration(&s))
}

func getMakeMoneyDuration(s *bots.Situation) int64 {
	s.Params.Partition = strings.Split(strings.Replace(s.Command, "/make_money_", "", 1), "?")[0]

	breakTime := 0
	makeMoneyLevel := db.RdbGetMakeMoneyLevel(*s)
	data := strings.Split(strings.Replace(makeMoneyLevel, "/make_money_", "", 1), "?")
	if len(data) > 1 {
		breakTime, _ = strconv.Atoi(data[1])
	}
	return int64(breakTime)
}

type SpendMoneyCommand struct {
}

func NewSpendMoneyCommand() *SpendMoneyCommand {
	return &SpendMoneyCommand{}
}

func (c *SpendMoneyCommand) Serve(s bots.Situation) {
	text := assets.LangText(s.UserLang, "make_money_text")
	db.RdbSetUser(s.BotLang, s.UserID, "main")

	msg := tgbotapi.NewMessage(int64(s.UserID), text)
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("spend_money_withdrawal")),
		msgs2.NewRow(msgs2.NewDataButton("spend_money_promotion")),
		msgs2.NewRow(msgs2.NewDataButton("back_to_main_menu_button")),
	).Build(s.UserLang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type SpendMoneyWithdrawalCommand struct {
}

func NewSpendMoneyWithdrawalCommand() *SpendMoneyWithdrawalCommand {
	return &SpendMoneyWithdrawalCommand{}
}

func (c *SpendMoneyWithdrawalCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)
	db.RdbSetUser(s.BotLang, s.UserID, "withdrawal")

	text := msgs2.GetFormatText(user.Language, "withdrawal_money", user.Balance)
	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdrawal_method_1"),
			msgs2.NewDataButton("withdrawal_method_2")),
		msgs2.NewRow(msgs2.NewDataButton("withdrawal_method_3"),
			msgs2.NewDataButton("withdrawal_method_4")),
		msgs2.NewRow(msgs2.NewDataButton("withdrawal_method_5"),
			msgs2.NewDataButton("withdrawal_method_6")),
		msgs2.NewRow(msgs2.NewDataButton("back_to_main_menu_button")),
	).Build(user.Language)

	msgs2.NewParseMarkUpMessage(s.BotLang, int64(s.UserID), &markUp, text)
}

type PaypalReqCommand struct {
}

func NewPaypalReqCommand() *PaypalReqCommand {
	return &PaypalReqCommand{}
}

func (c *PaypalReqCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "/withdrawal_req_amount")

	lang := auth.GetLang(s.BotLang, s.UserID)
	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(lang, "paypal_email"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type CreditCardReqCommand struct {
}

func NewCreditCardReqCommand() *CreditCardReqCommand {
	return &CreditCardReqCommand{}
}

func (c *CreditCardReqCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "/withdrawal_req_amount")

	lang := auth.GetLang(s.BotLang, s.UserID)
	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(lang, "credit_card_number"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalMethodCommand struct {
}

func NewWithdrawalMethodCommand() *WithdrawalMethodCommand {
	return &WithdrawalMethodCommand{}
}

func (c *WithdrawalMethodCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "/withdrawal_req_amount")

	lang := auth.GetLang(s.BotLang, s.UserID)
	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(lang, "request_number_email"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type ReqWithdrawalAmountCommand struct {
}

func NewReqWithdrawalAmountCommand() *ReqWithdrawalAmountCommand {
	return &ReqWithdrawalAmountCommand{}
}

func (c *ReqWithdrawalAmountCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "/withdrawal_exit")

	lang := auth.GetLang(s.BotLang, s.UserID)
	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(lang, "req_withdrawal_amount"))

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalAmountCommand struct {
}

func NewWithdrawalAmountCommand() *WithdrawalAmountCommand {
	return &WithdrawalAmountCommand{}
}

func (c *WithdrawalAmountCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)
	user.WithdrawMoneyFromBalance(s, s.Message.Text)
}

type SendProfileCommand struct {
}

func NewSendProfileCommand() *SendProfileCommand {
	return &SendProfileCommand{}
}

func (c *SendProfileCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)
	db.RdbSetUser(s.BotLang, s.UserID, "main")

	text := msgs2.GetFormatText(user.Language, "profile_text",
		s.Message.From.FirstName, user.Balance, user.Completed, user.ReferralCount)

	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)

	//markUp := msgs2.NewIlMarkUp(
	//	msgs2.NewIlRow(msgs2.NewIlDataButton("change_lang_button", "/send_change_lang")),
	//).Build(user.Language)
	//
	//msgs2.NewParseMarkUpMessage(s.BotLang, int64(user.ID), markUp, text)
}

type SendStatisticsCommand struct {
}

func NewSendStatisticsCommand() *SendStatisticsCommand {
	return &SendStatisticsCommand{}
}

func (c *SendStatisticsCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "main")
	text := assets.LangText(s.UserLang, "statistic_to_user")

	text = fillDate(text)
	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
}

type MoneyForAFriendCommand struct {
}

func NewMoneyForAFriendCommand() *MoneyForAFriendCommand {
	return &MoneyForAFriendCommand{}
}

func (c *MoneyForAFriendCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)
	db.RdbSetUser(s.BotLang, s.UserID, "main")

	text := msgs2.GetFormatText(user.Language, "referral_text", bots.GetGlobalBot(s.BotLang).BotLink,
		user.ID, assets.AdminSettings.Parameters[s.BotLang].ReferralAmount, user.ReferralCount)

	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
}

type MoreMoneyCommand struct {
}

func NewMoreMoneyCommand() *MoreMoneyCommand {
	return &MoreMoneyCommand{}
}

func (c *MoreMoneyCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "main")
	text := msgs2.GetFormatText(s.UserLang, "more_money_text",
		assets.AdminSettings.Parameters[s.BotLang].BonusAmount, assets.AdminSettings.Parameters[s.BotLang].BonusAmount)

	msg := tgbotapi.NewMessage(int64(s.UserID), text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.UserLang].Url)),
		msgs2.NewIlRow(msgs2.NewIlDataButton("get_bonus_button", "/send_bonus_to_user")),
	).Build(s.UserLang)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type PromotionCommand struct {
}

func NewPromotionCommand() *PromotionCommand {
	return &PromotionCommand{}
}

func (c *PromotionCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "main")
	text := assets.LangText(s.UserLang, "promotion_main_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlDataButton("promotion_case_1", "/promotion_case?500")),
		msgs2.NewIlRow(msgs2.NewIlDataButton("promotion_case_2", "/promotion_case?1200")),
		msgs2.NewIlRow(msgs2.NewIlDataButton("promotion_case_3", "/promotion_case?2000")),
	).Build(s.UserLang)

	msgs2.NewParseMarkUpMessage(s.BotLang, int64(s.UserID), markUp, text)
}

type PromotionCaseAnswerCommand struct {
}

func NewPromotionCaseAnswerCommand() *PromotionCaseAnswerCommand {
	return &PromotionCaseAnswerCommand{}
}

func (c *PromotionCaseAnswerCommand) Serve(s bots.Situation) {
	cost, err := strconv.Atoi(strings.Split(s.Params.Level, "?")[1])
	if err != nil {
		log.Println(err)
		return
	}

	user := auth.GetUser(s.BotLang, s.UserID)
	if user.Balance < cost {
		smtWentWrong(s)

		sendMainMenu(s)
		return
	}
	user.Balance -= cost
	dataBase := bots.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateBalanceQuery, user.Balance, user.ID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(s.UserLang, "promotion_successfully_order"))
	msgs2.SendMsgToUser(s.BotLang, msg)

	sendMainMenu(s)
}

func smtWentWrong(s bots.Situation) {
	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(s.UserLang, "something_went_wrong"))

	msgs2.SendMsgToUser(s.BotLang, msg)
}

type AdminLogOutCommand struct {
}

func NewAdminLogOutCommand() *AdminLogOutCommand {
	return &AdminLogOutCommand{}
}

func (c *AdminLogOutCommand) Serve(s bots.Situation) {
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)
	simpleAdminMsg(s, "admin_log_out")

	sendMainMenu(s)
}

func simpleAdminMsg(s bots.Situation, key string) {
	text := assets.AdminText(s.UserLang, key)
	msg := tgbotapi.NewMessage(int64(s.UserID), text)

	msgs2.SendMsgToUser(s.BotLang, msg)
}

func sendMainMenu(s bots.Situation) {
	s.Command = "/start"
	s.Err = nil
	checkMessage(s) //Send start menu
}

func fillDate(text string) string {
	currentTime := time.Now()
	//formatTime := currentTime.Format("02.01.2006 15.04")

	users := currentTime.Unix()/6000 - 265000
	totalEarned := currentTime.Unix()/5*5 - 1622000000
	totalVoice := totalEarned / 7
	return fmt.Sprintf(text /*formatTime,*/, users, totalEarned, totalVoice)
}
