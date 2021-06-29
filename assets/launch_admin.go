package assets

import (
	"encoding/json"
	"fmt"
	"os"
)

type Admin struct {
	AdminID         map[int]*AdminUser
	Parameters      map[string]*Params
	AdvertisingChan map[string]*AdvertChannel
	BlockedUsers    map[string]int
	LangSelectedMap map[string]bool
	AdvertisingText map[string]string
}

type AdminUser struct {
	Language           string
	FirstName          string
	SpecialPossibility bool
}

type Params struct {
	BonusAmount         int
	MinWithdrawalAmount int
	WatchReward         int
	SecondBetweenViews  int64
	MaxOfVideoPerDay    int
	ReferralAmount      int
}

type AdvertChannel struct {
	Url       string
	ChannelID int64
}

var AdminSettings *Admin

func UploadAdminSettings() {
	var settings *Admin
	data, err := os.ReadFile("assets/admin.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println(err)
	}

	AdminSettings = settings
}

func SaveAdminSettings() {
	data, err := json.MarshalIndent(AdminSettings, "", "  ")
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile("assets/admin.json", data, 0600); err != nil {
		panic(err)
	}
}

type UpdateInfo struct {
	Counter int
	Day     int
}

var UpdateStatistic *UpdateInfo

func UploadUpdateStatistic() {
	var info *UpdateInfo
	data, err := os.ReadFile("assets/statistic.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &info)
	if err != nil {
		fmt.Println(err)
	}

	UpdateStatistic = info
}

func SaveUpdateStatistic() {
	data, err := json.MarshalIndent(UpdateStatistic, "", "  ")
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile("assets/statistic.json", data, 0600); err != nil {
		panic(err)
	}
}

func SaveTasks(botLang string) {
	data, err := json.MarshalIndent(Tasks[botLang], "", "  ")
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile("assets/task/"+botLang+".json", data, 0600); err != nil {
		panic(err)
	}
}
