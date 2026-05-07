package flashalpha

// Typed response model for `GET /optionquote/{ticker}` (Growth+, live only).
//
// Live option quote with full BSM + extended greeks, plus optional SVI-fitted
// vol from the surface. The endpoint accepts up to three filters (expiry,
// strike, type); when all three are supplied the response is a single
// OptionQuote object, otherwise it's a slice. Use OptionQuoteListResponse
// to decode the list shape.
//
// Wire format quirks (these match the live API exactly):
//   - bidSize / askSize / lastUpdate are camelCase on the wire.
//   - svi_vol is nullable and gated — when svi_vol_gated=true the fitted
//     vol is suppressed (insufficient surface coverage).
//   - underlying is an optional echo of the parent ticker — present when
//     the request was made with a single quote shape, absent otherwise.
//
// Requires Growth+ plan; returns 403 tier_restricted for Basic/Free.

// OptionQuote is one option quote row (or the single-object response when
// expiry, strike, and type were all supplied).
type OptionQuote struct {
	// Type is the option type ("call" or "put").
	Type *string `json:"type"`
	// Expiry is the expiration date (YYYY-MM-DD).
	Expiry *string `json:"expiry"`
	// Strike is the option strike price.
	Strike *float64 `json:"strike"`
	// Bid is the current best bid.
	Bid *float64 `json:"bid"`
	// Ask is the current best ask.
	Ask *float64 `json:"ask"`
	// Mid is (bid + ask) / 2.
	Mid *float64 `json:"mid"`
	// BidSize is the size at the best bid (camelCase on the wire).
	BidSize *int `json:"bidSize"`
	// AskSize is the size at the best ask (camelCase on the wire).
	AskSize *int `json:"askSize"`
	// LastUpdate is the wire-format last-quote timestamp (camelCase).
	LastUpdate *string `json:"lastUpdate"`
	// ImpliedVol is the BSM implied volatility from the mid price (annualised %).
	ImpliedVol *float64 `json:"implied_vol"`
	// IvBid is the IV implied from the bid (annualised %).
	IvBid *float64 `json:"iv_bid"`
	// IvAsk is the IV implied from the ask (annualised %).
	IvAsk *float64 `json:"iv_ask"`
	// Delta is the BSM delta.
	Delta *float64 `json:"delta"`
	// Gamma is the BSM gamma.
	Gamma *float64 `json:"gamma"`
	// Theta is the BSM theta (per-day decay convention).
	Theta *float64 `json:"theta"`
	// Vega is the BSM vega (per-1-vol-point convention).
	Vega *float64 `json:"vega"`
	// Rho is the BSM rho.
	Rho *float64 `json:"rho"`
	// Vanna is the cross greek d²V / dSpot dVol.
	Vanna *float64 `json:"vanna"`
	// Charm is the time-decay-of-delta greek d²V / dSpot dT.
	Charm *float64 `json:"charm"`
	// SviVol is the SVI-fitted IV from the surface (annualised %, nullable).
	// When SviVolGated=true, this is suppressed regardless.
	SviVol *float64 `json:"svi_vol"`
	// SviVolGated is true when surface coverage is insufficient and SviVol
	// has been gated off.
	SviVolGated *bool `json:"svi_vol_gated"`
	// OpenInterest is the open interest at this contract.
	OpenInterest *int `json:"open_interest"`
	// Volume is the trading volume at this contract.
	Volume *int `json:"volume"`
	// Underlying is the parent ticker (optional echo; only present in some
	// response shapes).
	Underlying *string `json:"underlying,omitempty"`
}
