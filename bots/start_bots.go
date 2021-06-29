package bots

import (
	"database/sql"
	"encoding/json"
	"github.com/Stepan1328/youtube-assist-bot/cfg"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
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
}

type GlobalHandlers interface {
	GetHandler(command string) Handler
}

type Handler interface {
	Serve(situation Situation)
}

type Situation struct {
	Message       *tgbotapi.Message
	CallbackQuery *tgbotapi.CallbackQuery
	BotLang       string
	UserID        int
	UserLang      string
	Command       string
	Params        Parameters
	Err           error
}

type Parameters struct {
	ReplyText string
	Level     string
	Partition string
	Link      *LinkInfo
}

type LinkInfo struct {
	Url      string
	FileID   string
	Duration int
}

func UploadDataBase(dbLang string) *sql.DB {
	dataBase, err := sql.Open("mysql",
		cfg.DBCfg.User+cfg.DBCfg.Password+"@/"+cfg.DBCfg.Names[dbLang])
	if err != nil {
		panic(err.Error())
	}
	return dataBase
}

func StartRedis(k int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
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
	bytes, _ := os.ReadFile("./cfg/tokens.json")
	err := json.Unmarshal(bytes, &Bots)
	if err != nil {
		panic(err)
	}
}

func GetGlobalBot(botLang string) *GlobalBot {
	return Bots[botLang]
}
