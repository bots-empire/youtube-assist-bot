package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Stepan1328/youtube-assist-bot/assets"
	"github.com/Stepan1328/youtube-assist-bot/log"
	"github.com/Stepan1328/youtube-assist-bot/model"
	msgs "github.com/Stepan1328/youtube-assist-bot/msgs"
	"github.com/Stepan1328/youtube-assist-bot/services"
	"github.com/Stepan1328/youtube-assist-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	rand.Seed(time.Now().Unix())

	logger := log.NewDefaultLogger().Prefix("YouTube Bot")
	log.PrintLogo("YouTube Bot", []string{"FF5655"})

	startServices(logger)
	startAllBot(logger)
	assets.UploadUpdateStatistic()

	startHandlers(logger)
}

func startAllBot(logger log.Logger) {
	k := 0
	for lang, globalBot := range model.Bots {
		startBot(globalBot, logger, lang, k)
		model.Bots[lang].MessageHandler = NewMessagesHandler()
		model.Bots[lang].CallbackHandler = NewCallbackHandler()
		model.Bots[lang].AdminMessageHandler = NewAdminMessagesHandler()
		model.Bots[lang].AdminCallBackHandler = NewAdminCallbackHandler()
		k++
	}

	logger.Ok("All bots is running")
}

func startServices(logger log.Logger) {
	model.FillBotsConfig()
	assets.ParseLangMap()
	assets.ParseTasks()
	assets.ParseAdminMap()
	assets.UploadAdminSettings()
	assets.ParseCommandsList()

	logger.Ok("All services are running successfully")
}

func startBot(b *model.GlobalBot, logger log.Logger, lang string, k int) {
	var err error
	b.Bot, err = tgbotapi.NewBotAPI(b.BotToken)
	if err != nil {
		logger.Fatal("error start bot: %s", err.Error())
	}

	u := tgbotapi.NewUpdate(0)

	b.Chanel = b.Bot.GetUpdatesChan(u)

	b.Rdb = model.StartRedis(k)
	b.DataBase = model.UploadDataBase(lang)
}

func startHandlers(logger log.Logger) {
	wg := new(sync.WaitGroup)

	for botLang, handler := range model.Bots {
		wg.Add(1)
		go func(botLang string, handler *model.GlobalBot, wg *sync.WaitGroup) {
			defer wg.Done()
			services.ActionsWithUpdates(botLang, handler.Chanel, logger)
		}(botLang, handler, wg)
	}

	logger.Ok("All handlers are running")
	_ = msgs.NewParseMessage("it", 1418862576, "All bots are restart")
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
