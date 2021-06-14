package services

import (
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
)

type CallBackHandlers struct {
	Handlers map[string]bots.Handler
}

func (h *CallBackHandlers) GetHandler(command string) bots.Handler {
	return h.Handlers[command]
}

func (h *CallBackHandlers) Init() {
	//Money command
	h.OnCommand("/send_bonus_to_user", NewGetBonusCommand())
	h.OnCommand("/make_money_advertisement", NewRepeatTaskCommand())
	h.OnCommand("/make_money_youtube", NewRepeatTaskCommand())
	h.OnCommand("/make_money_tiktok", NewRepeatTaskCommand())
	//h.OnCommand("/withdrawal_main", NewWithdrawalMainCommand())
	h.OnCommand("/withdrawal_money", NewRecheckSubscribeCommand())
	h.OnCommand("/promotion_case", NewPromotionCaseCommand())

	//Change language command
	h.OnCommand("/send_change_lang", NewSendLanguageCommand())
	h.OnCommand("/change_lang", NewChangeLanguageCommand())

	log.Println("CallBack Handlers Initialized")
}

func (h *CallBackHandlers) OnCommand(command string, handler bots.Handler) {
	h.Handlers[command] = handler
}

func checkCallbackQuery(s bots.Situation) {
	if strings.Contains(s.Params.Level, "admin") {
		administrator.CheckAdminCallback(s)
	}

	Handler := bots.Bots[s.BotLang].CallbackHandler.
		GetHandler(s.Command)

	if Handler != nil {
		Handler.Serve(s)
		return
	}
}

type GetBonusCommand struct {
}

func NewGetBonusCommand() *GetBonusCommand {
	return &GetBonusCommand{}
}

func (c *GetBonusCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)

	user.GetABonus(s)
}

type RepeatTaskCommand struct {
}

func NewRepeatTaskCommand() *RepeatTaskCommand {
	return &RepeatTaskCommand{}
}

func (c *RepeatTaskCommand) Serve(s bots.Situation) {
	s.Command = strings.Split(s.CallbackQuery.Data, "?")[0] + "?"

	msgs2.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.UserLang, "watch_previous_video")
	checkMessage(s)
}

type WithdrawalMainCommand struct {
}

func NewWithdrawalMainCommand() *WithdrawalMainCommand {
	return &WithdrawalMainCommand{}
}

func (c *WithdrawalMainCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "withdrawal")

	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(s.UserLang, "select_payment"))

	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("paypal_method"),
			msgs2.NewDataButton("credit_card_method")),
		msgs2.NewRow(msgs2.NewDataButton("back_to_main_menu_button")),
	).Build(s.UserLang)

	msgs2.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.UserLang, "make_a_choice")
	msgs2.SendMsgToUser(s.BotLang, msg)
}

type ChangeLanguageCommand struct {
}

func NewChangeLanguageCommand() *ChangeLanguageCommand {
	return &ChangeLanguageCommand{}
}

func (c *ChangeLanguageCommand) Serve(s bots.Situation) {
	newLang := getNewLang(s)

	setLanguage(s, newLang)
	msgs2.SendAnswerCallback(s.BotLang, s.CallbackQuery, newLang, "language_successful_set")
	db.DeleteTemporaryMessages(s.BotLang, s.CallbackQuery.From.ID)
}

func getNewLang(s bots.Situation) string {
	return strings.Split(s.CallbackQuery.Data, "?")[1]
}

func setLanguage(s bots.Situation, newLang string) {
	db.RdbSetUser(s.BotLang, s.UserID, "main")

	if newLang == "back" {
		s.Command = "/start"
		checkMessage(s)
		return
	}

	dataBase := bots.GetDB(s.BotLang)
	_, err := dataBase.Query("UPDATE users SET lang = ? WHERE id = ?;", newLang, s.UserID)
	if err != nil {
		panic(err.Error())
	}

	s.Command = "/start"
	checkMessage(s)
}

type RecheckSubscribeCommand struct {
}

func NewRecheckSubscribeCommand() *RecheckSubscribeCommand {
	return &RecheckSubscribeCommand{}
}

func (c *RecheckSubscribeCommand) Serve(s bots.Situation) {
	amount := strings.Split(s.CallbackQuery.Data, "?")[1]
	s.Message = &tgbotapi.Message{
		Text: amount,
	}
	msgs2.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.UserLang, "invitation_to_subscribe")
	u := auth.GetUser(s.BotLang, s.UserID)
	amountInt, _ := strconv.Atoi(amount)

	if u.CheckSubscribeToWithdrawal(s, amountInt) {
		db.RdbSetUser(s.BotLang, s.UserID, "main")

		sendMainMenu(s)
	}
}

type PromotionCaseCommand struct {
}

func NewPromotionCaseCommand() *PromotionCaseCommand {
	return &PromotionCaseCommand{}
}

func (c *PromotionCaseCommand) Serve(s bots.Situation) {
	user := auth.GetUser(s.BotLang, s.UserID)
	cost, err := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])
	if err != nil {
		log.Println(err)
		return
	}

	if user.Balance < cost {
		msgs2.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.UserLang, "not_enough_money")
		return
	}

	db.RdbSetUser(s.BotLang, s.UserID, s.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(s.UserLang, "invitation_to_send_link_text"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(s.UserLang)

	msgs2.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.UserLang, "invitation_to_send_link")
	msgs2.SendMsgToUser(s.BotLang, msg)
}

type SendLanguageCommand struct {
}

func NewSendLanguageCommand() *SendLanguageCommand {
	return &SendLanguageCommand{}
}

func (c *SendLanguageCommand) Serve(s bots.Situation) {
	msg := tgbotapi.NewMessage(int64(s.UserID), assets.LangText(s.UserLang, "select_language"))

	markUp := parseChangeLanguageButton()
	markUp.Rows = append(markUp.Rows, msgs2.NewIlRow(
		msgs2.NewIlDataButton("back_to_main_menu_button", "/change_lang?back")),
	)
	msg.ReplyMarkup = markUp.Build(s.UserLang)

	bot := bots.GetBot(s.BotLang)
	data, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	msgs2.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.UserLang, "make_a_choice")
	db.RdbSetTemporary(s.BotLang, s.UserID, data.MessageID)
}

func parseChangeLanguageButton() *msgs2.InlineMarkUp {
	markUp := msgs2.NewIlMarkUp()

	for _, lang := range assets.AvailableLang {
		markUp.Rows = append(markUp.Rows,
			msgs2.NewIlRow(msgs2.NewIlDataButton("lang_"+lang, "/change_lang?"+lang)),
		)
	}
	return &markUp
}
