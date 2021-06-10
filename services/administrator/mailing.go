package administrator

import (
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/db"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

type StartMailingCommand struct {
}

func NewStartMailingCommand() *StartMailingCommand {
	return &StartMailingCommand{}
}

func (c *StartMailingCommand) Serve(s bots.Situation) {
	if !selectedLangAreNotEmpty() {
		msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "no_language_selected")
		return
	}

	db.StartMailing(s.BotLang)
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "mailing_successful")
	resendAdvertisementMenuLevel(s.BotLang, s.CallbackQuery.From.ID)
}

type SelectedLangCommand struct {
}

func NewSelectedLangCommand() *SelectedLangCommand {
	return &SelectedLangCommand{}
}

func (c *SelectedLangCommand) Serve(s bots.Situation) {
	data := strings.Split(s.CallbackQuery.Data, "?")
	partition := data[1]
	lang := data[2]
	switch partition {
	case "switch_lang":
		switchLangOnKeyboard(lang)
		msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
		sendMailingMenu(s.BotLang, s.UserID)
	case "switch_all":
		switchedSelectedLanguages()
		msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
		sendMailingMenu(s.BotLang, s.UserID)
	}
}

func sendMailingMenu(botLang string, userID int) {
	lang := assets.AdminLang(userID)

	text := assets.AdminText(lang, "change_text_of_advertisement_text")
	markUp := createMailingMarkUp(lang)

	if db.RdbGetAdminMsgID(botLang, userID) == 0 {
		msgID := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), &markUp, text)
		db.RdbSetAdminMsgID(botLang, userID, msgID)
		return
	}
	msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &markUp, text)
}

func createMailingMarkUp(lang string) tgbotapi.InlineKeyboardMarkup {
	markUp := parseMainLanguageButton()

	text := "select_all_language"
	data := "admin/send_advertisement?switch_all?select_all"
	if selectedAllLanguage() {
		text = "deselect_all_selections"
		data = strings.Replace(data, "select_all", "deselect_all", 1)
	}

	markUp.Rows = append(markUp.Rows,
		msgs2.NewIlRow(msgs2.NewIlAdminButton(text, data)),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("start_mailing_button", "admin/start_mailing")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement")),
	)
	return markUp.Build(lang)
}

func parseMainLanguageButton() *msgs2.InlineMarkUp {
	markUp := msgs2.NewIlMarkUp()

	for _, lang := range assets.AvailableLang {
		button := "button_"
		if assets.AdminSettings.LangSelectedMap[lang] {
			button += "on_" + lang
		} else {
			button += "off_" + lang
		}
		markUp.Rows = append(markUp.Rows,
			msgs2.NewIlRow(msgs2.NewIlAdminButton(button, "admin/send_advertisement?switch_lang?"+lang)),
		)
	}
	return &markUp
}

func switchLangOnKeyboard(lang string) {
	assets.AdminSettings.LangSelectedMap[lang] = !assets.AdminSettings.LangSelectedMap[lang]
	assets.SaveAdminSettings()
}

func resendAdvertisementMenuLevel(botLang string, userID int) {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	inlineMarkUp, text := getAdvertisementMenu(botLang, userID)
	msgID := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), inlineMarkUp, text)
	db.RdbSetAdminMsgID(botLang, userID, msgID)
}

func switchedSelectedLanguages() {
	if selectedAllLanguage() {
		resetSelectedLang()
		return
	}
	chooseAllLanguages()
}

func resetSelectedLang() {
	for lang := range assets.AdminSettings.LangSelectedMap {
		assets.AdminSettings.LangSelectedMap[lang] = false
	}
	assets.SaveAdminSettings()
}

func chooseAllLanguages() {
	for lang := range assets.AdminSettings.LangSelectedMap {
		assets.AdminSettings.LangSelectedMap[lang] = true
	}
	assets.SaveAdminSettings()
}

func selectedAllLanguage() bool {
	for _, lang := range assets.AvailableLang {
		if !assets.AdminSettings.LangSelectedMap[lang] {
			return false
		}
	}
	return true
}

func selectedLangAreNotEmpty() bool {
	for _, lang := range assets.AvailableLang {
		if assets.AdminSettings.LangSelectedMap[lang] {
			return true
		}
	}
	return false
}