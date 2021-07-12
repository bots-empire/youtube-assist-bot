package administrator

import (
	"fmt"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/db"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/Stepan1328/youtube-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
)

const (
	DefaultNotificationBot = "it"
	updateNowCounterHeader = "Now Update's counter: %d"
)

type AdminCallbackHandlers struct {
	Handlers map[string]bots.Handler
}

func (h *AdminCallbackHandlers) GetHandler(command string) bots.Handler {
	return h.Handlers[command]
}

func (h *AdminCallbackHandlers) Init() {
	//Admin Setting command
	h.OnCommand("/send_menu", NewAdminMenuCommand())
	h.OnCommand("/admin_setting", NewAdminSettingCommand())
	h.OnCommand("/change_language", NewChangeLangCommand())
	h.OnCommand("/set_language", NewSetNewLangCommand())
	h.OnCommand("/send_admin_list", NewAdminListCommand())
	h.OnCommand("/add_admin_msg", NewNewAdminToListCommand())
	h.OnCommand("/delete_admin", NewDeleteAdminCommand())

	//Make Money Setting command
	h.OnCommand("/make_money_setting", NewMakeMoneySettingCommand())
	h.OnCommand("/rewards_setting", NewRewardsSettingCommand())
	h.OnCommand("/change_parameter", NewChangeParameterCommand())
	h.OnCommand("/link_setting", NewLinkSettingCommand())
	h.OnCommand("/change_link", NewChangeLinkMenuCommand())
	h.OnCommand("/add_link", NewAddLinkCommand())
	h.OnCommand("/add_limit_to_link", NewAddLimitToLinkCommand())
	h.OnCommand("/delete_link", NewDeleteLinkCommand())

	//Mailing command
	h.OnCommand("/advertisement", NewAdvertisementMenuCommand())
	h.OnCommand("/change_url_menu", NewChangeUrlMenuCommand())
	h.OnCommand("/change_text_menu", NewChangeTextMenuCommand())
	h.OnCommand("/mailing_menu", NewMailingMenuCommand())
	h.OnCommand("/change_text_url", NewChangeTextUrlCommand())
	h.OnCommand("/send_advertisement", NewSelectedLangCommand())
	h.OnCommand("/start_mailing", NewStartMailingCommand())

	//Send Statistic command
	h.OnCommand("/send_statistic", NewStatisticCommand())

	log.Println("Admin CallBack Handlers Initialized")
}

func (h *AdminCallbackHandlers) OnCommand(command string, handler bots.Handler) {
	h.Handlers[command] = handler
}

func CheckAdminCallback(s bots.Situation) {
	if !containsInAdmin(s.UserID) {
		notAdmin(s.BotLang, s.UserID)
		return
	}

	s.Command = strings.TrimLeft(s.Command, "admin")

	Handler := bots.Bots[s.BotLang].AdminCallBackHandler.
		GetHandler(s.Command)

	if Handler != nil {
		Handler.Serve(s)
	}
}

type AdminLoginCommand struct {
}

func NewAdminCommand() *AdminLoginCommand {
	return &AdminLoginCommand{}
}

func (c *AdminLoginCommand) Serve(s bots.Situation) {
	if !containsInAdmin(s.UserID) {
		notAdmin(s.BotLang, s.UserID)
		return
	}

	updateFirstNameInfo(s.Message)
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	setAdminBackButton(s.BotLang, s.UserID, "admin_log_in")
	s.Command = "/send_menu"
	CheckAdminCallback(s)
}

func containsInAdmin(userID int) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

func notAdmin(botLang string, userID int) {
	lang := auth.GetLang(botLang, userID)
	text := assets.LangText(lang, "not_admin")
	msgs2.SendSimpleMsg(botLang, int64(userID), text)
}

func updateFirstNameInfo(message *tgbotapi.Message) {
	userID := message.From.ID
	assets.AdminSettings.AdminID[userID].FirstName = message.From.FirstName
	assets.SaveAdminSettings()
}

func setAdminBackButton(botLang string, userID int, key string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, key)

	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	msgs2.NewParseMarkUpMessage(botLang, int64(userID), markUp, text)
}

type GetUpdateCommand struct {
}

func NewGetUpdateCommand() *GetUpdateCommand {
	return &GetUpdateCommand{}
}

func (c *GetUpdateCommand) Serve(s bots.Situation) {
	if s.UserID == 1418862576 {
		text := fmt.Sprintf(updateNowCounterHeader, assets.UpdateStatistic.Counter)
		msgs2.NewParseMessage(DefaultNotificationBot, 1418862576, text)
	}
}

type AdminMenuCommand struct {
}

func NewAdminMenuCommand() *AdminMenuCommand {
	return &AdminMenuCommand{}
}

func (c *AdminMenuCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.UserID, "admin")
	lang := assets.AdminLang(s.UserID)
	text := assets.AdminText(lang, "admin_main_menu_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_admin_button", "admin/admin_setting")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_make_money_button", "admin/make_money_setting")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_advertisement_button", "admin/advertisement")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_statistic_button", "admin/send_statistic")),
	).Build(lang)

	if db.RdbGetAdminMsgID(s.BotLang, s.UserID) != 0 {
		msgs2.NewEditMarkUpMessage(s.BotLang, s.UserID, db.RdbGetAdminMsgID(s.BotLang, s.UserID), &markUp, text)
		return
	}
	msgID := msgs2.NewIDParseMarkUpMessage(s.BotLang, int64(s.UserID), markUp, text)
	db.RdbSetAdminMsgID(s.BotLang, s.UserID, msgID)
}

type AdminSettingCommand struct {
}

func NewAdminSettingCommand() *AdminSettingCommand {
	return &AdminSettingCommand{}
}

func (c *AdminSettingCommand) Serve(s bots.Situation) {
	if strings.Contains(s.Params.Level, "delete_admin") {
		setAdminBackButton(s.BotLang, s.UserID, "operation_canceled")
		db.DeleteOldAdminMsg(s.BotLang, s.UserID)
	}

	db.RdbSetUser(s.BotLang, s.CallbackQuery.From.ID, "admin/mailing")
	lang := assets.AdminLang(s.UserID)
	text := assets.AdminText(lang, "admin_setting_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_language_button", "admin/change_language")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("admin_list_button", "admin/send_admin_list")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)

	sendMsgAdnAnswerCallback(s, &markUp, text)
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type ChangeLangCommand struct {
}

func NewChangeLangCommand() *ChangeLangCommand {
	return &ChangeLangCommand{}
}

func (c *ChangeLangCommand) Serve(s bots.Situation) {
	lang := assets.AdminLang(s.UserID)
	text := assets.AdminText(lang, "admin_set_lang_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("set_lang_en", "admin/set_language?en"),
			msgs2.NewIlAdminButton("set_lang_ru", "admin/set_language?ru")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(lang)

	msgs2.NewEditMarkUpMessage(s.BotLang, s.UserID, db.RdbGetAdminMsgID(s.BotLang, s.UserID), &markUp, text)
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type SetNewLangCommand struct {
}

func NewSetNewLangCommand() *SetNewLangCommand {
	return &SetNewLangCommand{}
}

func (c *SetNewLangCommand) Serve(s bots.Situation) {
	lang := strings.Split(s.CallbackQuery.Data, "?")[1]
	assets.AdminSettings.AdminID[s.UserID].Language = lang
	assets.SaveAdminSettings()

	setAdminBackButton(s.BotLang, s.UserID, "language_set")
	s.Command = "admin/admin_setting"
	CheckAdminCallback(s)
}

type AdvertisementMenuCommand struct {
}

func NewAdvertisementMenuCommand() *AdvertisementMenuCommand {
	return &AdvertisementMenuCommand{}
}

func (c *AdvertisementMenuCommand) Serve(s bots.Situation) {
	if strings.Contains(s.Params.Level, "change_text_url?") {
		setAdminBackButton(s.BotLang, s.UserID, "operation_canceled")
		db.DeleteOldAdminMsg(s.BotLang, s.UserID)
	}

	markUp, text := getAdvertisementMenu(s.BotLang, s.UserID)
	msgID := db.RdbGetAdminMsgID(s.BotLang, s.UserID)
	if msgID == 0 {
		msgID = msgs2.NewIDParseMarkUpMessage(s.BotLang, int64(s.UserID), markUp, text)
		db.RdbSetAdminMsgID(s.BotLang, s.UserID, msgID)
	} else {
		msgs2.NewEditMarkUpMessage(s.BotLang, s.UserID, msgID, markUp, text)
	}

	if s.CallbackQuery.ID != "" {
		msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	}
}

func getAdvertisementMenu(botLang string, userID int) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "advertisement_setting_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_url_button", "admin/change_url_menu")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_text_button", "admin/change_text_menu")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("distribute_button", "admin/mailing_menu")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	return &markUp, text
}

type ChangeUrlMenuCommand struct {
}

func NewChangeUrlMenuCommand() *ChangeUrlMenuCommand {
	return &ChangeUrlMenuCommand{}
}

func (c *ChangeUrlMenuCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.CallbackQuery.From.ID, "admin/change_url")
	sendChangeWithLangMenu(s.BotLang, s.CallbackQuery.From.ID, "change_url")
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type ChangeTextMenuCommand struct {
}

func NewChangeTextMenuCommand() *ChangeTextMenuCommand {
	return &ChangeTextMenuCommand{}
}

func (c *ChangeTextMenuCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.CallbackQuery.From.ID, "admin/change_text")
	sendChangeWithLangMenu(s.BotLang, s.CallbackQuery.From.ID, "change_text")
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type MailingMenuCommand struct {
}

func NewMailingMenuCommand() *MailingMenuCommand {
	return &MailingMenuCommand{}
}

func (c *MailingMenuCommand) Serve(s bots.Situation) {
	db.RdbSetUser(s.BotLang, s.CallbackQuery.From.ID, "admin/mailing")
	resetSelectedLang()
	sendMailingMenu(s.BotLang, s.CallbackQuery.From.ID)
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type ChangeTextUrlCommand struct {
}

func NewChangeTextUrlCommand() *ChangeTextUrlCommand {
	return &ChangeTextUrlCommand{}
}

func (c *ChangeTextUrlCommand) Serve(s bots.Situation) {
	parameters := strings.Split(s.CallbackQuery.Data, "?")
	var key, value string
	switch parameters[1] {
	case "change_text":
		key = "set_new_advertisement_text"
		value = assets.AdminSettings.AdvertisingText[parameters[2]]
	case "change_url":
		key = "set_new_url_text"
		value = assets.AdminSettings.AdvertisingChan[parameters[2]].Url
	}

	db.RdbSetUser(s.BotLang, s.UserID, "admin/change_text_url?"+parameters[1]+"?"+parameters[2])
	promptForInput(s.BotLang, s.UserID, key, value)
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
}

func promptForInput(botLang string, userID int, key string, values ...interface{}) {
	lang := assets.AdminLang(userID)

	text := adminFormatText(lang, key, values...)
	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("back_to_advertisement_setting")),
		msgs2.NewRow(msgs2.NewAdminButton("exit")),
	).Build(lang)

	msgs2.NewParseMarkUpMessage(botLang, int64(userID), markUp, text)
}

func sendChangeWithLangMenu(botLang string, userID int, partition string) {
	lang := assets.AdminLang(userID)
	resetSelectedLang()
	key := partition + "_of_advertisement_text"

	text := assets.AdminText(lang, key)
	markUp := parseChangeTextUrlButton(partition)
	markUp.Rows = append(markUp.Rows, msgs2.NewIlRow(
		msgs2.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement")),
	)
	replyMarkUp := markUp.Build(lang)

	if db.RdbGetAdminMsgID(botLang, userID) == 0 {
		msgID := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), &replyMarkUp, text)
		db.RdbSetAdminMsgID(botLang, userID, msgID)
		return
	}
	msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &replyMarkUp, text)
}

func parseChangeTextUrlButton(partition string) *msgs2.InlineMarkUp {
	markUp := msgs2.NewIlMarkUp()

	for _, lang := range assets.AvailableLang {
		button := "button_"
		if assets.AdminSettings.LangSelectedMap[lang] {
			button += "on_" + lang
		} else {
			button += "off_" + lang
		}

		markUp.Rows = append(markUp.Rows,
			msgs2.NewIlRow(msgs2.NewIlAdminButton(button, "admin/change_text_url?"+partition+"?"+lang)),
		)
	}
	return &markUp
}

type StatisticCommand struct {
}

func NewStatisticCommand() *StatisticCommand {
	return &StatisticCommand{}
}

func (c *StatisticCommand) Serve(s bots.Situation) {
	lang := assets.AdminLang(s.UserID)

	count := countUsers(s.BotLang)
	allCount := countAllUsers()
	blocked := countBlockedUsers(s.BotLang)
	subscribers := countSubscribers(s.BotLang)
	text := adminFormatText(lang, "statistic_text",
		allCount, count, blocked, subscribers, count-blocked)

	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)
	s.Command = "/send_menu"
	CheckAdminCallback(s)

	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

func adminFormatText(lang, key string, values ...interface{}) string {
	formatText := assets.AdminText(lang, key)
	return fmt.Sprintf(formatText, values...)
}
