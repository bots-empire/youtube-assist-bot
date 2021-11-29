package main

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/model"
	msgs2 "github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/Stepan1328/youtube-assist-bot/services"
	"github.com/Stepan1328/youtube-assist-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	rand.Seed(time.Now().Unix())

	startServices()
	startAllBot()
	assets.UploadUpdateStatistic()

	startHandlers()
}

func startAllBot() {
	k := 0
	for lang, globalBot := range model.Bots {
		StartBot(globalBot, lang, k)
		model.Bots[lang].MessageHandler = NewMessagesHandler()
		model.Bots[lang].CallbackHandler = NewCallbackHandler()
		model.Bots[lang].AdminMessageHandler = NewAdminMessagesHandler()
		model.Bots[lang].AdminCallBackHandler = NewAdminCallbackHandler()
		k++
	}

	log.Println("All bots is running")
}

func startServices() {
	model.FillBotsConfig()
	assets.ParseLangMap()
	assets.ParseTasks()
	assets.ParseAdminMap()
	assets.UploadAdminSettings()
	assets.ParseCommandsList()

	log.Println("All services are running successfully")
}

func StartBot(b *model.GlobalBot, lang string, k int) {
	var err error
	b.Bot, err = tgbotapi.NewBotAPI(b.BotToken)
	if err != nil {
		log.Println(err)
		return
	}

	u := tgbotapi.NewUpdate(0)

	b.Chanel = b.Bot.GetUpdatesChan(u)

	b.Rdb = model.StartRedis(k)
	b.DataBase = model.UploadDataBase(lang)
}

func startHandlers() {
	wg := new(sync.WaitGroup)

	for botLang, handler := range model.Bots {
		wg.Add(1)
		go func(botLang string, handler *model.GlobalBot, wg *sync.WaitGroup) {
			defer wg.Done()
			services.ActionsWithUpdates(botLang, handler.Chanel)
		}(botLang, handler, wg)
	}

	log.Println("All handlers are running")
	msgs2.NewParseMessage("it", 1418862576, "All bots are restart")
	wg.Wait()
}

func NewMessagesHandler() *services.MessagesHandlers {
	handle := services.MessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewCallbackHandler() *services.CallBackHandlers {
	handle := services.CallBackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminMessagesHandler() *administrator.AdminMessagesHandlers {
	handle := administrator.AdminMessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminCallbackHandler() *administrator.AdminCallbackHandlers {
	handle := administrator.AdminCallbackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}
