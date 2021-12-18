package services

import (
	"strconv"
	"strings"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/log"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/Stepan1328/youtube-assist-bot/services/administrator"
	"github.com/Stepan1328/youtube-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallBackHandlers struct {
	Handlers map[string]model.Handler
}

func (h *CallBackHandlers) GetHandler(command string) model.Handler {
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
	//h.OnCommand("/send_change_lang", NewSendLanguageCommand())
	//h.OnCommand("/change_lang", NewChangeLanguageCommand())
}

func (h *CallBackHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func checkCallbackQuery(s model.Situation, logger log.Logger) {
	if strings.Contains(s.Params.Level, "admin") {
		if err := administrator.CheckAdminCallback(s); err != nil {
			logger.Warn("error with serve admin callback command: %s", err.Error())
		}
		return
	}

	Handler := model.Bots[s.BotLang].CallbackHandler.
		GetHandler(s.Command)

	if Handler != nil {
		if err := Handler.Serve(s); err != nil {
			logger.Warn("error with serve user callback command: %s", err.Error())
			smthWentWrong(s.BotLang, s.CallbackQuery.Message.Chat.ID, s.User.Language)
		}
		return
	}

	logger.Warn("get callback data='%s', but they didn't react in any way", s.CallbackQuery.Data)
}

type GetBonusCommand struct {
}

func NewGetBonusCommand() *GetBonusCommand {
	return &GetBonusCommand{}
}

func (c *GetBonusCommand) Serve(s model.Situation) error {
	return auth.GetABonus(s)
}

type RepeatTaskCommand struct {
}

func NewRepeatTaskCommand() *RepeatTaskCommand {
	return &RepeatTaskCommand{}
}

func (c *RepeatTaskCommand) Serve(s model.Situation) error {
	s.Command = strings.Split(s.CallbackQuery.Data, "?")[0] + "?"

	_ = msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "watch_previous_video")

	Handler := model.Bots[s.BotLang].MessageHandler.
		GetHandler(s.Command)

	if Handler != nil {
		return Handler.Serve(s)
	}

	return nil
}

type WithdrawalMainCommand struct {
}

func NewWithdrawalMainCommand() *WithdrawalMainCommand {
	return &WithdrawalMainCommand{}
}

func (c *WithdrawalMainCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "withdrawal")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "select_payment"))

	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("paypal_method"),
			msgs.NewDataButton("credit_card_method")),
		msgs.NewRow(msgs.NewDataButton("back_to_main_menu_button")),
	).Build(s.User.Language)

	_ = msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "make_a_choice")
	return msgs.SendMsgToUser(s.BotLang, msg)
}

type RecheckSubscribeCommand struct {
}

func NewRecheckSubscribeCommand() *RecheckSubscribeCommand {
	return &RecheckSubscribeCommand{}
}

func (c *RecheckSubscribeCommand) Serve(s model.Situation) error {
	amount := strings.Split(s.CallbackQuery.Data, "?")[1]
	s.Message = &tgbotapi.Message{
		Text: amount,
	}
	_ = msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "invitation_to_subscribe")

	amountInt, _ := strconv.Atoi(amount)

	if auth.CheckSubscribeToWithdrawal(s, amountInt) {
		db.RdbSetUser(s.BotLang, s.User.ID, "main")

		return NewStartCommand().Serve(s)
	}

	return nil
}

type PromotionCaseCommand struct {
}

func NewPromotionCaseCommand() *PromotionCaseCommand {
	return &PromotionCaseCommand{}
}

func (c *PromotionCaseCommand) Serve(s model.Situation) error {
	user, err := auth.GetUser(s.BotLang, s.User.ID)
	if err != nil {
		return err
	}

	cost, err := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])
	if err != nil {
		return err
	}

	if user.Balance < cost {
		return msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "not_enough_money")
	}

	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "invitation_to_send_link_text"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	_ = msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "invitation_to_send_link")
	return msgs.SendMsgToUser(s.BotLang, msg)
}
