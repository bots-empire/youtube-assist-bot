package assets

import (
	"encoding/json"
	"github.com/Stepan1328/youtube-assist-bot/bots"
	"github.com/Stepan1328/youtube-assist-bot/err"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"math/rand"
	"os"
	"strings"
)

var (
	AvailableAdminLang = []string{"en", "ru"}
	AvailableLang      = []string{"en", "it", "pt", "es" /*, "de"*/}

	Commands     = make(map[string]string)
	Language     = make([]map[string]string, 5)
	AdminLibrary = make([]map[string]string, 2)
	Tasks        = make(map[string]Task, 5)
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

//type Quest interface {
//	GetTask() string
//	IsLimited() bool
//	GetImpressionsLeft() int
//	DecreaseImpressionsLeft()
//}
//
//func (l *Link) GetTask() string {
//	return l.Url
//}
//
//func (l *Link) IsLimited() bool {
//	return l.Limited
//}
//
//func (l *Link) GetImpressionsLeft() int {
//	return l.ImpressionsLeft
//}
//
//func (l *Link) DecreaseImpressionsLeft() {
//	l.ImpressionsLeft--
//}

func ParseLangMap() {
	for i, lang := range AvailableLang {
		bytes, _ := os.ReadFile("./assets/language/" + lang + ".json")
		_ = json.Unmarshal(bytes, &Language[i])
	}
}

func LangText(lang, key string) string {
	index := findLangIndex(lang)
	return Language[index][key]
}

func ParseTasks() {
	for _, lang := range AvailableLang {
		task := Task{}
		bytes, _ := os.ReadFile("./assets/task/" + lang + ".json")
		_ = json.Unmarshal(bytes, &task)
		Tasks[lang] = task
	}
}

func GetTask(s bots.Situation) (*bots.LinkInfo, error) {
	if len(Tasks[s.BotLang].Partition[s.Params.Partition]) == 0 {
		return nil, err.ErrTaskNotFound
	}

	num := rand.Intn(len(Tasks[s.BotLang].Partition[s.Params.Partition]))
	if checkLimitedLink(s, num) {
		link := Tasks[s.BotLang].Partition[s.Params.Partition][num]
		return &bots.LinkInfo{
			Url:      link.Url,
			FileID:   link.FileID,
			Duration: link.Duration,
		}, nil
	}
	return GetTask(s)
}

func checkLimitedLink(s bots.Situation, num int) bool {
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
	bytes, _ := os.ReadFile("./assets/commands.json")
	_ = json.Unmarshal(bytes, &Commands)
}

func ParseAdminMap() {
	for i, lang := range AvailableAdminLang {
		bytes, _ := os.ReadFile("./assets/admin/" + lang + ".json")
		_ = json.Unmarshal(bytes, &AdminLibrary[i])
	}
}

func AdminText(lang, key string) string {
	index := findAdminLangIndex(lang)
	return AdminLibrary[index][key]
}

func GetCommandFromText(s bots.Situation) (string, error) {
	searchText := getSearchText(s.Message)
	for key, text := range Language[findLangIndex(s.UserLang)] {
		if text == searchText {
			return Commands[key], nil
		}
	}

	if command := searchInCommandAdmins(s, searchText); command != "" {
		return command, nil
	}

	command := Commands[searchText]
	if command != "" {
		return command, nil
	}

	return "", err.ErrCommandNotConverted
}

func getSearchText(message *tgbotapi.Message) string {
	if message.Command() != "" {
		return strings.Split(message.Text, " ")[0]
	}
	return message.Text
}

func searchInCommandAdmins(s bots.Situation, searchText string) string {
	lang := getAdminLang(s.UserID)
	for key, text := range AdminLibrary[findAdminLangIndex(lang)] {
		if text == searchText {
			return Commands[key]
		}
	}
	return ""
}

func getAdminLang(userID int) string {
	for key := range AdminSettings.AdminID {
		if key == userID {
			return AdminSettings.AdminID[key].Language
		}
	}
	return ""
}

func findLangIndex(lang string) int {
	for i, elem := range AvailableLang {
		if elem == lang {
			return i
		}
	}
	return 0
}

func findAdminLangIndex(lang string) int {
	for i, elem := range AvailableAdminLang {
		if elem == lang {
			return i
		}
	}
	return 0
}

func AdminLang(userID int) string {
	return AdminSettings.AdminID[userID].Language
}
