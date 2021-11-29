package administrator

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/model"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminMessagesHandlers struct {
	Handlers map[string]model.Handler
}

func (h *AdminMessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *AdminMessagesHandlers) Init() {
	//Delete Admin command
	h.OnCommand("/delete_admin", NewRemoveAdminCommand())

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

func (h *AdminMessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

type RemoveAdminCommand struct {
}

func NewRemoveAdminCommand() *RemoveAdminCommand {
	return &RemoveAdminCommand{}
}

func (c *RemoveAdminCommand) Serve(s model.Situation) {
	lang := assets.AdminLang(s.UserID)
	adminId, err := strconv.ParseInt(s.Message.Text, 10, 64)
	if err != nil {
		text := assets.AdminText(lang, "incorrect_admin_id_text")
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	if !checkAdminIDInTheList(adminId) {
		text := assets.AdminText(lang, "incorrect_admin_id_text")
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	delete(assets.AdminSettings.AdminID, adminId)
	assets.SaveAdminSettings()
	setAdminBackButton(s.BotLang, s.UserID, "admin_removed_status")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	s.Command = "admin/send_admin_list"
	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/send_admin_list"}
	CheckAdminCallback(s)
}

func checkAdminIDInTheList(adminID int64) bool {
	_, inMap := assets.AdminSettings.AdminID[adminID]
	return inMap
}

type AdvertisementSettingCommand struct {
}

func NewAdvertisementSettingCommand() *AdvertisementSettingCommand {
	return &AdvertisementSettingCommand{}
}

func (c *AdvertisementSettingCommand) Serve(s model.Situation) {
	s.CallbackQuery = &tgbotapi.CallbackQuery{
		Data: "admin/change_text_url?",
	}
	s.Command = "admin/advertisement"
	CheckAdminCallback(s)
}

func CheckAdminMessage(s model.Situation) bool {
	if !containsInAdmin(s.UserID) {
		notAdmin(s.BotLang, s.UserID)
		return true
	}

	s.Command, s.Err = assets.GetCommandFromText(s)
	if s.Err == nil {
		Handler := model.Bots[s.BotLang].AdminMessageHandler.
			GetHandler(s.Command)

		if Handler != nil {
			Handler.Serve(s)
			return true
		}
	}

	s.Command = strings.TrimLeft(strings.Split(s.Params.Level, "?")[0], "admin")

	Handler := model.Bots[s.BotLang].AdminMessageHandler.
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

func (c *SetNewTextUrlCommand) Serve(s model.Situation) {
	capitation := strings.Split(s.Params.Level, "?")[1]
	lang := assets.AdminLang(s.UserID)
	status := "operation_canceled"

	switch capitation {
	case "change_url":
		advertChan := getUrlAndChatID(s.Message)
		if advertChan.ChannelID == 0 {
			text := assets.AdminText(lang, "chat_id_not_update")
			msgs2.NewParseMessage(s.BotLang, s.UserID, text)
			return
		}

		assets.AdminSettings.AdvertisingChan[s.BotLang] = advertChan
	case "change_text":
		assets.AdminSettings.AdvertisingText[s.BotLang] = s.Message.Text
	}
	assets.SaveAdminSettings()
	status = "operation_completed"

	setAdminBackButton(s.BotLang, s.UserID, status)
	db.RdbSetUser(s.BotLang, s.UserID, "admin")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)
	s.Command = "admin/advertisement"
	s.Params.Level = "admin/change_url"
	CheckAdminCallback(s)
}

func getUrlAndChatID(message *tgbotapi.Message) *assets.AdvertChannel {
	data := strings.Split(message.Text, "\n")
	if len(data) != 2 {
		return &assets.AdvertChannel{}
	}

	chatId, err := strconv.Atoi(data[1])
	if err != nil {
		return &assets.AdvertChannel{}
	}

	return &assets.AdvertChannel{
		Url:       data[0],
		ChannelID: int64(chatId),
	}
}

type UpdateParameterCommand struct {
}

func NewUpdateParameterCommand() *UpdateParameterCommand {
	return &UpdateParameterCommand{}
}

func (c *UpdateParameterCommand) Serve(s model.Situation) {
	lang := assets.AdminLang(s.UserID)

	newAmount, err := strconv.Atoi(s.Message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]

	switch partition {
	case bonusAmountName:
		assets.AdminSettings.Parameters[s.BotLang].BonusAmount = newAmount
	case minWithdrawalName:
		assets.AdminSettings.Parameters[s.BotLang].MinWithdrawalAmount = newAmount
	case watchAmountName:
		assets.AdminSettings.Parameters[s.BotLang].WatchReward = newAmount
	case breakAmountName:
		assets.AdminSettings.Parameters[s.BotLang].SecondBetweenViews = int64(newAmount)
	case watchPdTAmountName:
		assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDayT = newAmount
	case watchPdYAmountName:
		assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDayY = newAmount
	case watchPdAAmountName:
		assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDayA = newAmount
	case referralAmountName:
		assets.AdminSettings.Parameters[s.BotLang].ReferralAmount = newAmount
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

func (c *AddLinkToListCommand) Serve(s model.Situation) {
	if s.Message.Text == assets.AdminText(assets.AdminLang(s.UserID), "back_to_link_list_menu") {
		s.Command = "admin/link_setting"
		CheckAdminCallback(s)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	if partition == "youtube" {
		link := s.Message.Text
		assets.Tasks[s.BotLang].Partition[partition] = append(assets.Tasks[s.BotLang].Partition[partition], &assets.Link{
			Url: link,
		})
		updateTasks(s, partition)
		return
	}

	if s.Message.Video == nil {
		bytes, err := json.MarshalIndent(s.Message.Video, "", "  ")
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(bytes))

		lang := assets.AdminLang(s.UserID)
		text := assets.AdminText(lang, "incorrect_new_video")
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	assets.Tasks[s.BotLang].Partition[partition] = append(assets.Tasks[s.BotLang].Partition[partition], &assets.Link{
		FileID:   s.Message.Video.FileID,
		Duration: s.Message.Video.Duration,
	})
	updateTasks(s, partition)
}

func updateTasks(s model.Situation, partition string) {
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

func (c *SetLimitToLinkCommand) Serve(s model.Situation) {
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
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	if len(assets.Tasks[s.BotLang].Partition[partition]) < linkNumber {
		text := assets.AdminText(assets.AdminLang(s.UserID), "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	db.RdbSetUser(s.BotLang, s.UserID, "admin/set_limit_to_link?"+partition+"?"+strconv.Itoa(linkNumber))
	text := adminFormatText(lang, "invitation_to_send_limit",
		assets.Tasks[s.BotLang].Partition[partition][linkNumber-1].Url)
	msgs2.NewParseMessage(s.BotLang, s.UserID, text)
}

type UpdateLimitToLinkCommand struct {
}

func NewUpdateLimitToLinkCommand() *UpdateLimitToLinkCommand {
	return &UpdateLimitToLinkCommand{}
}

func (c *UpdateLimitToLinkCommand) Serve(s model.Situation) {
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
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
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

func (c *DeleteLinkFromListCommand) Serve(s model.Situation) {
	if s.Message.Text == assets.AdminText(assets.AdminLang(s.UserID), "back_to_link_list_menu") {
		s.Command = "admin/link_setting"
		CheckAdminCallback(s)
		return
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	linkNumber, err := strconv.Atoi(s.Message.Text)
	if err != nil {
		text := assets.AdminText(assets.AdminLang(s.UserID), "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	err = deleteLinkFromList(s, partition, linkNumber-1)
	if err != nil {
		text := assets.AdminText(assets.AdminLang(s.UserID), "incorrect_make_money_change_input")
		msgs2.NewParseMessage(s.BotLang, s.UserID, text)
		return
	}

	assets.SaveTasks(s.BotLang)
	setAdminBackButton(s.BotLang, s.UserID, "operation_completed")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	s.Command = "admin/change_link"
	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/change_link?" + partition}
	CheckAdminCallback(s)
}

func deleteLinkFromList(s model.Situation, partition string, linkNumber int) error {
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
