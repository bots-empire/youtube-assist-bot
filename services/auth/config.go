package auth

type User struct {
	ID            int    `json:"id"`
	Balance       int    `json:"balance"`
	Completed     int    `json:"completed"`
	CompletedT    int    `json:"completed_t"`
	CompletedY    int    `json:"completed_y"`
	CompletedA    int    `json:"completed_a"`
	LastViewT     int64  `json:"last_view_t"`
	LastViewY     int64  `json:"last_view_y"`
	LastViewA     int64  `json:"last_view_a"`
	ReferralCount int    `json:"referral_count"`
	TakeBonus     bool   `json:"take_bonus"`
	Language      string `json:"language"`
}
