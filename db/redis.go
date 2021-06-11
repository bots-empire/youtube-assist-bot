package db

import (
	"github.com/Stepan1328/youtube-assist-bot/bots"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
)

func RdbSetUser(botLang string, ID int, level string) {
	userID := userIDToRdb(ID)
	_, err := bots.Bots[botLang].Rdb.Set(userID, level, 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func userIDToRdb(userID int) string {
	return "user:" + strconv.Itoa(userID)
}

func GetLevel(botLang string, id int) string {
	userID := userIDToRdb(id)
	have, err := bots.Bots[botLang].Rdb.Exists(userID).Result()
	if err != nil {
		log.Println(err)
	}
	if have == 0 {
		return "empty"
	}

	value, err := bots.Bots[botLang].Rdb.Get(userID).Result()
	if err != nil {
		log.Println(err)
	}
	return value
}

func RdbSetTemporary(botLang string, userID, msgID int) {
	temporaryID := temporaryIDToRdb(userID)
	_, err := bots.Bots[botLang].Rdb.Set(temporaryID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func temporaryIDToRdb(userID int) string {
	return "message:" + strconv.Itoa(userID)
}

func RdbGetTemporary(botLang string, userID int) string {
	temporaryID := temporaryIDToRdb(userID)
	result, err := bots.Bots[botLang].Rdb.Get(temporaryID).Result()
	if err != nil {
		log.Println(err)
	}
	return result
}

func RdbSetAdminMsgID(botLang string, userID, msgID int) {
	adminMsgID := adminMsgIDToRdb(userID)
	_, err := bots.Bots[botLang].Rdb.Set(adminMsgID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func adminMsgIDToRdb(userID int) string {
	return "admin_msg_id:" + strconv.Itoa(userID)
}

func RdbGetAdminMsgID(botLang string, userID int) int {
	adminMsgID := adminMsgIDToRdb(userID)
	result, err := bots.Bots[botLang].Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}
	msgID, _ := strconv.Atoi(result)
	return msgID
}

func DeleteOldAdminMsg(botLang string, userID int) {
	adminMsgID := adminMsgIDToRdb(userID)
	result, err := bots.Bots[botLang].Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}

	if oldMsgID, _ := strconv.Atoi(result); oldMsgID != 0 {
		msg := tgbotapi.NewDeleteMessage(int64(userID), oldMsgID)

		if _, err = bots.Bots[botLang].Bot.Send(msg); err != nil && err.Error() != "message to delete not found" {
			log.Println(err)
		}
		RdbSetAdminMsgID(botLang, userID, 0)
	}
}

func DeleteTemporaryMessages(botLang string, userID int) {
	result := RdbGetTemporary(botLang, userID)

	if result == "" {
		return
	}

	msgID, err := strconv.Atoi(result)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewDeleteMessage(int64(userID), msgID)

	bot := bots.GetBot(botLang)
	if _, err = bot.Send(msg); err != nil && err.Error() != "message to delete not found" {
		log.Println(err)
	}
}

func RdbSetMakeMoneyLevel(s bots.Situation, breakTime string) {
	makeMoneyID := makeMoneyLevelKey(s.Params.Partition, s.UserID)
	_, err := bots.Bots[s.BotLang].Rdb.Set(makeMoneyID, "/make_money_"+s.Params.Partition+"?"+breakTime, 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func makeMoneyLevelKey(partition string, userID int) string {
	return partition + "_level:" + strconv.Itoa(userID)
}

func RdbGetMakeMoneyLevel(s bots.Situation) string {
	makeMoneyID := makeMoneyLevelKey(s.Params.Partition, s.UserID)
	result, err := bots.Bots[s.BotLang].Rdb.Get(makeMoneyID).Result()
	if err != nil && err.Error() != "redis: nil" {
		log.Println(err)
	}

	if result == "" {
		result = "/make_money_" + s.Params.Partition + "?"
	}
	return result
}
