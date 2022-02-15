package administrator

import (
	"strconv"
	"strings"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
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
	h.OnCommand("/get_new_source", NewGetNewSourceCommand())

	//Change Link List command
	h.OnCommand("/add_link", NewAddLinkToListCommand())
	h.OnCommand("/add_limit_to_link", NewSetLimitToLinkCommand())
	h.OnCommand("/set_limit_to_link", NewUpdateLimitToLinkCommand())
	h.OnCommand("/delete_link", NewDeleteLinkFromListCommand())
}

func (h *AdminMessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

type RemoveAdminCommand struct {
}

func NewRemoveAdminCommand() *RemoveAdminCommand {
	return &RemoveAdminCommand{}
}

func (c *RemoveAdminCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	adminId, err := strconv.ParseInt(s.Message.Text, 10, 64)
	if err != nil {
		text := assets.AdminText(lang, "incorrect_admin_id_text")

		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	if !checkAdminIDInTheList(adminId) {
		text := assets.AdminText(lang, "incorrect_admin_id_text")

		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	delete(assets.AdminSettings.AdminID, adminId)
	assets.SaveAdminSettings()
	err = setAdminBackButton(s.BotLang, s.User.ID, "admin_removed_status")
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/send_admin_list"}
	return NewAdminListCommand().Serve(s)
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

func (c *AdvertisementSettingCommand) Serve(s model.Situation) error {
	s.CallbackQuery = &tgbotapi.CallbackQuery{
		Data: "admin/change_text_url?",
	}

	return NewAdvertisementMenuCommand().Serve(s)
}

func CheckAdminMessage(s model.Situation) error {
	if !containsInAdmin(s.User.ID) {
		return notAdmin(s.BotLang, s.User)
	}

	s.Command, s.Err = assets.GetCommandFromText(s)
	if s.Err == nil {
		Handler := model.Bots[s.BotLang].AdminMessageHandler.
			GetHandler(s.Command)

		if Handler != nil {
			return Handler.Serve(s)
		}
	}

	s.Command = strings.TrimLeft(strings.Split(s.Params.Level, "?")[0], "admin")

	Handler := model.Bots[s.BotLang].AdminMessageHandler.
		GetHandler(s.Command)

	if Handler != nil {
		return Handler.Serve(s)
	}

	return model.ErrCommandNotConverted
}

type SetNewTextUrlCommand struct {
}

func NewSetNewTextUrlCommand() *SetNewTextUrlCommand {
	return &SetNewTextUrlCommand{}
}

func (c *SetNewTextUrlCommand) Serve(s model.Situation) error {
	capitation := strings.Split(s.Params.Level, "?")[1]
	lang := assets.AdminLang(s.User.ID)
	status := "operation_canceled"

	switch capitation {
	case "change_url":
		advertChan := getUrlAndChatID(s.Message)
		if advertChan.ChannelID == 0 {
			text := assets.AdminText(lang, "chat_id_not_update")
			return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
		}

		assets.AdminSettings.AdvertisingChan[s.BotLang] = advertChan
	case "change_text":
		assets.AdminSettings.AdvertisingText[s.BotLang] = s.Message.Text
	}
	assets.SaveAdminSettings()
	status = "operation_completed"

	err := setAdminBackButton(s.BotLang, s.User.ID, status)
	if err != nil {
		return err
	}
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	s.Params.Level = "admin/change_url"
	return NewAdvertisementMenuCommand().Serve(s)
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

func (c *UpdateParameterCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)

	newAmount, err := strconv.Atoi(s.Message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
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
	err = setAdminBackButton(s.BotLang, s.User.ID, "operation_completed")
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	return NewRewardsSettingCommand().Serve(s)
}

type AddLinkToListCommand struct {
}

func NewAddLinkToListCommand() *AddLinkToListCommand {
	return &AddLinkToListCommand{}
}

func (c *AddLinkToListCommand) Serve(s model.Situation) error {
	if s.Message.Text == assets.AdminText(assets.AdminLang(s.User.ID), "back_to_link_list_menu") {
		return NewLinkSettingCommand().Serve(s)
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	if partition == "youtube" {
		link := s.Message.Text
		assets.Tasks[s.BotLang].Partition[partition] = append(assets.Tasks[s.BotLang].Partition[partition], &assets.Link{
			Url: link,
		})

		return updateTasks(s, partition)
	}

	if s.Message.Video == nil {
		lang := assets.AdminLang(s.User.ID)
		text := assets.AdminText(lang, "incorrect_new_video")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	assets.Tasks[s.BotLang].Partition[partition] = append(assets.Tasks[s.BotLang].Partition[partition], &assets.Link{
		FileID:   s.Message.Video.FileID,
		Duration: s.Message.Video.Duration,
	})

	return updateTasks(s, partition)
}

func updateTasks(s model.Situation, partition string) error {
	assets.SaveTasks(s.BotLang)
	err := setAdminBackButton(s.BotLang, s.User.ID, "operation_completed")
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/change_link?" + partition}
	return NewChangeLinkMenuCommand().Serve(s)
}

type SetLimitToLinkCommand struct {
}

func NewSetLimitToLinkCommand() *SetLimitToLinkCommand {
	return &SetLimitToLinkCommand{}
}

func (c *SetLimitToLinkCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	if s.Message.Text == assets.AdminText(lang, "back_to_link_list_menu") {
		s.Command = "admin/link_setting"
		return CheckAdminCallback(s)
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	linkNumber, err := strconv.Atoi(s.Message.Text)
	if err != nil {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	if len(assets.Tasks[s.BotLang].Partition[partition]) < linkNumber {
		text := assets.AdminText(assets.AdminLang(s.User.ID), "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/set_limit_to_link?"+partition+"?"+strconv.Itoa(linkNumber))
	text := adminFormatText(lang, "invitation_to_send_limit",
		assets.Tasks[s.BotLang].Partition[partition][linkNumber-1].Url)
	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type UpdateLimitToLinkCommand struct {
}

func NewUpdateLimitToLinkCommand() *UpdateLimitToLinkCommand {
	return &UpdateLimitToLinkCommand{}
}

func (c *UpdateLimitToLinkCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	if s.Message.Text == assets.AdminText(lang, "back_to_link_list_menu") {
		return NewLinkSettingCommand().Serve(s)
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	linkNumber, _ := strconv.Atoi(strings.Split(s.Params.Level, "?")[2])
	newImpression, err := strconv.Atoi(s.Message.Text)
	if err != nil || newImpression < 1 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	assets.Tasks[s.BotLang].Partition[partition][linkNumber-1] = &assets.Link{
		Url:    assets.Tasks[s.BotLang].Partition[partition][linkNumber-1].Url,
		FileID: assets.Tasks[s.BotLang].Partition[partition][linkNumber-1].FileID,

		Limited:         true,
		ImpressionsLeft: newImpression,
	}

	assets.SaveTasks(s.BotLang)
	err = setAdminBackButton(s.BotLang, s.User.ID, "operation_completed")
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/change_link?" + partition}
	return NewChangeLinkMenuCommand().Serve(s)
}

type DeleteLinkFromListCommand struct {
}

func NewDeleteLinkFromListCommand() *DeleteLinkFromListCommand {
	return &DeleteLinkFromListCommand{}
}

func (c *DeleteLinkFromListCommand) Serve(s model.Situation) error {
	if s.Message.Text == assets.AdminText(assets.AdminLang(s.User.ID), "back_to_link_list_menu") {
		return NewLinkSettingCommand().Serve(s)
	}

	partition := strings.Split(s.Params.Level, "?")[1]
	linkNumber, err := strconv.Atoi(s.Message.Text)
	if err != nil {
		text := assets.AdminText(assets.AdminLang(s.User.ID), "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	err = deleteLinkFromList(s, partition, linkNumber-1)
	if err != nil {
		text := assets.AdminText(assets.AdminLang(s.User.ID), "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	assets.SaveTasks(s.BotLang)
	err = setAdminBackButton(s.BotLang, s.User.ID, "operation_completed")
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/change_link?" + partition}
	return NewChangeLinkMenuCommand().Serve(s)
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
