package db

import (
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/model"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	getLangIDQuery = "SELECT id, lang FROM users ORDER BY id LIMIT ? OFFSET ?;"
)

var (
	message           = make(map[string]tgbotapi.MessageConfig, 10)
	photoMessage      = make(map[string]tgbotapi.PhotoConfig, 10)
	videoMessage      = make(map[string]tgbotapi.VideoConfig, 10)
	usersPerIteration = 100
)

func StartMailing(botLang string, initiator *model.User) {
	startTime := time.Now()
	fillMessageMap()

	var (
		sendToUsers  int
		blockedUsers int
	)

	msgs.SendNotificationToDeveloper(
		fmt.Sprintf("%s // mailing started", botLang),
	)

	for offset := 0; ; offset += usersPerIteration {
		countSend, errCount := mailToUserWithPagination(botLang, offset)
		if countSend == -1 {
			sendRespMsgToMailingInitiator(botLang, initiator, "failing_mailing_text", sendToUsers)
			break
		}

		if countSend == 0 && errCount == 0 {
			break
		}

		sendToUsers += countSend
		blockedUsers += errCount
	}

	msgs.SendNotificationToDeveloper(
		fmt.Sprintf("%s // send to %d users mail; latency: %v", botLang, sendToUsers, time.Now().Sub(startTime)),
	)

	sendRespMsgToMailingInitiator(botLang, initiator, "complete_mailing_text", sendToUsers)

	assets.AdminSettings.BlockedUsers[botLang] = blockedUsers
	assets.SaveAdminSettings()
}

func sendRespMsgToMailingInitiator(botLang string, user *model.User, key string, countOfSends int) {
	lang := assets.AdminLang(user.ID)
	text := fmt.Sprintf(assets.AdminText(lang, key), countOfSends)

	_ = msgs.NewParseMessage(botLang, user.ID, text)
}

func mailToUserWithPagination(botLang string, offset int) (int, int) {
	users, err := getUsersWithPagination(botLang, offset)
	if err != nil {
		msgs.SendNotificationToDeveloper(errors.Wrap(err, "get users with pagination").Error())
		return -1, 0
	}

	totalCount := len(users)
	if totalCount == 0 {
		return 0, 0
	}

	responseChan := make(chan bool)
	var sendToUsers int

	fmt.Println(users)

	for _, user := range users {
		go sendMailToUser(botLang, user, responseChan)
	}

	for countOfResp := 0; countOfResp < len(users); countOfResp++ {
		select {
		case resp := <-responseChan:
			if resp {
				sendToUsers++
			}
		}
	}

	return sendToUsers, totalCount - sendToUsers
}

func getUsersWithPagination(botLang string, offset int) ([]*model.User, error) {
	rows, err := model.GetDB(botLang).Query(getLangIDQuery, usersPerIteration, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed execute query")
	}

	var users []*model.User

	for rows.Next() {
		user := &model.User{}

		if err := rows.Scan(&user.ID, &user.Language); err != nil {
			return nil, errors.Wrap(err, "failed scan row")
		}

		if containsInAdmin(user.ID) {
			continue
		}

		users = append(users, user)
	}

	return users, nil
}

func sendMailToUser(botLang string, user *model.User, respChan chan<- bool) {
	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[user.Language].Url)),
	).Build(user.Language)
	button := &markUp

	if !assets.AdminSettings.Parameters[botLang].ButtonUnderAdvert {
		button = nil
	}
	baseChat := tgbotapi.BaseChat{
		ChatID:      user.ID,
		ReplyMarkup: button,
	}

	switch assets.AdminSettings.AdvertisingChoice[botLang] {
	case "photo":
		msg := photoMessage[user.Language]
		msg.BaseChat = baseChat
		respChan <- msgs.SendMsgToChat(botLang, msg)
	case "video":
		msg := videoMessage[user.Language]
		msg.BaseChat = baseChat
		respChan <- msgs.SendMsgToChat(botLang, msg)
	default:
		msg := message[user.Language]
		msg.BaseChat = baseChat
		respChan <- msgs.SendMessageToChat(botLang, msg)
	}
}

func containsInAdmin(userID int64) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

func copyBlockedMap() map[string]int {
	blockedUsers := make(map[string]int, 5)
	for _, lang := range assets.AvailableLang {
		if assets.AdminSettings.LangSelectedMap[lang] {
			blockedUsers[lang] = 0
		}
	}
	return blockedUsers
}

func clearSelectedLang(blockedUsers map[string]int) {
	for _, lang := range assets.AvailableLang {
		if assets.AdminSettings.LangSelectedMap[lang] {
			blockedUsers[lang] = 0
		}
	}
}

//func createAStringOfLang() string {
//	var str string
//
//	for _, lang := range assets.AvailableLang {
//		if assets.AdminSettings.LangSelectedMap[lang] {
//			str += " lang = '" + lang + "' OR"
//		}
//	}
//	return strings.TrimRight(str, " OR")
//}

func fillMessageMap() {
	var markUp tgbotapi.InlineKeyboardMarkup
	for _, lang := range assets.AvailableLang {
		text := assets.AdminSettings.AdvertisingText[lang]

		if assets.AdminSettings.Parameters[lang].ButtonUnderAdvert {
			markUp = tgbotapi.InlineKeyboardMarkup{}
		} else {
			markUp = msgs.NewIlMarkUp(
				msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[lang].Url)),
			).Build(lang)

		}

		switch assets.AdminSettings.AdvertisingChoice[lang] {
		case "photo":
			photoMessage[lang] = tgbotapi.PhotoConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ReplyMarkup: markUp,
					},
					File: tgbotapi.FileID(assets.AdminSettings.AdvertisingPhoto[lang]),
				},
				Caption:   text,
				ParseMode: "HTML",
			}
		case "video":
			videoMessage[lang] = tgbotapi.VideoConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ReplyMarkup: markUp,
					},
					File: tgbotapi.FileID(assets.AdminSettings.AdvertisingVideo[lang]),
				},
				Caption:   text,
				ParseMode: "HTML",
			}
		default:
			message[lang] = tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ReplyMarkup: markUp,
				},
				Text: text,
			}
		}
	}
}
