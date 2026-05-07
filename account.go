package flashalpha

// Typed response model for `GET /v1/account` (all plans).
//
// Account profile + daily quota usage. DailyLimit and Remaining are returned
// as STRINGS on the wire — they are integer counts on numeric plans
// (e.g. "1000") but the literal string "unlimited" for Business /
// Institutional / Enterprise tiers. UsageToday is an integer (always 0 on
// unlimited plans).

// AccountResponse is the typed body of GET /v1/account.
type AccountResponse struct {
	// UserID is the account's unique identifier (GUID).
	UserID string `json:"user_id"`
	// Email is the user's registered email address.
	Email string `json:"email"`
	// Plan is the plan name: "free", "basic", "growth", "alpha", "enterprise", etc.
	Plan string `json:"plan"`
	// DailyLimit is the daily request limit as a string. Numeric on bounded
	// plans (e.g. "1000"); literal "unlimited" on uncapped tiers.
	DailyLimit string `json:"daily_limit"`
	// UsageToday is the number of API requests made so far today. Always 0 on
	// unlimited plans.
	UsageToday int `json:"usage_today"`
	// Remaining is requests remaining today as a string. Numeric on bounded
	// plans (e.g. "958"); literal "unlimited" on uncapped tiers.
	Remaining string `json:"remaining"`
	// ResetsAt is the ISO 8601 timestamp when the daily quota resets
	// (midnight UTC).
	ResetsAt string `json:"resets_at"`
}
