package model

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/Stepan1328/youtube-assist-bot/cfg"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	tokensPath       = "./cfg/tokens.json"
	dbDriver         = "mysql"
	redisDefaultAddr = "127.0.0.1:6379"
)

var Bots = make(map[string]*GlobalBot)

type GlobalBot struct {
	Bot      *tgbotapi.BotAPI
	Chanel   tgbotapi.UpdatesChannel
	Rdb      *redis.Client
	DataBase *sql.DB

	MessageHandler  GlobalHandlers
	CallbackHandler GlobalHandlers

	AdminMessageHandler  GlobalHandlers
	AdminCallBackHandler GlobalHandlers

	BotToken      string
	BotLink       string
	LanguageInBot string

	MaintenanceMode bool
}

type GlobalHandlers interface {
	GetHandler(command string) Handler
}

type Handler interface {
	Serve(situation Situation) error
}

func UploadDataBase(dbLang string) *sql.DB {
	dataBase, err := sql.Open(dbDriver, cfg.DBCfg.User+cfg.DBCfg.Password+"@/"+cfg.DBCfg.Names[dbLang])
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	err = dataBase.Ping()
	if err != nil {
		log.Fatalf("Failed upload database: %s\n", err.Error())
	}

	return dataBase
}

func StartRedis(k int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisDefaultAddr,
		Password: "", // no password set
		DB:       k,  // use default DB
	})
	return rdb
}

func GetBot(botLang string) *tgbotapi.BotAPI {
	return Bots[botLang].Bot
}

func GetDB(botLang string) *sql.DB {
	return Bots[botLang].DataBase
}

func FillBotsConfig() {
	bytes, _ := os.ReadFile(tokensPath)
	err := json.Unmarshal(bytes, &Bots)
	if err != nil {
		panic(err)
	}
}

func GetGlobalBot(botLang string) *GlobalBot {
	return Bots[botLang]
}
