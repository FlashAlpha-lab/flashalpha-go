package flashalpha

// Typed response model for `GET /v1/pricing/greeks`.
//
// Black-Scholes-Merton greeks calculator. Given (spot, strike, DTE, sigma,
// type) — with optional risk-free rate and dividend yield — the endpoint
// returns the theoretical option price plus first-, second-, and third-order
// greeks.
//
// Useful as a reference implementation: lets you compute greeks server-side
// without maintaining your own BSM library, and gives you the same numbers
// FlashAlpha uses internally to dollarize chain-wide GEX/DEX/VEX/CHEX
// exposures.

// PricingGreeksResponse is the typed body of GET /v1/pricing/greeks.
type PricingGreeksResponse struct {
	// Inputs echoes the BSM inputs the server actually used (after applying
	// defaults for r and q if omitted in the request).
	Inputs *PricingGreeksInputs `json:"inputs"`
	// TheoreticalPrice is the BSM fair value of the option (per share, in
	// dollars). Multiply by 100 for the per-contract dollar value.
	TheoreticalPrice *float64 `json:"theoretical_price"`
	// FirstOrder is the first-order greeks block (delta, gamma, theta, vega, rho).
	FirstOrder *PricingFirstOrderGreeks `json:"first_order"`
	// SecondOrder is the second-order greeks block (vanna, charm, vomma, dual delta).
	SecondOrder *PricingSecondOrderGreeks `json:"second_order"`
	// ThirdOrder is the third-order greeks block (speed, zomma, color, ultima).
	ThirdOrder *PricingThirdOrderGreeks `json:"third_order"`
	// Additional is the additional-greeks block (lambda, veta).
	Additional *PricingAdditionalGreeks `json:"additional"`
}

// PricingGreeksInputs echoes the BSM inputs used for the calculation.
type PricingGreeksInputs struct {
	// Spot is the underlying price (dollars).
	Spot *float64 `json:"spot"`
	// Strike is the option strike (dollars).
	Strike *float64 `json:"strike"`
	// Dte is days to expiration.
	Dte *float64 `json:"dte"`
	// Sigma is the annualised implied volatility (decimal — e.g. 0.20 = 20%).
	Sigma *float64 `json:"sigma"`
	// Type is "call" or "put".
	Type *string `json:"type"`
	// RiskFreeRate is the annualised risk-free rate used (decimal).
	RiskFreeRate *float64 `json:"risk_free_rate"`
	// DividendYield is the annualised continuous dividend yield (decimal).
	DividendYield *float64 `json:"dividend_yield"`
}

// PricingFirstOrderGreeks is the first-order BSM greeks block.
type PricingFirstOrderGreeks struct {
	// Delta is dV/dS — sensitivity of option price to a $1 spot move
	// (signed: ~+1 deep ITM call, ~-1 deep ITM put).
	Delta *float64 `json:"delta"`
	// Gamma is d²V/dS² — rate of change of delta per $1 spot move
	// (always non-negative for vanilla options).
	Gamma *float64 `json:"gamma"`
	// Theta is dV/dt — time decay (typically negative for long options;
	// quoted per CALENDAR DAY here, not annualised).
	Theta *float64 `json:"theta"`
	// Vega is dV/dσ — sensitivity per 1-vol-point (so a 1% vol change
	// changes price by Vega).
	Vega *float64 `json:"vega"`
	// Rho is dV/dr — sensitivity per 1% change in the risk-free rate.
	Rho *float64 `json:"rho"`
}

// PricingSecondOrderGreeks is the second-order BSM greeks block.
type PricingSecondOrderGreeks struct {
	// Vanna is d²V/(dS dσ) — cross-sensitivity of delta to a vol change
	// (drives vanna-driven dealer hedging flow).
	Vanna *float64 `json:"vanna"`
	// Charm is -d²V/(dS dt) — drift in delta from time decay
	// (drives charm-driven dealer hedging into the close).
	Charm *float64 `json:"charm"`
	// Vomma is d²V/dσ² — sensitivity of vega to vol changes
	// (the "vol-of-vol" greek).
	Vomma *float64 `json:"vomma"`
	// DualDelta is dV/dK — sensitivity to the strike. Equal to the
	// risk-neutral cumulative density at the strike — useful for
	// implied-distribution work.
	DualDelta *float64 `json:"dual_delta"`
}

// PricingThirdOrderGreeks is the third-order BSM greeks block — used for
// curvature analysis and advanced hedging strategies.
type PricingThirdOrderGreeks struct {
	// Speed is d³V/dS³ — rate of change of gamma per $1 spot move
	// (drives the convexity of the gamma surface).
	Speed *float64 `json:"speed"`
	// Zomma is d³V/(dS² dσ) — sensitivity of gamma to a vol change.
	Zomma *float64 `json:"zomma"`
	// Color is -d³V/(dS² dt) — rate of change of gamma over time
	// (the "gamma decay").
	Color *float64 `json:"color"`
	// Ultima is d³V/dσ³ — sensitivity of vomma to vol changes.
	Ultima *float64 `json:"ultima"`
}

// PricingAdditionalGreeks is the additional-greeks block.
type PricingAdditionalGreeks struct {
	// Lambda is the option's elasticity = (Delta * Spot) / Price — the
	// percentage change in option price per 1% change in spot. Nil when the
	// theoretical price is <= 0 (would otherwise divide by zero).
	Lambda *float64 `json:"lambda"`
	// Veta is dVega/dt — the rate of change of vega over time
	// (vega's time decay).
	Veta *float64 `json:"veta"`
}
