// Package flashalpha provides a Go client for the FlashAlpha options analytics API.
//
// FlashAlpha delivers institutional-grade options analytics including gamma exposure
// (GEX), delta exposure (DEX), vanna/charm exposure, volatility surfaces, 0DTE
// analytics, and BSM pricing utilities.
//
// See https://flashalpha.com for API documentation and subscription plans.
package flashalpha

import (
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

// Health checks whether the API is operational. This endpoint is public and does
// not require authentication.
func (c *Client) Health(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/health", nil)
}
