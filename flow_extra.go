package flashalpha

// Typed models + client methods for the additional flow surfaces:
//
//   - FlowDealerPremium     — GET /v1/flow/options/{symbol}/dealer-premium   (Alpha+)
//   - FlowStockBars         — GET /v1/flow/stocks/{symbol}/bars              (Alpha+)
//   - Zero-DTE Flow group   — GET /v1/flow/zero-dte/{snapshot,series,
//                              hedge-flow,heatmap,strike-flow}/{symbol}       (Growth+/Alpha+)
//
// Every untyped method returns map[string]interface{}; the *Typed wrappers
// decode into the structs below. The intraday 0DTE-flow series share a small
// set of options (bar size / lookback minutes / side / metric / mode).

import (
	"context"
	"net/url"
	"strconv"
)

// ── flow/options/{symbol}/dealer-premium ─────────────────────────────────────

// FlowDealerPremiumResponse is the body of
// GET /v1/flow/options/{symbol}/dealer-premium.
type FlowDealerPremiumResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol             string  `json:"symbol"`
	AsOf               string  `json:"as_of"`
	WindowMinutes      int     `json:"window_minutes"`
	Expiry             *string `json:"expiry"`
	DealerBuyPremium   float64 `json:"dealer_buy_premium"`
	DealerWritePremium float64 `json:"dealer_write_premium"`
	NetDealerPremium   float64 `json:"net_dealer_premium"`
	TotalPremium       float64 `json:"total_premium"`
	TradeCount         int     `json:"trade_count"`
	BucketCount        int     `json:"bucket_count"`
}

// FlowDealerPremium returns the full-tape Net Dealer Premium roll-up over a
// configurable window (VWAP-weighted per minute bucket). Requires Alpha+.
// Optional: WithFlowWindowMinutes, WithFlowExpiry.
func (c *Client) FlowDealerPremium(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.windowMinutes != nil {
		params.Set("windowMinutes", strconv.Itoa(*cfg.windowMinutes))
	}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/options/"+seg(symbol)+"/dealer-premium", params)
}

// FlowDealerPremiumTyped is the strongly-typed variant of FlowDealerPremium.
func (c *Client) FlowDealerPremiumTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowDealerPremiumResponse, error) {
	raw, err := c.FlowDealerPremium(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowDealerPremiumResponse{}
	if err := decodeTyped("flow dealer-premium", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── flow/stocks/{symbol}/bars ────────────────────────────────────────────────

// FlowStockBar is one OHLCV+flow bar from the live trade tape.
type FlowStockBar struct {
	Ts           string   `json:"ts"`
	Closed       bool     `json:"closed"`
	Open         *float64 `json:"open"`
	High         *float64 `json:"high"`
	Low          *float64 `json:"low"`
	Close        *float64 `json:"close"`
	Vwap         *float64 `json:"vwap"`
	BuyVolume    float64  `json:"buyVolume"`
	SellVolume   float64  `json:"sellVolume"`
	MidVolume    float64  `json:"midVolume"`
	NetVolume    float64  `json:"netVolume"`
	TradeCount   int      `json:"tradeCount"`
	BiggestTrade *float64 `json:"biggestTrade"`
}

// FlowStockBarsResponse is the body of GET /v1/flow/stocks/{symbol}/bars.
type FlowStockBarsResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol       string         `json:"symbol"`
	Resolution   string         `json:"resolution"`
	Minutes      int            `json:"minutes"`
	Count        int            `json:"count"`
	DataStartUtc *string        `json:"dataStartUtc"`
	Bars         []FlowStockBar `json:"bars"`
}

// BarsOption configures a FlowStockBars request.
type BarsOption func(*barsConfig)

type barsConfig struct {
	minutes *int
}

// WithBarsMinutes sets the look-back window in minutes (clamped 1–1440).
func WithBarsMinutes(minutes int) BarsOption {
	return func(c *barsConfig) { c.minutes = &minutes }
}

// FlowStockBars returns multi-resolution OHLCV+flow bars (oldest-first) from
// the live trade tape. Requires Alpha+. resolution is required: one of
// 1s / 1m / 5m / 15m / 30m / 1h / 4h. Optional: WithBarsMinutes.
func (c *Client) FlowStockBars(ctx context.Context, symbol, resolution string, opts ...BarsOption) (map[string]interface{}, error) {
	cfg := &barsConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	params.Set("resolution", resolution)
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	return c.get(ctx, "/v1/flow/stocks/"+seg(symbol)+"/bars", params)
}

// FlowStockBarsTyped is the strongly-typed variant of FlowStockBars.
func (c *Client) FlowStockBarsTyped(ctx context.Context, symbol, resolution string, opts ...BarsOption) (*FlowStockBarsResponse, error) {
	raw, err := c.FlowStockBars(ctx, symbol, resolution, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowStockBarsResponse{}
	if err := decodeTyped("flow stock bars", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── Zero-DTE Flow group ──────────────────────────────────────────────────────

// ZeroDteFlowOption configures the intraday 0DTE-flow series endpoints
// (series / hedge-flow / heatmap / strike-flow). Each endpoint reads only the
// parameters it supports; unsupported options are ignored.
type ZeroDteFlowOption func(*zeroDteFlowConfig)

type zeroDteFlowConfig struct {
	bar     string
	minutes *int
	side    string
	metric  string
	mode    string
	n       *int
	expiry  string
}

// WithZeroDteFlowBar sets the bar size. /series and /hedge-flow accept
// 30s / 1m / 5m / 15m; /heatmap and /strike-flow accept only 1m.
func WithZeroDteFlowBar(bar string) ZeroDteFlowOption {
	return func(c *zeroDteFlowConfig) { c.bar = bar }
}

// WithZeroDteFlowMinutes sets the lookback window in minutes (clamped 1–390).
func WithZeroDteFlowMinutes(minutes int) ZeroDteFlowOption {
	return func(c *zeroDteFlowConfig) { c.minutes = &minutes }
}

// WithZeroDteFlowSide projects the hedge-flow leg: "all", "calls", or "puts".
func WithZeroDteFlowSide(side string) ZeroDteFlowOption {
	return func(c *zeroDteFlowConfig) { c.side = side }
}

// WithZeroDteFlowMetric sets the heatmap per-strike metric: one of
// gex / dex / vex / chex / oi / signed_flow.
func WithZeroDteFlowMetric(metric string) ZeroDteFlowOption {
	return func(c *zeroDteFlowConfig) { c.metric = metric }
}

// WithZeroDteFlowMode sets the heatmap mode: "raw" or "delta".
func WithZeroDteFlowMode(mode string) ZeroDteFlowOption {
	return func(c *zeroDteFlowConfig) { c.mode = mode }
}

// WithZeroDteFlowN sets the number of ranked rows (1–100) for the leaderboard.
func WithZeroDteFlowN(n int) ZeroDteFlowOption {
	return func(c *zeroDteFlowConfig) { c.n = &n }
}

// WithZeroDteFlowExpiry pins the snapshot to a specific expiry (YYYY-MM-DD).
func WithZeroDteFlowExpiry(expiry string) ZeroDteFlowOption {
	return func(c *zeroDteFlowConfig) { c.expiry = expiry }
}

func zeroDteFlowConfigFrom(opts []ZeroDteFlowOption) *zeroDteFlowConfig {
	cfg := &zeroDteFlowConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// FlowZeroDteSnapshot returns the current live 0DTE shape — the same fields as
// ZeroDte plus a flow_direction block (settled vs live net GEX, flow GEX
// adjustment, amplifying/dampening label). Requires Growth+. Untyped because it
// returns the full 0DTE payload; access flow_direction via the map. Optional:
// WithZeroDteFlowExpiry.
func (c *Client) FlowZeroDteSnapshot(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (map[string]interface{}, error) {
	cfg := zeroDteFlowConfigFrom(opts)
	params := url.Values{}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/zero-dte/snapshot/"+seg(symbol), params)
}

// FlowZeroDteSnapshotTyped decodes the snapshot into a *FlowZeroDteSnapshotResponse.
// The full 0DTE payload is preserved in Raw; FlowDirection is surfaced typed.
// Optional: WithZeroDteFlowExpiry.
func (c *Client) FlowZeroDteSnapshotTyped(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (*FlowZeroDteSnapshotResponse, error) {
	raw, err := c.FlowZeroDteSnapshot(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowZeroDteSnapshotResponse{}
	if err := decodeTyped("flow zero-dte snapshot", raw, out); err != nil {
		return nil, err
	}
	out.Raw = raw
	return out, nil
}

// FlowDirectionBlock is the flow_direction block on the 0DTE-flow snapshot.
type FlowDirectionBlock struct {
	// Label is no_flow / neutral / amplifying / dampening / regime_flip.
	Label                  string   `json:"label"`
	SettledNetGex          *float64 `json:"settled_net_gex"`
	LiveNetGex             *float64 `json:"live_net_gex"`
	FlowGexAdjustment      *float64 `json:"flow_gex_adjustment"`
	FlowGexPctShift        *float64 `json:"flow_gex_pct_shift"`
	ContractsWithFlow      int      `json:"contracts_with_flow"`
	TotalAbsDeltaContracts float64  `json:"total_abs_delta_contracts"`
	Description            string   `json:"description"`
}

// FlowZeroDteSnapshotResponse is the typed envelope of the 0DTE-flow snapshot.
// Only the headline + flow_direction fields are modeled; Raw holds the full
// ZeroDte payload (regime, exposures, pin_risk, hedging, decay, levels, ...).
type FlowZeroDteSnapshotResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol          string                 `json:"symbol"`
	UnderlyingPrice *float64               `json:"underlying_price"`
	AsOf            string                 `json:"as_of"`
	MarketOpen      bool                   `json:"market_open"`
	NoZeroDte       bool                   `json:"no_zero_dte"`
	SessionClosed   bool                   `json:"session_closed"`
	FlowDirection   *FlowDirectionBlock    `json:"flow_direction"`
	Raw             map[string]interface{} `json:"-"`
}

// ZeroDteSeriesBar is one intraday 0DTE-flow series bar.
type ZeroDteSeriesBar struct {
	T                       string   `json:"t"`
	Spot                    *float64 `json:"spot"`
	NetGex                  *float64 `json:"net_gex"`
	NetDex                  *float64 `json:"net_dex"`
	GammaFlip               *float64 `json:"gamma_flip"`
	CallWall                *float64 `json:"call_wall"`
	PutWall                 *float64 `json:"put_wall"`
	Magnet                  *float64 `json:"magnet"`
	PinScore                *int     `json:"pin_score"`
	PinProbabilityPct       *float64 `json:"pin_probability_pct"`
	Regime                  string   `json:"regime"`
	AtmIv                   *float64 `json:"atm_iv"`
	CharmDollarsPerHour     *float64 `json:"charm_dollars_per_hour"`
	HedgeFlowCallCumulative *float64 `json:"hedge_flow_call_cumulative"`
	HedgeFlowPutCumulative  *float64 `json:"hedge_flow_put_cumulative"`
	HedgeFlowCumulativeAll  *float64 `json:"hedge_flow_cumulative_all"`
}

// ZeroDteSeriesResponse is the body of GET /v1/flow/zero-dte/series/{symbol}.
type ZeroDteSeriesResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol     string             `json:"symbol"`
	Expiration string             `json:"expiration"`
	AsOf       string             `json:"as_of"`
	BarSize    string             `json:"bar_size"`
	Bars       []ZeroDteSeriesBar `json:"bars"`
}

// FlowZeroDteSeries returns an intraday time series of today's 0DTE flow — one
// bar per sampled interval (net GEX/DEX, gamma flip, walls, magnet, pin score,
// regime, ATM IV, charm, cumulative hedge-flow). Requires Growth+. Optional:
// WithZeroDteFlowBar (30s/1m/5m/15m), WithZeroDteFlowMinutes.
func (c *Client) FlowZeroDteSeries(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (map[string]interface{}, error) {
	cfg := zeroDteFlowConfigFrom(opts)
	params := url.Values{}
	if cfg.bar != "" {
		params.Set("bar", cfg.bar)
	}
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	return c.get(ctx, "/v1/flow/zero-dte/series/"+seg(symbol), params)
}

// FlowZeroDteSeriesTyped is the strongly-typed variant of FlowZeroDteSeries.
func (c *Client) FlowZeroDteSeriesTyped(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (*ZeroDteSeriesResponse, error) {
	raw, err := c.FlowZeroDteSeries(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ZeroDteSeriesResponse{}
	if err := decodeTyped("flow zero-dte series", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ZeroDteHedgeFlowBar is one bar of dealer hedge-flow.
type ZeroDteHedgeFlowBar struct {
	T          string  `json:"t"`
	Bar        float64 `json:"bar"`
	Cumulative float64 `json:"cumulative"`
}

// ZeroDteHedgeFlowResponse is the body of GET /v1/flow/zero-dte/hedge-flow/{symbol}.
type ZeroDteHedgeFlowResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol     string                `json:"symbol"`
	Expiration string                `json:"expiration"`
	AsOf       string                `json:"as_of"`
	Side       string                `json:"side"`
	BarSize    string                `json:"bar_size"`
	Bars       []ZeroDteHedgeFlowBar `json:"bars"`
}

// FlowZeroDteHedgeFlow returns the estimated dealer hedge-flow time series for
// today's 0DTE chain — signed delta-dollars, per-bar and running cumulative —
// projectable to calls / puts / all. Requires Growth+. Optional:
// WithZeroDteFlowSide, WithZeroDteFlowBar, WithZeroDteFlowMinutes.
func (c *Client) FlowZeroDteHedgeFlow(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (map[string]interface{}, error) {
	cfg := zeroDteFlowConfigFrom(opts)
	params := url.Values{}
	if cfg.side != "" {
		params.Set("side", cfg.side)
	}
	if cfg.bar != "" {
		params.Set("bar", cfg.bar)
	}
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	return c.get(ctx, "/v1/flow/zero-dte/hedge-flow/"+seg(symbol), params)
}

// FlowZeroDteHedgeFlowTyped is the strongly-typed variant of FlowZeroDteHedgeFlow.
func (c *Client) FlowZeroDteHedgeFlowTyped(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (*ZeroDteHedgeFlowResponse, error) {
	raw, err := c.FlowZeroDteHedgeFlow(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ZeroDteHedgeFlowResponse{}
	if err := decodeTyped("flow zero-dte hedge-flow", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ZeroDteHeatmapBar is one column of the strike × time heatmap. Values is
// index-aligned to StrikesGrid on the parent response.
type ZeroDteHeatmapBar struct {
	T      string    `json:"t"`
	Spot   *float64  `json:"spot"`
	Values []float64 `json:"values"`
}

// ZeroDteHeatmapResponse is the body of GET /v1/flow/zero-dte/heatmap/{symbol}.
type ZeroDteHeatmapResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol          string              `json:"symbol"`
	UnderlyingPrice *float64            `json:"underlying_price"`
	Expiration      string              `json:"expiration"`
	Metric          string              `json:"metric"`
	Mode            string              `json:"mode"`
	BarSize         string              `json:"bar_size"`
	AsOf            string              `json:"as_of"`
	TierUsed        string              `json:"tier_used"`
	StrikesGrid     []float64           `json:"strikes_grid"`
	Bars            []ZeroDteHeatmapBar `json:"bars"`
	GapIntervals    []interface{}       `json:"gap_intervals"`
}

// FlowZeroDteHeatmap returns the per-strike value matrix over today's 0DTE
// session — the data behind a strike × time heatmap. Requires Alpha+. Optional:
// WithZeroDteFlowMetric (gex/dex/vex/chex/oi/signed_flow), WithZeroDteFlowMode
// (raw/delta), WithZeroDteFlowBar (1m only), WithZeroDteFlowMinutes.
func (c *Client) FlowZeroDteHeatmap(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (map[string]interface{}, error) {
	cfg := zeroDteFlowConfigFrom(opts)
	params := url.Values{}
	if cfg.metric != "" {
		params.Set("metric", cfg.metric)
	}
	if cfg.mode != "" {
		params.Set("mode", cfg.mode)
	}
	if cfg.bar != "" {
		params.Set("bar", cfg.bar)
	}
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	return c.get(ctx, "/v1/flow/zero-dte/heatmap/"+seg(symbol), params)
}

// FlowZeroDteHeatmapTyped is the strongly-typed variant of FlowZeroDteHeatmap.
func (c *Client) FlowZeroDteHeatmapTyped(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (*ZeroDteHeatmapResponse, error) {
	raw, err := c.FlowZeroDteHeatmap(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ZeroDteHeatmapResponse{}
	if err := decodeTyped("flow zero-dte heatmap", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ZeroDteStrikeFlowBar is one bar of per-strike signed aggressor flow. The
// three arrays are index-aligned to StrikesGrid on the parent response.
type ZeroDteStrikeFlowBar struct {
	T                  string    `json:"t"`
	Spot               *float64  `json:"spot"`
	SignedDeltaDollars []float64 `json:"signed_delta_dollars"`
	SignedGammaDollars []float64 `json:"signed_gamma_dollars"`
	Contracts          []float64 `json:"contracts"`
}

// ZeroDteStrikeFlowResponse is the body of GET /v1/flow/zero-dte/strike-flow/{symbol}.
type ZeroDteStrikeFlowResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol          string                 `json:"symbol"`
	UnderlyingPrice *float64               `json:"underlying_price"`
	Expiration      string                 `json:"expiration"`
	BarSize         string                 `json:"bar_size"`
	AsOf            string                 `json:"as_of"`
	TierUsed        string                 `json:"tier_used"`
	StrikesGrid     []float64              `json:"strikes_grid"`
	Bars            []ZeroDteStrikeFlowBar `json:"bars"`
	GapIntervals    []interface{}          `json:"gap_intervals"`
}

// FlowZeroDteStrikeFlow returns per-strike signed aggressor flow over today's
// 0DTE session — signed delta-dollars, signed gamma-dollars, and contract count
// per bar and strike. Requires Alpha+. Optional: WithZeroDteFlowBar (1m only),
// WithZeroDteFlowMinutes.
func (c *Client) FlowZeroDteStrikeFlow(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (map[string]interface{}, error) {
	cfg := zeroDteFlowConfigFrom(opts)
	params := url.Values{}
	if cfg.bar != "" {
		params.Set("bar", cfg.bar)
	}
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	return c.get(ctx, "/v1/flow/zero-dte/strike-flow/"+seg(symbol), params)
}

// FlowZeroDteStrikeFlowTyped is the strongly-typed variant of FlowZeroDteStrikeFlow.
func (c *Client) FlowZeroDteStrikeFlowTyped(ctx context.Context, symbol string, opts ...ZeroDteFlowOption) (*ZeroDteStrikeFlowResponse, error) {
	raw, err := c.FlowZeroDteStrikeFlow(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ZeroDteStrikeFlowResponse{}
	if err := decodeTyped("flow zero-dte strike-flow", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ZeroDteLeaderboardEntry is one ranked symbol on the 0DTE-flow leaderboard.
type ZeroDteLeaderboardEntry struct {
	Rank   int     `json:"rank"`
	Symbol string  `json:"symbol"`
	Value  float64 `json:"value"`
}

// ZeroDteLeaderboardResponse is the body of GET /v1/flow/zero-dte/leaderboard.
type ZeroDteLeaderboardResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Metric     string                    `json:"metric"`
	N          int                       `json:"n"`
	AsOf       string                    `json:"as_of"`
	MarketOpen bool                      `json:"market_open"`
	Entries    []ZeroDteLeaderboardEntry `json:"entries"`
}

// FlowZeroDteLeaderboard returns the cross-symbol 0DTE-flow leaderboard — the
// top-N symbols ranked by a chosen intraday metric. Cross-symbol, so it takes no
// symbol path segment. Requires Alpha+. Optional: WithZeroDteFlowMetric
// (heat/pin_risk/abs_flow/charm_intensity), WithZeroDteFlowN (1–100).
func (c *Client) FlowZeroDteLeaderboard(ctx context.Context, opts ...ZeroDteFlowOption) (map[string]interface{}, error) {
	cfg := zeroDteFlowConfigFrom(opts)
	params := url.Values{}
	if cfg.metric != "" {
		params.Set("metric", cfg.metric)
	}
	if cfg.n != nil {
		params.Set("n", strconv.Itoa(*cfg.n))
	}
	return c.get(ctx, "/v1/flow/zero-dte/leaderboard", params)
}

// FlowZeroDteLeaderboardTyped is the strongly-typed variant of FlowZeroDteLeaderboard.
func (c *Client) FlowZeroDteLeaderboardTyped(ctx context.Context, opts ...ZeroDteFlowOption) (*ZeroDteLeaderboardResponse, error) {
	raw, err := c.FlowZeroDteLeaderboard(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &ZeroDteLeaderboardResponse{}
	if err := decodeTyped("flow zero-dte leaderboard", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
