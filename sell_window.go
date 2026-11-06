package flashalpha

// sell-window: GET /v1/strategies/sell-window/{symbol}.
//
// Should you sell options today? One answer per session for SPY, QQQ and IWM, built
// from more than 5 million backtests. Docs:
// https://flashalpha.com/docs/lab-api-strategies-sell-window

import (
	"context"
)

// SellWindowLeg is one contract of a sell-window trade.
type SellWindowLeg struct {
	Side string `json:"side"` // sell or buy
	// Quantity is contracts per unit of the trade: 2 for the sold body of put_bwb, otherwise 1.
	Quantity int     `json:"quantity"`
	Type     string  `json:"type"` // call or put
	Strike   float64 `json:"strike"`
	// Delta is the contract's own delta when the answer was served (negative for puts).
	Delta float64 `json:"delta"`
	Bid   float64 `json:"bid"`
	Ask   float64 `json:"ask"`
}

// SellWindowCandidate is one trade to consider on a sell window.
type SellWindowCandidate struct {
	// Structure is short_put, short_call, short_strangle, short_straddle,
	// tail_hedged_put, put_credit_spread or put_bwb.
	Structure string `json:"structure"`
	// Exit is expiry (hold to expiration) or tp50 (close at half of the credit, otherwise hold).
	Exit string `json:"exit"`
	// Legs is empty when no option chain could be loaded for the answer.
	Legs []SellWindowLeg `json:"legs,omitempty"`
	// MidCredit is the net credit per share at mid. Multiply by 100 for one unit in dollars.
	MidCredit *float64 `json:"mid_credit,omitempty"`
}

// SellWindowWindow is one of the four windows, named by its listed expiry.
type SellWindowWindow struct {
	// Expiry (yyyy-MM-dd) is nil when no listed expiry falls in the window's range.
	Expiry *string `json:"expiry,omitempty"`
	// DaysToExpiry is calendar days from the session to Expiry.
	DaysToExpiry *int `json:"days_to_expiry,omitempty"`
	// Action is sell, skip or unavailable. Treat unavailable like a skip.
	Action string `json:"action"`
	// Candidates is set only when Action is sell. It can be empty: a trade whose
	// legs cannot all be found in the chain is left out.
	Candidates []SellWindowCandidate `json:"candidates,omitempty"`
}

// SellWindowResponse is the sell-window answer. It has its own shape, not the
// StrategyDecisionResponse envelope. Spot, QuotedAt and LegsIndicative are set when a
// window sells and an option chain could be loaded.
type SellWindowResponse struct {
	Symbol string `json:"symbol"`
	// Session (yyyy-MM-dd) is the trading session the answer applies to.
	Session string `json:"session"`
	// Spot is the underlying mid the strikes were picked against.
	Spot *float64 `json:"spot,omitempty"`
	// QuotedAt is the UTC time of the quotes behind the strikes.
	QuotedAt *string `json:"quoted_at,omitempty"`
	// LegsIndicative is true when the quotes are from before that session's 09:35 ET
	// entry: call again at or after 09:35 ET for the strikes to trade.
	LegsIndicative *bool `json:"legs_indicative,omitempty"`
	// Windows always has four entries: same day, then about one week, two weeks and one month out.
	Windows []SellWindowWindow `json:"windows"`
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
}

// SellWindow answers "should you sell options today?" for four listed expiries (same
// day, then the expiries closest to one week, two weeks and one month out). Each window
// says sell, skip or unavailable. On a sell it lists the trades to place, with strikes,
// quotes and the exit rule.
//
// The answer is set from the previous session's close and stays fixed for the session.
// Any symbol other than SPY, QQQ or IWM returns every window as unavailable. Requires
// Basic+ through November 2026, Growth+ from December 1, 2026. Backtest any session with
// flashalpha-historical-go (Alpha+).
func (c *Client) SellWindow(ctx context.Context, symbol string) (*SellWindowResponse, error) {
	raw, err := c.get(ctx, "/v1/strategies/sell-window/"+seg(symbol), nil)
	if err != nil {
		return nil, err
	}
	out := &SellWindowResponse{}
	if err := decodeTyped("sell-window", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
