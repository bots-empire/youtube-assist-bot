package administrator

import (
	"fmt"
	"strings"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	DefaultNotificationBot = "it"
	updateNowCounterHeader = "Now Update's counter: %d"
)

type AdminCallbackHandlers struct {
	Handlers map[string]model.Handler
}

func (h *AdminCallbackHandlers) GetHandler(command string) model.Handler {
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
	h.OnCommand("/send_advert_source_menu", NewAdvertSourceMenuCommand())
	h.OnCommand("/add_new_source", NewAddNewSourceCommand())

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
	//h.OnCommand("/change_text_url", NewChangeTextUrlCommand())
	h.OnCommand("/send_advertisement", NewSelectedLangCommand())
	h.OnCommand("/start_mailing", NewStartMailingCommand())

	//Send Statistic command
	h.OnCommand("/send_statistic", NewStatisticCommand())
}

func (h *AdminCallbackHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func CheckAdminCallback(s model.Situation) error {
	if !containsInAdmin(s.User.ID) {
		return notAdmin(s.BotLang, s.User)
	}

	s.Command = strings.TrimLeft(s.Command, "admin")

	Handler := model.Bots[s.BotLang].AdminCallBackHandler.
		GetHandler(s.Command)

	if Handler != nil {
		return Handler.Serve(s)
	}

	return model.ErrCommandNotConverted
}

type AdminLoginCommand struct {
}

func NewAdminCommand() *AdminLoginCommand {
	return &AdminLoginCommand{}
}

func (c *AdminLoginCommand) Serve(s model.Situation) error {
	if !containsInAdmin(s.User.ID) {
		return notAdmin(s.BotLang, s.User)
	}

	updateFirstNameInfo(s.Message)
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	err := setAdminBackButton(s.BotLang, s.User.ID, "admin_log_in")
	if err != nil {
		return err
	}
	return NewAdminMenuCommand().Serve(s)
}

func containsInAdmin(userID int64) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

func notAdmin(botLang string, user *model.User) error {
	text := assets.LangText(user.Language, "not_admin")
	return msgs.SendSimpleMsg(botLang, user.ID, text)
}

func updateFirstNameInfo(message *tgbotapi.Message) {
	userID := message.From.ID
	assets.AdminSettings.AdminID[userID].FirstName = message.From.FirstName
	assets.SaveAdminSettings()
}

func setAdminBackButton(botLang string, userID int64, key string) error {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, key)

	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(botLang, userID, markUp, text)
}

type GetUpdateCommand struct {
}

func NewGetUpdateCommand() *GetUpdateCommand {
	return &GetUpdateCommand{}
}

func (c *GetUpdateCommand) Serve(s model.Situation) error {
	if s.User.ID == 1418862576 {
		text := fmt.Sprintf(updateNowCounterHeader, assets.UpdateStatistic.Counter)
		return msgs.NewParseMessage(DefaultNotificationBot, 1418862576, text)
	}

	return nil
}

type AdminMenuCommand struct {
}

func NewAdminMenuCommand() *AdminMenuCommand {
	return &AdminMenuCommand{}
}

func (c *AdminMenuCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "admin_main_menu_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_admin_button", "admin/admin_setting")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_make_money_button", "admin/make_money_setting")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_advertisement_button", "admin/advertisement")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_statistic_button", "admin/send_statistic")),
	).Build(lang)

	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		return msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), &markUp, text)
	}
	msgID, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	return nil
}

type AdminSettingCommand struct {
}

func NewAdminSettingCommand() *AdminSettingCommand {
	return &AdminSettingCommand{}
}

func (c *AdminSettingCommand) Serve(s model.Situation) error {
	if strings.Contains(s.Params.Level, "delete_admin") {
		err := setAdminBackButton(s.BotLang, s.User.ID, "operation_canceled")
		if err != nil {
			return err
		}
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	}

	db.RdbSetUser(s.BotLang, s.CallbackQuery.From.ID, "admin/mailing")
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "admin_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_language_button", "admin/change_language")),
		msgs.NewIlRow(msgs.NewIlAdminButton("admin_list_button", "admin/send_admin_list")),
		msgs.NewIlRow(msgs.NewIlAdminButton("advertisement_source_button", "admin/send_advert_source_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)

	err := sendMsgAdnAnswerCallback(s, &markUp, text)
	if err != nil {
		return err
	}
	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type ChangeLangCommand struct {
}

func NewChangeLangCommand() *ChangeLangCommand {
	return &ChangeLangCommand{}
}

func (c *ChangeLangCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "admin_set_lang_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("set_lang_en", "admin/set_language?en"),
			msgs.NewIlAdminButton("set_lang_ru", "admin/set_language?ru")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(lang)

	err := msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), &markUp, text)
	if err != nil {
		return err
	}
	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type SetNewLangCommand struct {
}

func NewSetNewLangCommand() *SetNewLangCommand {
	return &SetNewLangCommand{}
}

func (c *SetNewLangCommand) Serve(s model.Situation) error {
	lang := strings.Split(s.CallbackQuery.Data, "?")[1]
	assets.AdminSettings.AdminID[s.User.ID].Language = lang
	assets.SaveAdminSettings()

	err := setAdminBackButton(s.BotLang, s.User.ID, "language_set")
	if err != nil {
		return err
	}

	return NewAdminSettingCommand().Serve(s)
}

type AdvertisementMenuCommand struct {
}

func NewAdvertisementMenuCommand() *AdvertisementMenuCommand {
	return &AdvertisementMenuCommand{}
}

func (c *AdvertisementMenuCommand) Serve(s model.Situation) error {
	if strings.Contains(s.Params.Level, "change_text_url?") {
		err := setAdminBackButton(s.BotLang, s.User.ID, "operation_canceled")
		if err != nil {
			return err
		}
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	}

	markUp, text := getAdvertisementMenu(s.BotLang, s.User.ID)
	msgID := db.RdbGetAdminMsgID(s.BotLang, s.User.ID)
	if msgID == 0 {
		msgID, _ = msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
		db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	} else {
		err := msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, msgID, markUp, text)
		if err != nil {
			return err
		}
	}

	if s.CallbackQuery != nil {
		if s.CallbackQuery.ID != "" {
			err := msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getAdvertisementMenu(botLang string, userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "advertisement_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_url_button", "admin/change_url_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_text_button", "admin/change_text_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("distribute_button", "admin/mailing_menu")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	return &markUp, text
}

type ChangeUrlMenuCommand struct {
}

func NewChangeUrlMenuCommand() *ChangeUrlMenuCommand {
	return &ChangeUrlMenuCommand{}
}

func (c *ChangeUrlMenuCommand) Serve(s model.Situation) error {
	key := "set_new_url_text"
	value := assets.AdminSettings.AdvertisingChan[s.BotLang].Url

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_text_url?change_url")
	err := promptForInput(s.BotLang, s.User.ID, key, value)
	if err != nil {
		return err
	}

	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
}

type ChangeTextMenuCommand struct {
}

func NewChangeTextMenuCommand() *ChangeTextMenuCommand {
	return &ChangeTextMenuCommand{}
}

func (c *ChangeTextMenuCommand) Serve(s model.Situation) error {
	//db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_text")
	//sendChangeWithLangMenu(s.BotLang, s.User.ID, "change_text")
	//msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")

	key := "set_new_advertisement_text"
	value := assets.AdminSettings.AdvertisingText[s.BotLang]

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_text_url?change_text")
	err := promptForInput(s.BotLang, s.User.ID, key, value)
	if err != nil {
		return err
	}

	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
}

type MailingMenuCommand struct {
}

func NewMailingMenuCommand() *MailingMenuCommand {
	return &MailingMenuCommand{}
}

func (c *MailingMenuCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.CallbackQuery.From.ID, "admin/mailing")
	resetSelectedLang()
	err := sendMailingMenu(s.BotLang, s.CallbackQuery.From.ID)
	if err != nil {
		return err
	}
	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

type ChangeTextUrlCommand struct {
}

func NewChangeTextUrlCommand() *ChangeTextUrlCommand {
	return &ChangeTextUrlCommand{}
}

func (c *ChangeTextUrlCommand) Serve(s model.Situation) error {
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

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_text_url?"+parameters[1]+"?"+parameters[2])
	err := promptForInput(s.BotLang, s.User.ID, key, value)
	if err != nil {
		return err
	}
	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
}

func promptForInput(botLang string, userID int64, key string, values ...interface{}) error {
	lang := assets.AdminLang(userID)

	text := adminFormatText(lang, key, values...)
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_advertisement_setting")),
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(botLang, userID, markUp, text)
}

func sendChangeWithLangMenu(botLang string, userID int64, partition string) error {
	lang := assets.AdminLang(userID)
	resetSelectedLang()
	key := partition + "_of_advertisement_text"

	text := assets.AdminText(lang, key)
	markUp := parseChangeTextUrlButton(partition)
	markUp.Rows = append(markUp.Rows, msgs.NewIlRow(
		msgs.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement")),
	)
	replyMarkUp := markUp.Build(lang)

	if db.RdbGetAdminMsgID(botLang, userID) == 0 {
		msgID, err := msgs.NewIDParseMarkUpMessage(botLang, userID, &replyMarkUp, text)
		if err != nil {
			return err
		}
		db.RdbSetAdminMsgID(botLang, userID, msgID)
		return nil
	}
	return msgs.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &replyMarkUp, text)
}

func parseChangeTextUrlButton(partition string) *msgs.InlineMarkUp {
	markUp := msgs.NewIlMarkUp()

	for _, lang := range assets.AvailableLang {
		button := "button_"
		if assets.AdminSettings.LangSelectedMap[lang] {
			button += "on_" + lang
		} else {
			button += "off_" + lang
		}

		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton(button, "admin/change_text_url?"+partition+"?"+lang)),
		)
	}
	return &markUp
}

type StatisticCommand struct {
}

func NewStatisticCommand() *StatisticCommand {
	return &StatisticCommand{}
}

func (c *StatisticCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)

	allCount := countAllUsers()
	count := countUsers(s.BotLang)
	referrals := countReferrals(s.BotLang, count)
	blocked := countBlockedUsers(s.BotLang)
	subscribers := countSubscribers(s.BotLang)
	text := adminFormatText(lang, "statistic_text",
		allCount, count, referrals, blocked, subscribers, count-blocked)

	err := msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")

	return NewAdminMenuCommand().Serve(s)
}

func adminFormatText(lang, key string, values ...interface{}) string {
	formatText := assets.AdminText(lang, key)
	return fmt.Sprintf(formatText, values...)
}
