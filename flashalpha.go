// Package flashalpha provides a Go client for the FlashAlpha options analytics API.
//
// FlashAlpha delivers institutional-grade options analytics including gamma exposure
// (GEX), delta exposure (DEX), vanna/charm exposure, volatility surfaces, 0DTE
// analytics, and BSM pricing utilities.
//
// See https://flashalpha.com for API documentation and subscription plans.
package flashalpha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultBaseURL = "https://lab.flashalpha.com"
const defaultTimeout = 30 * time.Second

// Client is a thread-safe HTTP client for the FlashAlpha REST API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewClient creates a Client using the default production base URL.
func NewClient(apiKey string) *Client {
	return NewClientWithURL(apiKey, defaultBaseURL)
}

// NewClientWithURL creates a Client with a custom base URL (useful for testing).
func NewClientWithURL(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
}

// ── internal ─────────────────────────────────────────────────────────────────

func (c *Client) get(ctx context.Context, path string, params url.Values) (map[string]interface{}, error) {
	rawURL := c.baseURL + path
	if len(params) > 0 {
		rawURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("flashalpha: build request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flashalpha: request: %w", err)
	}
	defer resp.Body.Close()

	return c.handle(resp)
}

func (c *Client) post(ctx context.Context, path string, body interface{}) (map[string]interface{}, error) {
	var reader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("flashalpha: marshal body: %w", err)
		}
		reader = bytes.NewReader(jsonBody)
	}

	rawURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, reader)
	if err != nil {
		return nil, fmt.Errorf("flashalpha: build request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flashalpha: request: %w", err)
	}
	defer resp.Body.Close()

	return c.handle(resp)
}

func (c *Client) handle(resp *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("flashalpha: read body: %w", err)
	}

	var parsed map[string]interface{}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &parsed)
	}

	if resp.StatusCode == http.StatusOK {
		if parsed == nil {
			parsed = make(map[string]interface{})
		}
		return parsed, nil
	}

	msg := extractMessage(parsed, string(body))

	baseErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		Response:   parsed,
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, &AuthenticationError{APIError: baseErr}
	case http.StatusForbidden:
		te := &TierRestrictedError{APIError: baseErr}
		if parsed != nil {
			if v, ok := parsed["current_plan"].(string); ok {
				te.CurrentPlan = v
			}
			if v, ok := parsed["required_plan"].(string); ok {
				te.RequiredPlan = v
			}
		}
		return nil, te
	case http.StatusNotFound:
		return nil, &NotFoundError{APIError: baseErr}
	case http.StatusTooManyRequests:
		re := &RateLimitError{APIError: baseErr}
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if n, err := strconv.Atoi(ra); err == nil {
				re.RetryAfter = n
			}
		}
		return nil, re
	default:
		if resp.StatusCode >= 500 {
			return nil, &ServerError{APIError: baseErr}
		}
		return nil, baseErr
	}
}

func extractMessage(body map[string]interface{}, raw string) string {
	if body != nil {
		if v, ok := body["message"].(string); ok && v != "" {
			return v
		}
		if v, ok := body["detail"].(string); ok && v != "" {
			return v
		}
	}
	if raw != "" {
		return raw
	}
	return "unknown error"
}

// ── Functional option types ───────────────────────────────────────────────────

// GexOption configures a Gex request.
type GexOption func(*gexConfig)

type gexConfig struct {
	expiration string
	minOI      *int
}

// WithExpiration filters results to a specific option expiration date (YYYY-MM-DD).
func WithExpiration(exp string) GexOption {
	return func(c *gexConfig) { c.expiration = exp }
}

// WithMinOI filters strikes below a minimum open-interest threshold.
func WithMinOI(minOI int) GexOption {
	return func(c *gexConfig) { c.minOI = &minOI }
}

// DexOption configures a Dex request.
type DexOption func(*dexConfig)

type dexConfig struct {
	expiration string
}

// WithDexExpiration filters delta exposure results to a specific expiration date.
func WithDexExpiration(exp string) DexOption {
	return func(c *dexConfig) { c.expiration = exp }
}

// VexOption configures a Vex request.
type VexOption func(*vexConfig)

type vexConfig struct {
	expiration string
}

// WithVexExpiration filters vanna exposure results to a specific expiration date.
func WithVexExpiration(exp string) VexOption {
	return func(c *vexConfig) { c.expiration = exp }
}

// ChexOption configures a Chex request.
type ChexOption func(*chexConfig)

type chexConfig struct {
	expiration string
}

// WithChexExpiration filters charm exposure results to a specific expiration date.
func WithChexExpiration(exp string) ChexOption {
	return func(c *chexConfig) { c.expiration = exp }
}

// ZeroDteOption configures a ZeroDte request.
type ZeroDteOption func(*zeroDteConfig)

type zeroDteConfig struct {
	strikeRange *float64
}

// WithStrikeRange sets the strike range around the spot price for 0DTE analytics.
func WithStrikeRange(r float64) ZeroDteOption {
	return func(c *zeroDteConfig) { c.strikeRange = &r }
}

// HistoryOption configures an ExposureHistory request.
type HistoryOption func(*historyConfig)

type historyConfig struct {
	days *int
}

// WithDays sets the number of historical days to retrieve.
func WithDays(days int) HistoryOption {
	return func(c *historyConfig) { c.days = &days }
}

// OptionQuoteOption configures an OptionQuote request.
type OptionQuoteOption func(*optionQuoteConfig)

type optionQuoteConfig struct {
	expiry string
	strike *float64
	typ    string
}

// WithOptionExpiry filters option quotes by expiration date.
func WithOptionExpiry(expiry string) OptionQuoteOption {
	return func(c *optionQuoteConfig) { c.expiry = expiry }
}

// WithStrike filters option quotes by strike price.
func WithStrike(strike float64) OptionQuoteOption {
	return func(c *optionQuoteConfig) { c.strike = &strike }
}

// WithOptionType filters option quotes by type ("call" or "put").
func WithOptionType(typ string) OptionQuoteOption {
	return func(c *optionQuoteConfig) { c.typ = typ }
}

// HistOptOption configures a HistoricalOptionQuote request.
type HistOptOption func(*histOptConfig)

type histOptConfig struct {
	time   string
	expiry string
	strike *float64
	typ    string
}

// WithHistTime filters historical option quotes to a specific intraday time (HH:MM).
func WithHistTime(t string) HistOptOption {
	return func(c *histOptConfig) { c.time = t }
}

// WithHistExpiry filters historical option quotes by expiration date.
func WithHistExpiry(expiry string) HistOptOption {
	return func(c *histOptConfig) { c.expiry = expiry }
}

// WithHistStrike filters historical option quotes by strike price.
func WithHistStrike(strike float64) HistOptOption {
	return func(c *histOptConfig) { c.strike = &strike }
}

// WithHistType filters historical option quotes by type ("call" or "put").
func WithHistType(typ string) HistOptOption {
	return func(c *histOptConfig) { c.typ = typ }
}

// GreeksParams holds required and optional parameters for the Greeks endpoint.
type GreeksParams struct {
	Spot   float64
	Strike float64
	DTE    float64
	Sigma  float64
	Type   string  // "call" or "put"; defaults to "call"
	R      *float64
	Q      *float64
}

// IVParams holds required and optional parameters for the IV endpoint.
type IVParams struct {
	Spot   float64
	Strike float64
	DTE    float64
	Price  float64
	Type   string  // "call" or "put"; defaults to "call"
	R      *float64
	Q      *float64
}

// KellyParams holds required and optional parameters for the Kelly endpoint.
type KellyParams struct {
	Spot    float64
	Strike  float64
	DTE     float64
	Sigma   float64
	Premium float64
	Mu      float64
	Type    string  // "call" or "put"; defaults to "call"
	R       *float64
	Q       *float64
}

// ── Exposure Analytics ────────────────────────────────────────────────────────

// Gex returns gamma exposure by strike for the given symbol.
// Optional: WithExpiration, WithMinOI.
func (c *Client) Gex(ctx context.Context, symbol string, opts ...GexOption) (map[string]interface{}, error) {
	cfg := &gexConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.expiration != "" {
		params.Set("expiration", cfg.expiration)
	}
	if cfg.minOI != nil {
		params.Set("min_oi", strconv.Itoa(*cfg.minOI))
	}
	return c.get(ctx, "/v1/exposure/gex/"+symbol, params)
}

// Dex returns delta exposure by strike for the given symbol.
// Optional: WithDexExpiration.
func (c *Client) Dex(ctx context.Context, symbol string, opts ...DexOption) (map[string]interface{}, error) {
	cfg := &dexConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.expiration != "" {
		params.Set("expiration", cfg.expiration)
	}
	return c.get(ctx, "/v1/exposure/dex/"+symbol, params)
}

// Vex returns vanna exposure by strike for the given symbol.
// Optional: WithVexExpiration.
func (c *Client) Vex(ctx context.Context, symbol string, opts ...VexOption) (map[string]interface{}, error) {
	cfg := &vexConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.expiration != "" {
		params.Set("expiration", cfg.expiration)
	}
	return c.get(ctx, "/v1/exposure/vex/"+symbol, params)
}

// Chex returns charm exposure by strike for the given symbol.
// Optional: WithChexExpiration.
func (c *Client) Chex(ctx context.Context, symbol string, opts ...ChexOption) (map[string]interface{}, error) {
	cfg := &chexConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.expiration != "" {
		params.Set("expiration", cfg.expiration)
	}
	return c.get(ctx, "/v1/exposure/chex/"+symbol, params)
}

// ExposureLevels returns key support/resistance levels derived from options exposure.
func (c *Client) ExposureLevels(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/exposure/levels/"+symbol, nil)
}

// ExposureSummary returns a full exposure summary (GEX/DEX/VEX/CHEX + hedging).
// Requires Growth+ plan.
func (c *Client) ExposureSummary(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/exposure/summary/"+symbol, nil)
}

// Narrative returns a verbal narrative analysis of options exposure.
// Requires Growth+ plan.
func (c *Client) Narrative(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/exposure/narrative/"+symbol, nil)
}

// ZeroDte returns real-time 0DTE analytics: regime, expected move, pin risk, hedging, decay.
// Requires Growth+ plan. Optional: WithStrikeRange.
func (c *Client) ZeroDte(ctx context.Context, symbol string, opts ...ZeroDteOption) (map[string]interface{}, error) {
	cfg := &zeroDteConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.strikeRange != nil {
		params.Set("strike_range", strconv.FormatFloat(*cfg.strikeRange, 'f', -1, 64))
	}
	return c.get(ctx, "/v1/exposure/zero-dte/"+symbol, params)
}

// ExposureHistory returns daily exposure snapshots for trend analysis.
// Requires Growth+ plan. Optional: WithDays.
func (c *Client) ExposureHistory(ctx context.Context, symbol string, opts ...HistoryOption) (map[string]interface{}, error) {
	cfg := &historyConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.days != nil {
		params.Set("days", strconv.Itoa(*cfg.days))
	}
	return c.get(ctx, "/v1/exposure/history/"+symbol, params)
}

// ── Market Data ───────────────────────────────────────────────────────────────

// StockQuote returns a live stock quote (bid/ask/mid/last).
func (c *Client) StockQuote(ctx context.Context, ticker string) (map[string]interface{}, error) {
	return c.get(ctx, "/stockquote/"+ticker, nil)
}

// OptionQuote returns option quotes with greeks for the given ticker.
// Requires Growth+ plan. Optional: WithOptionExpiry, WithStrike, WithOptionType.
func (c *Client) OptionQuote(ctx context.Context, ticker string, opts ...OptionQuoteOption) (map[string]interface{}, error) {
	cfg := &optionQuoteConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	if cfg.strike != nil {
		params.Set("strike", strconv.FormatFloat(*cfg.strike, 'f', -1, 64))
	}
	if cfg.typ != "" {
		params.Set("type", cfg.typ)
	}
	return c.get(ctx, "/optionquote/"+ticker, params)
}

// StockSummary returns a comprehensive stock summary (price, vol, exposure, macro).
func (c *Client) StockSummary(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/stock/"+symbol+"/summary", nil)
}

// Surface returns the volatility surface grid. This endpoint is public and does not
// require authentication.
func (c *Client) Surface(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/surface/"+symbol, nil)
}

// ── Historical Data ───────────────────────────────────────────────────────────

// HistoricalStockQuote returns historical stock quotes (minute-by-minute).
// date must be in YYYY-MM-DD format. time is optional (HH:MM).
func (c *Client) HistoricalStockQuote(ctx context.Context, ticker, date string, timeStr ...string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("date", date)
	if len(timeStr) > 0 && timeStr[0] != "" {
		params.Set("time", timeStr[0])
	}
	return c.get(ctx, "/historical/stockquote/"+ticker, params)
}

// HistoricalOptionQuote returns historical option quotes (minute-by-minute).
// date must be in YYYY-MM-DD format. Optional: WithHistTime, WithHistExpiry,
// WithHistStrike, WithHistType.
func (c *Client) HistoricalOptionQuote(ctx context.Context, ticker, date string, opts ...HistOptOption) (map[string]interface{}, error) {
	cfg := &histOptConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	params.Set("date", date)
	if cfg.time != "" {
		params.Set("time", cfg.time)
	}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	if cfg.strike != nil {
		params.Set("strike", strconv.FormatFloat(*cfg.strike, 'f', -1, 64))
	}
	if cfg.typ != "" {
		params.Set("type", cfg.typ)
	}
	return c.get(ctx, "/historical/optionquote/"+ticker, params)
}

// ── Pricing & Sizing ──────────────────────────────────────────────────────────

// Greeks computes full BSM greeks (first, second, and third order) for an option.
func (c *Client) Greeks(ctx context.Context, p GreeksParams) (map[string]interface{}, error) {
	typ := p.Type
	if typ == "" {
		typ = "call"
	}
	params := url.Values{}
	params.Set("spot", strconv.FormatFloat(p.Spot, 'f', -1, 64))
	params.Set("strike", strconv.FormatFloat(p.Strike, 'f', -1, 64))
	params.Set("dte", strconv.FormatFloat(p.DTE, 'f', -1, 64))
	params.Set("sigma", strconv.FormatFloat(p.Sigma, 'f', -1, 64))
	params.Set("type", typ)
	if p.R != nil {
		params.Set("r", strconv.FormatFloat(*p.R, 'f', -1, 64))
	}
	if p.Q != nil {
		params.Set("q", strconv.FormatFloat(*p.Q, 'f', -1, 64))
	}
	return c.get(ctx, "/v1/pricing/greeks", params)
}

// IV computes implied volatility from a market price using the BSM model.
func (c *Client) IV(ctx context.Context, p IVParams) (map[string]interface{}, error) {
	typ := p.Type
	if typ == "" {
		typ = "call"
	}
	params := url.Values{}
	params.Set("spot", strconv.FormatFloat(p.Spot, 'f', -1, 64))
	params.Set("strike", strconv.FormatFloat(p.Strike, 'f', -1, 64))
	params.Set("dte", strconv.FormatFloat(p.DTE, 'f', -1, 64))
	params.Set("price", strconv.FormatFloat(p.Price, 'f', -1, 64))
	params.Set("type", typ)
	if p.R != nil {
		params.Set("r", strconv.FormatFloat(*p.R, 'f', -1, 64))
	}
	if p.Q != nil {
		params.Set("q", strconv.FormatFloat(*p.Q, 'f', -1, 64))
	}
	return c.get(ctx, "/v1/pricing/iv", params)
}

// Kelly computes the Kelly criterion optimal position size for an option trade.
// Requires Growth+ plan.
func (c *Client) Kelly(ctx context.Context, p KellyParams) (map[string]interface{}, error) {
	typ := p.Type
	if typ == "" {
		typ = "call"
	}
	params := url.Values{}
	params.Set("spot", strconv.FormatFloat(p.Spot, 'f', -1, 64))
	params.Set("strike", strconv.FormatFloat(p.Strike, 'f', -1, 64))
	params.Set("dte", strconv.FormatFloat(p.DTE, 'f', -1, 64))
	params.Set("sigma", strconv.FormatFloat(p.Sigma, 'f', -1, 64))
	params.Set("premium", strconv.FormatFloat(p.Premium, 'f', -1, 64))
	params.Set("mu", strconv.FormatFloat(p.Mu, 'f', -1, 64))
	params.Set("type", typ)
	if p.R != nil {
		params.Set("r", strconv.FormatFloat(*p.R, 'f', -1, 64))
	}
	if p.Q != nil {
		params.Set("q", strconv.FormatFloat(*p.Q, 'f', -1, 64))
	}
	return c.get(ctx, "/v1/pricing/kelly", params)
}

// ── Volatility Analytics ──────────────────────────────────────────────────────

// Volatility returns comprehensive volatility analysis for the given symbol.
// Requires Growth+ plan.
func (c *Client) Volatility(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/volatility/"+symbol, nil)
}

// AdvVolatility returns advanced volatility analytics: SVI parameters, variance
// surface, arbitrage detection, greeks surfaces, and variance swap pricing.
// Requires Alpha+ plan.
func (c *Client) AdvVolatility(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/adv_volatility/"+symbol, nil)
}

// ── Reference Data ────────────────────────────────────────────────────────────

// Tickers returns all available stock tickers.
func (c *Client) Tickers(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/tickers", nil)
}

// Options returns option chain metadata (expirations and strikes) for the ticker.
func (c *Client) Options(ctx context.Context, ticker string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/options/"+ticker, nil)
}

// Symbols returns all currently queried symbols with live data.
func (c *Client) Symbols(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/symbols", nil)
}

// ── Account & System ──────────────────────────────────────────────────────────

// Account returns account information and quota usage.
func (c *Client) Account(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/account", nil)
}

// MaxPainOption configures the MaxPain request.
type MaxPainOption func(*maxPainConfig)

type maxPainConfig struct {
	expiration string
}

// WithMaxPainExpiration filters max pain to a single expiry (YYYY-MM-DD).
func WithMaxPainExpiration(exp string) MaxPainOption {
	return func(c *maxPainConfig) { c.expiration = exp }
}

// MaxPain returns max pain analysis with dealer alignment, pain curve, OI
// breakdown, expected move, pin probability, and multi-expiry calendar.
// Requires Growth+ plan.
func (c *Client) MaxPain(ctx context.Context, symbol string, opts ...MaxPainOption) (map[string]interface{}, error) {
	cfg := &maxPainConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.expiration != "" {
		params.Set("expiration", cfg.expiration)
	}
	return c.get(ctx, "/v1/maxpain/"+symbol, params)
}

// ScreenerRequest is the request body for the live options screener. All fields
// are optional — an empty request returns the default universe for your tier.
//
// See https://flashalpha.com/docs/lab-api-screener for the full field reference.
type ScreenerRequest struct {
	Filters  interface{}       `json:"filters,omitempty"`
	Sort     []ScreenerSort    `json:"sort,omitempty"`
	Select   []string          `json:"select,omitempty"`
	Formulas []ScreenerFormula `json:"formulas,omitempty"`
	Limit    *int              `json:"limit,omitempty"`
	Offset   *int              `json:"offset,omitempty"`
}

// ScreenerLeaf is a leaf filter condition. Use Field xor Formula.
type ScreenerLeaf struct {
	Field    string      `json:"field,omitempty"`
	Formula  string      `json:"formula,omitempty"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value,omitempty"`
}

// ScreenerGroup is an AND/OR filter group.
type ScreenerGroup struct {
	Op         string        `json:"op"`
	Conditions []interface{} `json:"conditions"`
}

// ScreenerSort is a sort spec. Use Field xor Formula. Direction is "asc" or "desc".
type ScreenerSort struct {
	Field     string `json:"field,omitempty"`
	Formula   string `json:"formula,omitempty"`
	Direction string `json:"direction"`
}

// ScreenerFormula defines a computed field (Alpha tier only).
type ScreenerFormula struct {
	Alias      string `json:"alias"`
	Expression string `json:"expression"`
}

// Screener runs the live options screener — filter and rank symbols by gamma
// exposure, VRP, volatility, greeks, and more. Powered by an in-memory store
// updated every 5-10s from live market data.
//
// Growth: 10-symbol universe, up to 10 rows. Alpha: ~250 symbols, up to 50
// rows, formulas, and harvest/dealer-flow-risk scores.
//
// Example:
//
//	limit := 20
//	req := flashalpha.ScreenerRequest{
//	    Filters: flashalpha.ScreenerGroup{
//	        Op: "and",
//	        Conditions: []interface{}{
//	            flashalpha.ScreenerLeaf{Field: "regime", Operator: "eq", Value: "positive_gamma"},
//	            flashalpha.ScreenerLeaf{Field: "harvest_score", Operator: "gte", Value: 65},
//	        },
//	    },
//	    Sort:   []flashalpha.ScreenerSort{{Field: "harvest_score", Direction: "desc"}},
//	    Select: []string{"symbol", "price", "harvest_score", "dealer_flow_risk"},
//	    Limit:  &limit,
//	}
//	result, err := client.Screener(ctx, req)
func (c *Client) Screener(ctx context.Context, req ScreenerRequest) (map[string]interface{}, error) {
	return c.post(ctx, "/v1/screener", req)
}

// ScreenerRaw runs the live screener with a raw request body (map, struct, or
// any JSON-serializable value). Use this when you need full control over the
// payload shape.
func (c *Client) ScreenerRaw(ctx context.Context, body interface{}) (map[string]interface{}, error) {
	return c.post(ctx, "/v1/screener", body)
}

// Health checks whether the API is operational. This endpoint is public and does
// not require authentication.
func (c *Client) Health(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/health", nil)
}

// ── VRP (Variance Risk Premium) ───────────────────────────────────────────────

// VrpCore holds the core variance-risk-premium metrics under response.vrp.
//
// Customers must access these via the nested path (e.g. resp.Vrp.ZScore) —
// the fields are NOT promoted to the top level.
type VrpCore struct {
	AtmIv       *float64 `json:"atm_iv"`
	Rv5d        *float64 `json:"rv_5d"`
	Rv10d       *float64 `json:"rv_10d"`
	Rv20d       *float64 `json:"rv_20d"`
	Rv30d       *float64 `json:"rv_30d"`
	Vrp5d       *float64 `json:"vrp_5d"`
	Vrp10d      *float64 `json:"vrp_10d"`
	Vrp20d      *float64 `json:"vrp_20d"`
	Vrp30d      *float64 `json:"vrp_30d"`
	ZScore      *float64 `json:"z_score"`
	Percentile  *int     `json:"percentile"`
	HistoryDays int      `json:"history_days"`
}

// VrpDirectional holds put/call wing IV and downside/upside skew metrics.
//
// Canonical names are DownsideVrp / UpsideVrp — NOT put_vrp / call_vrp
// (a common naming trap from other APIs).
type VrpDirectional struct {
	PutWingIv25d  *float64 `json:"put_wing_iv_25d"`
	CallWingIv25d *float64 `json:"call_wing_iv_25d"`
	DownsideRv20d *float64 `json:"downside_rv_20d"`
	UpsideRv20d   *float64 `json:"upside_rv_20d"`
	DownsideVrp   *float64 `json:"downside_vrp"`
	UpsideVrp     *float64 `json:"upside_vrp"`
}

// VrpTermPoint is one point in the term-structure curve (response.term_vrp[]).
type VrpTermPoint struct {
	Dte int     `json:"dte"`
	Iv  float64 `json:"iv"`
	Rv  float64 `json:"rv"`
	Vrp float64 `json:"vrp"`
}

// VrpGexConditioned holds gamma-regime-conditioned harvest scoring under
// response.gex_conditioned (nullable).
type VrpGexConditioned struct {
	Regime         string  `json:"regime"`
	HarvestScore   float64 `json:"harvest_score"`
	Interpretation string  `json:"interpretation"`
}

// VrpVannaConditioned holds vanna-regime-conditioned outlook under
// response.vanna_conditioned (nullable).
type VrpVannaConditioned struct {
	Outlook        string `json:"outlook"`
	Interpretation string `json:"interpretation"`
}

// VrpRegime is the regime snapshot under response.regime. NetGex and
// GammaFlip live HERE — not at the top level (a customer trap).
type VrpRegime struct {
	Gamma     string   `json:"gamma"`
	VrpRegime *string  `json:"vrp_regime"`
	NetGex    float64  `json:"net_gex"`
	GammaFlip *float64 `json:"gamma_flip"`
}

// VrpStrategyScores holds 0–100 scores per strategy under
// response.strategy_scores (nullable).
type VrpStrategyScores struct {
	ShortPutSpread *int `json:"short_put_spread"`
	ShortStrangle  *int `json:"short_strangle"`
	IronCondor     *int `json:"iron_condor"`
	CalendarSpread *int `json:"calendar_spread"`
}

// VrpMacro is the macro context block under response.macro (nullable).
type VrpMacro struct {
	Vix          *float64 `json:"vix"`
	Vix3m        *float64 `json:"vix_3m"`
	VixTermSlope *float64 `json:"vix_term_slope"`
	Dgs10        *float64 `json:"dgs10"`
	HySpread     *float64 `json:"hy_spread"`
	FedFunds     *float64 `json:"fed_funds"`
}

// VrpResponse is the full payload from GET /v1/vrp/{symbol}.
//
// Nested access paths (these are the canonical, only paths — fields are NOT
// duplicated at the top level):
//
//   - resp.Vrp.ZScore, .Percentile, .AtmIv, .Rv20d, .Vrp20d (core metrics)
//   - resp.Directional.DownsideVrp, .UpsideVrp (NOT put_vrp/call_vrp)
//   - resp.Regime.NetGex, .Gamma, .GammaFlip (NOT top-level)
//   - resp.GexConditioned.HarvestScore, .Regime, .Interpretation (nullable)
//   - resp.StrategyScores.ShortPutSpread, .ShortStrangle, ... (nullable)
//   - resp.TermVrp[i].Dte, .Iv, .Rv, .Vrp
//   - resp.Macro.Vix, .Vix3m, ... (nullable)
//
// Top-level composite scores (the only top-level VRP scalars):
// resp.NetHarvestScore and resp.DealerFlowRisk.
//
// Raw holds the underlying decoded JSON for any field not modeled above.
type VrpResponse struct {
	Symbol           string                 `json:"symbol"`
	UnderlyingPrice  float64                `json:"underlying_price"`
	AsOf             string                 `json:"as_of"`
	MarketOpen       bool                   `json:"market_open"`
	Vrp              VrpCore                `json:"vrp"`
	VarianceRiskPrem *float64               `json:"variance_risk_premium"`
	ConvexityPremium *float64               `json:"convexity_premium"`
	FairVol          *float64               `json:"fair_vol"`
	Directional      VrpDirectional         `json:"directional"`
	TermVrp          []VrpTermPoint         `json:"term_vrp"`
	GexConditioned   *VrpGexConditioned     `json:"gex_conditioned"`
	VannaConditioned *VrpVannaConditioned   `json:"vanna_conditioned"`
	Regime           VrpRegime              `json:"regime"`
	StrategyScores   *VrpStrategyScores     `json:"strategy_scores"`
	NetHarvestScore  *float64               `json:"net_harvest_score"`
	DealerFlowRisk   *float64               `json:"dealer_flow_risk"`
	Warnings         []string               `json:"warnings"`
	Macro            *VrpMacro              `json:"macro"`
	Raw              map[string]interface{} `json:"-"`
}

// Vrp returns variance-risk-premium analytics — the implied-vs-realized vol
// spread, conditioned on dealer gamma and vanna regime, plus strategy scores
// for harvesting. Requires Alpha+ plan.
//
// The response is nested. Common access paths:
//
//   - resp.Vrp.ZScore           (NOT a top-level field)
//   - resp.Regime.NetGex        (NOT resp.NetGex)
//   - resp.GexConditioned.HarvestScore  (NOT resp.HarvestScore)
//   - resp.Directional.DownsideVrp / .UpsideVrp  (NOT put_vrp/call_vrp)
//
// Top-level composite scores: resp.NetHarvestScore, resp.DealerFlowRisk.
func (c *Client) Vrp(ctx context.Context, symbol string) (*VrpResponse, error) {
	raw, err := c.get(ctx, "/v1/vrp/"+symbol, nil)
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("flashalpha: re-encode vrp: %w", err)
	}
	out := &VrpResponse{}
	if err := json.Unmarshal(buf, out); err != nil {
		return nil, fmt.Errorf("flashalpha: decode vrp: %w", err)
	}
	out.Raw = raw
	return out, nil
}
