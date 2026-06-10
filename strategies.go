package flashalpha

// Strategy Signals — GET /v1/strategies/* (Basic+/Growth+/Alpha+).
//
// Ten endpoints share ONE uniform decision envelope (StrategyDecisionResponse).
// Only `metrics` (a free-form key/value bag) and `regime` vary per strategy;
// everything else has a fixed shape. Each named method below returns the shared
// typed envelope. The signal endpoints (skew / term-structure / tail-pricing)
// are pure-signal — `best_structures` is empty and `decision` is "neutral".

import (
	"context"
	"net/url"
	"strconv"
)

// ── Shared decision envelope ──────────────────────────────────────────────────

// StrategyLeg is one leg of a proposed structure.
type StrategyLeg struct {
	Action   string   `json:"action"`
	Type     string   `json:"type"`
	Strike   float64  `json:"strike"`
	Delta    *float64 `json:"delta"`
	Premium  *float64 `json:"premium"`
	Quantity int      `json:"quantity"`
}

// StrategyStructure is one ranked candidate structure.
type StrategyStructure struct {
	Rank           int           `json:"rank"`
	Structure      string        `json:"structure"`
	Expiry         string        `json:"expiry"`
	Legs           []StrategyLeg `json:"legs"`
	Credit         *float64      `json:"credit"`
	Debit          *float64      `json:"debit"`
	MaxProfit      *float64      `json:"max_profit"`
	MaxLoss        *float64      `json:"max_loss"`
	Breakevens     []float64     `json:"breakevens"`
	EdgeScore      *float64      `json:"edge_score"`
	LiquidityScore *float64      `json:"liquidity_score"`
}

// StrategyRiskFlag is one optional risk callout.
type StrategyRiskFlag struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// StrategyDataQuality gates whether the decision should be acted on.
type StrategyDataQuality struct {
	Score    int      `json:"score"`
	Warnings []string `json:"warnings"`
}

// StrategyDecisionResponse is the uniform envelope returned by every
// /v1/strategies/* endpoint.
//
// Only Metrics and Regime change between strategies — Metrics is a
// strategy-specific key/value bag (always includes underlying_price). Decision
// is one of insufficient_data / avoid / neutral / candidate. Gate on
// DataQuality before trading.
type StrategyDecisionResponse struct {
	Strategy       string                 `json:"strategy"`
	Symbol         string                 `json:"symbol"`
	Timestamp      string                 `json:"timestamp"`
	Decision       string                 `json:"decision"`
	Score          int                    `json:"score"`
	Confidence     float64                `json:"confidence"`
	Regime         string                 `json:"regime"`
	BestStructures []StrategyStructure    `json:"best_structures"`
	Metrics        map[string]interface{} `json:"metrics"`
	RiskFlags      []StrategyRiskFlag     `json:"risk_flags"`
	Why            []string               `json:"why"`
	AvoidIf        []string               `json:"avoid_if"`
	DataQuality    *StrategyDataQuality   `json:"data_quality"`
}

// ── Shared options ───────────────────────────────────────────────────────────

// StrategyOption configures a strategy request. A single option type is shared
// by all strategy endpoints; each method reads only the parameters it supports
// (documented on the method). Unsupported options are ignored.
type StrategyOption func(*strategyConfig)

type strategyConfig struct {
	expiry                      string
	minOpenInterest             *int
	wingWidth                   *float64
	targetShortDelta            *float64
	maxWidth                    *float64
	minCredit                   *float64
	targetDelta                 *float64
	structure                   string
	excludeEarningsBeforeExpiry *bool
}

// WithStrategyExpiry restricts a strategy to a single expiry (YYYY-MM-DD).
func WithStrategyExpiry(expiry string) StrategyOption {
	return func(c *strategyConfig) { c.expiry = expiry }
}

// WithStrategyMinOpenInterest sets the minimum OI a leg must have.
func WithStrategyMinOpenInterest(minOI int) StrategyOption {
	return func(c *strategyConfig) { c.minOpenInterest = &minOI }
}

// WithStrategyWingWidth sets the iron-fly wing width in strike points.
func WithStrategyWingWidth(w float64) StrategyOption {
	return func(c *strategyConfig) { c.wingWidth = &w }
}

// WithStrategyTargetShortDelta sets the target delta for the short leg
// (vol-carry).
func WithStrategyTargetShortDelta(d float64) StrategyOption {
	return func(c *strategyConfig) { c.targetShortDelta = &d }
}

// WithStrategyMaxWidth sets the maximum spread/wing width (vol-carry).
func WithStrategyMaxWidth(w float64) StrategyOption {
	return func(c *strategyConfig) { c.maxWidth = &w }
}

// WithStrategyMinCredit sets the minimum credit a candidate must collect
// (vol-carry).
func WithStrategyMinCredit(cr float64) StrategyOption {
	return func(c *strategyConfig) { c.minCredit = &cr }
}

// WithStrategyTargetDelta sets the target delta for the short option
// (yield-enhancement).
func WithStrategyTargetDelta(d float64) StrategyOption {
	return func(c *strategyConfig) { c.targetDelta = &d }
}

// WithStrategyStructure selects "covered_call" or "cash_secured_put"
// (yield-enhancement).
func WithStrategyStructure(structure string) StrategyOption {
	return func(c *strategyConfig) { c.structure = structure }
}

// WithStrategyExcludeEarningsBeforeExpiry penalises/flags picks whose expiry
// spans an earnings event (yield-enhancement).
func WithStrategyExcludeEarningsBeforeExpiry(exclude bool) StrategyOption {
	return func(c *strategyConfig) { c.excludeEarningsBeforeExpiry = &exclude }
}

func strategyConfigFrom(opts []StrategyOption) *strategyConfig {
	cfg := &strategyConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// strategyGet builds params from the supported keys and decodes the shared
// envelope. The keys slice names which option fields this endpoint honours.
func (c *Client) strategyGet(ctx context.Context, label, path string, cfg *strategyConfig, keys ...string) (*StrategyDecisionResponse, error) {
	params := url.Values{}
	for _, k := range keys {
		switch k {
		case "expiry":
			if cfg.expiry != "" {
				params.Set("expiry", cfg.expiry)
			}
		case "minOpenInterest":
			if cfg.minOpenInterest != nil {
				params.Set("minOpenInterest", strconv.Itoa(*cfg.minOpenInterest))
			}
		case "wingWidth":
			if cfg.wingWidth != nil {
				params.Set("wingWidth", strconv.FormatFloat(*cfg.wingWidth, 'f', -1, 64))
			}
		case "targetShortDelta":
			if cfg.targetShortDelta != nil {
				params.Set("targetShortDelta", strconv.FormatFloat(*cfg.targetShortDelta, 'f', -1, 64))
			}
		case "maxWidth":
			if cfg.maxWidth != nil {
				params.Set("maxWidth", strconv.FormatFloat(*cfg.maxWidth, 'f', -1, 64))
			}
		case "minCredit":
			if cfg.minCredit != nil {
				params.Set("minCredit", strconv.FormatFloat(*cfg.minCredit, 'f', -1, 64))
			}
		case "targetDelta":
			if cfg.targetDelta != nil {
				params.Set("targetDelta", strconv.FormatFloat(*cfg.targetDelta, 'f', -1, 64))
			}
		case "structure":
			if cfg.structure != "" {
				params.Set("structure", cfg.structure)
			}
		case "excludeEarningsBeforeExpiry":
			if cfg.excludeEarningsBeforeExpiry != nil {
				params.Set("excludeEarningsBeforeExpiry", strconv.FormatBool(*cfg.excludeEarningsBeforeExpiry))
			}
		}
	}
	raw, err := c.get(ctx, path, params)
	if err != nil {
		return nil, err
	}
	out := &StrategyDecisionResponse{}
	if err := decodeTyped(label, raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── Named strategy methods ────────────────────────────────────────────────────

// StrategyFlowAnomaly scores directional options-flow imbalance and proposes a
// matching short vertical spread. Requires Growth+. Optional: WithStrategyExpiry.
func (c *Client) StrategyFlowAnomaly(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy flow-anomaly", "/v1/strategies/flow-anomaly/"+seg(symbol), strategyConfigFrom(opts), "expiry")
}

// StrategyExpiryPositioning scores OPEX pin risk for a single expiry and
// proposes an iron fly when a pin is likely. Requires Basic+. Optional:
// WithStrategyExpiry, WithStrategyMinOpenInterest, WithStrategyWingWidth.
func (c *Client) StrategyExpiryPositioning(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy expiry-positioning", "/v1/strategies/expiry-positioning/"+seg(symbol), strategyConfigFrom(opts), "expiry", "minOpenInterest", "wingWidth")
}

// StrategyZeroDte is the same-day range-compression read combining pin
// positioning with intraday time-of-day context. Requires Growth+ and 0DTE
// access. Optional: WithStrategyExpiry, WithStrategyMinOpenInterest,
// WithStrategyWingWidth.
func (c *Client) StrategyZeroDte(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy zero-dte", "/v1/strategies/zero-dte/"+seg(symbol), strategyConfigFrom(opts), "expiry", "minOpenInterest", "wingWidth")
}

// StrategyDealerRegime classifies the dealer gamma regime and proposes an iron
// condor in compressive regimes. Requires Growth+. Optional: WithStrategyExpiry.
func (c *Client) StrategyDealerRegime(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy dealer-regime", "/v1/strategies/dealer-regime/"+seg(symbol), strategyConfigFrom(opts), "expiry")
}

// StrategyVolCarry scores vol-risk-premium carry (implied vs realized) and
// proposes short put spreads / iron condors when carry is favourable.
// Requires Alpha+. Optional: WithStrategyExpiry, WithStrategyMinOpenInterest,
// WithStrategyTargetShortDelta, WithStrategyMaxWidth, WithStrategyMinCredit.
func (c *Client) StrategyVolCarry(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy vol-carry", "/v1/strategies/vol-carry/"+seg(symbol), strategyConfigFrom(opts), "expiry", "minOpenInterest", "targetShortDelta", "maxWidth", "minCredit")
}

// StrategyYieldEnhancement finds the best covered call or cash-secured put at a
// target delta with annualized yield, assignment probability, and an earnings
// filter. Requires Growth+. Optional: WithStrategyExpiry, WithStrategyTargetDelta,
// WithStrategyMinOpenInterest, WithStrategyStructure,
// WithStrategyExcludeEarningsBeforeExpiry.
func (c *Client) StrategyYieldEnhancement(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy yield-enhancement", "/v1/strategies/yield-enhancement/"+seg(symbol), strategyConfigFrom(opts), "expiry", "targetDelta", "minOpenInterest", "structure", "excludeEarningsBeforeExpiry")
}

// StrategySurfaceAnomaly compares observed IVs against the calibrated SVI fit
// to find rich/cheap wings and proposes the obvious vertical-credit sale on a
// rich wing. Requires Alpha+. Optional: WithStrategyExpiry.
func (c *Client) StrategySurfaceAnomaly(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy surface-anomaly", "/v1/strategies/surface-anomaly/"+seg(symbol), strategyConfigFrom(opts), "expiry")
}

// StrategySkew is a pure-signal read of 25-delta skew (put vs call wing
// richness and the risk reversal) for one expiry. Requires Growth+. Optional:
// WithStrategyExpiry.
func (c *Client) StrategySkew(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy skew", "/v1/strategies/skew/"+seg(symbol), strategyConfigFrom(opts), "expiry")
}

// StrategyTermStructure is a pure-signal read of the ATM IV term structure
// across all upcoming expiries (contango vs backwardation). Requires Growth+.
// Takes no tuning parameters.
func (c *Client) StrategyTermStructure(ctx context.Context, symbol string) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy term-structure", "/v1/strategies/term-structure/"+seg(symbol), &strategyConfig{})
}

// StrategyTailPricing is a pure-signal read of downside-tail pricing for one
// expiry (10-delta put IV vs ATM, tail asymmetry, rich/cheap crash tail).
// Requires Growth+. Optional: WithStrategyExpiry.
func (c *Client) StrategyTailPricing(ctx context.Context, symbol string, opts ...StrategyOption) (*StrategyDecisionResponse, error) {
	return c.strategyGet(ctx, "strategy tail-pricing", "/v1/strategies/tail-pricing/"+seg(symbol), strategyConfigFrom(opts), "expiry")
}
