package flashalpha

// Typed models + client methods for the additional volatility, macro, universe,
// surface, and expected-move surfaces:
//
//   - Liquidity            — GET /v1/liquidity/{symbol}                 (Growth+)
//   - SkewTerm             — GET /v1/volatility/skew-term/{symbol}      (Growth+)
//   - SpotVolCorrelation   — GET /v1/volatility/spot-vol-correlation/{symbol} (Growth+)
//   - Dispersion           — GET /v1/dispersion                         (Alpha+)
//   - VixState             — GET /v1/macro/vix-state                    (Growth+)
//   - Universe             — GET /v1/universe                           (Public)
//   - SurfaceSvi           — GET /v1/surface/svi/{symbol}               (Alpha+)
//   - ExpectedMove         — GET /v1/expected-move/{symbol}             (Basic+)
//   - VrpHistory           — GET /v1/vrp/{symbol}/history               (Alpha+)
//
// Every untyped method returns map[string]interface{}; the *Typed wrappers
// decode into the structs below.

import (
	"context"
	"net/url"
	"strconv"
)

// ── liquidity ────────────────────────────────────────────────────────────────

// LiquidityExpiry is one expiry's execution-quality breakdown.
type LiquidityExpiry struct {
	Expiration        string   `json:"expiration"`
	Dte               int      `json:"dte"`
	AtmSpreadPct      *float64 `json:"atm_spread_pct"`
	WeightedSpreadPct *float64 `json:"weighted_spread_pct"`
	AtmOi             float64  `json:"atm_oi"`
	ExecutionScore    int      `json:"execution_score"`
	// Label is one of tight / normal / wide / illiquid.
	Label string `json:"label"`
}

// LiquidityResponse is the body of GET /v1/liquidity/{symbol}.
type LiquidityResponse struct {
	Symbol              string            `json:"symbol"`
	UnderlyingPrice     *float64          `json:"underlying_price"`
	AsOf                string            `json:"as_of"`
	ChainExecutionScore int               `json:"chain_execution_score"`
	BestExpiry          *string           `json:"best_expiry"`
	WorstExpiry         *string           `json:"worst_expiry"`
	ThinExpiryCount     int               `json:"thin_expiry_count"`
	Expiries            []LiquidityExpiry `json:"expiries"`
}

// Liquidity returns per-expiry execution scores, ATM/OI-weighted spread %, ATM
// OI depth, and chain-level score with best/worst expiry. Requires Growth+.
func (c *Client) Liquidity(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/liquidity/"+seg(symbol), nil)
}

// LiquidityTyped is the strongly-typed variant of Liquidity.
func (c *Client) LiquidityTyped(ctx context.Context, symbol string) (*LiquidityResponse, error) {
	raw, err := c.Liquidity(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &LiquidityResponse{}
	if err := decodeTyped("liquidity", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── volatility/skew-term ─────────────────────────────────────────────────────

// SkewTermExpiry is one expiry's skew term-structure point.
type SkewTermExpiry struct {
	Expiry          string   `json:"expiry"`
	Dte             int      `json:"dte"`
	AtmIv           *float64 `json:"atm_iv"`
	Put25dIv        *float64 `json:"put_25d_iv"`
	Call25dIv       *float64 `json:"call_25d_iv"`
	Put10dIv        *float64 `json:"put_10d_iv"`
	Call10dIv       *float64 `json:"call_10d_iv"`
	Skew25d         *float64 `json:"skew_25d"`
	RiskReversal25d *float64 `json:"risk_reversal_25d"`
	Butterfly25d    *float64 `json:"butterfly_25d"`
	TailConvexity   *float64 `json:"tail_convexity"`
}

// SkewTermResponse is the body of GET /v1/volatility/skew-term/{symbol}.
type SkewTermResponse struct {
	Symbol          string           `json:"symbol"`
	UnderlyingPrice *float64         `json:"underlying_price"`
	AsOf            string           `json:"as_of"`
	Expiries        []SkewTermExpiry `json:"expiries"`
}

// SkewTerm returns the skew term structure with vol-desk naming conventions:
// ATM/25Δ/10Δ wing IVs plus skew_25d, risk_reversal_25d, butterfly_25d, and
// tail_convexity per expiry. Requires Growth+.
func (c *Client) SkewTerm(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/volatility/skew-term/"+seg(symbol), nil)
}

// SkewTermTyped is the strongly-typed variant of SkewTerm.
func (c *Client) SkewTermTyped(ctx context.Context, symbol string) (*SkewTermResponse, error) {
	raw, err := c.SkewTerm(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &SkewTermResponse{}
	if err := decodeTyped("skew-term", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── volatility/spot-vol-correlation ──────────────────────────────────────────

// SpotVolCorrelationResponse is the body of
// GET /v1/volatility/spot-vol-correlation/{symbol}.
type SpotVolCorrelationResponse struct {
	Symbol                string   `json:"symbol"`
	AsOf                  string   `json:"as_of"`
	SpotVolCorrelation20d *float64 `json:"spot_vol_correlation_20d"`
	SpotVolCorrelation60d *float64 `json:"spot_vol_correlation_60d"`
	DataPoints20d         int      `json:"data_points_20d"`
	DataPoints60d         int      `json:"data_points_60d"`
	Interpretation        string   `json:"interpretation"`
}

// SpotVolCorrelation returns the daily Pearson correlation between spot
// log-returns and ATM-IV first differences over 20d and 60d windows.
// Requires Growth+.
func (c *Client) SpotVolCorrelation(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/volatility/spot-vol-correlation/"+seg(symbol), nil)
}

// SpotVolCorrelationTyped is the strongly-typed variant of SpotVolCorrelation.
func (c *Client) SpotVolCorrelationTyped(ctx context.Context, symbol string) (*SpotVolCorrelationResponse, error) {
	raw, err := c.SpotVolCorrelation(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &SpotVolCorrelationResponse{}
	if err := decodeTyped("spot-vol-correlation", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── dispersion ───────────────────────────────────────────────────────────────

// DispersionContributor is one constituent's contribution to basket vol.
type DispersionContributor struct {
	Symbol                  string  `json:"symbol"`
	Weight                  float64 `json:"weight"`
	Iv                      float64 `json:"iv"`
	ContributionToBasketVol float64 `json:"contribution_to_basket_vol"`
}

// DispersionResponse is the body of GET /v1/dispersion.
type DispersionResponse struct {
	AsOf                string                  `json:"as_of"`
	Index               string                  `json:"index"`
	ConstituentCount    int                     `json:"constituent_count"`
	MissingSymbols      []string                `json:"missing_symbols"`
	HorizonDays         int                     `json:"horizon_days"`
	ImpliedCorrelation  *float64                `json:"implied_correlation"`
	RealizedCorrelation *float64                `json:"realized_correlation"`
	CorrelationPremium  *float64                `json:"correlation_premium"`
	ImpliedVolIndex     *float64                `json:"implied_vol_index"`
	ImpliedVolBasket    *float64                `json:"implied_vol_basket"`
	TopContributors     []DispersionContributor `json:"top_contributors"`
}

// DispersionOption configures a Dispersion request.
type DispersionOption func(*dispersionConfig)

type dispersionConfig struct {
	weights     string
	horizonDays *int
}

// WithDispersionWeights sets comma-separated constituent weights (normalised to
// sum 1). Defaults to equal weight.
func WithDispersionWeights(weights string) DispersionOption {
	return func(c *dispersionConfig) { c.weights = weights }
}

// WithDispersionHorizonDays sets the realized-correlation lookback window in
// days (clamped 5–252; default 20).
func WithDispersionHorizonDays(days int) DispersionOption {
	return func(c *dispersionConfig) { c.horizonDays = &days }
}

// Dispersion returns Demeterfi-Derman-Kani implied correlation between an index
// and a user-supplied basket vs a 1-factor realized correlation, plus the
// correlation premium and per-constituent vol contributions. Requires Alpha+.
// index and symbols (comma-separated constituents) are required.
func (c *Client) Dispersion(ctx context.Context, index, symbols string, opts ...DispersionOption) (map[string]interface{}, error) {
	cfg := &dispersionConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	params.Set("index", index)
	params.Set("symbols", symbols)
	if cfg.weights != "" {
		params.Set("weights", cfg.weights)
	}
	if cfg.horizonDays != nil {
		params.Set("horizon_days", strconv.Itoa(*cfg.horizonDays))
	}
	return c.get(ctx, "/v1/dispersion", params)
}

// DispersionTyped is the strongly-typed variant of Dispersion.
func (c *Client) DispersionTyped(ctx context.Context, index, symbols string, opts ...DispersionOption) (*DispersionResponse, error) {
	raw, err := c.Dispersion(ctx, index, symbols, opts...)
	if err != nil {
		return nil, err
	}
	out := &DispersionResponse{}
	if err := decodeTyped("dispersion", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── macro/vix-state ──────────────────────────────────────────────────────────

// VixStateResponse is the body of GET /v1/macro/vix-state.
type VixStateResponse struct {
	AsOf     string   `json:"as_of"`
	Vix      *float64 `json:"vix"`
	SpxRv20d *float64 `json:"spx_rv_20d"`
	Spread   *float64 `json:"spread"`
	Ratio    *float64 `json:"ratio"`
	// State is overvixing / undervixing / neutral.
	State          string `json:"state"`
	Interpretation string `json:"interpretation"`
}

// VixState returns the "overvixing / undervixing" regime label for the index
// complex (VIX spot vs SPX 20-day realized vol). Requires Growth+.
func (c *Client) VixState(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/macro/vix-state", nil)
}

// VixStateTyped is the strongly-typed variant of VixState.
func (c *Client) VixStateTyped(ctx context.Context) (*VixStateResponse, error) {
	raw, err := c.VixState(ctx)
	if err != nil {
		return nil, err
	}
	out := &VixStateResponse{}
	if err := decodeTyped("vix-state", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── universe ─────────────────────────────────────────────────────────────────

// UniverseSymbol is one curated-universe member.
type UniverseSymbol struct {
	Symbol      string `json:"symbol"`
	Tier        int    `json:"tier"`
	IsPreWarmed bool   `json:"is_pre_warmed"`
}

// UniverseResponse is the body of GET /v1/universe.
type UniverseResponse struct {
	AsOf     string           `json:"as_of"`
	Count    int              `json:"count"`
	Returned int              `json:"returned"`
	Limit    int              `json:"limit"`
	Sort     string           `json:"sort"`
	Symbols  []UniverseSymbol `json:"symbols"`
}

// UniverseOption configures a Universe request.
type UniverseOption func(*universeConfig)

type universeConfig struct {
	sort  string
	limit *int
}

// WithUniverseSort sets the sort order: "tier" (default) or "symbol".
func WithUniverseSort(sort string) UniverseOption {
	return func(c *universeConfig) { c.sort = sort }
}

// WithUniverseLimit caps the number of symbols returned (clamped 1–1000).
func WithUniverseLimit(limit int) UniverseOption {
	return func(c *universeConfig) { c.limit = &limit }
}

// Universe returns the curated tier-1 / tier-2 symbol directory the screener
// loop keeps pre-warmed. Public — no auth required. Optional: WithUniverseSort,
// WithUniverseLimit.
func (c *Client) Universe(ctx context.Context, opts ...UniverseOption) (map[string]interface{}, error) {
	cfg := &universeConfig{}
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
	return c.get(ctx, "/v1/universe", params)
}

// UniverseTyped is the strongly-typed variant of Universe.
func (c *Client) UniverseTyped(ctx context.Context, opts ...UniverseOption) (*UniverseResponse, error) {
	raw, err := c.Universe(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &UniverseResponse{}
	if err := decodeTyped("universe", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── surface/svi ──────────────────────────────────────────────────────────────

// SviSlice is one expiry's calibrated raw-SVI parameters.
type SviSlice struct {
	Expiry           string  `json:"expiry"`
	DaysToExpiry     int     `json:"days_to_expiry"`
	Forward          float64 `json:"forward"`
	A                float64 `json:"a"`
	B                float64 `json:"b"`
	Rho              float64 `json:"rho"`
	M                float64 `json:"m"`
	Sigma            float64 `json:"sigma"`
	AtmTotalVariance float64 `json:"atm_total_variance"`
	AtmIv            float64 `json:"atm_iv"`
}

// SurfaceSviResponse is the body of GET /v1/surface/svi/{symbol}.
type SurfaceSviResponse struct {
	Symbol          string     `json:"symbol"`
	UnderlyingPrice *float64   `json:"underlying_price"`
	AsOf            string     `json:"as_of"`
	MarketOpen      bool       `json:"market_open"`
	SviParameters   []SviSlice `json:"svi_parameters"`
}

// SurfaceSvi returns the live SVI-fitted volatility surface — calibrated
// (a, b, rho, m, sigma) parameters per expiry slice with ATM total variance
// and ATM IV. Requires Alpha+.
func (c *Client) SurfaceSvi(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/surface/svi/"+seg(symbol), nil)
}

// SurfaceSviTyped is the strongly-typed variant of SurfaceSvi.
func (c *Client) SurfaceSviTyped(ctx context.Context, symbol string) (*SurfaceSviResponse, error) {
	raw, err := c.SurfaceSvi(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &SurfaceSviResponse{}
	if err := decodeTyped("surface svi", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── expected-move ────────────────────────────────────────────────────────────

// ExpectedMoveExpiry is one expiry's straddle-implied expected move. NOTE:
// these item fields are camelCase on the wire (the top-level keys are
// snake_case).
type ExpectedMoveExpiry struct {
	Expiry          string   `json:"expiry"`
	DaysToExpiry    int      `json:"daysToExpiry"`
	AtmIv           *float64 `json:"atmIv"`
	ExpectedMove    float64  `json:"expectedMove"`
	ExpectedMovePct float64  `json:"expectedMovePct"`
	LowerBound      float64  `json:"lowerBound"`
	UpperBound      float64  `json:"upperBound"`
}

// ExpectedMoveResponse is the body of GET /v1/expected-move/{symbol}.
type ExpectedMoveResponse struct {
	Symbol          string               `json:"symbol"`
	UnderlyingPrice *float64             `json:"underlying_price"`
	AsOf            string               `json:"as_of"`
	ExpectedMoves   []ExpectedMoveExpiry `json:"expected_moves"`
}

// ExpectedMoveOption configures an ExpectedMove request.
type ExpectedMoveOption func(*expectedMoveConfig)

type expectedMoveConfig struct {
	expiry string
}

// WithExpectedMoveExpiry restricts the result to a single expiry (YYYY-MM-DD).
func WithExpectedMoveExpiry(expiry string) ExpectedMoveOption {
	return func(c *expectedMoveConfig) { c.expiry = expiry }
}

// ExpectedMove returns the straddle-implied expected move per expiry, derived
// from ATM implied volatility. Requires Basic+. Optional: WithExpectedMoveExpiry.
func (c *Client) ExpectedMove(ctx context.Context, symbol string, opts ...ExpectedMoveOption) (map[string]interface{}, error) {
	cfg := &expectedMoveConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/expected-move/"+seg(symbol), params)
}

// ExpectedMoveTyped is the strongly-typed variant of ExpectedMove.
func (c *Client) ExpectedMoveTyped(ctx context.Context, symbol string, opts ...ExpectedMoveOption) (*ExpectedMoveResponse, error) {
	raw, err := c.ExpectedMove(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ExpectedMoveResponse{}
	if err := decodeTyped("expected-move", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── vrp/{symbol}/history ─────────────────────────────────────────────────────

// VrpHistoryPoint is one daily VRP time-series point.
type VrpHistoryPoint struct {
	Date           string   `json:"date"`
	Spot           *float64 `json:"spot"`
	AtmIv          *float64 `json:"atm_iv"`
	Rv5d           *float64 `json:"rv_5d"`
	Rv10d          *float64 `json:"rv_10d"`
	Rv20d          *float64 `json:"rv_20d"`
	Rv30d          *float64 `json:"rv_30d"`
	Vrp20d         *float64 `json:"vrp_20d"`
	Straddle       *float64 `json:"straddle"`
	ExpectedMove1d *float64 `json:"expected_move_1d"`
}

// VrpHistoryResponse is the body of GET /v1/vrp/{symbol}/history.
type VrpHistoryResponse struct {
	Symbol     string            `json:"symbol"`
	Days       int               `json:"days"`
	DataPoints int               `json:"data_points"`
	History    []VrpHistoryPoint `json:"history"`
}

// VrpHistoryOption configures a VrpHistory request.
type VrpHistoryOption func(*vrpHistoryConfig)

type vrpHistoryConfig struct {
	days *int
}

// WithVrpHistoryDays sets the lookback window in days (1–365; default 30).
func WithVrpHistoryDays(days int) VrpHistoryOption {
	return func(c *vrpHistoryConfig) { c.days = &days }
}

// VrpHistory returns the daily VRP time series for charting/backtesting.
// Requires Alpha+. Optional: WithVrpHistoryDays.
func (c *Client) VrpHistory(ctx context.Context, symbol string, opts ...VrpHistoryOption) (map[string]interface{}, error) {
	cfg := &vrpHistoryConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.days != nil {
		params.Set("days", strconv.Itoa(*cfg.days))
	}
	return c.get(ctx, "/v1/vrp/"+seg(symbol)+"/history", params)
}

// VrpHistoryTyped is the strongly-typed variant of VrpHistory.
func (c *Client) VrpHistoryTyped(ctx context.Context, symbol string, opts ...VrpHistoryOption) (*VrpHistoryResponse, error) {
	raw, err := c.VrpHistory(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &VrpHistoryResponse{}
	if err := decodeTyped("vrp history", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
