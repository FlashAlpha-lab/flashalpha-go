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

// ── /v1/pricing/iv (Free+) ───────────────────────────────────────────────────

// PricingIvInputs echoes the inputs the server actually used for the IV
// calculation (after defaults are applied for r and q).
type PricingIvInputs struct {
	// Spot is the underlying price (dollars).
	Spot *float64 `json:"spot"`
	// Strike is the option strike (dollars).
	Strike *float64 `json:"strike"`
	// Dte is days to expiration.
	Dte *float64 `json:"dte"`
	// Price is the market option price (mid) used as the inversion target.
	Price *float64 `json:"price"`
	// Type is "call" or "put".
	Type *string `json:"type"`
	// RiskFreeRate is the annualised risk-free rate used (decimal).
	RiskFreeRate *float64 `json:"risk_free_rate"`
	// DividendYield is the annualised continuous dividend yield (decimal).
	DividendYield *float64 `json:"dividend_yield"`
}

// PricingIvResponse is the typed body of GET /v1/pricing/iv.
//
// Newton-Raphson root-find on the BSM model that returns the implied
// volatility consistent with the supplied market price. ImpliedVolatility is
// the decimal vol (0.18 = 18%); ImpliedVolatilityPct is the same number in
// percent.
type PricingIvResponse struct {
	// Inputs echoes the BSM inputs the server actually used.
	Inputs *PricingIvInputs `json:"inputs"`
	// ImpliedVolatility is the annualised implied vol as a decimal
	// (e.g. 0.180042 = 18.0%).
	ImpliedVolatility *float64 `json:"implied_volatility"`
	// ImpliedVolatilityPct is ImpliedVolatility × 100 (e.g. 18.0).
	ImpliedVolatilityPct *float64 `json:"implied_volatility_pct"`
}

// ── /v1/pricing/kelly (Growth+) ──────────────────────────────────────────────

// PricingKellyInputs echoes the inputs the server actually used for the Kelly
// calculation.
type PricingKellyInputs struct {
	// Spot is the underlying price (dollars).
	Spot *float64 `json:"spot"`
	// Strike is the option strike (dollars).
	Strike *float64 `json:"strike"`
	// Dte is days to expiration.
	Dte *float64 `json:"dte"`
	// Sigma is the annualised implied volatility used (decimal).
	Sigma *float64 `json:"sigma"`
	// Premium is the option premium paid per share (dollars).
	Premium *float64 `json:"premium"`
	// Mu is the user-supplied real-world annualised drift (decimal) used to
	// build the lognormal distribution Kelly integrates over.
	Mu *float64 `json:"mu"`
	// Type is "call" or "put".
	Type *string `json:"type"`
	// RiskFreeRate is the annualised risk-free rate used (decimal).
	RiskFreeRate *float64 `json:"risk_free_rate"`
	// DividendYield is the annualised continuous dividend yield (decimal).
	DividendYield *float64 `json:"dividend_yield"`
}

// PricingKellySizing is the Kelly-fraction sizing block.
//
// Half-Kelly is the standard practical recommendation. On negative
// expected-value setups the server returns zeros across all five fields and
// PricingKellyResponse.Recommendation explains why.
type PricingKellySizing struct {
	// KellyFraction is the full-Kelly optimal bankroll fraction (0–1).
	KellyFraction *float64 `json:"kelly_fraction"`
	// HalfKelly is KellyFraction / 2 — the standard practical recommendation.
	HalfKelly *float64 `json:"half_kelly"`
	// QuarterKelly is KellyFraction / 4 — a conservative practical sizing.
	QuarterKelly *float64 `json:"quarter_kelly"`
	// KellyPct is KellyFraction expressed as a percent (e.g. 7.68 = 7.68%).
	KellyPct *float64 `json:"kelly_pct"`
	// HalfKellyPct is HalfKelly as a percent.
	HalfKellyPct *float64 `json:"half_kelly_pct"`
}

// PricingKellyAnalysis is the analysis block — expected ROI, win
// probabilities, breakeven, and the expected log-growth rate.
//
// All probabilities are real-world (computed under Mu), not risk-neutral.
type PricingKellyAnalysis struct {
	// ExpectedRoi is the expected return on investment (decimal,
	// payoff/premium - 1).
	ExpectedRoi *float64 `json:"expected_roi"`
	// ExpectedRoiPct is ExpectedRoi as a percent.
	ExpectedRoiPct *float64 `json:"expected_roi_pct"`
	// ExpectedPayoff is the expected payoff per share (dollars).
	ExpectedPayoff *float64 `json:"expected_payoff"`
	// ProbabilityOfProfit is the real-world probability the trade is
	// profitable at expiry (decimal).
	ProbabilityOfProfit *float64 `json:"probability_of_profit"`
	// ProbabilityOfProfitPct is ProbabilityOfProfit as a percent.
	ProbabilityOfProfitPct *float64 `json:"probability_of_profit_pct"`
	// ProbabilityItm is the real-world probability the option is ITM at
	// expiry (decimal).
	ProbabilityItm *float64 `json:"probability_itm"`
	// ProbabilityItmPct is ProbabilityItm as a percent.
	ProbabilityItmPct *float64 `json:"probability_itm_pct"`
	// MaxLoss is the maximum loss per share (= premium paid).
	MaxLoss *float64 `json:"max_loss"`
	// Breakeven is the underlying price needed at expiry to break even.
	Breakeven *float64 `json:"breakeven"`
	// ExpectedGrowthRate is the expected log-growth rate of bankroll at the
	// Kelly fraction.
	ExpectedGrowthRate *float64 `json:"expected_growth_rate"`
}

// PricingKellyResponse is the typed body of GET /v1/pricing/kelly.
//
// Numerical-integration Kelly criterion sizing for an option trade. Uses the
// real-world (Mu-conditioned) lognormal distribution rather than the
// risk-neutral measure. On negative-EV setups Sizing fields are all zero and
// Recommendation explains the no-position outcome.
type PricingKellyResponse struct {
	// Inputs echoes the inputs the server actually used.
	Inputs *PricingKellyInputs `json:"inputs"`
	// Sizing is the Kelly-fraction block (full / half / quarter).
	Sizing *PricingKellySizing `json:"sizing"`
	// Analysis is the expected-value / probability / breakeven block.
	Analysis *PricingKellyAnalysis `json:"analysis"`
	// Recommendation is a human-readable sizing recommendation safe to
	// surface verbatim (e.g. "Risk 3.8% of bankroll (half-Kelly).").
	Recommendation *string `json:"recommendation"`
}
