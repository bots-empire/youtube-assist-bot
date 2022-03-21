package assets

import (
	"encoding/json"
	"math/rand"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Stepan1328/youtube-assist-bot/model"
)

const (
	commandsPath             = "assets/commands"
	beginningOfAdminLangPath = "assets/admin/"
	beginningOfUserLangPath  = "assets/language/"
)

var (
	AvailableAdminLang = []string{"en", "ru"}
	AvailableLang      = []string{"it", "it2", "br", "es", "mx", "en", "fr" /*, "de"*/}

	Commands     = make(map[string]string)
	Language     = make(map[string]map[string]string)
	AdminLibrary = make(map[string]map[string]string)
	Tasks        = make(map[string]Task, len(AvailableLang))
)

type Task struct {
	Partition map[string][]*Link
}

type Link struct {
	Url             string
	FileID          string
	Duration        int
	Limited         bool
	ImpressionsLeft int
}

func ParseLangMap() {
	for _, lang := range AvailableLang {
		bytes, _ := os.ReadFile(beginningOfUserLangPath + lang + jsonFormatName)
		dictionary := make(map[string]string)
		_ = json.Unmarshal(bytes, &dictionary)

		Language[lang] = dictionary
	}
}

func LangText(lang, key string) string {
	return Language[lang][key]
}

func ParseTasks() {
	for _, lang := range AvailableLang {
		task := Task{}
		bytes, _ := os.ReadFile(beginningOfTaskPath + lang + jsonFormatName)
		_ = json.Unmarshal(bytes, &task)
		Tasks[lang] = task
	}
}

func GetTask(s model.Situation) (*model.LinkInfo, error) {
	if len(Tasks[s.BotLang].Partition[s.Params.Partition]) == 0 {
		return nil, model.ErrTaskNotFound
	}

	num := rand.Intn(len(Tasks[s.BotLang].Partition[s.Params.Partition]))
	if checkLimitedLink(s, num) {
		link := Tasks[s.BotLang].Partition[s.Params.Partition][num]
		return &model.LinkInfo{
			Url:      link.Url,
			FileID:   link.FileID,
			Duration: link.Duration,
		}, nil
	}
	return GetTask(s)
}

func checkLimitedLink(s model.Situation, num int) bool {
	if !Tasks[s.BotLang].Partition[s.Params.Partition][num].Limited {
		return true
	}

	defer SaveTasks(s.BotLang)
	if Tasks[s.BotLang].Partition[s.Params.Partition][num].ImpressionsLeft == 0 {
		length := len(Tasks[s.BotLang].Partition[s.Params.Partition])
		Tasks[s.BotLang].Partition[s.Params.Partition][num] = Tasks[s.BotLang].Partition[s.Params.Partition][length-1]
		Tasks[s.BotLang].Partition[s.Params.Partition] = Tasks[s.BotLang].Partition[s.Params.Partition][:length-1]
		return false
	}
	Tasks[s.BotLang].Partition[s.Params.Partition][num].ImpressionsLeft--
	return true
}

func ParseCommandsList() {
	bytes, _ := os.ReadFile(commandsPath + jsonFormatName)
	_ = json.Unmarshal(bytes, &Commands)
}

func ParseAdminMap() {
	for _, lang := range AvailableAdminLang {
		bytes, _ := os.ReadFile(beginningOfAdminLangPath + lang + jsonFormatName)
		dictionary := make(map[string]string)
		_ = json.Unmarshal(bytes, &dictionary)

		AdminLibrary[lang] = dictionary
	}
}

func AdminText(lang, key string) string {
	return AdminLibrary[lang][key]
}

func GetCommandFromText(s model.Situation) (string, error) {
	searchText := getSearchText(s.Message)
	for key, text := range Language[s.User.Language] {
		if text == searchText {
			return Commands[key], nil
		}
	}

	if command := searchInCommandAdmins(s.User.ID, searchText); command != "" {
		return command, nil
	}

	command := Commands[searchText]
	if command != "" {
		return command, nil
	}

	return "", model.ErrCommandNotConverted
}

func getSearchText(message *tgbotapi.Message) string {
	if message.Command() != "" {
		return strings.Split(message.Text, " ")[0]
	}
	return message.Text
}

func searchInCommandAdmins(userID int64, searchText string) string {
	lang := getAdminLang(userID)
	for key, text := range AdminLibrary[lang] {
		if text == searchText {
			return Commands[key]
		}
	}
	return ""
}

func getAdminLang(userID int64) string {
	for key := range AdminSettings.AdminID {
		if key == userID {
			return AdminSettings.AdminID[key].Language
		}
	}
	return ""
}

func AdminLang(userID int64) string {
	return AdminSettings.AdminID[userID].Language
}
