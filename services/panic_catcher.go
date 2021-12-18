package services

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/Stepan1328/youtube-assist-bot/log"
	"github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/Stepan1328/youtube-assist-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	panicLogger = log.NewDefaultLogger().Prefix("panic cather")
)

func panicCather(botLang string, update *tgbotapi.Update) {
	msg := recover()
	if msg == nil {
		return
	}

	panicText := fmt.Sprintf("%s\npanic in backend: message = %s\n%s",
		botLang,
		msg,
		string(debug.Stack()),
	)
	panicLogger.Warn(panicText)

	alertMsg := tgbotapi.NewMessage(notificationChatID, panicText)
	_ = msgs.SendMsgToUser(administrator.DefaultNotificationBot, alertMsg)

	data, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return
	}

	updateDataMsg := tgbotapi.NewMessage(notificationChatID, string(data))
	_ = msgs.SendMsgToUser(administrator.DefaultNotificationBot, updateDataMsg)
}
