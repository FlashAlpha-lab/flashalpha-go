package flashalpha

// Typed response model for `GET /v1/exposure/narrative/{symbol}` (Growth+).
//
// FlashAlpha's "LLM-friendly" verbal-output endpoint — a structured,
// pre-written narrative summary of the dealer-exposure regime for an
// underlying. Every string in the Narrative block is server-generated from
// the numeric exposures, prior-day GEX delta, VIX context, and top OI changes,
// and is SAFE TO SURFACE VERBATIM in customer-facing UIs (newsletters,
// dashboards, chat bots, voice assistants, LLM tool responses).
//
// Pair with the raw Data sub-block when you want both the human-readable
// narrative AND the underlying numbers in a single round-trip.

// NarrativeResponse is the typed body of GET /v1/exposure/narrative/{symbol}.
type NarrativeResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// UnderlyingPrice is the spot mid in dollars at AsOf.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// AsOf is the ET wall-clock timestamp this snapshot was computed for.
	AsOf string `json:"as_of"`
	// Narrative is the verbal-output block + the underlying Data numerics.
	Narrative *NarrativeBlock `json:"narrative"`
}

// NarrativeBlock is the verbal-output block.
//
// All string fields (Regime, GexChange, KeyLevels, Flow, Vanna, Charm,
// ZeroDte, Outlook) are server-generated narratives — safe to surface
// verbatim. The Data sub-block holds the underlying numbers used to author
// those narratives.
type NarrativeBlock struct {
	// Regime is the dealer-positioning narrative — describes the current
	// gamma regime in plain English (e.g. "SPY is in a positive-gamma
	// regime — dealers long gamma, intraday moves likely dampened").
	Regime string `json:"regime"`
	// GexChange is the day-over-day GEX change narrative — explains how
	// dealer positioning shifted from the prior session close.
	GexChange string `json:"gex_change"`
	// KeyLevels is the call-wall / put-wall / gamma-flip narrative.
	KeyLevels string `json:"key_levels"`
	// Flow is the OI / volume / flow narrative — describes today's options
	// activity in context of the regime.
	Flow string `json:"flow"`
	// Vanna is the dealer-vanna narrative — how the hedge book responds to
	// vol shocks.
	Vanna string `json:"vanna"`
	// Charm is the dealer-charm narrative — time-decay-driven hedging into
	// the close.
	Charm string `json:"charm"`
	// ZeroDte is the same-day-expiry narrative — 0DTE share of GEX, key
	// 0DTE strikes, pin risk.
	ZeroDte string `json:"zero_dte"`
	// Outlook is the forward-looking narrative — what to expect for the
	// rest of the session given the current regime + key levels + flow.
	Outlook string `json:"outlook"`
	// Data is the underlying numerics used to author the narratives — useful
	// for cross-checking or building your own strings on top.
	Data *NarrativeData `json:"data"`
}

// NarrativeData is the numerics block backing the narrative strings.
type NarrativeData struct {
	// NetGex is net dealer gamma exposure ($/1% spot move) at AsOf.
	NetGex *float64 `json:"net_gex"`
	// NetGexPrior is the prior-session-close net dealer GEX — the
	// denominator for NetGexChangePct.
	NetGexPrior *float64 `json:"net_gex_prior"`
	// NetGexChangePct is (NetGex - NetGexPrior) / |NetGexPrior| * 100 —
	// percent change in dealer gamma vs prior session close.
	NetGexChangePct *float64 `json:"net_gex_change_pct"`
	// Vix is the CBOE VIX index level — macro vol context for the narrative.
	Vix *float64 `json:"vix"`
	// GammaFlip is the strike where net dealer gamma crosses zero.
	GammaFlip *float64 `json:"gamma_flip"`
	// CallWall is the strike with the largest absolute call GEX
	// (dealer-side resistance).
	CallWall *float64 `json:"call_wall"`
	// PutWall is the strike with the largest absolute put GEX
	// (dealer-side support).
	PutWall *float64 `json:"put_wall"`
	// Regime is the dealer-positioning classifier:
	//   "positive_gamma" | "negative_gamma" | "unknown"
	Regime string `json:"regime"`
	// ZeroDtePct is the 0DTE share of full-chain net GEX (%).
	ZeroDtePct *float64 `json:"zero_dte_pct"`
	// TopOiChanges is the per-strike top of OI changes vs the prior session —
	// the strikes where positioning shifted the most.
	TopOiChanges []NarrativeOiChange `json:"top_oi_changes"`
}

// NarrativeOiChange is one row in NarrativeData.TopOiChanges — a strike whose
// OI shifted significantly vs the prior session.
type NarrativeOiChange struct {
	// Strike is the strike price.
	Strike *float64 `json:"strike"`
	// Type is "call" or "put".
	Type string `json:"type"`
	// OiChange is the change in OI vs the prior session close (signed:
	// positive = OI added, negative = OI closed).
	OiChange *int `json:"oi_change"`
	// Volume is today's session volume at this strike+type.
	Volume *int `json:"volume"`
}
