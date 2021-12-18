package db

import (
	"log"
	"strconv"

	"github.com/Stepan1328/youtube-assist-bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	emptyLevelName = "empty"
	nilRedisErr    = "redis: nil"
)

func RdbSetUser(botLang string, ID int64, level string) {
	userID := userIDToRdb(ID)
	_, err := model.Bots[botLang].Rdb.Set(userID, level, 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func userIDToRdb(userID int64) string {
	return "user:" + strconv.FormatInt(userID, 10)
}

func GetLevel(botLang string, id int64) string {
	userID := userIDToRdb(id)
	have, err := model.Bots[botLang].Rdb.Exists(userID).Result()
	if err != nil {
		log.Println(err)
	}
	if have == 0 {
		return emptyLevelName
	}

	value, err := model.Bots[botLang].Rdb.Get(userID).Result()
	if err != nil {
		log.Println(err)
	}
	return value
}

func RdbSetTemporary(botLang string, userID int64, msgID int) {
	temporaryID := temporaryIDToRdb(userID)
	_, err := model.Bots[botLang].Rdb.Set(temporaryID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func temporaryIDToRdb(userID int64) string {
	return "message:" + strconv.FormatInt(userID, 10)
}

func RdbGetTemporary(botLang string, userID int64) string {
	temporaryID := temporaryIDToRdb(userID)
	result, err := model.Bots[botLang].Rdb.Get(temporaryID).Result()
	if err != nil {
		log.Println(err)
	}
	return result
}

func RdbSetAdminMsgID(botLang string, userID int64, msgID int) {
	adminMsgID := adminMsgIDToRdb(userID)
	_, err := model.Bots[botLang].Rdb.Set(adminMsgID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func adminMsgIDToRdb(userID int64) string {
	return "admin_msg_id:" + strconv.FormatInt(userID, 10)
}

func RdbGetAdminMsgID(botLang string, userID int64) int {
	adminMsgID := adminMsgIDToRdb(userID)
	result, err := model.Bots[botLang].Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}
	msgID, _ := strconv.Atoi(result)
	return msgID
}

func DeleteOldAdminMsg(botLang string, userID int64) {
	adminMsgID := adminMsgIDToRdb(userID)
	result, err := model.Bots[botLang].Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}

	if oldMsgID, _ := strconv.Atoi(result); oldMsgID != 0 {
		msg := tgbotapi.NewDeleteMessage(int64(userID), oldMsgID)

		if _, err = model.Bots[botLang].Bot.Send(msg); err != nil && err.Error() != "message to delete not found" {
			log.Println(err)
		}
		RdbSetAdminMsgID(botLang, userID, 0)
	}
}

func DeleteTemporaryMessages(botLang string, userID int64) {
	result := RdbGetTemporary(botLang, userID)

	if result == "" {
		return
	}

	msgID, err := strconv.Atoi(result)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewDeleteMessage(userID, msgID)

	bot := model.GetBot(botLang)
	if _, err = bot.Send(msg); err != nil && err.Error() != "message to delete not found" {
		log.Println(err)
	}
}

func RdbSetMakeMoneyLevel(s model.Situation, breakTime string) {
	makeMoneyID := makeMoneyLevelKey(s.Params.Partition, s.User.ID)
	_, err := model.Bots[s.BotLang].Rdb.Set(makeMoneyID, "/make_money_"+s.Params.Partition+"?"+breakTime, 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func makeMoneyLevelKey(partition string, userID int64) string {
	return partition + "_level:" + strconv.FormatInt(userID, 10)
}

func RdbGetMakeMoneyLevel(s model.Situation) string {
	makeMoneyID := makeMoneyLevelKey(s.Params.Partition, s.User.ID)
	result, err := model.Bots[s.BotLang].Rdb.Get(makeMoneyID).Result()
	if err != nil && err.Error() != nilRedisErr {
		log.Println(err)
	}

	if result == "" {
		result = "/make_money_" + s.Params.Partition + "?"
	}
	return result
}
