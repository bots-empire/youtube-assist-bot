package administrator

import (
	"errors"
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/db"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

type AdminMessagesHandlers struct {
	Handlers map[string]bots.Handler
}

func (h *AdminMessagesHandlers) GetHandler(command string) bots.Handler {
	return h.Handlers[command]
}

func (h *AdminMessagesHandlers) Init() {
	//Change Advertisement parameters command
	h.OnCommand("/advertisement_setting", NewAdvertisementSettingCommand())
	h.OnCommand("/change_text_url", NewSetNewTextUrlCommand())
	h.OnCommand("/change_parameter", NewUpdateParameterCommand())

	//Change Link List command
	h.OnCommand("/add_link", NewAddLinkToListCommand())
	h.OnCommand("/add_limit_to_link", NewSetLimitToLinkCommand())
	h.OnCommand("/set_limit_to_link", NewUpdateLimitToLinkCommand())
	h.OnCommand("/delete_link", NewDeleteLinkFromListCommand())

	log.Println("Admin Message Handlers Initialized")
}

func (h *AdminMessagesHandlers) OnCommand(command string, handler bots.Handler) {
	h.Handlers[command] = handler
}

type AdvertisementSettingCommand struct {
}

func NewAdvertisementSettingCommand() *AdvertisementSettingCommand {
	return &AdvertisementSettingCommand{}
}

func (c *AdvertisementSettingCommand) Serve(s bots.Situation) {
	s.CallbackQuery = &tgbotapi.CallbackQuery{
		Data: "admin/change_text_url?",
	}
	s.Command = "admin/advertisement"
	CheckAdminCallback(s)
}

func CheckAdminMessage(s bots.Situation) bool {
	s.Command, s.Err = assets.GetCommandFromText(s)
	if s.Err == nil {
		Handler := bots.Bots[s.BotLang].AdminMessageHandler.
			GetHandler(s.Command)

		if Handler != nil {
			Handler.Serve(s)
			return true
		}
	}

	s.Command = strings.TrimLeft(strings.Split(s.Params.Level, "?")[0], "admin")

	Handler := bots.Bots[s.BotLang].AdminMessageHandler.
		GetHandler(s.Command)

	if Handler != nil {
		Handler.Serve(s)
		return true
	}

	return false
}

type SetNewTextUrlCommand struct {
}

func NewSetNewTextUrlCommand() *SetNewTextUrlCommand {
	return &SetNewTextUrlCommand{}
}

func (c *SetNewTextUrlCommand) Serve(s bots.Situation) {
	data := strings.Split(s.Params.Level, "?")
	textLang := data[2]
	capitation := data[1]
	status := "operation_canceled"

	switch capitation {
	case "change_url":
		assets.AdminSettings.AdvertisingURL[textLang] = s.Message.Text
		status = "operation_completed"
	case "change_text":
		assets.AdminSettings.AdvertisingText[textLang] = s.Message.Text
		status = "operation_completed"
	}
	assets.SaveAdminSettings()

	setAdminBackButton(s.BotLang, s.UserID, status)
	db.RdbSetUser(s.BotLang, s.UserID, "admin/"+capitation)
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)
	sendChangeWithLangMenu(s.BotLang, s.UserID, capitation)
}

type UpdateParameterCommand struct {
}

func NewUpdateParameterCommand() *UpdateParameterCommand {
	return &UpdateParameterCommand{}
}

func (c *UpdateParameterCommand) Serve(s bots.Situation) {
	lang := assets.AdminLang(s.UserID)

	newAmount, err := strconv.Atoi(s.Message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]

	switch partition {
	case "bonus_amount":
		assets.AdminSettings.BonusAmount = newAmount
	case "min_withdrawal_amount":
		assets.AdminSettings.MinWithdrawalAmount = newAmount
	case "watch_amount":
		assets.AdminSettings.WatchReward = newAmount
	case "break_amount":
		assets.AdminSettings.SecondBetweenViews = int64(newAmount)
	case "watch_pd_amount":
		assets.AdminSettings.MaxOfVideoPerDay = newAmount
	case "referral_amount":
		assets.AdminSettings.ReferralAmount = newAmount
	}
	assets.SaveAdminSettings()
	setAdminBackButton(s.BotLang, s.UserID, "operation_completed")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	s.Command = "admin/rewards_setting"
	CheckAdminCallback(s)
}

type AddLinkToListCommand struct {
}

func NewAddLinkToListCommand() *AddLinkToListCommand {
	return &AddLinkToListCommand{}
}

func (c *AddLinkToListCommand) Serve(s bots.Situation) {
	if s.Message.Text == assets.AdminText(assets.AdminLang(s.UserID), "back_to_link_list_menu") {
		s.Command = "admin/link_setting"
		CheckAdminCallback(s)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	defer updateTasks(s, partition)
	if partition == "youtube" {

		link := s.Message.Text
		assets.Tasks[s.BotLang].Partition[partition] = append(assets.Tasks[s.BotLang].Partition[partition], &assets.Link{
			Url: link,
		})
		return
	}

	if s.Message.Video == nil {
		lang := assets.AdminLang(s.UserID)
		text := assets.AdminText(lang, "incorrect_new_video")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	assets.Tasks[s.BotLang].Partition[partition] = append(assets.Tasks[s.BotLang].Partition[partition], &assets.Link{
		FileID: s.Message.Video.FileID,
	})
}

func updateTasks(s bots.Situation, partition string) {
	assets.SaveTasks(s.BotLang)
	setAdminBackButton(s.BotLang, s.UserID, "operation_completed")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	s.Command = "admin/change_link"
	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/change_link?" + partition}
	CheckAdminCallback(s)
}

type SetLimitToLinkCommand struct {
}

func NewSetLimitToLinkCommand() *SetLimitToLinkCommand {
	return &SetLimitToLinkCommand{}
}

func (c *SetLimitToLinkCommand) Serve(s bots.Situation) {
	lang := assets.AdminLang(s.UserID)
	if s.Message.Text == assets.AdminText(lang, "back_to_link_list_menu") {
		s.Command = "admin/link_setting"
		CheckAdminCallback(s)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	linkNumber, err := strconv.Atoi(s.Message.Text)
	if err != nil {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	if len(assets.Tasks[s.BotLang].Partition[partition]) < linkNumber {
		text := assets.AdminText(assets.AdminLang(s.UserID), "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	db.RdbSetUser(s.BotLang, s.UserID, "admin/set_limit_to_link?"+partition+"?"+strconv.Itoa(linkNumber))
	text := adminFormatText(lang, "invitation_to_send_limit",
		assets.Tasks[s.BotLang].Partition[partition][linkNumber-1].Url)
	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
}

type UpdateLimitToLinkCommand struct {
}

func NewUpdateLimitToLinkCommand() *UpdateLimitToLinkCommand {
	return &UpdateLimitToLinkCommand{}
}

func (c *UpdateLimitToLinkCommand) Serve(s bots.Situation) {
	lang := assets.AdminLang(s.UserID)
	if s.Message.Text == assets.AdminText(lang, "back_to_link_list_menu") {
		s.Command = "admin/link_setting"
		CheckAdminCallback(s)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	linkNumber, _ := strconv.Atoi(strings.Split(s.Params.Level, "?")[2])
	newImpression, err := strconv.Atoi(s.Message.Text)
	if err != nil || newImpression < 1 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	assets.Tasks[s.BotLang].Partition[partition][linkNumber-1] = &assets.Link{
		Url:    assets.Tasks[s.BotLang].Partition[partition][linkNumber-1].Url,
		FileID: assets.Tasks[s.BotLang].Partition[partition][linkNumber-1].FileID,

		Limited:         true,
		ImpressionsLeft: newImpression,
	}

	assets.SaveTasks(s.BotLang)
	setAdminBackButton(s.BotLang, s.UserID, "operation_completed")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	s.Command = "admin/change_link"
	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/change_link?" + partition}
	CheckAdminCallback(s)
}

type DeleteLinkFromListCommand struct {
}

func NewDeleteLinkFromListCommand() *DeleteLinkFromListCommand {
	return &DeleteLinkFromListCommand{}
}

func (c *DeleteLinkFromListCommand) Serve(s bots.Situation) {
	if s.Message.Text == assets.AdminText(assets.AdminLang(s.UserID), "back_to_link_list_menu") {
		s.Command = "admin/link_setting"
		CheckAdminCallback(s)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	linkNumber, err := strconv.Atoi(s.Message.Text)
	if err != nil {
		text := assets.AdminText(assets.AdminLang(s.UserID), "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	err = deleteLinkFromList(s, partition, linkNumber-1)
	if err != nil {
		text := assets.AdminText(assets.AdminLang(s.UserID), "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	assets.SaveTasks(s.BotLang)
	setAdminBackButton(s.BotLang, s.UserID, "operation_completed")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	s.Command = "admin/change_link"
	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/change_link?" + partition}
	CheckAdminCallback(s)
}

func deleteLinkFromList(s bots.Situation, partition string, linkNumber int) error {
	listLength := len(assets.Tasks[s.BotLang].Partition[partition])
	if linkNumber > listLength || linkNumber < 0 || listLength == 0 {
		return errors.New("index out of range")
	}

	for i := linkNumber; i < listLength-1; i++ {
		assets.Tasks[s.BotLang].Partition[partition][i] = assets.Tasks[s.BotLang].Partition[partition][i+1]
	}

	assets.Tasks[s.BotLang].Partition[partition] = assets.Tasks[s.BotLang].Partition[partition][:listLength-1]
	return nil
}
