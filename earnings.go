package flashalpha

// Typed models + client methods for the Earnings analytics family
// (/v1/earnings/*, Growth+/Alpha+):
//
//   - EarningsCalendar           — GET /v1/earnings/calendar                 (Growth+)
//   - EarningsExpectedMove       — GET /v1/earnings/expected-move/{symbol}   (Growth+)
//   - EarningsHistory            — GET /v1/earnings/history/{symbol}         (Growth+)
//   - EarningsIvCrush            — GET /v1/earnings/iv-crush/{symbol}        (Growth+)
//   - EarningsVrp                — GET /v1/earnings/vrp/{symbol}             (Alpha+)
//   - EarningsDealerPositioning  — GET /v1/earnings/dealer-positioning/{symbol} (Alpha+)
//   - EarningsStrategies         — GET /v1/earnings/strategies/{symbol}      (Alpha+)
//   - EarningsScreener           — GET /v1/earnings/screener                 (Alpha+)
//
// Every untyped method returns map[string]interface{}; the *Typed wrappers
// decode into the structs below.

import (
	"context"
	"net/url"
	"strconv"
)

// ── earnings/calendar ────────────────────────────────────────────────────────

// EarningsEvent is one upcoming earnings calendar entry.
type EarningsEvent struct {
	Symbol         string   `json:"symbol"`
	CompanyName    string   `json:"company_name"`
	EarningsDate   string   `json:"earnings_date"`
	Timing         *string  `json:"timing"`
	IsConfirmed    bool     `json:"is_confirmed"`
	FiscalPeriod   *string  `json:"fiscal_period"`
	FiscalYear     *int     `json:"fiscal_year"`
	Importance     *int     `json:"importance"`
	EpsEstimate    *float64 `json:"eps_estimate"`
	ImpliedMovePct *float64 `json:"implied_move_pct"`
	DaysToEvent    int      `json:"days_to_event"`
}

// EarningsCalendarResponse is the body of GET /v1/earnings/calendar.
type EarningsCalendarResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Events []EarningsEvent `json:"events"`
	Count  int             `json:"count"`
}

// EarningsCalendarOption configures an EarningsCalendar request.
type EarningsCalendarOption func(*earningsCalendarConfig)

type earningsCalendarConfig struct {
	days       *int
	symbols    string
	importance *int
}

// WithEarningsCalendarDays sets the forward window in days (clamped 1–90).
func WithEarningsCalendarDays(days int) EarningsCalendarOption {
	return func(c *earningsCalendarConfig) { c.days = &days }
}

// WithEarningsCalendarSymbols filters to a comma-separated symbol list.
func WithEarningsCalendarSymbols(symbols string) EarningsCalendarOption {
	return func(c *earningsCalendarConfig) { c.symbols = symbols }
}

// WithEarningsCalendarImportance returns only events with importance >= this.
func WithEarningsCalendarImportance(importance int) EarningsCalendarOption {
	return func(c *earningsCalendarConfig) { c.importance = &importance }
}

// EarningsCalendar returns the upcoming earnings calendar over a forward
// window, optionally filtered to symbols and a minimum importance. Requires
// Growth+.
func (c *Client) EarningsCalendar(ctx context.Context, opts ...EarningsCalendarOption) (map[string]interface{}, error) {
	cfg := &earningsCalendarConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.days != nil {
		params.Set("days", strconv.Itoa(*cfg.days))
	}
	if cfg.symbols != "" {
		params.Set("symbols", cfg.symbols)
	}
	if cfg.importance != nil {
		params.Set("importance", strconv.Itoa(*cfg.importance))
	}
	return c.get(ctx, "/v1/earnings/calendar", params)
}

// EarningsCalendarTyped is the strongly-typed variant of EarningsCalendar.
func (c *Client) EarningsCalendarTyped(ctx context.Context, opts ...EarningsCalendarOption) (*EarningsCalendarResponse, error) {
	raw, err := c.EarningsCalendar(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &EarningsCalendarResponse{}
	if err := decodeTyped("earnings calendar", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── earnings/expected-move ───────────────────────────────────────────────────

// EarningsExpectedMoveBlock decomposes the front-straddle expected move.
type EarningsExpectedMoveBlock struct {
	RawStraddlePct     *float64 `json:"raw_straddle_pct"`
	EarningsImpliedPct *float64 `json:"earnings_implied_pct"`
	BaselineDriftPct   *float64 `json:"baseline_drift_pct"`
	EarningsIv         *float64 `json:"earnings_iv"`
	TermIvPostEvent    *float64 `json:"term_iv_post_event"`
	TermKinkPct        *float64 `json:"term_kink_pct"`
}

// EarningsExpectedMoveResponse is the body of GET /v1/earnings/expected-move/{symbol}.
type EarningsExpectedMoveResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol          string                     `json:"symbol"`
	UnderlyingPrice *float64                   `json:"underlying_price"`
	AsOf            string                     `json:"as_of"`
	EarningsDate    string                     `json:"earnings_date"`
	Session         *string                    `json:"session"`
	DaysToEvent     int                        `json:"days_to_event"`
	ExpectedMove    *EarningsExpectedMoveBlock `json:"expected_move"`
}

// EarningsExpectedMove returns the live earnings-implied move decomposition for
// the next event (earnings-jump vs baseline diffusion). Requires Growth+.
func (c *Client) EarningsExpectedMove(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/earnings/expected-move/"+seg(symbol), nil)
}

// EarningsExpectedMoveTyped is the strongly-typed variant of EarningsExpectedMove.
func (c *Client) EarningsExpectedMoveTyped(ctx context.Context, symbol string) (*EarningsExpectedMoveResponse, error) {
	raw, err := c.EarningsExpectedMove(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &EarningsExpectedMoveResponse{}
	if err := decodeTyped("earnings expected-move", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── earnings/history ─────────────────────────────────────────────────────────

// EarningsHistoryEvent is one past earnings event with actuals and reactions.
type EarningsHistoryEvent struct {
	Date               string   `json:"date"`
	FiscalPeriod       *string  `json:"fiscal_period"`
	FiscalYear         *int     `json:"fiscal_year"`
	EpsEstimate        *float64 `json:"eps_estimate"`
	EpsActual          *float64 `json:"eps_actual"`
	EpsSurprisePct     *float64 `json:"eps_surprise_pct"`
	RevenueActual      *float64 `json:"revenue_actual"`
	RevenueSurprisePct *float64 `json:"revenue_surprise_pct"`
	ImpliedMovePct     *float64 `json:"implied_move_pct"`
	ActualMovePct      *float64 `json:"actual_move_pct"`
	IvCrushPct         *float64 `json:"iv_crush_pct"`
	PreAtmIv           *float64 `json:"pre_atm_iv"`
	PostAtmIv          *float64 `json:"post_atm_iv"`
}

// EarningsHistoryResponse is the body of GET /v1/earnings/history/{symbol}.
type EarningsHistoryResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol  string                 `json:"symbol"`
	Count   int                    `json:"count"`
	History []EarningsHistoryEvent `json:"history"`
}

// EarningsHistoryOption configures an EarningsHistory request.
type EarningsHistoryOption func(*earningsHistoryConfig)

type earningsHistoryConfig struct {
	limit *int
}

// WithEarningsHistoryLimit sets the number of most-recent events (clamped 1–40).
func WithEarningsHistoryLimit(limit int) EarningsHistoryOption {
	return func(c *earningsHistoryConfig) { c.limit = &limit }
}

// EarningsHistory returns past earnings events with EPS/revenue actuals and
// surprises, implied vs actual moves, and realized IV crush. Requires Growth+.
func (c *Client) EarningsHistory(ctx context.Context, symbol string, opts ...EarningsHistoryOption) (map[string]interface{}, error) {
	cfg := &earningsHistoryConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.limit != nil {
		params.Set("limit", strconv.Itoa(*cfg.limit))
	}
	return c.get(ctx, "/v1/earnings/history/"+seg(symbol), params)
}

// EarningsHistoryTyped is the strongly-typed variant of EarningsHistory.
func (c *Client) EarningsHistoryTyped(ctx context.Context, symbol string, opts ...EarningsHistoryOption) (*EarningsHistoryResponse, error) {
	raw, err := c.EarningsHistory(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &EarningsHistoryResponse{}
	if err := decodeTyped("earnings history", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── earnings/iv-crush ────────────────────────────────────────────────────────

// EarningsIvCrushEstimate is the live IV-crush estimate for the next event.
type EarningsIvCrushEstimate struct {
	ExpectedCrushPct *float64 `json:"expected_crush_pct"`
	PreIv            *float64 `json:"pre_iv"`
	PostIv           *float64 `json:"post_iv"`
}

// EarningsIvCrushDistribution is the historical IV-crush distribution.
type EarningsIvCrushDistribution struct {
	Median *float64 `json:"median"`
	P25    *float64 `json:"p25"`
	P75    *float64 `json:"p75"`
	Worst  *float64 `json:"worst"`
	Best   *float64 `json:"best"`
	Count  int      `json:"count"`
}

// EarningsIvCrushResponse is the body of GET /v1/earnings/iv-crush/{symbol}.
type EarningsIvCrushResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol          string                       `json:"symbol"`
	AsOf            string                       `json:"as_of"`
	EarningsDate    *string                      `json:"earnings_date"`
	CurrentEstimate *EarningsIvCrushEstimate     `json:"current_estimate"`
	Distribution    *EarningsIvCrushDistribution `json:"distribution"`
}

// EarningsIvCrush returns the expected IV crush for the next event plus the
// symbol's historical IV-crush distribution. Requires Growth+.
func (c *Client) EarningsIvCrush(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/earnings/iv-crush/"+seg(symbol), nil)
}

// EarningsIvCrushTyped is the strongly-typed variant of EarningsIvCrush.
func (c *Client) EarningsIvCrushTyped(ctx context.Context, symbol string) (*EarningsIvCrushResponse, error) {
	raw, err := c.EarningsIvCrush(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &EarningsIvCrushResponse{}
	if err := decodeTyped("earnings iv-crush", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── earnings/vrp ─────────────────────────────────────────────────────────────

// EarningsVrpBlock is the earnings VRP block.
type EarningsVrpBlock struct {
	ImpliedMovePct  *float64 `json:"implied_move_pct"`
	RealizedMedian  *float64 `json:"realized_median"`
	RealizedMean    *float64 `json:"realized_mean"`
	PremiumRatio    *float64 `json:"premium_ratio"`
	ZScore          *float64 `json:"z_score"`
	Percentile      *int     `json:"percentile"`
	Assessment      string   `json:"assessment"`
	DirectionalBias *string  `json:"directional_bias"`
}

// EarningsSurpriseReaction is the surprise-reaction breakdown.
type EarningsSurpriseReaction struct {
	BeatAvgMovePct   *float64 `json:"beat_avg_move_pct"`
	MissAvgMovePct   *float64 `json:"miss_avg_move_pct"`
	InlineAvgMovePct *float64 `json:"inline_avg_move_pct"`
}

// EarningsVrpResponse is the body of GET /v1/earnings/vrp/{symbol}.
type EarningsVrpResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol           string                    `json:"symbol"`
	UnderlyingPrice  *float64                  `json:"underlying_price"`
	AsOf             string                    `json:"as_of"`
	EarningsDate     string                    `json:"earnings_date"`
	DaysToEvent      int                       `json:"days_to_event"`
	EarningsVrp      *EarningsVrpBlock         `json:"earnings_vrp"`
	SurpriseReaction *EarningsSurpriseReaction `json:"surprise_reaction"`
}

// EarningsVrp returns the earnings volatility-risk-premium: live event-implied
// move vs the symbol's realized history, with a richness assessment and
// surprise-reaction breakdown. Requires Alpha+.
func (c *Client) EarningsVrp(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/earnings/vrp/"+seg(symbol), nil)
}

// EarningsVrpTyped is the strongly-typed variant of EarningsVrp.
func (c *Client) EarningsVrpTyped(ctx context.Context, symbol string) (*EarningsVrpResponse, error) {
	raw, err := c.EarningsVrp(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &EarningsVrpResponse{}
	if err := decodeTyped("earnings vrp", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── earnings/dealer-positioning ──────────────────────────────────────────────

// EarningsDealerLevels holds the event-week dealer levels.
type EarningsDealerLevels struct {
	GammaFlip       *float64 `json:"gamma_flip"`
	CallWall        *float64 `json:"call_wall"`
	PutWall         *float64 `json:"put_wall"`
	HighestOiStrike *float64 `json:"highest_oi_strike"`
}

// EarningsGexBucket is one pre/event/post net-GEX bucket.
type EarningsGexBucket struct {
	Bucket        string  `json:"bucket"`
	NetGex        float64 `json:"net_gex"`
	ContractCount int     `json:"contract_count"`
}

// EarningsTopStrike is one of the top strikes by absolute net GEX.
type EarningsTopStrike struct {
	Strike float64 `json:"strike"`
	NetGex float64 `json:"net_gex"`
	CallOi float64 `json:"call_oi"`
	PutOi  float64 `json:"put_oi"`
}

// EarningsDealerPositioningResponse is the body of
// GET /v1/earnings/dealer-positioning/{symbol}.
type EarningsDealerPositioningResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol            string                `json:"symbol"`
	UnderlyingPrice   *float64              `json:"underlying_price"`
	AsOf              string                `json:"as_of"`
	EarningsDate      string                `json:"earnings_date"`
	EventExpiry       *string               `json:"event_expiry"`
	Levels            *EarningsDealerLevels `json:"levels"`
	GexByDteBucket    []EarningsGexBucket   `json:"gex_by_dte_bucket"`
	TopStrikes        []EarningsTopStrike   `json:"top_strikes"`
	CharmAcceleration *float64              `json:"charm_acceleration"`
	Regime            string                `json:"regime"`
}

// EarningsDealerPositioning returns dealer exposure analysis scoped to the
// earnings event (event-week levels, pre/event/post net-GEX buckets, charm
// acceleration, top strikes). Requires Alpha+.
func (c *Client) EarningsDealerPositioning(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/earnings/dealer-positioning/"+seg(symbol), nil)
}

// EarningsDealerPositioningTyped is the strongly-typed variant.
func (c *Client) EarningsDealerPositioningTyped(ctx context.Context, symbol string) (*EarningsDealerPositioningResponse, error) {
	raw, err := c.EarningsDealerPositioning(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &EarningsDealerPositioningResponse{}
	if err := decodeTyped("earnings dealer-positioning", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── earnings/strategies ──────────────────────────────────────────────────────

// EarningsStrategyScores holds 0–100 suitability scores per structure.
type EarningsStrategyScores struct {
	LongStraddle     *int `json:"long_straddle"`
	ShortStrangle    *int `json:"short_strangle"`
	IronCondor       *int `json:"iron_condor"`
	CalendarSpread   *int `json:"calendar_spread"`
	EarningsDiagonal *int `json:"earnings_diagonal"`
}

// EarningsStrategyContext is the scoring-input context block.
type EarningsStrategyContext struct {
	PremiumRatio   *float64 `json:"premium_ratio"`
	IvCrushMedian  *float64 `json:"iv_crush_median"`
	Regime         string   `json:"regime"`
	ImpliedMovePct *float64 `json:"implied_move_pct"`
}

// EarningsStrategiesResponse is the body of GET /v1/earnings/strategies/{symbol}.
type EarningsStrategiesResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol       string                   `json:"symbol"`
	AsOf         string                   `json:"as_of"`
	EarningsDate string                   `json:"earnings_date"`
	Scores       *EarningsStrategyScores  `json:"scores"`
	Context      *EarningsStrategyContext `json:"context"`
}

// EarningsStrategies returns strategy-suitability scores (0–100) for the
// upcoming event across common earnings structures. Requires Alpha+.
func (c *Client) EarningsStrategies(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/earnings/strategies/"+seg(symbol), nil)
}

// EarningsStrategiesTyped is the strongly-typed variant of EarningsStrategies.
func (c *Client) EarningsStrategiesTyped(ctx context.Context, symbol string) (*EarningsStrategiesResponse, error) {
	raw, err := c.EarningsStrategies(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &EarningsStrategiesResponse{}
	if err := decodeTyped("earnings strategies", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── earnings/screener ────────────────────────────────────────────────────────

// EarningsScreenerEvent is one ranked upcoming earnings event.
type EarningsScreenerEvent struct {
	Symbol         string   `json:"symbol"`
	CompanyName    string   `json:"company_name"`
	EarningsDate   string   `json:"earnings_date"`
	DaysToEvent    int      `json:"days_to_event"`
	Timing         *string  `json:"timing"`
	Importance     *int     `json:"importance"`
	ImpliedMovePct *float64 `json:"implied_move_pct"`
	PremiumRatio   *float64 `json:"premium_ratio"`
	IvCrushMedian  *float64 `json:"iv_crush_median"`
	Assessment     *string  `json:"assessment"`
}

// EarningsScreenerResponse is the body of GET /v1/earnings/screener.
type EarningsScreenerResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Events []EarningsScreenerEvent `json:"events"`
	Count  int                     `json:"count"`
}

// EarningsScreenerOption configures an EarningsScreener request.
type EarningsScreenerOption func(*earningsScreenerConfig)

type earningsScreenerConfig struct {
	sort          string
	limit         *int
	days          *int
	minImportance *int
}

// WithEarningsScreenerSort sets the ranking: vrp_richest (default) /
// cheapest_move / highest_crush / importance.
func WithEarningsScreenerSort(sort string) EarningsScreenerOption {
	return func(c *earningsScreenerConfig) { c.sort = sort }
}

// WithEarningsScreenerLimit caps rows returned (clamped 1–50).
func WithEarningsScreenerLimit(limit int) EarningsScreenerOption {
	return func(c *earningsScreenerConfig) { c.limit = &limit }
}

// WithEarningsScreenerDays sets the forward window in days (clamped 1–60).
func WithEarningsScreenerDays(days int) EarningsScreenerOption {
	return func(c *earningsScreenerConfig) { c.days = &days }
}

// WithEarningsScreenerMinImportance returns only events with importance >= this.
func WithEarningsScreenerMinImportance(minImportance int) EarningsScreenerOption {
	return func(c *earningsScreenerConfig) { c.minImportance = &minImportance }
}

// EarningsScreener runs the cross-sectional earnings screener over upcoming
// events, ranked by VRP richness, cheapest implied move, highest historical
// crush, or importance. Requires Alpha+.
func (c *Client) EarningsScreener(ctx context.Context, opts ...EarningsScreenerOption) (map[string]interface{}, error) {
	cfg := &earningsScreenerConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.sort != "" {
		params.Set("sort", cfg.sort)
	}
	if cfg.limit != nil {
		params.Set("limit", strconv.Itoa(*cfg.limit))
	}
	if cfg.days != nil {
		params.Set("days", strconv.Itoa(*cfg.days))
	}
	if cfg.minImportance != nil {
		params.Set("min_importance", strconv.Itoa(*cfg.minImportance))
	}
	return c.get(ctx, "/v1/earnings/screener", params)
}

// EarningsScreenerTyped is the strongly-typed variant of EarningsScreener.
func (c *Client) EarningsScreenerTyped(ctx context.Context, opts ...EarningsScreenerOption) (*EarningsScreenerResponse, error) {
	raw, err := c.EarningsScreener(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &EarningsScreenerResponse{}
	if err := decodeTyped("earnings screener", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
