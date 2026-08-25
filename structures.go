package flashalpha

// Pure-math multi-leg structure utilities (POST, Basic+):
//
//   - StructurePnl    — POST /v1/structures/pnl     — at-expiry P&L curve, breakevens, max P/L
//   - StructureGreeks — POST /v1/structures/greeks  — aggregate BSM Greeks across legs
//
// Every result is a deterministic function of the legs supplied: no
// market-data lookup, no symbol resolution, no live IV.

import "context"

// ── Request types ────────────────────────────────────────────────────────────

// StructurePnlLeg is one leg of a P&L structure. Action is "buy"/"sell"
// (aliases long/short); Type is "call"/"put" (aliases c/p).
type StructurePnlLeg struct {
	Action   string  `json:"action"`
	Type     string  `json:"type"`
	Strike   float64 `json:"strike"`
	Premium  float64 `json:"premium"`
	Quantity int     `json:"quantity,omitempty"`
}

// StructurePnlRequest is the request body for POST /v1/structures/pnl.
// Legs must be non-empty. MinUnderlying/MaxUnderlying/Points are optional —
// when omitted the curve range is derived from the leg strikes ±30% and Points
// defaults to 81.
type StructurePnlRequest struct {
	Legs          []StructurePnlLeg `json:"legs"`
	MinUnderlying *float64          `json:"minUnderlying,omitempty"`
	MaxUnderlying *float64          `json:"maxUnderlying,omitempty"`
	Points        *int              `json:"points,omitempty"`
}

// StructureGreeksLeg is one leg of a Greeks structure — each carries its own
// expiry and impliedVol so calendars/diagonals aggregate correctly.
type StructureGreeksLeg struct {
	Action     string  `json:"action"`
	Type       string  `json:"type"`
	Strike     float64 `json:"strike"`
	Expiry     string  `json:"expiry"`
	ImpliedVol float64 `json:"impliedVol"`
	Quantity   int     `json:"quantity,omitempty"`
}

// StructureGreeksRequest is the request body for POST /v1/structures/greeks.
// Legs and Spot are required; Today/Rate/DividendYield are optional.
type StructureGreeksRequest struct {
	Legs          []StructureGreeksLeg `json:"legs"`
	Spot          float64              `json:"spot"`
	Today         string               `json:"today,omitempty"`
	Rate          *float64             `json:"rate,omitempty"`
	DividendYield *float64             `json:"dividendYield,omitempty"`
}

// ── Response types ───────────────────────────────────────────────────────────

// PnlPoint is one (underlying, pnl) sample on the at-expiry curve.
type PnlPoint struct {
	Underlying float64 `json:"underlying"`
	Pnl        float64 `json:"pnl"`
}

// StructurePnlResponse is the body of POST /v1/structures/pnl. MaxProfit /
// MaxLoss are nil when unbounded on that side.
type StructurePnlResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Legs       []StructurePnlLeg `json:"legs"`
	PnlCurve   []PnlPoint        `json:"pnl_curve"`
	Breakevens []float64         `json:"breakevens"`
	MaxProfit  *float64          `json:"max_profit"`
	MaxLoss    *float64          `json:"max_loss"`
}

// PositionGreeks holds the aggregated, quantity-scaled, direction-signed Greeks.
type PositionGreeks struct {
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Theta float64 `json:"theta"`
	Vega  float64 `json:"vega"`
	Rho   float64 `json:"rho"`
	Vanna float64 `json:"vanna"`
	Charm float64 `json:"charm"`
}

// StructureGreeksResponse is the body of POST /v1/structures/greeks.
type StructureGreeksResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Spot           float64              `json:"spot"`
	AsOf           string               `json:"as_of"`
	ValuationDate  string               `json:"valuation_date"`
	Rate           float64              `json:"rate"`
	DividendYield  float64              `json:"dividend_yield"`
	Legs           []StructureGreeksLeg `json:"legs"`
	PositionGreeks *PositionGreeks      `json:"position_greeks"`
}

// ── Client methods ───────────────────────────────────────────────────────────

// StructurePnl computes the at-expiry P&L curve, breakevens, and max profit/loss
// for an arbitrary multi-leg structure. Requires Basic+. Returns the raw map;
// use StructurePnlTyped for a typed response.
func (c *Client) StructurePnl(ctx context.Context, req StructurePnlRequest) (map[string]interface{}, error) {
	return c.post(ctx, "/v1/structures/pnl", req)
}

// StructurePnlTyped is the strongly-typed variant of StructurePnl.
func (c *Client) StructurePnlTyped(ctx context.Context, req StructurePnlRequest) (*StructurePnlResponse, error) {
	raw, err := c.StructurePnl(ctx, req)
	if err != nil {
		return nil, err
	}
	out := &StructurePnlResponse{}
	if err := decodeTyped("structure pnl", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// StructureGreeks aggregates Black-Scholes Greeks across a multi-leg position.
// Requires Basic+. Returns the raw map; use StructureGreeksTyped for a typed
// response.
func (c *Client) StructureGreeks(ctx context.Context, req StructureGreeksRequest) (map[string]interface{}, error) {
	return c.post(ctx, "/v1/structures/greeks", req)
}

// StructureGreeksTyped is the strongly-typed variant of StructureGreeks.
func (c *Client) StructureGreeksTyped(ctx context.Context, req StructureGreeksRequest) (*StructureGreeksResponse, error) {
	raw, err := c.StructureGreeks(ctx, req)
	if err != nil {
		return nil, err
	}
	out := &StructureGreeksResponse{}
	if err := decodeTyped("structure greeks", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
