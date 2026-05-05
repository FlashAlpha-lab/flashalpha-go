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
type ExposureSummaryResponse struct {
	Symbol          string   `json:"symbol"`
	UnderlyingPrice *float64 `json:"underlying_price"`
	AsOf            string   `json:"as_of"`
	GammaFlip       *float64 `json:"gamma_flip"`
	// Confirmed live values across all 5 SDK integration test suites:
	//   positive_gamma | negative_gamma | neutral
	// Documented fourth: undetermined (no usable options data). `neutral`
	// appears in edge cases where net_gex straddles zero.
	Regime          string                          `json:"regime"`
	Exposures       *ExposureSummaryExposures       `json:"exposures"`
	Interpretation  *ExposureSummaryInterpretation  `json:"interpretation"`
	HedgingEstimate *ExposureSummaryHedgingEstimate `json:"hedging_estimate"`
	ZeroDte         *ExposureSummaryZeroDte         `json:"zero_dte"`
}

// ExposureSummaryExposures aggregates net GEX/DEX/VEX/CHEX across the chain.
type ExposureSummaryExposures struct {
	NetGex  *float64 `json:"net_gex"`
	NetDex  *float64 `json:"net_dex"`
	NetVex  *float64 `json:"net_vex"`
	NetChex *float64 `json:"net_chex"`
}

// ExposureSummaryInterpretation holds the verbal gamma/vanna/charm regime
// interpretations.
type ExposureSummaryInterpretation struct {
	Gamma string `json:"gamma"`
	Vanna string `json:"vanna"`
	Charm string `json:"charm"`
}

// ExposureSummaryHedgingMove is one side (up or down) of a dealer-hedging
// estimate. Direction is "buy" or "sell" (lowercase on both this endpoint
// and zero-dte).
type ExposureSummaryHedgingMove struct {
	DealerSharesToTrade *float64 `json:"dealer_shares_to_trade"`
	Direction           string   `json:"direction"`
	NotionalUsd         *float64 `json:"notional_usd"`
}

// ExposureSummaryHedgingEstimate holds the estimated dealer hedging flow at
// +/- 1% spot moves.
type ExposureSummaryHedgingEstimate struct {
	SpotUp1Pct   *ExposureSummaryHedgingMove `json:"spot_up_1pct"`
	SpotDown1Pct *ExposureSummaryHedgingMove `json:"spot_down_1pct"`
}

// ExposureSummaryZeroDte is the same-day-expiration contribution to total GEX.
type ExposureSummaryZeroDte struct {
	NetGex        *float64 `json:"net_gex"`
	PctOfTotalGex *float64 `json:"pct_of_total_gex"`
	Expiration    *string  `json:"expiration"`
}
