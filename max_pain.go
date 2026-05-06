package flashalpha

// Typed response model for `GET /v1/maxpain/{symbol}` (Basic+).
//
// Max pain is the strike where total option-holder intrinsic value across
// all OI in the chain is minimized — equivalently, the strike at which
// dealers (the counterparty) lose the least to expiring contracts. The
// endpoint also overlays GEX-based dealer alignment, a multi-expiry
// calendar (full chain only), and a 0-100 pin probability score.
//
// The endpoint accepts an optional `expiration` query filter (yyyy-MM-dd).
// When present, the response is scoped to that single expiry and
// MaxPainByExpiration is nil. When absent, the full-chain max pain is
// returned alongside the multi-expiry calendar.
//
// Returns 403 tier_restricted for Free-tier users.

// MaxPainResponse is the typed body of GET /v1/maxpain/{symbol}.
type MaxPainResponse struct {
	Symbol          string   `json:"symbol"`
	UnderlyingPrice *float64 `json:"underlying_price"`
	AsOf            string   `json:"as_of"`
	// The headline number. Strike where total chain pain is minimized.
	MaxPainStrike *float64 `json:"max_pain_strike"`
	// Distance from spot to MaxPainStrike (absolute, percent, direction).
	Distance *MaxPainDistance `json:"distance"`
	// "bullish" (spot >= 5% below max_pain — pin attracts upside),
	// "bearish" (>= 5% above), or "neutral" (within 5%).
	Signal *string `json:"signal"`
	// Expiration this view is scoped to. When the request omits the
	// `expiration` filter, this is the front-month expiry that the
	// full-chain max pain landed on.
	Expiration *string `json:"expiration"`
	// Total put OI / total call OI. >1.0 = put-heavy chain.
	PutCallOiRatio *float64 `json:"put_call_oi_ratio"`
	// Strike-by-strike pain curve. Minimum is at MaxPainStrike.
	PainCurve []MaxPainCurveRow `json:"pain_curve"`
	// Per-strike OI + volume breakdown. Same strike grid as PainCurve.
	OiByStrike []MaxPainOiRow `json:"oi_by_strike"`
	// Per-expiry calendar. nil when the request specified an expiry.
	MaxPainByExpiration []MaxPainByExpirationRow `json:"max_pain_by_expiration"`
	// GEX-based dealer alignment overlay.
	DealerAlignment *MaxPainDealerAlignment `json:"dealer_alignment"`
	// Same gamma classification as on exposure_summary.regime:
	// "positive_gamma" | "negative_gamma" | "unknown".
	Regime *string `json:"regime"`
	// Expected move from the ATM straddle, contextualized vs max pain.
	ExpectedMove *MaxPainExpectedMove `json:"expected_move"`
	// 0-100 composite — likelihood of pinning to MaxPainStrike. Inputs:
	// OI concentration (30%), magnet proximity (25%), time remaining (25%),
	// gamma magnitude (20%). Most meaningful for near-term expiries.
	PinProbability *int `json:"pin_probability"`
}

// MaxPainDistance is the distance from spot to the max-pain strike.
type MaxPainDistance struct {
	// Dollar distance: |underlying_price - max_pain_strike|.
	Absolute *float64 `json:"absolute"`
	// Percent of spot: absolute / underlying_price * 100.
	Percent *float64 `json:"percent"`
	// "above", "below", or "at" — spot relative to max-pain.
	Direction *string `json:"direction"`
}

// MaxPainCurveRow is one row of the strike-by-strike pain curve.
//
// Each row is the dollar pain (intrinsic value × OI × 100 contract
// multiplier) summed across all expirations at that strike. The strike
// where TotalPain is minimized is the max-pain strike.
type MaxPainCurveRow struct {
	Strike *float64 `json:"strike"`
	// Dollar intrinsic value of all calls at this strike.
	CallPain *float64 `json:"call_pain"`
	// Dollar intrinsic value of all puts at this strike.
	PutPain *float64 `json:"put_pain"`
	// CallPain + PutPain. The pain curve's minimum identifies max pain.
	TotalPain *float64 `json:"total_pain"`
}

// MaxPainOiRow is one row of the OI-by-strike breakdown.
type MaxPainOiRow struct {
	Strike     *float64 `json:"strike"`
	CallOi     *int     `json:"call_oi"`
	PutOi      *int     `json:"put_oi"`
	TotalOi    *int     `json:"total_oi"`
	CallVolume *int     `json:"call_volume"`
	PutVolume  *int     `json:"put_volume"`
}

// MaxPainByExpirationRow is one row of the per-expiry max-pain breakdown
// returned when no `expiration` filter is applied.
//
// The parent slice is nil when the request specified an expiration filter
// — the response is then scoped to that single expiry and the multi-expiry
// view is suppressed.
type MaxPainByExpirationRow struct {
	Expiration    *string  `json:"expiration"`
	MaxPainStrike *float64 `json:"max_pain_strike"`
	// Days to expiry (counting from AsOf).
	Dte     *int `json:"dte"`
	TotalOi *int `json:"total_oi"`
}

// MaxPainDealerAlignment is the GEX-based dealer-alignment overlay on the
// max-pain view. The Alignment label tells you whether dealer hedging will
// REINFORCE the max-pain pin or fight it:
//   - "converging": max pain near gamma flip and between walls — strongest pin.
//   - "moderate":   between walls but far from flip.
//   - "diverging":  max pain outside the wall range.
//   - "unknown":    insufficient data.
type MaxPainDealerAlignment struct {
	Alignment *string `json:"alignment"`
	// Plain-English explanation. Safe to surface verbatim.
	Description *string `json:"description"`
	// Strike where net dealer gamma crosses zero.
	GammaFlip *float64 `json:"gamma_flip"`
	// Strike with highest absolute call GEX (dealer-side resistance).
	CallWall *float64 `json:"call_wall"`
	// Strike with highest absolute put GEX (dealer-side support).
	PutWall *float64 `json:"put_wall"`
}

// MaxPainExpectedMove is the implied move from the ATM straddle,
// contextualized vs max pain.
type MaxPainExpectedMove struct {
	// ATM straddle mid in dollars. Rough proxy for the 1σ implied move.
	StraddlePrice *float64 `json:"straddle_price"`
	// ATM implied volatility (annualised %, e.g. 18.5 = 18.5%).
	AtmIv *float64 `json:"atm_iv"`
	// True when |spot - max_pain_strike| <= straddle_price.
	MaxPainWithinExpectedRange *bool `json:"max_pain_within_expected_range"`
}
