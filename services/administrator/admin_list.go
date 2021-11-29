package administrator

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/model"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
)

const (
	AvailableSymbolInKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	AdminKeyLength       = 24
	LinkLifeTime         = 180
	GodUserID            = 138814168
)

var availableKeys = make(map[string]string)

type AdminListCommand struct {
}

func NewAdminListCommand() *AdminListCommand {
	return &AdminListCommand{}
}

func (c *AdminListCommand) Serve(s model.Situation) {
	lang := assets.AdminLang(s.UserID)
	text := assets.AdminText(lang, "admin_list_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("add_admin_button", "admin/add_admin_msg")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("delete_admin_button", "admin/delete_admin")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(lang)

	sendMsgAdnAnswerCallback(s, &markUp, text)
}

func CheckNewAdmin(s model.Situation) {
	key := strings.Replace(s.Command, "/start new_admin_", "", 1)
	if availableKeys[key] != "" {
		assets.AdminSettings.AdminID[s.UserID] = &assets.AdminUser{
			Language:  "ru",
			FirstName: s.Message.From.FirstName,
		}
		if s.UserID == GodUserID {
			assets.AdminSettings.AdminID[s.UserID].SpecialPossibility = true
		}
		assets.SaveAdminSettings()

		text := assets.AdminText(s.UserLang, "welcome_to_admin")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		availableKeys[key] = ""
		return
	}

	text := assets.LangText(s.UserLang, "invalid_link_err")
	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
}

type NewAdminToListCommand struct {
}

func NewNewAdminToListCommand() *NewAdminToListCommand {
	return &NewAdminToListCommand{}
}

func (c *NewAdminToListCommand) Serve(s model.Situation) {
	lang := assets.AdminLang(s.UserID)

	link := createNewAdminLink(s.BotLang)
	text := adminFormatText(lang, "new_admin_key_text", link, LinkLifeTime)

	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)
	s.Command = "/send_admin_list"
	CheckAdminCallback(s)

	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
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

func (c *DeleteAdminCommand) Serve(s model.Situation) {
	if !adminHavePrivileges(s) {
		_ = msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "admin_dont_have_permissions")
		return
	}

	lang := assets.AdminLang(s.UserID)
	db.RdbSetUser(s.BotLang, s.UserID, s.CallbackQuery.Data)

	_ = msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	_ = msgs2.NewParseMessage(s.BotLang, int64(s.UserID), createListOfAdminText(lang))

	//markUp := msgs2.NewMarkUp(
	//	msgs2.NewRow(msgs2.NewAdminButton("back_to_link_list_menu")),
	//	msgs2.NewRow(msgs2.NewAdminButton("admin_log_out_text")),
	//).Build(lang)

	//msgs2.NewParseMarkUpMessage(s.BotLang, int64(s.UserID), markUp, createListOfAdminText(lang))
}

func adminHavePrivileges(s model.Situation) bool {
	return assets.AdminSettings.AdminID[s.UserID].SpecialPossibility
}

func createListOfAdminText(lang string) string {
	var listOfAdmins string
	for id, admin := range assets.AdminSettings.AdminID {
		listOfAdmins += strconv.FormatInt(id, 10) + ") " + admin.FirstName + "\n"
	}

	return adminFormatText(lang, "delete_admin_body_text", listOfAdmins)
}
