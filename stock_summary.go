package flashalpha

// Typed response model for `GET /v1/stock/{symbol}/summary`.
//
// Composite snapshot of price, volatility, options flow, dealer exposure, and
// macro context for an underlying — one round-trip in lieu of calling the
// stock-quote, volatility, options-flow, exposure-summary, and VIX-context
// endpoints separately. Designed for dashboards and LLM tool-use where one
// JSON blob per symbol is more useful than five.
//
// Class-level: dual mode — authenticated requests return a LIVE snapshot;
// unauthenticated requests return the previous-day cached snapshot. The
// payload shape is identical in both modes.
//
// Nullability conventions:
//   - All numeric fields are *float64 / *int / *string so nil represents
//     values the API could not compute (insufficient data, market closed,
//     external data source unavailable, etc.).
//   - Top-level Exposure is nil when the symbol has no options/greeks loaded
//     (small-caps, ETFs without listed options, freshly listed names).
//   - Macro fields can each be nil independently when an external feed
//     (CBOE VIX/VVIX/SKEW/MOVE, Polygon SPX, FRED) is unavailable.
//
// IMPORTANT — sign convention diff vs zero-dte:
//   On THIS endpoint, Exposure.HedgingEstimate.{SpotUp1Pct,SpotDown1Pct}
//   .DealerShares is the MAGNITUDE (always non-negative). The signed direction
//   is carried by the sibling Direction string ("buy" / "sell"). The
//   /v1/exposure/zero-dte response uses signed values for its hedging buckets
//   instead — don't conflate them.

// StockSummaryResponse is the typed body of GET /v1/stock/{symbol}/summary.
type StockSummaryResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// UnderlyingPrice is the spot mid in dollars at AsOf — the reference price
	// for all dollar-denominated fields below.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// AsOf is the ET wall-clock timestamp this snapshot was computed for.
	AsOf string `json:"as_of"`
	// MarketOpen is true if NYSE was open at AsOf. False overnight, weekends,
	// and holidays — when MarketOpen=false the response is a previous-day
	// cached snapshot.
	MarketOpen bool `json:"market_open"`
	// Price is the bid/ask/mid/last quote block. See StockSummaryPrice.
	Price *StockSummaryPrice `json:"price"`
	// Volatility is the IV / historical-vol / VRP / skew / term-structure
	// block. See StockSummaryVolatility.
	Volatility *StockSummaryVolatility `json:"volatility"`
	// OptionsFlow is the chain-wide OI + volume aggregates. See
	// StockSummaryOptionsFlow.
	OptionsFlow *StockSummaryOptionsFlow `json:"options_flow"`
	// Exposure is the dealer-Greek block (GEX/DEX/VEX/CHEX, key levels,
	// hedging estimate, 0DTE attribution). See StockSummaryExposure.
	//
	// Nil when no options/greeks data is loaded for this symbol (small-caps,
	// ETFs without listed options, etc.).
	Exposure *StockSummaryExposure `json:"exposure"`
	// Macro is the macro-context block (VIX, VVIX, SKEW, SPX, MOVE, term
	// structure, futures, fear & greed). Individual fields can be nil when
	// the underlying external data source is unavailable. See StockSummaryMacro.
	Macro *StockSummaryMacro `json:"macro"`
}

// StockSummaryPrice is the bid/ask/mid/last quote block.
type StockSummaryPrice struct {
	// Bid is the NBBO bid price.
	Bid *float64 `json:"bid"`
	// Ask is the NBBO ask price.
	Ask *float64 `json:"ask"`
	// Mid is the NBBO mid (Bid + Ask) / 2 — the canonical reference price.
	Mid *float64 `json:"mid"`
	// Last is the last trade price.
	Last *float64 `json:"last"`
	// LastUpdate is the ET wall-clock timestamp of the last quote update.
	LastUpdate *string `json:"last_update"`
}

// StockSummaryVolatility is the IV / historical-vol / VRP / skew / term-
// structure block.
//
// All vol fields are PERCENT — e.g. AtmIv = 18.45 means 18.45% annualised IV,
// NOT 0.1845 decimal. Same convention across AtmIv, Hv20, Hv60, Vrp, and the
// term-structure / skew sub-blocks.
type StockSummaryVolatility struct {
	// AtmIv is the at-the-money implied volatility (annualised %, e.g. 18.45
	// = 18.45%).
	AtmIv *float64 `json:"atm_iv"`
	// Hv20 is the trailing 20-day realized vol (annualised %).
	Hv20 *float64 `json:"hv_20"`
	// Hv60 is the trailing 60-day realized vol (annualised %).
	Hv60 *float64 `json:"hv_60"`
	// Vrp is the variance risk premium = AtmIv - Hv20 (percentage points).
	// Positive = IV richer than realized → premium for selling vol.
	Vrp *float64 `json:"vrp"`
	// Skew25d is the 25-delta wing skew block — put wing IV vs ATM vs call
	// wing IV at the front-month expiry.
	Skew25d *StockSummarySkew25d `json:"skew_25d"`
	// IvTermStructure is the at-the-money IV at each available expiration —
	// gives the contango/backwardation view across the term.
	IvTermStructure []StockSummaryIvTermPoint `json:"iv_term_structure"`
}

// StockSummarySkew25d is the 25-delta wing skew block (put / ATM / call IVs +
// composite skew + smile ratio).
type StockSummarySkew25d struct {
	// Expiry is the ISO date (yyyy-MM-dd) used for the skew measurement
	// (typically the front-month).
	Expiry *string `json:"expiry"`
	// DaysToExpiry is calendar days from AsOf to Expiry.
	DaysToExpiry *int `json:"days_to_expiry"`
	// Put25dIv is the IV at the 25-delta put wing (annualised %).
	Put25dIv *float64 `json:"put_25d_iv"`
	// AtmIv is the at-the-money IV at this expiry (annualised %).
	AtmIv *float64 `json:"atm_iv"`
	// Call25dIv is the IV at the 25-delta call wing (annualised %).
	Call25dIv *float64 `json:"call_25d_iv"`
	// Skew25d is Put25dIv - Call25dIv in vol points. Positive = downside
	// skew (puts pricing richer than calls — classic equity-index pattern).
	Skew25d *float64 `json:"skew_25d"`
	// SmileRatio is (Put25dIv + Call25dIv) / (2 * AtmIv) — a measure of how
	// curved the smile is. >1 = wings priced richer than ATM (typical);
	// closer to 1 = flat smile.
	SmileRatio *float64 `json:"smile_ratio"`
}

// StockSummaryIvTermPoint is one point on the IV term structure curve.
type StockSummaryIvTermPoint struct {
	// Expiry is the ISO date (yyyy-MM-dd) of this expiration.
	Expiry *string `json:"expiry"`
	// Iv is the at-the-money IV at this expiry (annualised %).
	Iv *float64 `json:"iv"`
	// DaysToExpiry is calendar days from AsOf to Expiry.
	DaysToExpiry *int `json:"days_to_expiry"`
}

// StockSummaryOptionsFlow is the chain-wide OI + volume aggregate block.
type StockSummaryOptionsFlow struct {
	// TotalCallOi is total open interest across all call contracts in the chain.
	TotalCallOi *int `json:"total_call_oi"`
	// TotalPutOi is total open interest across all put contracts in the chain.
	TotalPutOi *int `json:"total_put_oi"`
	// TotalCallVolume is total session volume across all call contracts.
	TotalCallVolume *int `json:"total_call_volume"`
	// TotalPutVolume is total session volume across all put contracts.
	TotalPutVolume *int `json:"total_put_volume"`
	// PcRatioOi is TotalPutOi / TotalCallOi — positioning lean. >1 = put-heavy.
	PcRatioOi *float64 `json:"pc_ratio_oi"`
	// PcRatioVolume is TotalPutVolume / TotalCallVolume — flow lean.
	PcRatioVolume *float64 `json:"pc_ratio_volume"`
	// ActiveExpirations is the count of expirations with non-zero OI.
	ActiveExpirations *int `json:"active_expirations"`
}

// StockSummaryExposure is the dealer-Greek block — net GEX/DEX/VEX/CHEX, key
// levels, regime label, hedging estimates, 0DTE attribution, and per-strike
// top.
//
// Nil at the top level when no options data is loaded.
type StockSummaryExposure struct {
	// NetGex is net dealer gamma exposure in dollars per 1% spot move.
	// Positive (dealers long gamma) → moves dampened, mean-reversion likely.
	// Negative (short gamma) → moves amplified, trend-following likely.
	NetGex *float64 `json:"net_gex"`
	// NetDex is net dealer delta exposure in dollars.
	NetDex *float64 `json:"net_dex"`
	// NetVex is net dealer vanna exposure in dollars per 1-vol-point.
	NetVex *float64 `json:"net_vex"`
	// NetChex is net dealer charm exposure in dollars per day.
	NetChex *float64 `json:"net_chex"`
	// GammaFlip is the strike where net dealer gamma crosses zero.
	GammaFlip *float64 `json:"gamma_flip"`
	// CallWall is the strike with the largest absolute call GEX (dealer-side
	// resistance).
	CallWall *float64 `json:"call_wall"`
	// PutWall is the strike with the largest absolute put GEX (dealer-side
	// support).
	PutWall *float64 `json:"put_wall"`
	// MaxPain is the strike where total option-holder intrinsic value across
	// the chain is minimized.
	MaxPain *float64 `json:"max_pain"`
	// HighestOiStrike is the strike with the largest total OI (calls + puts).
	HighestOiStrike *float64 `json:"highest_oi_strike"`
	// Regime is the dealer-positioning classifier. Confirmed values:
	//   "positive_gamma" | "negative_gamma" | "undetermined"
	Regime string `json:"regime"`
	// Interpretation is the plain-English narrative for each Greek regime —
	// safe to surface verbatim.
	Interpretation *StockSummaryInterpretation `json:"interpretation"`
	// HedgingEstimate is the estimated dealer hedging flow at +/- 1% spot
	// moves.
	//
	// IMPORTANT: on this endpoint DealerShares is the MAGNITUDE (always
	// non-negative). The Direction string ("buy" / "sell") carries the sign.
	// This differs from /v1/exposure/zero-dte where the equivalent buckets
	// store signed values directly.
	HedgingEstimate *StockSummaryHedgingEstimate `json:"hedging_estimate"`
	// ZeroDte is the same-day-expiry attribution block.
	ZeroDte *StockSummaryExposureZeroDte `json:"zero_dte"`
	// TopStrikes is the per-strike top of the dealer-gamma footprint —
	// largest |NetGex| strikes with their OI breakdown.
	TopStrikes []StockSummaryTopStrike `json:"top_strikes"`
	// OiWeightedDte is the OI-weighted average days-to-expiry across the
	// chain — positioning-duration metric.
	OiWeightedDte *float64 `json:"oi_weighted_dte"`
}

// StockSummaryInterpretation holds the verbal Greek-regime narratives — safe
// to surface verbatim in customer-facing UIs.
type StockSummaryInterpretation struct {
	// Gamma is the gamma-regime narrative
	// (e.g. "Dealers long gamma — moves dampened, mean reversion likely").
	Gamma string `json:"gamma"`
	// Vanna is the vanna-regime narrative.
	Vanna string `json:"vanna"`
	// Charm is the charm-regime narrative.
	Charm string `json:"charm"`
}

// StockSummaryHedgingMove is one side (up or down) of the dealer-hedging
// estimate.
//
// IMPORTANT: DealerShares is the MAGNITUDE (always non-negative on this
// endpoint). The signed direction is carried by the Direction string.
type StockSummaryHedgingMove struct {
	// DealerShares is the estimated MAGNITUDE of underlying shares dealers
	// must trade to remain delta-neutral if spot moved by 1%. Always
	// non-negative on this endpoint — see Direction for the sign.
	DealerShares *float64 `json:"dealer_shares"`
	// Direction is the lowercase signed label — "buy" or "sell". Combine
	// with DealerShares to recover the signed flow.
	Direction string `json:"direction"`
	// NotionalUsd is DealerShares × current_spot — the dollar size of the
	// hedge.
	NotionalUsd *float64 `json:"notional_usd"`
}

// StockSummaryHedgingEstimate holds the estimated dealer hedging flow at
// +/- 1% spot moves. Symmetric in magnitude (linearised from net_dex).
type StockSummaryHedgingEstimate struct {
	// SpotUp1Pct is the hedging flow if spot rises 1%.
	SpotUp1Pct *StockSummaryHedgingMove `json:"spot_up_1pct"`
	// SpotDown1Pct is the hedging flow if spot falls 1%. Equal magnitude to
	// SpotUp1Pct, opposite Direction.
	SpotDown1Pct *StockSummaryHedgingMove `json:"spot_down_1pct"`
}

// StockSummaryExposureZeroDte is the same-day-expiry contribution block.
type StockSummaryExposureZeroDte struct {
	// NetGex is net 0DTE gamma exposure ($/1% spot move).
	NetGex *float64 `json:"net_gex"`
	// PctOfTotal is the 0DTE share of full-chain net GEX (%).
	PctOfTotal *float64 `json:"pct_of_total"`
	// Expiration is the ISO date (yyyy-MM-dd) of today's 0DTE if one exists.
	Expiration *string `json:"expiration"`
}

// StockSummaryTopStrike is one row in Exposure.TopStrikes — the largest
// |NetGex| strikes with their OI breakdown.
type StockSummaryTopStrike struct {
	// Strike is the strike price.
	Strike *float64 `json:"strike"`
	// NetGex is the dealer gamma exposure at this strike ($/1% spot move).
	NetGex *float64 `json:"net_gex"`
	// CallOi is open interest on the call side at this strike.
	CallOi *int `json:"call_oi"`
	// PutOi is open interest on the put side at this strike.
	PutOi *int `json:"put_oi"`
	// TotalOi is CallOi + PutOi.
	TotalOi *int `json:"total_oi"`
}

// StockSummaryMacro is the macro-context block.
//
// Each field is independently nullable — when an external data source (CBOE
// VIX/VVIX/SKEW/MOVE feed, Polygon SPX, FRED) is unavailable, the
// corresponding field is nil while the rest of the response is unaffected.
type StockSummaryMacro struct {
	// Vix is the CBOE VIX index level/change/change%.
	Vix *StockSummaryMacroQuote `json:"vix"`
	// Vvix is the VVIX (vol of vol).
	Vvix *StockSummaryMacroQuote `json:"vvix"`
	// Skew is the CBOE SKEW index.
	Skew *StockSummaryMacroQuote `json:"skew"`
	// Spx is the S&P 500 index.
	Spx *StockSummaryMacroQuote `json:"spx"`
	// Move is the ICE BofA MOVE index (Treasury vol).
	Move *StockSummaryMacroQuote `json:"move"`
	// VixTermStructure is the VIX9D / VIX / VIX3M / VIX6M curve plus a
	// near-slope and contango/backwardation classifier.
	VixTermStructure *StockSummaryVixTermStructure `json:"vix_term_structure"`
	// VixFutures is the VIX-futures basis block (front-month vs spot).
	VixFutures *StockSummaryVixFutures `json:"vix_futures"`
	// FearAndGreed is the CNN Fear & Greed score / rating.
	FearAndGreed *StockSummaryFearAndGreed `json:"fear_and_greed"`
}

// StockSummaryMacroQuote is the value/change/change% triple used for each
// macro index.
type StockSummaryMacroQuote struct {
	// Value is the current index level.
	Value *float64 `json:"value"`
	// Change is the absolute change vs prior session close.
	Change *float64 `json:"change"`
	// ChangePct is the percent change vs prior session close (e.g. 1.25 = +1.25%).
	ChangePct *float64 `json:"change_pct"`
}

// StockSummaryVixTermStructure is the VIX term-structure block.
type StockSummaryVixTermStructure struct {
	// Levels are the per-tenor VIX levels (VIX9D, VIX, VIX3M, VIX6M).
	Levels *StockSummaryVixTermLevels `json:"levels"`
	// NearSlopePct is the near-term slope: (Vix3m - Vix) / Vix * 100.
	// Positive = contango (VIX3M > VIX, calmer near-term); negative =
	// backwardation (VIX > VIX3M, near-term stress).
	NearSlopePct *float64 `json:"near_slope_pct"`
	// Structure is the term-structure classifier: "contango" |
	// "backwardation" | "flat".
	Structure *string `json:"structure"`
}

// StockSummaryVixTermLevels is the VIX9D / VIX / VIX3M / VIX6M tenor levels.
type StockSummaryVixTermLevels struct {
	// Vix9d is the CBOE VIX9D (9-day forward variance).
	Vix9d *float64 `json:"vix_9d"`
	// Vix is the CBOE VIX (30-day forward variance).
	Vix *float64 `json:"vix"`
	// Vix3m is the CBOE VIX3M (3-month forward variance).
	Vix3m *float64 `json:"vix_3m"`
	// Vix6m is the CBOE VIX6M (6-month forward variance).
	Vix6m *float64 `json:"vix_6m"`
}

// StockSummaryVixFutures is the VIX-futures basis block.
//
// IMPORTANT: Basis on this endpoint is APPROXIMATED from VIX3M vs VIX spot —
// it is NOT computed from actual front-month VIX-futures prices. Treat as a
// proxy.
type StockSummaryVixFutures struct {
	// FrontMonth is the proxy front-month VIX futures level (uses VIX3M as a
	// stand-in).
	FrontMonth *float64 `json:"front_month"`
	// Spot is the VIX spot level for reference.
	Spot *float64 `json:"spot"`
	// Spread is FrontMonth - Spot (vol points).
	Spread *float64 `json:"spread"`
	// BasisPct is the basis as a percent of spot: (FrontMonth - Spot) / Spot * 100.
	BasisPct *float64 `json:"basis_pct"`
	// Basis is the descriptive label: "contango" | "backwardation" | "flat".
	Basis *string `json:"basis"`
}

// StockSummaryFearAndGreed is the CNN Fear & Greed indicator.
type StockSummaryFearAndGreed struct {
	// Score is the 0-100 composite score (0 = extreme fear, 100 = extreme greed).
	Score *int `json:"score"`
	// Rating is the descriptive label (e.g. "Extreme Fear", "Greed").
	Rating *string `json:"rating"`
}
