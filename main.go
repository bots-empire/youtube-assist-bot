package main

import (
	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/services"
	"github.com/Stepan1328/youtube-assist-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())

	startServices()
	startAllBot()

	startHandlers()
}

func startAllBot() {
	k := 0
	for lang, globalBot := range bots.Bots {
		StartBot(globalBot, lang, k)
		bots.Bots[lang].MessageHandler = NewMessagesHandler()
		bots.Bots[lang].CallbackHandler = NewCallbackHandler()
		bots.Bots[lang].AdminMessageHandler = NewAdminMessagesHandler()
		bots.Bots[lang].AdminCallBackHandler = NewAdminCallbackHandler()
		k++
	}

	log.Println("All bots is running")
}

func startServices() {
	bots.FillBotsConfig()
	assets.ParseLangMap()
	assets.ParseTasks()
	assets.ParseAdminMap()
	assets.UploadAdminSettings()
	assets.ParseCommandsList()

	log.Println("All services are running successfully")
}

func StartBot(b *bots.GlobalBot, lang string, k int) {
	var err error
	b.Bot, err = tgbotapi.NewBotAPI(b.BotToken)
	if err != nil {
		log.Println(err)
		return
	}

	u := tgbotapi.NewUpdate(0)

	b.Chanel, err = b.Bot.GetUpdatesChan(u)
	if err != nil {
		panic("Failed to initialize bot: " + err.Error())
	}

	b.Rdb = bots.StartRedis(k)
	b.DataBase = bots.UploadDataBase(lang)
}

func startHandlers() {
	wg := new(sync.WaitGroup)

	for botLang, handler := range bots.Bots {
		wg.Add(1)
		go func(botLang string, handler *bots.GlobalBot, wg *sync.WaitGroup) {
			defer wg.Done()
			services.ActionsWithUpdates(botLang, handler.Chanel)
		}(botLang, handler, wg)
	}

	log.Println("All handlers are running")
	wg.Wait()
}

func NewMessagesHandler() *services.MessagesHandlers {
	handle := services.MessagesHandlers{
		Handlers: map[string]bots.Handler{},
	}

	handle.Init()
	return &handle
}

func NewCallbackHandler() *services.CallBackHandlers {
	handle := services.CallBackHandlers{
		Handlers: map[string]bots.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminMessagesHandler() *administrator.AdminMessagesHandlers {
	handle := administrator.AdminMessagesHandlers{
		Handlers: map[string]bots.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminCallbackHandler() *administrator.AdminCallbackHandlers {
	handle := administrator.AdminCallbackHandlers{
		Handlers: map[string]bots.Handler{},
	}

	handle.Init()
	return &handle
}
