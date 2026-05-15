package flashalpha

// Typed models + client methods for the live (simulation-aware) flow surface
// under /v1/flow/*. Two families:
//
//   - Analytics (/v1/flow/{levels,pin-risk,summary,oi,gex,dex,dealer-risk,
//     live}/{symbol}) — fold today's intraday trade tape onto the settled
//     book, so gamma flip / walls / GEX reflect *today's* flow. snake_case
//     wire shape; optional expiry=YYYY-MM-DD slices to one expiration cycle.
//
//   - Raw flow data (/v1/flow/options/*, /v1/flow/stocks/*) — the underlying
//     trade tape: prints, blocks, per-minute history, cumulative net-flow
//     series, cross-symbol leaderboards / outliers. Proxied verbatim so wire
//     keys are camelCase and timestamps are ISO-8601 UTC strings.
//
// All /v1/flow/* endpoints require the Alpha plan. Every untyped method
// returns map[string]interface{} (consistent with the rest of the client);
// the *Typed wrappers decode into the structs below. Flow gex/dex per-strike
// rows are the same wire shape as /v1/exposure/gex|dex, so they reuse
// GexStrike / DexStrike.

import (
	"context"
	"net/url"
	"strconv"
)

// ── Functional options ───────────────────────────────────────────────────────

// FlowOption configures a flow request. A single option type is shared by all
// flow endpoints; each endpoint reads only the parameters it supports
// (documented on the method). Unsupported options are ignored.
type FlowOption func(*flowConfig)

type flowConfig struct {
	expiry        string
	limit         *int
	minSize       *int
	minutes       *int
	n             *int
	windowMinutes *int
	minTrades     *int
}

// WithFlowExpiry slices a flow request to a single expiration cycle
// (format "YYYY-MM-DD"). Honoured by the analytics endpoints and the raw
// option endpoints; ignored by the stock endpoints.
func WithFlowExpiry(expiry string) FlowOption {
	return func(c *flowConfig) { c.expiry = expiry }
}

// WithFlowLimit sets the max number of trades to return (recent: 1–500).
func WithFlowLimit(limit int) FlowOption {
	return func(c *flowConfig) { c.limit = &limit }
}

// WithFlowMinSize sets the minimum trade size that qualifies as a block.
func WithFlowMinSize(minSize int) FlowOption {
	return func(c *flowConfig) { c.minSize = &minSize }
}

// WithFlowMinutes sets the lookback window in minutes (1–10080) for the
// history and cumulative endpoints.
func WithFlowMinutes(minutes int) FlowOption {
	return func(c *flowConfig) { c.minutes = &minutes }
}

// WithFlowN sets the number of ranked rows per side (1–50) for leaderboards.
func WithFlowN(n int) FlowOption {
	return func(c *flowConfig) { c.n = &n }
}

// WithFlowWindowMinutes sets the aggregation window in minutes (1–10080) for
// the leaderboard and outliers endpoints.
func WithFlowWindowMinutes(windowMinutes int) FlowOption {
	return func(c *flowConfig) { c.windowMinutes = &windowMinutes }
}

// WithFlowMinTrades sets the minimum trades a symbol needs to qualify for
// the outliers endpoints.
func WithFlowMinTrades(minTrades int) FlowOption {
	return func(c *flowConfig) { c.minTrades = &minTrades }
}

func flowParams(opts []FlowOption) (*flowConfig, url.Values) {
	cfg := &flowConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg, url.Values{}
}

// ── Analytics response structs (snake_case) ──────────────────────────────────

// FlowLevelsResponse is the typed body of GET /v1/flow/levels/{symbol}:
// gamma flip / call & put walls / max pain recomputed against the live
// (intraday-flow-adjusted) book. Each level is nil when it can't be located.
type FlowLevelsResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// UnderlyingPrice is the spot mid at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiry is the expiration filter echoed back ("YYYY-MM-DD"), or nil
	// when the whole chain was used.
	Expiry *string `json:"expiry"`
	// LiveGammaFlip is the spot where live net dealer gamma crosses zero.
	LiveGammaFlip *float64 `json:"live_gamma_flip"`
	// LiveCallWall is the strike of the largest live call-gamma
	// concentration (upside magnet).
	LiveCallWall *float64 `json:"live_call_wall"`
	// LivePutWall is the strike of the largest live put-gamma
	// concentration (downside magnet).
	LivePutWall *float64 `json:"live_put_wall"`
	// LiveMaxPain is the live max-pain strike (most option value expires
	// worthless).
	LiveMaxPain *float64 `json:"live_max_pain"`
}

// FlowPinRiskBreakdown holds the component scores (0–100) behind the
// LivePinRisk headline.
type FlowPinRiskBreakdown struct {
	// OiScore is the open-interest concentration around the magnet strike.
	OiScore *int `json:"oi_score"`
	// ProximityScore is how close spot is to the magnet strike.
	ProximityScore *int `json:"proximity_score"`
	// TimeScore is the time-to-close weighting (pin pressure rises into
	// the cash close).
	TimeScore *int `json:"time_score"`
	// GammaScore is the dealer-gamma intensity at the magnet strike.
	GammaScore *int `json:"gamma_score"`
}

// FlowPinRiskResponse is the typed body of GET /v1/flow/pin-risk/{symbol}:
// a 0–100 composite pin-risk score plus the magnet strike and breakdown.
type FlowPinRiskResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// UnderlyingPrice is the spot mid at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiry is the expiration filter echoed back, or nil.
	Expiry *string `json:"expiry"`
	// LivePinRisk is the composite 0–100 pin-risk score (higher = stronger
	// pin pull).
	LivePinRisk *int `json:"live_pin_risk"`
	// MagnetStrike is the strike acting as the pin magnet
	// (argmax|net gamma|), or nil when there is no dominant strike.
	MagnetStrike *float64 `json:"magnet_strike"`
	// DistanceToMagnetPct is the signed % distance from spot to the magnet
	// strike.
	DistanceToMagnetPct *float64 `json:"distance_to_magnet_pct"`
	// TimeToCloseHours is the hours remaining until the regular-session
	// cash close.
	TimeToCloseHours *float64 `json:"time_to_close_hours"`
	// Breakdown holds the four component scores behind LivePinRisk.
	Breakdown FlowPinRiskBreakdown `json:"breakdown"`
}

// FlowSummaryResponse is the typed body of GET /v1/flow/summary/{symbol}:
// an at-a-glance read on whether today's tape has shifted the dealer book.
type FlowSummaryResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// UnderlyingPrice is the spot mid at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiry is the expiration filter echoed back, or nil.
	Expiry *string `json:"expiry"`
	// FlowDirection is the net classified direction of intraday flow
	// (e.g. "bullish", "bearish", "neutral").
	FlowDirection string `json:"flow_direction"`
	// IntradayOiDelta is the net change in simulated open interest since
	// the open (contracts).
	IntradayOiDelta *int64 `json:"intraday_oi_delta"`
	// ContractsWithFlow is the number of contracts that have printed at
	// least one trade today.
	ContractsWithFlow *int `json:"contracts_with_flow"`
	// ContractsTotal is the total contracts tracked for the underlying.
	ContractsTotal *int `json:"contracts_total"`
	// LiveGex is the live (flow-adjusted) net GEX in dollars per 1% spot.
	LiveGex *float64 `json:"live_gex"`
	// FlowGexPctShift is the % shift in net GEX caused by today's flow vs
	// the settled book; nil when the settled baseline is zero.
	FlowGexPctShift *float64 `json:"flow_gex_pct_shift"`
}

// FlowOiResponse is the typed body of GET /v1/flow/oi/{symbol}: the settled
// (official) OI vs the intraday simulated OI. This endpoint does NOT return
// underlying_price.
type FlowOiResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// Expiry is the expiration filter echoed back, or nil.
	Expiry *string `json:"expiry"`
	// OfficialOi is the official exchange OI from the settled snapshot
	// (sum across the chain).
	OfficialOi *int64 `json:"official_oi"`
	// SimulatedOi is the intraday simulated OI (official + estimated
	// open/close from the tape).
	SimulatedOi *int64 `json:"simulated_oi"`
	// IntradayOiDelta is SimulatedOi - OfficialOi (signed).
	IntradayOiDelta *int64 `json:"intraday_oi_delta"`
	// OiDeltaConfidence is the confidence 0–1 in the intraday OI estimate
	// (trade-tape coverage).
	OiDeltaConfidence *float64 `json:"oi_delta_confidence"`
	// EffectiveOi is the OI actually used by the live analytics (blended).
	EffectiveOi *int64 `json:"effective_oi"`
	// ContractsTotal is the total contracts tracked for the underlying.
	ContractsTotal *int `json:"contracts_total"`
	// ContractsWithFlow is the contracts that printed at least one trade.
	ContractsWithFlow *int `json:"contracts_with_flow"`
}

// FlowGexResponse is the typed body of GET /v1/flow/gex/{symbol}: the live
// (flow-adjusted) GEX with the same per-strike shape as /v1/exposure/gex.
type FlowGexResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// UnderlyingPrice is the spot mid at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiry is the expiration filter echoed back, or nil.
	Expiry *string `json:"expiry"`
	// LiveNetGex is the live net GEX across the chain (dollars per 1% spot).
	LiveNetGex *float64 `json:"live_net_gex"`
	// LiveNetGexLabel is a categorical regime label (e.g. "positive",
	// "negative"). Safe to surface verbatim.
	LiveNetGexLabel string `json:"live_net_gex_label"`
	// LiveGammaFlip is the live gamma-flip spot, or nil if no sign change.
	LiveGammaFlip *float64 `json:"live_gamma_flip"`
	// Strikes is the per-strike breakdown (identical schema to settled GEX).
	Strikes []GexStrike `json:"strikes"`
}

// FlowDexResponse is the typed body of GET /v1/flow/dex/{symbol}: the live
// (flow-adjusted) DEX with the same per-strike shape as /v1/exposure/dex.
type FlowDexResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// UnderlyingPrice is the spot mid at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiry is the expiration filter echoed back, or nil.
	Expiry *string `json:"expiry"`
	// LiveNetDex is the live net DEX across the chain (dollars).
	LiveNetDex *float64 `json:"live_net_dex"`
	// Strikes is the per-strike DEX breakdown.
	Strikes []DexStrike `json:"strikes"`
}

// FlowDealerRiskResponse is the typed body of
// GET /v1/flow/dealer-risk/{symbol}: a side-by-side of the settled snapshot
// and the live flow-adjusted book, with the adjustment today's tape produced.
type FlowDealerRiskResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// UnderlyingPrice is the spot mid at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiry is the expiration filter echoed back, or nil.
	Expiry *string `json:"expiry"`
	// SettledNetGex is the net GEX from the settled (prior close) snapshot.
	SettledNetGex *float64 `json:"settled_net_gex"`
	// LiveNetGex is the net GEX from the live flow-adjusted book.
	LiveNetGex *float64 `json:"live_net_gex"`
	// FlowGexAdjustment is LiveNetGex - SettledNetGex (dollars).
	FlowGexAdjustment *float64 `json:"flow_gex_adjustment"`
	// FlowGexPctShift is the % GEX shift from flow; nil when the settled
	// baseline is zero.
	FlowGexPctShift *float64 `json:"flow_gex_pct_shift"`
	// SettledNetDex is the net DEX from the settled snapshot.
	SettledNetDex *float64 `json:"settled_net_dex"`
	// LiveNetDex is the net DEX from the live flow-adjusted book.
	LiveNetDex *float64 `json:"live_net_dex"`
	// FlowDexAdjustment is LiveNetDex - SettledNetDex (dollars).
	FlowDexAdjustment *float64 `json:"flow_dex_adjustment"`
	// FlowDexPctShift is the % DEX shift from flow; nil when baseline zero.
	FlowDexPctShift *float64 `json:"flow_dex_pct_shift"`
	// TotalAbsDeltaContracts is the absolute delta-weighted contracts
	// traded today (flow magnitude).
	TotalAbsDeltaContracts *int64 `json:"total_abs_delta_contracts"`
	// ContractsWithFlow is the contracts that printed at least one trade.
	ContractsWithFlow *int `json:"contracts_with_flow"`
	// FlowDirection is the net classified flow direction.
	FlowDirection string `json:"flow_direction"`
	// Description is a plain-English summary of whether flow has moved the
	// dealer book. Safe to surface verbatim.
	Description string `json:"description"`
}

// FlowAdjustedDealerRisk is the nested dealer-risk block inside
// FlowLiveResponse. Identical to FlowDealerRiskResponse minus
// ContractsWithFlow (carried on the parent live envelope instead).
type FlowAdjustedDealerRisk struct {
	// SettledNetGex is the net GEX from the settled snapshot.
	SettledNetGex *float64 `json:"settled_net_gex"`
	// LiveNetGex is the net GEX from the live flow-adjusted book.
	LiveNetGex *float64 `json:"live_net_gex"`
	// FlowGexAdjustment is LiveNetGex - SettledNetGex (dollars).
	FlowGexAdjustment *float64 `json:"flow_gex_adjustment"`
	// FlowGexPctShift is the % GEX shift from flow; nil when baseline zero.
	FlowGexPctShift *float64 `json:"flow_gex_pct_shift"`
	// SettledNetDex is the net DEX from the settled snapshot.
	SettledNetDex *float64 `json:"settled_net_dex"`
	// LiveNetDex is the net DEX from the live flow-adjusted book.
	LiveNetDex *float64 `json:"live_net_dex"`
	// FlowDexAdjustment is LiveNetDex - SettledNetDex (dollars).
	FlowDexAdjustment *float64 `json:"flow_dex_adjustment"`
	// FlowDexPctShift is the % DEX shift from flow; nil when baseline zero.
	FlowDexPctShift *float64 `json:"flow_dex_pct_shift"`
	// TotalAbsDeltaContracts is the absolute delta-weighted contracts
	// traded today (flow magnitude).
	TotalAbsDeltaContracts *int64 `json:"total_abs_delta_contracts"`
	// FlowDirection is the net classified flow direction.
	FlowDirection string `json:"flow_direction"`
	// Description is a plain-English summary. Safe to surface verbatim.
	Description string `json:"description"`
}

// FlowLiveResponse is the typed body of GET /v1/flow/live/{symbol}: an
// everything-at-once convenience bundle (OI simulator state + live exposure
// + live levels + pin risk + nested dealer-risk block).
type FlowLiveResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// AsOf is the timestamp this snapshot was computed for (ISO-8601 UTC).
	AsOf string `json:"as_of"`
	// UnderlyingPrice is the spot mid at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiry is the expiration filter echoed back, or nil.
	Expiry *string `json:"expiry"`
	// Contracts is the total contracts tracked for the underlying.
	Contracts *int `json:"contracts"`
	// ContractsWithFlow is the contracts that printed at least one trade.
	ContractsWithFlow *int `json:"contracts_with_flow"`
	// OfficialOi is the official exchange OI from the settled snapshot.
	OfficialOi *int64 `json:"official_oi"`
	// SimulatedOi is the intraday simulated OI.
	SimulatedOi *int64 `json:"simulated_oi"`
	// IntradayOiDelta is SimulatedOi - OfficialOi (signed).
	IntradayOiDelta *int64 `json:"intraday_oi_delta"`
	// OiDeltaConfidence is the confidence 0–1 in the intraday OI estimate.
	OiDeltaConfidence *float64 `json:"oi_delta_confidence"`
	// EffectiveOi is the OI actually used by the live analytics (blended).
	EffectiveOi *int64 `json:"effective_oi"`
	// LiveGex is the live net GEX (dollars per 1% spot move).
	LiveGex *float64 `json:"live_gex"`
	// LiveGexDelta is the live net DEX (dollars). Named live_gex_delta on
	// the wire.
	LiveGexDelta *float64 `json:"live_gex_delta"`
	// LiveGammaFlip is the live gamma-flip spot, or nil.
	LiveGammaFlip *float64 `json:"live_gamma_flip"`
	// LiveCallWall is the live call wall strike, or nil.
	LiveCallWall *float64 `json:"live_call_wall"`
	// LivePutWall is the live put wall strike, or nil.
	LivePutWall *float64 `json:"live_put_wall"`
	// LiveMaxPain is the live max-pain strike, or nil.
	LiveMaxPain *float64 `json:"live_max_pain"`
	// LivePinRisk is the composite 0–100 pin-risk score.
	LivePinRisk *int `json:"live_pin_risk"`
	// FlowAdjustedDealerRisk is the nested settled-vs-live dealer block.
	FlowAdjustedDealerRisk FlowAdjustedDealerRisk `json:"flow_adjusted_dealer_risk"`
}

// ── Raw flow data structs (camelCase wire keys) ──────────────────────────────

// FlowOptionTrade is a single option trade print (a trades[] element).
type FlowOptionTrade struct {
	// Ts is the trade timestamp (ISO-8601 UTC).
	Ts string `json:"ts"`
	// InstrumentID is the OPRA instrument id of the contract.
	InstrumentID *int64 `json:"instrumentId"`
	// Expiry is the contract expiration ("YYYY-MM-DD").
	Expiry string `json:"expiry"`
	// Strike is the contract strike price.
	Strike *float64 `json:"strike"`
	// Right is "C" (call) or "P" (put).
	Right string `json:"right"`
	// Price is the trade price.
	Price *float64 `json:"price"`
	// Size is the trade size in contracts.
	Size *int `json:"size"`
	// Side is the trade-side classification vs the NBBO at print
	// ("buy"/"sell"/"mid").
	Side string `json:"side"`
	// IsBlock is true when the print is at/above the block-size threshold.
	IsBlock *bool `json:"isBlock"`
	// Bid is the NBBO bid at the moment of the trade.
	Bid *float64 `json:"bid"`
	// Ask is the NBBO ask at the moment of the trade.
	Ask *float64 `json:"ask"`
}

// FlowOptionRecentResponse is the typed body of
// GET /v1/flow/options/{symbol}/recent: a newest-first option trade tape.
// Expiry is echoed only when the filter is supplied.
type FlowOptionRecentResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Expiry is the expiration filter echoed back when supplied, else nil.
	Expiry *string `json:"expiry"`
	// Count is the number of trades returned (capped by the limit).
	Count *int `json:"count"`
	// TotalAvailable is the unclamped total trade count.
	TotalAvailable *int `json:"totalAvailable"`
	// Trades is the newest-first list of trade prints.
	Trades []FlowOptionTrade `json:"trades"`
}

// FlowOptionSummaryResponse is the typed body of
// GET /v1/flow/options/{symbol}/summary: per-underlying option-flow aggregates.
type FlowOptionSummaryResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Expiry is the expiration filter echoed back when supplied, else nil.
	Expiry *string `json:"expiry"`
	// ContractsWithTrades is the distinct contracts that printed a trade.
	ContractsWithTrades *int `json:"contractsWithTrades"`
	// TotalTrades is the total number of trade prints.
	TotalTrades *int `json:"totalTrades"`
	// BuyVolume is the buy-classified contract volume.
	BuyVolume *int64 `json:"buyVolume"`
	// SellVolume is the sell-classified contract volume.
	SellVolume *int64 `json:"sellVolume"`
	// MidVolume is the volume classified at the mid (uninformed).
	MidVolume *int64 `json:"midVolume"`
	// NetVolume is BuyVolume - SellVolume.
	NetVolume *int64 `json:"netVolume"`
	// BiggestSingleTrade is the largest single trade size.
	BiggestSingleTrade *int `json:"biggestSingleTrade"`
	// LastTradeUtc is the timestamp of the most recent print; nil when
	// there are no trades.
	LastTradeUtc *string `json:"lastTradeUtc"`
}

// FlowOptionBlock is a single large option print (a blocks[] element).
type FlowOptionBlock struct {
	// Ts is the trade timestamp (ISO-8601 UTC).
	Ts string `json:"ts"`
	// Expiry is the contract expiration ("YYYY-MM-DD").
	Expiry string `json:"expiry"`
	// Strike is the contract strike price.
	Strike *float64 `json:"strike"`
	// Right is "C" (call) or "P" (put).
	Right string `json:"right"`
	// Price is the trade price.
	Price *float64 `json:"price"`
	// Size is the trade size in contracts.
	Size *int `json:"size"`
	// Side is the trade-side classification ("buy"/"sell"/"mid").
	Side string `json:"side"`
}

// FlowOptionBlocksResponse is the typed body of
// GET /v1/flow/options/{symbol}/blocks: all trades with size >= minSize.
type FlowOptionBlocksResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Expiry is the expiration filter echoed back when supplied, else nil.
	Expiry *string `json:"expiry"`
	// MinSize is the minimum trade size that qualified as a block (echoed).
	MinSize *int `json:"minSize"`
	// Count is the number of blocks returned.
	Count *int `json:"count"`
	// Blocks is the newest-first list of large prints.
	Blocks []FlowOptionBlock `json:"blocks"`
}

// FlowHistoryBucket is one per-minute option-flow bucket (a buckets[]
// element of the option history endpoint).
type FlowHistoryBucket struct {
	// Ts is the bucket start (ISO-8601 UTC, minute-aligned).
	Ts string `json:"ts"`
	// BuyVolume is the buy-classified volume in the bucket.
	BuyVolume *int64 `json:"buyVolume"`
	// SellVolume is the sell-classified volume in the bucket.
	SellVolume *int64 `json:"sellVolume"`
	// MidVolume is the mid-classified volume in the bucket.
	MidVolume *int64 `json:"midVolume"`
	// NetVolume is BuyVolume - SellVolume.
	NetVolume *int64 `json:"netVolume"`
	// TradeCount is the number of trades in the bucket.
	TradeCount *int `json:"tradeCount"`
	// BiggestTrade is the largest single trade size in the bucket.
	BiggestTrade *int `json:"biggestTrade"`
	// Vwap is the volume-weighted average trade price across the bucket.
	Vwap *float64 `json:"vwap"`
	// High is the highest trade price in the bucket.
	High *float64 `json:"high"`
	// Low is the lowest trade price in the bucket.
	Low *float64 `json:"low"`
}

// FlowOptionHistoryResponse is the typed body of
// GET /v1/flow/options/{symbol}/history: newest-first per-minute buckets.
type FlowOptionHistoryResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Expiry is the expiration filter echoed back when supplied, else nil.
	Expiry *string `json:"expiry"`
	// Minutes is the lookback window in minutes (echoed back).
	Minutes *int `json:"minutes"`
	// Count is the number of buckets returned.
	Count *int `json:"count"`
	// Buckets is the newest-first list of per-minute aggregates.
	Buckets []FlowHistoryBucket `json:"buckets"`
}

// FlowCumulativePoint is one point of a cumulative net-flow series (a
// points[] element). Shared by the option and stock cumulative endpoints.
type FlowCumulativePoint struct {
	// Ts is the bucket start (ISO-8601 UTC, minute-aligned).
	Ts string `json:"ts"`
	// NetVolume is the net volume in this minute bucket.
	NetVolume *int64 `json:"netVolume"`
	// Cumulative is the running sum of NetVolume from the start of the
	// window (the "HIRO-style" cumulative line).
	Cumulative *int64 `json:"cumulative"`
	// Vwap is the volume-weighted average price in the bucket.
	Vwap *float64 `json:"vwap"`
	// TradeCount is the number of trades in the bucket.
	TradeCount *int `json:"tradeCount"`
}

// FlowOptionCumulativeResponse is the typed body of
// GET /v1/flow/options/{symbol}/cumulative.
type FlowOptionCumulativeResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Expiry is the expiration filter echoed back when supplied, else nil.
	Expiry *string `json:"expiry"`
	// Minutes is the lookback window in minutes (echoed back).
	Minutes *int `json:"minutes"`
	// Count is the number of points returned.
	Count *int `json:"count"`
	// Points is the chronological cumulative net-flow series.
	Points []FlowCumulativePoint `json:"points"`
}

// FlowStockTrade is a single stock trade print (a trades[] element).
type FlowStockTrade struct {
	// Ts is the trade timestamp (ISO-8601 UTC).
	Ts string `json:"ts"`
	// Price is the trade price.
	Price *float64 `json:"price"`
	// Size is the trade size in shares.
	Size *int `json:"size"`
	// Side is the trade-side classification ("buy"/"sell"/"mid").
	Side string `json:"side"`
	// IsBlock is true when the print is at/above the block-size threshold.
	IsBlock *bool `json:"isBlock"`
	// Bid is the NBBO bid at the moment of the trade.
	Bid *float64 `json:"bid"`
	// Ask is the NBBO ask at the moment of the trade.
	Ask *float64 `json:"ask"`
}

// FlowStockRecentResponse is the typed body of
// GET /v1/flow/stocks/{symbol}/recent: a newest-first stock trade tape.
type FlowStockRecentResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Count is the number of trades returned (capped by the limit).
	Count *int `json:"count"`
	// TotalAvailable is the unclamped total trade count.
	TotalAvailable *int `json:"totalAvailable"`
	// Trades is the newest-first list of trade prints.
	Trades []FlowStockTrade `json:"trades"`
}

// FlowStockSummaryResponse is the typed body of
// GET /v1/flow/stocks/{symbol}/summary: per-symbol stock-flow aggregates.
type FlowStockSummaryResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// TotalTrades is the total number of trade prints.
	TotalTrades *int `json:"totalTrades"`
	// BuyVolume is the buy-classified share volume.
	BuyVolume *int64 `json:"buyVolume"`
	// SellVolume is the sell-classified share volume.
	SellVolume *int64 `json:"sellVolume"`
	// MidVolume is the volume classified at the mid (uninformed).
	MidVolume *int64 `json:"midVolume"`
	// NetVolume is BuyVolume - SellVolume.
	NetVolume *int64 `json:"netVolume"`
	// BiggestSingleTrade is the largest single trade size.
	BiggestSingleTrade *int `json:"biggestSingleTrade"`
	// LastTradeUtc is the timestamp of the most recent print; nil when
	// there are no trades.
	LastTradeUtc *string `json:"lastTradeUtc"`
}

// FlowStockBlock is a single large stock print (a blocks[] element).
type FlowStockBlock struct {
	// Ts is the trade timestamp (ISO-8601 UTC).
	Ts string `json:"ts"`
	// Price is the trade price.
	Price *float64 `json:"price"`
	// Size is the trade size in shares.
	Size *int `json:"size"`
	// Side is the trade-side classification ("buy"/"sell"/"mid").
	Side string `json:"side"`
	// Bid is the NBBO bid at the moment of the trade.
	Bid *float64 `json:"bid"`
	// Ask is the NBBO ask at the moment of the trade.
	Ask *float64 `json:"ask"`
}

// FlowStockBlocksResponse is the typed body of
// GET /v1/flow/stocks/{symbol}/blocks: all trades with size >= minSize.
type FlowStockBlocksResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// MinSize is the minimum trade size that qualified as a block (echoed).
	MinSize *int `json:"minSize"`
	// Count is the number of blocks returned.
	Count *int `json:"count"`
	// Blocks is the newest-first list of large prints.
	Blocks []FlowStockBlock `json:"blocks"`
}

// FlowStockHistoryBucket is one per-minute stock-flow bucket. Like
// FlowHistoryBucket but also carries OHLC of the print price.
type FlowStockHistoryBucket struct {
	// Ts is the bucket start (ISO-8601 UTC, minute-aligned).
	Ts string `json:"ts"`
	// BuyVolume is the buy-classified volume in the bucket.
	BuyVolume *int64 `json:"buyVolume"`
	// SellVolume is the sell-classified volume in the bucket.
	SellVolume *int64 `json:"sellVolume"`
	// MidVolume is the mid-classified volume in the bucket.
	MidVolume *int64 `json:"midVolume"`
	// NetVolume is BuyVolume - SellVolume.
	NetVolume *int64 `json:"netVolume"`
	// TradeCount is the number of trades in the bucket.
	TradeCount *int `json:"tradeCount"`
	// BiggestTrade is the largest single trade size in the bucket.
	BiggestTrade *int `json:"biggestTrade"`
	// Vwap is the volume-weighted average trade price across the bucket.
	Vwap *float64 `json:"vwap"`
	// Open is the first trade price in the bucket.
	Open *float64 `json:"open"`
	// Close is the last trade price in the bucket.
	Close *float64 `json:"close"`
	// High is the highest trade price in the bucket.
	High *float64 `json:"high"`
	// Low is the lowest trade price in the bucket.
	Low *float64 `json:"low"`
}

// FlowStockHistoryResponse is the typed body of
// GET /v1/flow/stocks/{symbol}/history: newest-first per-minute buckets.
type FlowStockHistoryResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Minutes is the lookback window in minutes (echoed back).
	Minutes *int `json:"minutes"`
	// Count is the number of buckets returned.
	Count *int `json:"count"`
	// Buckets is the newest-first list of per-minute aggregates.
	Buckets []FlowStockHistoryBucket `json:"buckets"`
}

// FlowStockCumulativeResponse is the typed body of
// GET /v1/flow/stocks/{symbol}/cumulative.
type FlowStockCumulativeResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Minutes is the lookback window in minutes (echoed back).
	Minutes *int `json:"minutes"`
	// Count is the number of points returned.
	Count *int `json:"count"`
	// Points is the chronological cumulative net-flow series.
	Points []FlowCumulativePoint `json:"points"`
}

// FlowOptionLeaderRow is one ranked underlying in the option-flow
// leaderboard. Option rows carry AvgPremium; the stock leaderboard uses
// Vwap instead.
type FlowOptionLeaderRow struct {
	// Symbol is the ranked underlying.
	Symbol string `json:"symbol"`
	// NetVolume is net contracts (BuyVolume - SellVolume).
	NetVolume *int64 `json:"netVolume"`
	// NetNotional is the net dollar option flow (≈ net contracts × avg
	// premium × 100).
	NetNotional *float64 `json:"netNotional"`
	// BuyVolume is the buy-classified contract volume.
	BuyVolume *int64 `json:"buyVolume"`
	// SellVolume is the sell-classified contract volume.
	SellVolume *int64 `json:"sellVolume"`
	// AvgPremium is the volume-weighted average option premium.
	AvgPremium *float64 `json:"avgPremium"`
	// TradeCount is the number of trades over the window.
	TradeCount *int `json:"tradeCount"`
	// LastTradeUtc is the timestamp of the most recent print.
	LastTradeUtc string `json:"lastTradeUtc"`
}

// FlowOptionLeaderboardResponse is the typed body of
// GET /v1/flow/options/leaderboard: top-N net-dollar buyers and sellers.
type FlowOptionLeaderboardResponse struct {
	// GeneratedUtc is when the cached snapshot was generated (ISO-8601 UTC).
	GeneratedUtc string `json:"generatedUtc"`
	// N is the number of ranked rows requested per side.
	N *int `json:"n"`
	// WindowMinutes is the aggregation window in minutes.
	WindowMinutes *int `json:"windowMinutes"`
	// Buyers is the top net-dollar buyers.
	Buyers []FlowOptionLeaderRow `json:"buyers"`
	// Sellers is the top net-dollar sellers.
	Sellers []FlowOptionLeaderRow `json:"sellers"`
}

// FlowOutlierRow is one flagged underlying in an outliers table (shared by
// the option and stock outliers endpoints).
type FlowOutlierRow struct {
	// Symbol is the flagged underlying.
	Symbol string `json:"symbol"`
	// TradeCount is the number of trades over the window.
	TradeCount *int `json:"tradeCount"`
	// BuyVolume is the buy-classified volume.
	BuyVolume *int64 `json:"buyVolume"`
	// SellVolume is the sell-classified volume.
	SellVolume *int64 `json:"sellVolume"`
	// MidVolume is the mid-classified volume.
	MidVolume *int64 `json:"midVolume"`
	// NetVolume is BuyVolume - SellVolume.
	NetVolume *int64 `json:"netVolume"`
	// ImbalancePct is |buy-sell| / (buy+sell) × 100: 0 = balanced,
	// 100 = one-sided.
	ImbalancePct *float64 `json:"imbalancePct"`
	// Skew is a tiered skew label (FLAT/MILD_BUY/BUY/STRONG_BUY/…).
	Skew string `json:"skew"`
	// Notional is the gross traded notional over the window (dollars).
	Notional *float64 `json:"notional"`
	// NetNotional is the net (signed) traded notional over the window.
	NetNotional *float64 `json:"netNotional"`
	// BiggestTrade is the largest single trade size.
	BiggestTrade *int `json:"biggestTrade"`
	// BiggestTradeUtc is the timestamp of the biggest print; nil if none.
	BiggestTradeUtc *string `json:"biggestTradeUtc"`
	// BiggestAgeSec is the age of the biggest print in seconds; -1 if none.
	BiggestAgeSec *int `json:"biggestAgeSec"`
	// LastVwap is the VWAP of the most recent activity.
	LastVwap *float64 `json:"lastVwap"`
	// LastTradeUtc is the timestamp of the last print; nil if none.
	LastTradeUtc *string `json:"lastTradeUtc"`
	// LastTradeAgeSec is the age of the last print in seconds; -1 if none.
	LastTradeAgeSec *int `json:"lastTradeAgeSec"`
}

// FlowOptionOutliersResponse is the typed body of
// GET /v1/flow/options/outliers.
type FlowOptionOutliersResponse struct {
	// GeneratedUtc is when the cached snapshot was generated (ISO-8601 UTC).
	GeneratedUtc string `json:"generatedUtc"`
	// WindowMinutes is the aggregation window in minutes.
	WindowMinutes *int `json:"windowMinutes"`
	// Tracked is the number of symbols evaluated.
	Tracked *int `json:"tracked"`
	// Qualified is the symbols that met minTrades and had non-zero volume.
	Qualified *int `json:"qualified"`
	// Limit is the max rows requested.
	Limit *int `json:"limit"`
	// Outliers is the imbalance-ranked flagged underlyings.
	Outliers []FlowOutlierRow `json:"outliers"`
}

// FlowStockLeaderRow is one ranked symbol in the stock-flow leaderboard.
// Stock rows carry Vwap; the option leaderboard uses AvgPremium instead.
type FlowStockLeaderRow struct {
	// Symbol is the ranked symbol.
	Symbol string `json:"symbol"`
	// NetVolume is net shares (BuyVolume - SellVolume).
	NetVolume *int64 `json:"netVolume"`
	// NetNotional is the net dollar flow (net shares × VWAP).
	NetNotional *float64 `json:"netNotional"`
	// BuyVolume is the buy-classified share volume.
	BuyVolume *int64 `json:"buyVolume"`
	// SellVolume is the sell-classified share volume.
	SellVolume *int64 `json:"sellVolume"`
	// Vwap is the volume-weighted average trade price over the window.
	Vwap *float64 `json:"vwap"`
	// TradeCount is the number of trades over the window.
	TradeCount *int `json:"tradeCount"`
	// LastTradeUtc is the timestamp of the most recent print.
	LastTradeUtc string `json:"lastTradeUtc"`
}

// FlowStockLeaderboardResponse is the typed body of
// GET /v1/flow/stocks/leaderboard: top-N net-dollar buyers and sellers.
type FlowStockLeaderboardResponse struct {
	// GeneratedUtc is when the cached snapshot was generated (ISO-8601 UTC).
	GeneratedUtc string `json:"generatedUtc"`
	// N is the number of ranked rows requested per side.
	N *int `json:"n"`
	// WindowMinutes is the aggregation window in minutes.
	WindowMinutes *int `json:"windowMinutes"`
	// Buyers is the top net-dollar buyers.
	Buyers []FlowStockLeaderRow `json:"buyers"`
	// Sellers is the top net-dollar sellers.
	Sellers []FlowStockLeaderRow `json:"sellers"`
}

// FlowStockOutliersResponse is the typed body of
// GET /v1/flow/stocks/outliers.
type FlowStockOutliersResponse struct {
	// GeneratedUtc is when the cached snapshot was generated (ISO-8601 UTC).
	GeneratedUtc string `json:"generatedUtc"`
	// WindowMinutes is the aggregation window in minutes.
	WindowMinutes *int `json:"windowMinutes"`
	// Tracked is the number of symbols evaluated.
	Tracked *int `json:"tracked"`
	// Qualified is the symbols that met minTrades and had non-zero volume.
	Qualified *int `json:"qualified"`
	// Limit is the max rows requested.
	Limit *int `json:"limit"`
	// Outliers is the imbalance-ranked flagged symbols.
	Outliers []FlowOutlierRow `json:"outliers"`
}

// ── Untyped client methods (map[string]interface{}) ──────────────────────────

// FlowLevels returns live gamma flip / call & put walls / max pain.
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowLevels(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/levels/"+seg(symbol), params)
}

// FlowPinRisk returns the 0DTE pin-risk score + component breakdown.
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowPinRisk(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/pin-risk/"+seg(symbol), params)
}

// FlowSummary returns at-a-glance flow direction + headline GEX shift.
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowSummary(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/summary/"+seg(symbol), params)
}

// FlowOi returns the open-interest simulator state (official vs intraday).
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowOi(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/oi/"+seg(symbol), params)
}

// FlowGex returns the live (flow-adjusted) GEX + per-strike profile.
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowGex(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/gex/"+seg(symbol), params)
}

// FlowDex returns the live (flow-adjusted) DEX + per-strike profile.
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowDex(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/dex/"+seg(symbol), params)
}

// FlowDealerRisk returns settled-vs-live dealer GEX/DEX + flow adjustment.
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowDealerRisk(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/dealer-risk/"+seg(symbol), params)
}

// FlowLive returns the everything-at-once live flow bundle (convenience).
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowLive(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/live/"+seg(symbol), params)
}

// FlowOptionRecent returns recent option trades, newest-first.
// Requires the Alpha plan. Optional: WithFlowLimit (1–500), WithFlowExpiry.
func (c *Client) FlowOptionRecent(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.limit != nil {
		params.Set("limit", strconv.Itoa(*cfg.limit))
	}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/options/"+seg(symbol)+"/recent", params)
}

// FlowOptionSummary returns per-underlying option-flow aggregates.
// Requires the Alpha plan. Optional: WithFlowExpiry.
func (c *Client) FlowOptionSummary(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/options/"+seg(symbol)+"/summary", params)
}

// FlowOptionBlocks returns large option prints (size >= minSize).
// Requires the Alpha plan. Optional: WithFlowMinSize, WithFlowExpiry.
func (c *Client) FlowOptionBlocks(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.minSize != nil {
		params.Set("minSize", strconv.Itoa(*cfg.minSize))
	}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/options/"+seg(symbol)+"/blocks", params)
}

// FlowOptionHistory returns per-minute option-flow buckets.
// Requires the Alpha plan. Optional: WithFlowMinutes (1–10080), WithFlowExpiry.
func (c *Client) FlowOptionHistory(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/options/"+seg(symbol)+"/history", params)
}

// FlowOptionCumulative returns the cumulative option net-flow series.
// Requires the Alpha plan. Optional: WithFlowMinutes, WithFlowExpiry.
func (c *Client) FlowOptionCumulative(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	if cfg.expiry != "" {
		params.Set("expiry", cfg.expiry)
	}
	return c.get(ctx, "/v1/flow/options/"+seg(symbol)+"/cumulative", params)
}

// FlowStockRecent returns recent stock trades, newest-first.
// Requires the Alpha plan. Optional: WithFlowLimit (1–500).
func (c *Client) FlowStockRecent(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.limit != nil {
		params.Set("limit", strconv.Itoa(*cfg.limit))
	}
	return c.get(ctx, "/v1/flow/stocks/"+seg(symbol)+"/recent", params)
}

// FlowStockSummary returns per-symbol stock-flow aggregates.
// Requires the Alpha plan.
func (c *Client) FlowStockSummary(ctx context.Context, symbol string) (map[string]interface{}, error) {
	return c.get(ctx, "/v1/flow/stocks/"+seg(symbol)+"/summary", nil)
}

// FlowStockBlocks returns large stock prints (size >= minSize).
// Requires the Alpha plan. Optional: WithFlowMinSize.
func (c *Client) FlowStockBlocks(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.minSize != nil {
		params.Set("minSize", strconv.Itoa(*cfg.minSize))
	}
	return c.get(ctx, "/v1/flow/stocks/"+seg(symbol)+"/blocks", params)
}

// FlowStockHistory returns per-minute stock-flow buckets with OHLC.
// Requires the Alpha plan. Optional: WithFlowMinutes (1–10080).
func (c *Client) FlowStockHistory(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	return c.get(ctx, "/v1/flow/stocks/"+seg(symbol)+"/history", params)
}

// FlowStockCumulative returns the cumulative stock net-flow series.
// Requires the Alpha plan. Optional: WithFlowMinutes.
func (c *Client) FlowStockCumulative(ctx context.Context, symbol string, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.minutes != nil {
		params.Set("minutes", strconv.Itoa(*cfg.minutes))
	}
	return c.get(ctx, "/v1/flow/stocks/"+seg(symbol)+"/cumulative", params)
}

// FlowOptionsLeaderboard returns the cross-symbol option-flow leaderboard.
// Requires the Alpha plan. Optional: WithFlowN (1–50), WithFlowWindowMinutes.
func (c *Client) FlowOptionsLeaderboard(ctx context.Context, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.n != nil {
		params.Set("n", strconv.Itoa(*cfg.n))
	}
	if cfg.windowMinutes != nil {
		params.Set("windowMinutes", strconv.Itoa(*cfg.windowMinutes))
	}
	return c.get(ctx, "/v1/flow/options/leaderboard", params)
}

// FlowOptionsOutliers returns the cross-symbol option-flow outliers table.
// Requires the Alpha plan. Optional: WithFlowLimit (1–200),
// WithFlowMinTrades, WithFlowWindowMinutes.
func (c *Client) FlowOptionsOutliers(ctx context.Context, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.limit != nil {
		params.Set("limit", strconv.Itoa(*cfg.limit))
	}
	if cfg.minTrades != nil {
		params.Set("minTrades", strconv.Itoa(*cfg.minTrades))
	}
	if cfg.windowMinutes != nil {
		params.Set("windowMinutes", strconv.Itoa(*cfg.windowMinutes))
	}
	return c.get(ctx, "/v1/flow/options/outliers", params)
}

// FlowStocksLeaderboard returns the cross-symbol stock-flow leaderboard.
// Requires the Alpha plan. Optional: WithFlowN (1–50), WithFlowWindowMinutes.
func (c *Client) FlowStocksLeaderboard(ctx context.Context, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.n != nil {
		params.Set("n", strconv.Itoa(*cfg.n))
	}
	if cfg.windowMinutes != nil {
		params.Set("windowMinutes", strconv.Itoa(*cfg.windowMinutes))
	}
	return c.get(ctx, "/v1/flow/stocks/leaderboard", params)
}

// FlowStocksOutliers returns the cross-symbol stock-flow outliers table.
// Requires the Alpha plan. Optional: WithFlowLimit (1–200),
// WithFlowMinTrades, WithFlowWindowMinutes.
func (c *Client) FlowStocksOutliers(ctx context.Context, opts ...FlowOption) (map[string]interface{}, error) {
	cfg, params := flowParams(opts)
	if cfg.limit != nil {
		params.Set("limit", strconv.Itoa(*cfg.limit))
	}
	if cfg.minTrades != nil {
		params.Set("minTrades", strconv.Itoa(*cfg.minTrades))
	}
	if cfg.windowMinutes != nil {
		params.Set("windowMinutes", strconv.Itoa(*cfg.windowMinutes))
	}
	return c.get(ctx, "/v1/flow/stocks/outliers", params)
}

// ── Strongly-typed wrappers ──────────────────────────────────────────────────

// FlowLevelsTyped is the strongly-typed variant of FlowLevels.
func (c *Client) FlowLevelsTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowLevelsResponse, error) {
	raw, err := c.FlowLevels(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowLevelsResponse{}
	if err := decodeTyped("flow levels", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowPinRiskTyped is the strongly-typed variant of FlowPinRisk.
func (c *Client) FlowPinRiskTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowPinRiskResponse, error) {
	raw, err := c.FlowPinRisk(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowPinRiskResponse{}
	if err := decodeTyped("flow pin-risk", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowSummaryTyped is the strongly-typed variant of FlowSummary.
func (c *Client) FlowSummaryTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowSummaryResponse, error) {
	raw, err := c.FlowSummary(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowSummaryResponse{}
	if err := decodeTyped("flow summary", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOiTyped is the strongly-typed variant of FlowOi.
func (c *Client) FlowOiTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowOiResponse, error) {
	raw, err := c.FlowOi(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOiResponse{}
	if err := decodeTyped("flow oi", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowGexTyped is the strongly-typed variant of FlowGex.
func (c *Client) FlowGexTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowGexResponse, error) {
	raw, err := c.FlowGex(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowGexResponse{}
	if err := decodeTyped("flow gex", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowDexTyped is the strongly-typed variant of FlowDex.
func (c *Client) FlowDexTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowDexResponse, error) {
	raw, err := c.FlowDex(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowDexResponse{}
	if err := decodeTyped("flow dex", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowDealerRiskTyped is the strongly-typed variant of FlowDealerRisk.
func (c *Client) FlowDealerRiskTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowDealerRiskResponse, error) {
	raw, err := c.FlowDealerRisk(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowDealerRiskResponse{}
	if err := decodeTyped("flow dealer-risk", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowLiveTyped is the strongly-typed variant of FlowLive.
func (c *Client) FlowLiveTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowLiveResponse, error) {
	raw, err := c.FlowLive(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowLiveResponse{}
	if err := decodeTyped("flow live", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOptionRecentTyped is the strongly-typed variant of FlowOptionRecent.
func (c *Client) FlowOptionRecentTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowOptionRecentResponse, error) {
	raw, err := c.FlowOptionRecent(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOptionRecentResponse{}
	if err := decodeTyped("flow option recent", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOptionSummaryTyped is the strongly-typed variant of FlowOptionSummary.
func (c *Client) FlowOptionSummaryTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowOptionSummaryResponse, error) {
	raw, err := c.FlowOptionSummary(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOptionSummaryResponse{}
	if err := decodeTyped("flow option summary", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOptionBlocksTyped is the strongly-typed variant of FlowOptionBlocks.
func (c *Client) FlowOptionBlocksTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowOptionBlocksResponse, error) {
	raw, err := c.FlowOptionBlocks(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOptionBlocksResponse{}
	if err := decodeTyped("flow option blocks", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOptionHistoryTyped is the strongly-typed variant of FlowOptionHistory.
func (c *Client) FlowOptionHistoryTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowOptionHistoryResponse, error) {
	raw, err := c.FlowOptionHistory(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOptionHistoryResponse{}
	if err := decodeTyped("flow option history", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOptionCumulativeTyped is the strongly-typed variant of
// FlowOptionCumulative.
func (c *Client) FlowOptionCumulativeTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowOptionCumulativeResponse, error) {
	raw, err := c.FlowOptionCumulative(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOptionCumulativeResponse{}
	if err := decodeTyped("flow option cumulative", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowStockRecentTyped is the strongly-typed variant of FlowStockRecent.
func (c *Client) FlowStockRecentTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowStockRecentResponse, error) {
	raw, err := c.FlowStockRecent(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowStockRecentResponse{}
	if err := decodeTyped("flow stock recent", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowStockSummaryTyped is the strongly-typed variant of FlowStockSummary.
func (c *Client) FlowStockSummaryTyped(ctx context.Context, symbol string) (*FlowStockSummaryResponse, error) {
	raw, err := c.FlowStockSummary(ctx, symbol)
	if err != nil {
		return nil, err
	}
	out := &FlowStockSummaryResponse{}
	if err := decodeTyped("flow stock summary", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowStockBlocksTyped is the strongly-typed variant of FlowStockBlocks.
func (c *Client) FlowStockBlocksTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowStockBlocksResponse, error) {
	raw, err := c.FlowStockBlocks(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowStockBlocksResponse{}
	if err := decodeTyped("flow stock blocks", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowStockHistoryTyped is the strongly-typed variant of FlowStockHistory.
func (c *Client) FlowStockHistoryTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowStockHistoryResponse, error) {
	raw, err := c.FlowStockHistory(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowStockHistoryResponse{}
	if err := decodeTyped("flow stock history", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowStockCumulativeTyped is the strongly-typed variant of
// FlowStockCumulative.
func (c *Client) FlowStockCumulativeTyped(ctx context.Context, symbol string, opts ...FlowOption) (*FlowStockCumulativeResponse, error) {
	raw, err := c.FlowStockCumulative(ctx, symbol, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowStockCumulativeResponse{}
	if err := decodeTyped("flow stock cumulative", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOptionsLeaderboardTyped is the strongly-typed variant of
// FlowOptionsLeaderboard.
func (c *Client) FlowOptionsLeaderboardTyped(ctx context.Context, opts ...FlowOption) (*FlowOptionLeaderboardResponse, error) {
	raw, err := c.FlowOptionsLeaderboard(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOptionLeaderboardResponse{}
	if err := decodeTyped("flow options leaderboard", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowOptionsOutliersTyped is the strongly-typed variant of
// FlowOptionsOutliers.
func (c *Client) FlowOptionsOutliersTyped(ctx context.Context, opts ...FlowOption) (*FlowOptionOutliersResponse, error) {
	raw, err := c.FlowOptionsOutliers(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowOptionOutliersResponse{}
	if err := decodeTyped("flow options outliers", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowStocksLeaderboardTyped is the strongly-typed variant of
// FlowStocksLeaderboard.
func (c *Client) FlowStocksLeaderboardTyped(ctx context.Context, opts ...FlowOption) (*FlowStockLeaderboardResponse, error) {
	raw, err := c.FlowStocksLeaderboard(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowStockLeaderboardResponse{}
	if err := decodeTyped("flow stocks leaderboard", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowStocksOutliersTyped is the strongly-typed variant of
// FlowStocksOutliers.
func (c *Client) FlowStocksOutliersTyped(ctx context.Context, opts ...FlowOption) (*FlowStockOutliersResponse, error) {
	raw, err := c.FlowStocksOutliers(ctx, opts...)
	if err != nil {
		return nil, err
	}
	out := &FlowStockOutliersResponse{}
	if err := decodeTyped("flow stocks outliers", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
