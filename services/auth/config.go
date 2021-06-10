package auth

type User struct {
	ID             int    `json:"id"`
	Balance        int    `json:"balance"`
	Completed      int    `json:"completed"`
	CompletedToday int    `json:"completed_today"`
	LastView       int64  `json:"last_view"`
	ReferralCount  int    `json:"referral_count"`
	TakeBonus      bool   `json:"take_bonus"`
	Language       string `json:"language"`
}
