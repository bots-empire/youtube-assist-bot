package administrator

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/pkg/errors"
)

const (
	AvailableSymbolInKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	AdminKeyLength       = 24
	LinkLifeTime         = 180
	GodUserID            = 872383555
)

var availableKeys = make(map[string]string)

type AdminListCommand struct {
}

func NewAdminListCommand() *AdminListCommand {
	return &AdminListCommand{}
}

func (c *AdminListCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "admin_list_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("add_admin_button", "admin/add_admin_msg")),
		msgs.NewIlRow(msgs.NewIlAdminButton("delete_admin_button", "admin/delete_admin")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(lang)

	return sendMsgAdnAnswerCallback(s, &markUp, text)
}

func CheckNewAdmin(s model.Situation) error {
	key := strings.Replace(s.Command, "/start new_admin_", "", 1)
	if availableKeys[key] != "" {
		assets.AdminSettings.AdminID[s.User.ID] = &assets.AdminUser{
			Language:  "ru",
			FirstName: s.Message.From.FirstName,
		}
		if s.User.ID == GodUserID {
			assets.AdminSettings.AdminID[s.User.ID].SpecialPossibility = true
		}
		assets.SaveAdminSettings()

		text := assets.AdminText("ru", "welcome_to_admin")
		availableKeys[key] = ""
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	text := assets.LangText(s.User.Language, "invalid_link_err")
	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type NewAdminToListCommand struct {
}

func NewNewAdminToListCommand() *NewAdminToListCommand {
	return &NewAdminToListCommand{}
}

func (c *NewAdminToListCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)

	link := createNewAdminLink(s.BotLang)
	text := adminFormatText(lang, "new_admin_key_text", link, LinkLifeTime)

	if err := msgs.NewParseMessage(s.BotLang, s.User.ID, text); err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	return NewAdminListCommand().Serve(s)
}

func createNewAdminLink(botLang string) string {
	key := generateKey()
	availableKeys[key] = key
	go deleteKey(key)
	return model.GetGlobalBot(botLang).BotLink + "?start=new_admin_" + key
}

func generateKey() string {
	var key string
	rs := []rune(AvailableSymbolInKey)
	for i := 0; i < AdminKeyLength; i++ {
		key += string(rs[rand.Intn(len(AvailableSymbolInKey))])
	}
	return key
}

func deleteKey(key string) {
	time.Sleep(time.Second * LinkLifeTime)
	availableKeys[key] = ""
}

type DeleteAdminCommand struct {
}

func NewDeleteAdminCommand() *DeleteAdminCommand {
	return &DeleteAdminCommand{}
}

func (c *DeleteAdminCommand) Serve(s model.Situation) error {
	if !adminHavePrivileges(s) {
		_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "admin_dont_have_permissions")
		return nil
	}

	lang := assets.AdminLang(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)

	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")

	//markUp := msgs.NewMarkUp(
	//	msgs.NewRow(msgs.NewAdminButton("back_to_link_list_menu")),
	//	msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	//).Build(lang)

	//msgs.NewParseMarkUpMessage(s.BotLang, int64(s.User.ID), markUp, createListOfAdminText(lang))

	return msgs.NewParseMessage(s.BotLang, s.User.ID, createListOfAdminText(lang))
}

func adminHavePrivileges(s model.Situation) bool {
	return assets.AdminSettings.AdminID[s.User.ID].SpecialPossibility
}

func createListOfAdminText(lang string) string {
	var listOfAdmins string
	for id, admin := range assets.AdminSettings.AdminID {
		listOfAdmins += strconv.FormatInt(id, 10) + ") " + admin.FirstName + "\n"
	}

	return adminFormatText(lang, "delete_admin_body_text", listOfAdmins)
}

type AdvertSourceMenuCommand struct {
}

func NewAdvertSourceMenuCommand() *AdvertSourceMenuCommand {
	return &AdvertSourceMenuCommand{}
}

func (c *AdvertSourceMenuCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "add_new_source_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("add_new_source_button", "admin/add_new_source")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(lang)

	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	return msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), &markUp, text)
}

type AddNewSourceCommand struct {
}

func NewAddNewSourceCommand() *AddNewSourceCommand {
	return &AddNewSourceCommand{}
}

func (c *AddNewSourceCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "input_new_source_text")
	db.RdbSetUser(s.BotLang, s.User.ID, "admin/get_new_source")

	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_admin_settings")),
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
}

type GetNewSourceCommand struct {
}

func NewGetNewSourceCommand() *GetNewSourceCommand {
	return &GetNewSourceCommand{}
}

func (c *GetNewSourceCommand) Serve(s model.Situation) error {
	link, err := model.EncodeLink(s.BotLang, &model.ReferralLinkInfo{
		Source: s.Message.Text,
	})
	if err != nil {
		return errors.Wrap(err, "encode link")
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin")

	if err := msgs.NewParseMessage(s.BotLang, s.User.ID, link); err != nil {
		return errors.Wrap(err, "send message with link")
	}

	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	return NewAdminMenuCommand().Serve(s)
}
