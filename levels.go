package flashalpha

// Typed response model for `GET /v1/exposure/levels/{symbol}`.
//
// Compact view of the most-watched dealer-exposure levels for an underlying:
// gamma flip, the strikes with the largest +/- net GEX, the call/put walls,
// the highest-OI strike, and the 0DTE magnet. Designed for UIs that need
// just the levels (e.g. chart overlays, alert systems) without the full
// exposure-summary dashboard.

// LevelsResponse is the typed body of GET /v1/exposure/levels/{symbol}.
type LevelsResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// UnderlyingPrice is the spot mid in dollars at AsOf — the reference
	// price for distance-to-level calculations.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// AsOf is the ET wall-clock timestamp this snapshot was computed for.
	AsOf string `json:"as_of"`
	// Levels is the key-strike block.
	Levels *LevelsBlock `json:"levels"`
}

// LevelsBlock is the key-strike block — every entry is a strike price (or nil
// when the metric is undefined / no options data is loaded).
type LevelsBlock struct {
	// GammaFlip is the strike where net dealer gamma crosses zero. The
	// single most-watched intraday level — spot crossing the flip flips the
	// dealer hedging regime (positive ↔ negative gamma).
	GammaFlip *float64 `json:"gamma_flip"`
	// MaxPositiveGamma is the strike with the largest positive net GEX —
	// typically the most call-heavy strike and a magnet for spot in
	// positive-gamma regimes.
	MaxPositiveGamma *float64 `json:"max_positive_gamma"`
	// MaxNegativeGamma is the strike with the largest negative net GEX —
	// typically the most put-heavy strike and a level where dealer hedging
	// flow can amplify moves.
	MaxNegativeGamma *float64 `json:"max_negative_gamma"`
	// CallWall is the strike with the largest absolute call GEX
	// (dealer-side resistance — spot tends to stall here in the
	// positive-gamma regime).
	CallWall *float64 `json:"call_wall"`
	// PutWall is the strike with the largest absolute put GEX
	// (dealer-side support — spot tends to bounce here in the
	// positive-gamma regime).
	PutWall *float64 `json:"put_wall"`
	// HighestOiStrike is the strike with the largest total OI (calls + puts).
	HighestOiStrike *float64 `json:"highest_oi_strike"`
	// ZeroDteMagnet is the single 0DTE strike with the largest absolute
	// dealer GEX — the strike most likely to "pin" spot into the close via
	// dealer hedging flow on today's expiry. Nil when no 0DTE chain is
	// active for this symbol.
	ZeroDteMagnet *float64 `json:"zero_dte_magnet"`
}
