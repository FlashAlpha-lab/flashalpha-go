package flashalpha

// Typed response model for `GET /v1/exposure/summary/{symbol}`.
//
// All numeric fields are *float64 / *int / *string so that nil represents
// values the API could not compute (insufficient data, market closed, etc.).
//
// Direction casing: /v1/exposure/summary/ and /v1/exposure/zero-dte/ both
// return lowercase "buy" / "sell". Docs and typed models use that casing
// consistently.

// ExposureSummaryResponse is the typed body of GET /v1/exposure/summary/{symbol}.
//
// One round-trip returns net dealer Greeks (gamma/delta/vanna/charm) across
// the entire chain, the gamma-flip strike, the dealer hedging-flow estimate
// at +/- 1% spot moves, verbal regime narratives, and a 0DTE attribution.
type ExposureSummaryResponse struct {
	// Underlying symbol echoed from the request path (e.g. "SPY").
	Symbol string `json:"symbol"`
	// Current spot mid in dollars. The reference price for all GEX/DEX/VEX/CHEX dollarisation.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// ET wall-clock timestamp this snapshot was computed for.
	AsOf string `json:"as_of"`
	// Strike where net dealer gamma exposure crosses zero. Spot ABOVE the
	// flip = positive-gamma regime (dealers dampen moves, mean-reversion
	// likely). Spot BELOW = negative-gamma (dealers amplify moves, trend-
	// following likely). One of the two or three numbers most experienced
	// users actually look at on this endpoint.
	GammaFlip *float64 `json:"gamma_flip"`
	// Dealer-positioning regime classifier. Confirmed live values across all
	// 5 SDK integration test suites:
	//   "positive_gamma" | "negative_gamma" | "neutral"
	// Documented fourth: "undetermined" (no usable options data). "neutral"
	// appears in edge cases where net_gex straddles zero. Don't conflate
	// with maxpain.signal (a separate bullish/bearish/neutral classifier).
	Regime string `json:"regime"`
	// Net Greek totals across the entire chain. See ExposureSummaryExposures.
	Exposures *ExposureSummaryExposures `json:"exposures"`
	// Plain-English narrative for each Greek regime — safe to surface to
	// end users verbatim.
	Interpretation *ExposureSummaryInterpretation `json:"interpretation"`
	// Estimated dealer hedging flow at +/- 1% spot moves.
	HedgingEstimate *ExposureSummaryHedgingEstimate `json:"hedging_estimate"`
	// Same-day-expiration contribution to total GEX.
	ZeroDte *ExposureSummaryZeroDte `json:"zero_dte"`
}

// ExposureSummaryExposures aggregates net GEX/DEX/VEX/CHEX across the chain.
//
// Each value is computed as Σ greek × OI × multiplier × spot_factor over
// every contract in the chain. Sign convention: positive means dealers are
// net long that Greek, negative means net short.
type ExposureSummaryExposures struct {
	// Net gamma exposure in dollars per 1% spot move. Positive (dealers
	// long gamma) → moves dampened, mean-reversion likely. Negative (short
	// gamma) → moves amplified, trend-following likely.
	NetGex *float64 `json:"net_gex"`
	// Net delta exposure in dollars. Sign is the direction of the dealer
	// hedge book against options inventory. Negative ≈ dealers net-short
	// the underlying as a delta hedge against long-call inventory.
	NetDex *float64 `json:"net_dex"`
	// Net vanna exposure in dollars per 1-vol-point. Positive = dealers
	// benefit from vol compression and tend to BUY stock when vol drops
	// (vanna-driven supportive bid); negative is the inverse.
	NetVex *float64 `json:"net_vex"`
	// Net charm exposure in dollars per day. Captures the time-decay drift
	// in dealer delta. Positive = dealers must BUY into close to stay
	// neutral (supportive); negative = SELL into close (pressure).
	NetChex *float64 `json:"net_chex"`
}

// ExposureSummaryInterpretation holds the verbal gamma/vanna/charm regime
// interpretations. Generated server-side from the numeric exposures and
// macro context; safe to surface verbatim in customer-facing UIs.
type ExposureSummaryInterpretation struct {
	// E.g. "Dealers long gamma — moves dampened, mean reversion likely".
	Gamma string `json:"gamma"`
	// E.g. "Vol up = dealers buy delta — downside dampened if vol spikes".
	// Conditional on prevailing VIX.
	Vanna string `json:"vanna"`
	// E.g. "Time decay pushing dealers to sell — pressure into close".
	Charm string `json:"charm"`
}

// ExposureSummaryHedgingMove is one side (up or down) of a dealer-hedging
// estimate. Direction is "buy" or "sell" (lowercase on both this endpoint
// and zero-dte).
//
// Estimates the order flow dealers would generate to remain delta-neutral
// if spot moved by 1%. Use this as a sizing reference for intraday momentum
// / mean-reversion setups: large positive DealerSharesToTrade = lots of
// forced BUYING by dealers if spot rises = self-reinforcing momentum.
type ExposureSummaryHedgingMove struct {
	// Estimated shares dealers must trade. Positive = buy, negative = sell.
	// SpotUp1Pct and SpotDown1Pct are equal in magnitude with opposite
	// signs (linearised from net_dex around current spot).
	DealerSharesToTrade *float64 `json:"dealer_shares_to_trade"`
	// Lowercase "buy" or "sell" — convenience label matching the sign of
	// DealerSharesToTrade.
	Direction string `json:"direction"`
	// |DealerSharesToTrade| × current_spot. Useful for cross-symbol
	// comparison: a 1M-share hedge in SPY ($600 spot) is much larger than
	// a 1M-share hedge in HOOD ($30 spot), and notional captures that.
	NotionalUsd *float64 `json:"notional_usd"`
}

// ExposureSummaryHedgingEstimate holds the estimated dealer hedging flow at
// +/- 1% spot moves. The two sides are symmetric: equal magnitude, opposite
// signs (linearised from net_dex). For a fully nonlinear view at multiple
// move sizes use the zero-dte endpoint (±10bp / ±25bp / ±50bp / ±1pct).
type ExposureSummaryHedgingEstimate struct {
	// Hedging flow if spot rises 1%.
	SpotUp1Pct *ExposureSummaryHedgingMove `json:"spot_up_1pct"`
	// Hedging flow if spot falls 1%. Equal magnitude to SpotUp1Pct, opposite sign.
	SpotDown1Pct *ExposureSummaryHedgingMove `json:"spot_down_1pct"`
}

// ExposureSummaryZeroDte is the same-day-expiration contribution to total GEX.
//
// 0DTE GEX is often the dominant intraday driver — gamma compresses to a
// delta function as expiry approaches, so even a small notional 0DTE book
// can swamp the rest of the chain in dealer-flow terms.
type ExposureSummaryZeroDte struct {
	// Net GEX contribution from same-day-expiration contracts only.
	NetGex *float64 `json:"net_gex"`
	// 0DTE share of full-chain GEX as a percentage. >50% means today's
	// expiry drives the dealer book; tradable signal.
	PctOfTotalGex *float64 `json:"pct_of_total_gex"`
	// ISO date of today's 0DTE if one exists (yyyy-MM-dd). nil on days
	// without a same-day expiry.
	Expiration *string `json:"expiration"`
}
