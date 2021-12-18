package administrator

import (
	"strconv"
	"strings"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/db"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	bonusAmountName    = "bonus_amount"
	minWithdrawalName  = "min_withdrawal_amount"
	watchAmountName    = "watch_amount"
	breakAmountName    = "break_amount"
	watchPdAmountName  = "watch_pd_amount"
	watchPdTAmountName = "watch_pd_t_amount"
	watchPdYAmountName = "watch_pd_y_amount"
	watchPdAAmountName = "watch_pd_a_amount"
	referralAmountName = "referral_amount"
)

type MakeMoneySettingCommand struct {
}

func NewMakeMoneySettingCommand() *MakeMoneySettingCommand {
	return &MakeMoneySettingCommand{}
}

func (c *MakeMoneySettingCommand) Serve(s model.Situation) error {
	if strings.Contains(s.Params.Level, "change_parameter?") {
		if err := setAdminBackButton(s.BotLang, s.User.ID, "operation_canceled"); err != nil {
			return err
		}

		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	}

	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "make_money_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("rewards_setting_button", "admin/rewards_setting")),
		msgs.NewIlRow(msgs.NewIlAdminButton("link_setting_button", "admin/link_setting")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(lang)

	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		err := msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), &markUp, text)
		if err != nil {
			return err
		}
		return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	}
	msgID, err := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	return nil
}

type RewardsSettingCommand struct {
}

func NewRewardsSettingCommand() *RewardsSettingCommand {
	return &RewardsSettingCommand{}
}

func (c *RewardsSettingCommand) Serve(s model.Situation) error {
	markUp, text := getRewardsMenu(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")

	sendMsgAdnAnswerCallback(s, markUp, text)
	return nil
}

func getRewardsMenu(userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "rewards_setting_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_bonus_amount_button", "admin/change_parameter?bonus_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_min_withdrawal_amount_button", "admin/change_parameter?min_withdrawal_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_watch_amount_button", "admin/change_parameter?watch_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_break_watch_button", "admin/change_parameter?break_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_watch_pd_amount_button", "admin/change_parameter?watch_pd_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_referral_amount_button", "admin/change_parameter?referral_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_make_money_setting", "admin/make_money_setting")),
	).Build(lang)

	return &markUp, text
}

func sendMsgAdnAnswerCallback(s model.Situation, markUp *tgbotapi.InlineKeyboardMarkup, text string) error {
	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		return msgs.NewEditMarkUpMessage(s.BotLang, s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), markUp, text)
	}
	msgID, _ := msgs.NewIDParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)

	if s.CallbackQuery != nil && s.CallbackQuery.ID != "" {
		return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
	}

	return nil
}

type ChangeParameterCommand struct {
}

func NewChangeParameterCommand() *ChangeParameterCommand {
	return &ChangeParameterCommand{}
}

func (c *ChangeParameterCommand) Serve(s model.Situation) error {
	changeParameter := strings.Split(s.CallbackQuery.Data, "?")[1]
	if changeParameter == watchPdAmountName {
		markUp, text := getChangeWatchPdAmountMenu(s.User.ID)
		db.RdbSetUser(s.BotLang, s.User.ID, "admin")

		return sendMsgAdnAnswerCallback(s, markUp, text)
	}

	lang := assets.AdminLang(s.User.ID)
	var parameter string
	var value int

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/change_parameter?"+changeParameter)
	switch changeParameter {
	case bonusAmountName:
		parameter = assets.AdminText(lang, "change_bonus_amount_button")
		value = assets.AdminSettings.Parameters[s.BotLang].BonusAmount
	case minWithdrawalName:
		parameter = assets.AdminText(lang, "change_min_withdrawal_amount_button")
		value = assets.AdminSettings.Parameters[s.BotLang].MinWithdrawalAmount
	case watchAmountName:
		parameter = assets.AdminText(lang, "change_watch_amount_button")
		value = assets.AdminSettings.Parameters[s.BotLang].WatchReward
	case breakAmountName:
		parameter = assets.AdminText(lang, "change_break_watch_button")
		value = int(assets.AdminSettings.Parameters[s.BotLang].SecondBetweenViews)
	case watchPdTAmountName:
		parameter = assets.AdminText(lang, "change_"+watchPdTAmountName+"_button")
		value = assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDayT
	case watchPdYAmountName:
		parameter = assets.AdminText(lang, "change_"+watchPdYAmountName+"_button")
		value = assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDayY
	case watchPdAAmountName:
		parameter = assets.AdminText(lang, "change_"+watchPdAAmountName+"_button")
		value = assets.AdminSettings.Parameters[s.BotLang].MaxOfVideoPerDayA
	case referralAmountName:
		parameter = assets.AdminText(lang, "change_referral_amount_button")
		value = assets.AdminSettings.Parameters[s.BotLang].ReferralAmount
	}

	err := msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	if err != nil {
		return err
	}
	text := adminFormatText(lang, "set_new_amount_text", parameter, value)
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_reward_setting_setting")),
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
}

func getChangeWatchPdAmountMenu(userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "rewards_setting_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_watch_pd_t_amount_button", "admin/change_parameter?"+watchPdTAmountName)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_watch_pd_y_amount_button", "admin/change_parameter?"+watchPdYAmountName)),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_watch_pd_a_amount_button", "admin/change_parameter?"+watchPdAAmountName)),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_make_money_setting", "admin/make_money_setting")),
	).Build(lang)

	return &markUp, text
}

type LinkSettingCommand struct {
}

func NewLinkSettingCommand() *LinkSettingCommand {
	return &LinkSettingCommand{}
}

func (c *LinkSettingCommand) Serve(s model.Situation) error {
	if strings.Contains(s.Params.Level, "/") {
		setAdminBackButton(s.BotLang, s.User.ID, "operation_canceled")
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	}

	markUp, text := getLinksMenu(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")

	return sendMsgAdnAnswerCallback(s, markUp, text)
}

func getLinksMenu(userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "links_menu_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_link_youtube", "admin/change_link?youtube")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_link_tiktok", "admin/change_link?tiktok")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_link_advertisement", "admin/change_link?advertisement")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_make_money_setting", "admin/make_money_setting")),
	).Build(lang)

	return &markUp, text
}

type ChangeLinkMenuCommand struct {
}

func NewChangeLinkMenuCommand() *ChangeLinkMenuCommand {
	return &ChangeLinkMenuCommand{}
}

func (c *ChangeLinkMenuCommand) Serve(s model.Situation) error {
	var text, buttonType string
	partition := strings.Split(s.CallbackQuery.Data, "?")[1]
	lang := assets.AdminLang(s.User.ID)
	switch partition {
	case "youtube":
		text = createLinkListText(s, partition)
		buttonType = "link"
	case "tiktok", "advertisement":
		text = createVideoListText(s, partition)
		buttonType = "video"
	}

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("add_"+buttonType+"_button", "admin/add_link?"+partition)),
		msgs.NewIlRow(msgs.NewIlAdminButton("add_limit_to_"+buttonType+"_button", "admin/add_limit_to_link?"+partition)),
		msgs.NewIlRow(msgs.NewIlAdminButton("delete_"+buttonType+"_button", "admin/delete_link?"+partition)),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_link_list_menu", "admin/link_setting")),
	).Build(lang)

	return sendMsgAdnAnswerCallback(s, &markUp, text)
}

func createLinkListText(s model.Situation, partition string) string {
	var text string
	lang := assets.AdminLang(s.User.ID)
	text = assets.AdminText(lang, "link_list_head") + partition + "⬇️" + "\n\n"

	if len(assets.Tasks[s.BotLang].Partition[partition]) == 0 {
		return text + assets.AdminText(lang, "link_list_empty")
	}

	for i, elem := range assets.Tasks[s.BotLang].Partition[partition] {
		text += strconv.Itoa(i+1) + ") " + elem.Url + "\n"
		if elem.Limited {
			text += assets.AdminText(lang, "impressions_left_text") +
				strconv.Itoa(elem.ImpressionsLeft) + "\n"
		}
		text += "\n"
	}

	return text
}

func createVideoListText(s model.Situation, partition string) string {
	var text string
	lang := assets.AdminLang(s.User.ID)
	text = assets.AdminText(lang, "link_list_head") + partition + "⬇️" + "\n\n"

	if len(assets.Tasks[s.BotLang].Partition[partition]) == 0 {
		return text + assets.AdminText(lang, "link_list_empty")
	}

	for i, elem := range assets.Tasks[s.BotLang].Partition[partition] {
		text += strconv.Itoa(i+1) + ") " + elem.FileID + "\n"
		if elem.Limited {
			text += assets.AdminText(lang, "impressions_left_text") +
				strconv.Itoa(elem.ImpressionsLeft) + "\n"
		}
		text += "\n"
	}

	return text
}

type AddLinkCommand struct {
}

func NewAddLinkCommand() *AddLinkCommand {
	return &AddLinkCommand{}
}

func (c *AddLinkCommand) Serve(s model.Situation) error {
	callBackText := "type_the_text"
	key := "invitation_to_send_new_link"

	partition := strings.Split(s.CallbackQuery.Data, "?")[1]
	if partition != "youtube" {
		callBackText = "send_the_video"
		key = "invitation_to_send_new_video"
	}

	lang := assets.AdminLang(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)

	err := msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, callBackText)
	if err != nil {
		return err
	}
	text := assets.AdminText(lang, key)
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_link_list_menu")),
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
}

type AddLimitToLinkCommand struct {
}

func NewAddLimitToLinkCommand() *AddLimitToLinkCommand {
	return &AddLimitToLinkCommand{}
}

func (c *AddLimitToLinkCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)

	err := msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	if err != nil {
		return err
	}
	text := assets.AdminText(lang, "invitation_to_send_number_link")
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_link_list_menu")),
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
}

type DeleteLinkCommand struct {
}

func NewDeleteLinkCommand() *DeleteLinkCommand {
	return &DeleteLinkCommand{}
}

func (c *DeleteLinkCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)

	err := msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	if err != nil {
		return err
	}
	text := assets.AdminText(lang, "invitation_to_send_delete_link")
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_link_list_menu")),
		msgs.NewRow(msgs.NewAdminButton("admin_log_out_text")),
	).Build(lang)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, markUp, text)
}
