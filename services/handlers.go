package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/log"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/Stepan1328/youtube-assist-bot/services/administrator"
	"github.com/Stepan1328/youtube-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	updateCounterHeader = "Today Update's counter: %d"
	updatePrintHeader   = "update number: %d    // youtube-bot-update:  %s %s"
	extraneousUpdate    = "extraneous update"
	notificationChatID  = 1418862576
	godUserID           = 1418862576

	defaultTimeInServiceMod = time.Hour * 2

	updateBalanceQuery = "UPDATE users SET balance = ? WHERE id = ?;"
)

type MessagesHandlers struct {
	Handlers map[string]model.Handler
}

func (h *MessagesHandlers) GetHandler(command string) model.Handler {
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
	h.OnCommand("/withdrawal_pix", NewWithdrawalMethodPixCommand())
	h.OnCommand("/withdrawal_req_amount", NewReqWithdrawalAmountCommand())
	h.OnCommand("/withdrawal_exit", NewWithdrawalAmountCommand())
	//h.OnCommand("/promotion_choice", NewPromotionCommand())
	//h.OnCommand("/promotion_case", NewPromotionCaseAnswerCommand())

	//Log out command
	h.OnCommand("/admin_log_out", NewAdminLogOutCommand())

	//Tech command
	h.OnCommand("/MaintenanceModeOn", NewMaintenanceModeOnCommand())
	h.OnCommand("/MaintenanceModeOff", NewMaintenanceModeOffCommand())
}

func (h *MessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func ActionsWithUpdates(botLang string, updates tgbotapi.UpdatesChannel, logger log.Logger) {
	for update := range updates {
		localUpdate := update

		go checkUpdate(botLang, &localUpdate, logger)
	}
}

func checkUpdate(botLang string, update *tgbotapi.Update, logger log.Logger) {
	defer panicCather(botLang, update)

	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if update.Message != nil && update.Message.PinnedMessage != nil {
		return
	}

	printNewUpdate(botLang, update, logger)
	if update.Message != nil {
		user, err := auth.CheckingTheUser(botLang, update.Message)
		if err != nil {
			emptyLevel(botLang, update.Message, botLang)
			logger.Warn("err with check user: %s", err.Error())
			return
		}

		situation := createSituationFromMsg(botLang, update.Message, user)
		checkMessage(situation, logger)
		return
	}

	if update.CallbackQuery != nil {
		situation, err := createSituationFromCallback(botLang, update.CallbackQuery)
		if err != nil {
			smthWentWrong(botLang, update.CallbackQuery.Message.Chat.ID, botLang)
			logger.Warn("err with create situation from callback: %s", err.Error())
			return
		}

		checkCallbackQuery(situation, logger)
		return
	}
}

func printNewUpdate(botLang string, update *tgbotapi.Update, logger log.Logger) {
	assets.UpdateStatistic.Mu.Lock()
	defer assets.UpdateStatistic.Mu.Unlock()

	if (time.Now().Unix())/86400 > int64(assets.UpdateStatistic.Day) {
		sendTodayUpdateMsg()
	}

	assets.UpdateStatistic.Counter++
	assets.SaveUpdateStatistic()

	model.HandleUpdates.WithLabelValues(
		model.GetGlobalBot(botLang).BotLink,
		botLang,
	).Inc()

	if update.Message != nil {
		if update.Message.Text != "" {
			logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, update.Message.Text)
			return
		}
	}

	if update.CallbackQuery != nil {
		logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, update.CallbackQuery.Data)
		return
	}

	logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, extraneousUpdate)
}

func sendTodayUpdateMsg() {
	text := fmt.Sprintf(updateCounterHeader, assets.UpdateStatistic.Counter)
	id := msgs.SendNotificationToDeveloper(text)
	msgs.PinMsgToDeveloper(id)

	assets.UpdateStatistic.Counter = 0
	assets.UpdateStatistic.Day = int(time.Now().Unix()) / 86400
}

func createSituationFromMsg(botLang string, message *tgbotapi.Message, user *model.User) model.Situation {
	return model.Situation{
		Message: message,
		BotLang: botLang,
		User:    user,
		Params: model.Parameters{
			Level: db.GetLevel(botLang, message.From.ID),
		},
	}
}

func createSituationFromCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery) (model.Situation, error) {
	user, err := auth.GetUser(botLang, callbackQuery.From.ID)
	if err != nil {
		return model.Situation{}, err
	}

	return model.Situation{
		CallbackQuery: callbackQuery,
		BotLang:       botLang,
		User:          user,
		Command:       strings.Split(callbackQuery.Data, "?")[0],
		Params: model.Parameters{
			Level: db.GetLevel(botLang, callbackQuery.From.ID),
		},
	}, nil
}

func checkMessage(situation model.Situation, logger log.Logger) {
	if model.Bots[situation.BotLang].MaintenanceMode {
		if situation.User.ID != godUserID {
			msg := tgbotapi.NewMessage(situation.User.ID, "The bot is under maintenance, please try again later")
			_ = msgs.SendMsgToUser(situation.BotLang, msg)
			return
		}
	}

	if situation.Command == "" {
		situation.Command, situation.Err = assets.GetCommandFromText(situation)
	}

	if situation.Err == nil {
		Handler := model.Bots[situation.BotLang].MessageHandler.
			GetHandler(situation.Command)

		if Handler != nil {
			err := Handler.Serve(situation)
			if err != nil {
				logger.Warn("error with serve user msg command: %s", err.Error())
				smthWentWrong(situation.BotLang, situation.Message.Chat.ID, situation.User.Language)
			}
			return
		}
	}

	situation.Command = strings.Split(situation.Params.Level, "?")[0]

	Handler := model.Bots[situation.BotLang].MessageHandler.
		GetHandler(situation.Command)

	if Handler != nil {
		err := Handler.Serve(situation)
		if err != nil {
			logger.Warn("error with serve user level command: %s", err.Error())
			smthWentWrong(situation.BotLang, situation.Message.Chat.ID, situation.User.Language)
		}
		return
	}

	if err := administrator.CheckAdminMessage(situation); err == nil {
		return
	}

	emptyLevel(situation.BotLang, situation.Message, situation.User.Language)
	if situation.Err != nil {
		logger.Info(situation.Err.Error())
	}
}

func smthWentWrong(botLang string, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, assets.LangText(lang, "user_level_not_defined"))
	_ = msgs.SendMsgToUser(botLang, msg)
}

func emptyLevel(botLang string, message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_msg_dont_recognize"))
	_ = msgs.SendMsgToUser(botLang, msg)
}

type StartCommand struct {
}

func NewStartCommand() *StartCommand {
	return &StartCommand{}
}

func (c *StartCommand) Serve(s model.Situation) error {
	if strings.Contains(s.Message.Text, "new_admin") {
		s.Command = s.Message.Text
		return administrator.CheckNewAdmin(s)
	}

	text := assets.LangText(s.User.Language, "main_select_menu")
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("main_make_money")),
		msgs.NewRow(msgs.NewDataButton("spend_money_withdrawal")),
		//msgs.NewRow(msgs.NewDataButton("main_spend_money")),
		msgs.NewRow(msgs.NewDataButton("main_money_for_a_friend"),
			msgs.NewDataButton("main_more_money")),
		msgs.NewRow(msgs.NewDataButton("main_profile"),
			msgs.NewDataButton("main_statistic")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type MakeMoneyCommand struct {
}

func NewMakeMoneyCommand() *MakeMoneyCommand {
	return &MakeMoneyCommand{}
}

func (c *MakeMoneyCommand) Serve(s model.Situation) error {
	text := assets.LangText(s.User.Language, "make_money_text")
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("make_money_youtube")),
		msgs.NewRow(msgs.NewDataButton("make_money_tiktok")),
		msgs.NewRow(msgs.NewDataButton("make_money_advertisement")),
		msgs.NewRow(msgs.NewDataButton("back_to_main_menu_button")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type LinkTaskCommand struct {
}

func NewLinkTaskCommand() *LinkTaskCommand {
	return &LinkTaskCommand{}
}

func (c *LinkTaskCommand) Serve(s model.Situation) error {
	s.Params.Partition = strings.Split(strings.Replace(s.Command, "/make_money_", "", 1), "?")[0]
	return auth.MakeMoney(s, assets.AdminSettings.Parameters[s.BotLang].SecondBetweenViews)
}

type VideoTaskCommand struct {
}

func NewVideoTaskCommand() *VideoTaskCommand {
	return &VideoTaskCommand{}
}

func (c *VideoTaskCommand) Serve(s model.Situation) error {
	return auth.MakeMoney(s, getMakeMoneyDuration(&s))
}

func getMakeMoneyDuration(s *model.Situation) int64 {
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

func (c *SpendMoneyCommand) Serve(s model.Situation) error {
	text := assets.LangText(s.User.Language, "make_money_text")
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("spend_money_withdrawal")),
		msgs.NewRow(msgs.NewDataButton("spend_money_promotion")),
		msgs.NewRow(msgs.NewDataButton("back_to_main_menu_button")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type SpendMoneyWithdrawalCommand struct {
}

func NewSpendMoneyWithdrawalCommand() *SpendMoneyWithdrawalCommand {
	return &SpendMoneyWithdrawalCommand{}
}

func (c *SpendMoneyWithdrawalCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "withdrawal")

	text := msgs.GetFormatText(s.User.Language, "withdrawal_money", s.User.Balance)
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_7")),
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_1"),
			msgs.NewDataButton("withdrawal_method_2")),
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_3"),
			msgs.NewDataButton("withdrawal_method_4")),
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_5"),
			msgs.NewDataButton("withdrawal_method_6")),
		msgs.NewRow(msgs.NewDataButton("back_to_main_menu_button")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markUp, text)
}

type PaypalReqCommand struct {
}

func NewPaypalReqCommand() *PaypalReqCommand {
	return &PaypalReqCommand{}
}

func (c *PaypalReqCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	lang := auth.GetLang(s.BotLang, s.User.ID)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(lang, "paypal_email"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(lang)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type CreditCardReqCommand struct {
}

func NewCreditCardReqCommand() *CreditCardReqCommand {
	return &CreditCardReqCommand{}
}

func (c *CreditCardReqCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	lang := auth.GetLang(s.BotLang, s.User.ID)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(lang, "credit_card_number"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(lang)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalMethodCommand struct {
}

func NewWithdrawalMethodCommand() *WithdrawalMethodCommand {
	return &WithdrawalMethodCommand{}
}

func (c *WithdrawalMethodCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	lang := auth.GetLang(s.BotLang, s.User.ID)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(lang, "request_number_email"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(lang)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalMethodPixCommand struct {
}

func NewWithdrawalMethodPixCommand() *WithdrawalMethodPixCommand {
	return &WithdrawalMethodPixCommand{}
}

func (c *WithdrawalMethodPixCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	lang := auth.GetLang(s.BotLang, s.User.ID)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(lang, "request_pix_code"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(lang)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type ReqWithdrawalAmountCommand struct {
}

func NewReqWithdrawalAmountCommand() *ReqWithdrawalAmountCommand {
	return &ReqWithdrawalAmountCommand{}
}

func (c *ReqWithdrawalAmountCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_exit")

	lang := auth.GetLang(s.BotLang, s.User.ID)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(lang, "req_withdrawal_amount"))

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalAmountCommand struct {
}

func NewWithdrawalAmountCommand() *WithdrawalAmountCommand {
	return &WithdrawalAmountCommand{}
}

func (c *WithdrawalAmountCommand) Serve(s model.Situation) error {
	return auth.WithdrawMoneyFromBalance(s, s.Message.Text)
}

type SendProfileCommand struct {
}

func NewSendProfileCommand() *SendProfileCommand {
	return &SendProfileCommand{}
}

func (c *SendProfileCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	text := msgs.GetFormatText(s.User.Language, "profile_text",
		s.Message.From.FirstName, s.User.Balance, s.User.Completed, s.User.ReferralCount)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)

	//markUp := msgs.NewIlMarkUp(
	//	msgs.NewIlRow(msgs.NewIlDataButton("change_lang_button", "/send_change_lang")),
	//).Build(user.Language)
	//
	//msgs.NewParseMarkUpMessage(s.BotLang, int64(user.ID), markUp, text)
}

type SendStatisticsCommand struct {
}

func NewSendStatisticsCommand() *SendStatisticsCommand {
	return &SendStatisticsCommand{}
}

func (c *SendStatisticsCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	text := assets.LangText(s.User.Language, "statistic_to_user")

	text = fillDate(text)
	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type MoneyForAFriendCommand struct {
}

func NewMoneyForAFriendCommand() *MoneyForAFriendCommand {
	return &MoneyForAFriendCommand{}
}

func (c *MoneyForAFriendCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	link, err := model.EncodeLink(s.BotLang, &model.ReferralLinkInfo{
		ReferralID: s.User.ID,
		Source:     "bot",
	})
	if err != nil {
		return err
	}

	text := msgs.GetFormatText(s.User.Language, "referral_text",
		link,
		assets.AdminSettings.Parameters[s.BotLang].ReferralAmount,
		s.User.ReferralCount)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type MoreMoneyCommand struct {
}

func NewMoreMoneyCommand() *MoreMoneyCommand {
	return &MoreMoneyCommand{}
}

func (c *MoreMoneyCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	text := msgs.GetFormatText(s.User.Language, "more_money_text",
		assets.AdminSettings.Parameters[s.BotLang].BonusAmount, assets.AdminSettings.Parameters[s.BotLang].BonusAmount)

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.User.Language].Url)),
		msgs.NewIlRow(msgs.NewIlDataButton("get_bonus_button", "/send_bonus_to_user")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type PromotionCommand struct {
}

func NewPromotionCommand() *PromotionCommand {
	return &PromotionCommand{}
}

func (c *PromotionCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	text := assets.LangText(s.User.Language, "promotion_main_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("promotion_case_1", "/promotion_case?500")),
		msgs.NewIlRow(msgs.NewIlDataButton("promotion_case_2", "/promotion_case?1200")),
		msgs.NewIlRow(msgs.NewIlDataButton("promotion_case_3", "/promotion_case?2000")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
}

type PromotionCaseAnswerCommand struct {
}

func NewPromotionCaseAnswerCommand() *PromotionCaseAnswerCommand {
	return &PromotionCaseAnswerCommand{}
}

func (c *PromotionCaseAnswerCommand) Serve(s model.Situation) error {
	cost, err := strconv.Atoi(strings.Split(s.Params.Level, "?")[1])
	if err != nil {
		return err
	}

	if s.User.Balance < cost {
		smtWentWrong(s)

		return NewStartCommand().Serve(s)
	}

	s.User.Balance -= cost
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateBalanceQuery, s.User.Balance, s.User.ID)
	if err != nil {
		panic(err.Error())
	}
	rows.Close()

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "promotion_successfully_order"))
	if err := msgs.SendMsgToUser(s.BotLang, msg); err != nil {
		return err
	}

	return NewStartCommand().Serve(s)
}

func smtWentWrong(s model.Situation) {
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "something_went_wrong"))

	_ = msgs.SendMsgToUser(s.BotLang, msg)
}

type AdminLogOutCommand struct {
}

func NewAdminLogOutCommand() *AdminLogOutCommand {
	return &AdminLogOutCommand{}
}

func (c *AdminLogOutCommand) Serve(s model.Situation) error {
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	if err := simpleAdminMsg(s, "admin_log_out"); err != nil {
		return err
	}

	return NewStartCommand().Serve(s)
}

type MaintenanceModeOnCommand struct {
}

func NewMaintenanceModeOnCommand() *MaintenanceModeOnCommand {
	return &MaintenanceModeOnCommand{}
}

func (c *MaintenanceModeOnCommand) Serve(s model.Situation) error {
	if s.User.ID != godUserID {
		return nil
	}

	for botLang := range model.Bots {
		model.Bots[botLang].MaintenanceMode = true
	}

	go func() {
		time.Sleep(defaultTimeInServiceMod)
		_ = NewMaintenanceModeOffCommand().Serve(s)
	}()

	msg := tgbotapi.NewMessage(s.User.ID, "Режим технического обслуживания включен")
	return msgs.SendMsgToUser(s.BotLang, msg)
}

type MaintenanceModeOffCommand struct {
}

func NewMaintenanceModeOffCommand() *MaintenanceModeOffCommand {
	return &MaintenanceModeOffCommand{}
}

func (c *MaintenanceModeOffCommand) Serve(s model.Situation) error {
	if s.User.ID != godUserID {
		return nil
	}

	for botLang := range model.Bots {
		model.Bots[botLang].MaintenanceMode = false
	}

	msg := tgbotapi.NewMessage(s.User.ID, "Режим технического обслуживания отключен")
	return msgs.SendMsgToUser(s.BotLang, msg)
}

func simpleAdminMsg(s model.Situation, key string) error {
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, key)
	msg := tgbotapi.NewMessage(s.User.ID, text)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func fillDate(text string) string {
	currentTime := time.Now()
	//formatTime := currentTime.Format("02.01.2006 15.04")

	users := currentTime.Unix()/6000 - 265000
	totalEarned := currentTime.Unix()/5*5 - 1622000000
	totalVoice := totalEarned / 7
	return fmt.Sprintf(text /*formatTime,*/, users, totalEarned, totalVoice)
}
