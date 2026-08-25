package flashalpha

// Typed models + client methods for the additional dealer-exposure surfaces:
//
//   - ExposureSheet           — GET /v1/exposure/sheet/{symbol}
//   - ExposureTermStructure   — GET /v1/exposure/term-structure/{symbol}
//   - ExposureBasket          — GET /v1/exposure/basket
//   - ExposureOiDiff          — GET /v1/exposure/oi-diff/{symbol}
//
// All four are settled-OI aggregates (use /v1/flow/* for the simulator-aware
// live versions). Every untyped method returns map[string]interface{}; the
// *Typed wrappers decode into the structs below. All require Growth+.

import (
	"context"
	"net/url"
	"strconv"
)

// ── Functional options ───────────────────────────────────────────────────────

// ExposureSheetOption configures an ExposureSheet request.
type ExposureSheetOption func(*exposureSheetConfig)

type exposureSheetConfig struct {
	expiration string
	minOI      *int
}

// WithSheetExpiration filters the exposure sheet to a single expiry
// (YYYY-MM-DD) and triggers the is_opex / is_triple_witching flags.
func WithSheetExpiration(exp string) ExposureSheetOption {
	return func(c *exposureSheetConfig) { c.expiration = exp }
}

// WithSheetMinOI drops strikes whose call_oi + put_oi < minOI.
func WithSheetMinOI(minOI int) ExposureSheetOption {
	return func(c *exposureSheetConfig) { c.minOI = &minOI }
}

// OiDiffOption configures an ExposureOiDiff request.
type OiDiffOption func(*oiDiffConfig)

type oiDiffConfig struct {
	topN *int
}

// WithOiDiffTopN sets how many top OI-change rows to return (clamped 1–100).
func WithOiDiffTopN(topN int) OiDiffOption {
	return func(c *oiDiffConfig) { c.topN = &topN }
}

// ── exposure/sheet ────────────────────────────────────────────────────────────

// ExposureSheetTotals holds the chain-level net-exposure totals on the sheet.
type ExposureSheetTotals struct {
	NetGex  float64 `json:"net_gex"`
	NetDex  float64 `json:"net_dex"`
	NetVex  float64 `json:"net_vex"`
	NetChex float64 `json:"net_chex"`
	// NetDag is the net delta-adjusted gamma across the chain.
	NetDag float64 `json:"net_dag"`
}

// ExposureSheetLis is the Line-in-the-Sand inflection strike — the strike with
// the largest second derivative of net GEX. Nil when undefined.
type ExposureSheetLis struct {
	Strike    float64 `json:"strike"`
	Magnitude float64 `json:"magnitude"`
}

// ExposureSheetPeak is one local maximum of |net_gex| (a gamma wall).
type ExposureSheetPeak struct {
	Strike   float64 `json:"strike"`
	NetGex   float64 `json:"net_gex"`
	Strength float64 `json:"strength"`
	// Side is "call_wall" when net_gex >= 0, else "put_wall".
	Side string `json:"side"`
}

// ExposureSheetStrike is a unified per-strike row joining GEX/DEX/VEX/CHEX/DAG.
type ExposureSheetStrike struct {
	Strike   float64 `json:"strike"`
	CallGex  float64 `json:"call_gex"`
	PutGex   float64 `json:"put_gex"`
	NetGex   float64 `json:"net_gex"`
	CallDex  float64 `json:"call_dex"`
	PutDex   float64 `json:"put_dex"`
	NetDex   float64 `json:"net_dex"`
	CallVex  float64 `json:"call_vex"`
	PutVex   float64 `json:"put_vex"`
	NetVex   float64 `json:"net_vex"`
	CallChex float64 `json:"call_chex"`
	PutChex  float64 `json:"put_chex"`
	NetChex  float64 `json:"net_chex"`
	Dag      float64 `json:"dag"`
	CallOi   float64 `json:"call_oi"`
	PutOi    float64 `json:"put_oi"`
}

// ExposureSheetResponse is the typed body of GET /v1/exposure/sheet/{symbol}.
type ExposureSheetResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol           string                `json:"symbol"`
	UnderlyingPrice  *float64              `json:"underlying_price"`
	AsOf             string                `json:"as_of"`
	Expiration       *string               `json:"expiration"`
	IsOpex           bool                  `json:"is_opex"`
	IsTripleWitching bool                  `json:"is_triple_witching"`
	Totals           ExposureSheetTotals   `json:"totals"`
	Lis              *ExposureSheetLis     `json:"lis"`
	Peaks            []ExposureSheetPeak   `json:"peaks"`
	Strikes          []ExposureSheetStrike `json:"strikes"`
}

// ExposureSheet returns the unified per-strike GEX/DEX/VEX/CHEX/DAG rowset plus
// chain totals, Line-in-the-Sand inflection strike, all gamma peaks, and OPEX /
// triple-witching flags. Requires Growth+. Optional: WithSheetExpiration,
// WithSheetMinOI.
func (c *Client) ExposureSheet(ctx context.Context, symbol string, opts ...ExposureSheetOption) (map[string]interface{}, error) {
	cfg := &exposureSheetConfig{}
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
	return c.get(ctx, "/v1/exposure/sheet/"+seg(symbol), params)
}

// ExposureSheetTyped is the strongly-typed variant of ExposureSheet.
func (c *Client) ExposureSheetTyped(ctx context.Context, symbol string, opts ...ExposureSheetOption) (*ExposureSheetResponse, error) {
	raw, err := c.ExposureSheet(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ExposureSheetResponse{}
	if err := decodeTyped("exposure sheet", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── exposure/term-structure ──────────────────────────────────────────────────

// ExposureDteBucket is one DTE bucket of aggregated per-greek exposure.
type ExposureDteBucket struct {
	Bucket        string  `json:"bucket"`
	DteRange      []int   `json:"dte_range"`
	NetGex        float64 `json:"net_gex"`
	NetDex        float64 `json:"net_dex"`
	NetVex        float64 `json:"net_vex"`
	NetChex       float64 `json:"net_chex"`
	ContractCount int     `json:"contract_count"`
}

// ExposureExpiryRow is one per-expiry rollup of net exposure.
type ExposureExpiryRow struct {
	Expiration       string  `json:"expiration"`
	Dte              int     `json:"dte"`
	IsOpex           bool    `json:"is_opex"`
	IsTripleWitching bool    `json:"is_triple_witching"`
	NetGex           float64 `json:"net_gex"`
	NetDex           float64 `json:"net_dex"`
	NetVex           float64 `json:"net_vex"`
	NetChex          float64 `json:"net_chex"`
	// PctOfChainGex is this expiry's |net_gex| share of the chain (0-100).
	PctOfChainGex float64 `json:"pct_of_chain_gex"`
}

// ExposureTermStructureResponse is the body of GET /v1/exposure/term-structure/{symbol}.
type ExposureTermStructureResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol          string              `json:"symbol"`
	UnderlyingPrice *float64            `json:"underlying_price"`
	AsOf            string              `json:"as_of"`
	ByDteBucket     []ExposureDteBucket `json:"by_dte_bucket"`
	ByExpiry        []ExposureExpiryRow `json:"by_expiry"`
}

// ExposureTermStructure returns per-greek exposure aggregated by DTE bucket and
// rolled up per expiry — one call in place of four full-chain exposure calls,
// grouped by time. Requires Growth+.
func (c *Client) ExposureTermStructure(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/exposure/term-structure/"+seg(symbol), nil)
}

// ExposureTermStructureTyped is the strongly-typed variant of ExposureTermStructure.
func (c *Client) ExposureTermStructureTyped(ctx context.Context, symbol string) (*ExposureTermStructureResponse, error) {
	raw, err := c.ExposureTermStructure(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &ExposureTermStructureResponse{}
	if err := decodeTyped("exposure term-structure", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── exposure/basket ──────────────────────────────────────────────────────────

// ExposureBasketAggregate is the weighted-aggregate net exposure of a basket.
type ExposureBasketAggregate struct {
	NetGex  float64 `json:"net_gex"`
	NetDex  float64 `json:"net_dex"`
	NetVex  float64 `json:"net_vex"`
	NetChex float64 `json:"net_chex"`
}

// ExposureBasketConstituent is one symbol's contribution to a basket aggregate.
type ExposureBasketConstituent struct {
	Symbol          string   `json:"symbol"`
	Weight          float64  `json:"weight"`
	UnderlyingPrice *float64 `json:"underlying_price"`
	NetGex          float64  `json:"net_gex"`
	NetDex          float64  `json:"net_dex"`
	NetVex          float64  `json:"net_vex"`
	NetChex         float64  `json:"net_chex"`
	// ContributionPct is the weighted-GEX share of the aggregate (0-100).
	ContributionPct float64 `json:"contribution_pct"`
	// Regime is "positive_gamma" when net_gex >= 0, else "negative_gamma".
	Regime string `json:"regime"`
}

// ExposureBasketResponse is the body of GET /v1/exposure/basket.
type ExposureBasketResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	AsOf             string                      `json:"as_of"`
	ConstituentCount int                         `json:"constituent_count"`
	MissingSymbols   []string                    `json:"missing_symbols"`
	Aggregate        ExposureBasketAggregate     `json:"aggregate"`
	Constituents     []ExposureBasketConstituent `json:"constituents"`
}

// BasketOption configures an ExposureBasket request.
type BasketOption func(*basketConfig)

type basketConfig struct {
	weights string
}

// WithBasketWeights sets comma-separated weights aligned to the symbols list.
// Defaults to equal weight. Negative values are rejected by the server.
func WithBasketWeights(weights string) BasketOption {
	return func(c *basketConfig) { c.weights = weights }
}

// ExposureBasket returns the weighted cross-symbol aggregate of GEX/DEX/VEX/CHEX
// across up to 50 symbols. Requires Growth+. symbols is comma-separated and
// required. Optional: WithBasketWeights.
func (c *Client) ExposureBasket(ctx context.Context, symbols string, opts ...BasketOption) (map[string]interface{}, error) {
	cfg := &basketConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	params.Set("symbols", symbols)
	if cfg.weights != "" {
		params.Set("weights", cfg.weights)
	}
	return c.get(ctx, "/v1/exposure/basket", params)
}

// ExposureBasketTyped is the strongly-typed variant of ExposureBasket.
func (c *Client) ExposureBasketTyped(ctx context.Context, symbols string, opts ...BasketOption) (*ExposureBasketResponse, error) {
	raw, err := c.ExposureBasket(ctx, symbols, opts...)
	if err != nil {
		return nil, err
	}
	out := &ExposureBasketResponse{}
	if err := decodeTyped("exposure basket", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── exposure/oi-diff ─────────────────────────────────────────────────────────

// OiChange is one per-contract day-over-day open-interest delta.
type OiChange struct {
	Strike   float64 `json:"strike"`
	Type     string  `json:"type"`
	Expiry   string  `json:"expiry"`
	TodayOi  float64 `json:"today_oi"`
	PriorOi  float64 `json:"prior_oi"`
	OiChange float64 `json:"oi_change"`
}

// ExposureOiDiffResponse is the body of GET /v1/exposure/oi-diff/{symbol}.
type ExposureOiDiffResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol                 string     `json:"symbol"`
	UnderlyingPrice        *float64   `json:"underlying_price"`
	AsOf                   string     `json:"as_of"`
	PriorSnapshotAvailable bool       `json:"prior_snapshot_available"`
	TotalCallOiChange      float64    `json:"total_call_oi_change"`
	TotalPutOiChange       float64    `json:"total_put_oi_change"`
	TopOiChanges           []OiChange `json:"top_oi_changes"`
}

// ExposureOiDiff returns day-over-day open-interest deltas — per-contract
// changes, top-N by absolute magnitude, and call/put aggregate totals.
// Requires Growth+. Optional: WithOiDiffTopN.
func (c *Client) ExposureOiDiff(ctx context.Context, symbol string, opts ...OiDiffOption) (map[string]interface{}, error) {
	cfg := &oiDiffConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.topN != nil {
		params.Set("topN", strconv.Itoa(*cfg.topN))
	}
	return c.get(ctx, "/v1/exposure/oi-diff/"+seg(symbol), params)
}

// ExposureOiDiffTyped is the strongly-typed variant of ExposureOiDiff.
func (c *Client) ExposureOiDiffTyped(ctx context.Context, symbol string, opts ...OiDiffOption) (*ExposureOiDiffResponse, error) {
	raw, err := c.ExposureOiDiff(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &ExposureOiDiffResponse{}
	if err := decodeTyped("exposure oi-diff", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
