package flashalpha

// Typed response model for `GET /stockquote/{ticker}` (Free+, live only).
//
// Live equity NBBO snapshot — bid, ask, mid, last, and the wire-format
// last-update timestamp. All fields are nullable to surface partial-data
// situations (e.g. before the open, halts, illiquid names).
//
// Wire format quirks (these match the live API exactly):
//   - lastPrice and lastUpdate are camelCase on the wire (the only
//     camelCase fields on this endpoint).
//
// Available on all paid plans plus Free.

// StockQuoteResponse is the typed body of GET /stockquote/{ticker}.
type StockQuoteResponse struct {
	// Ticker is the equity ticker echoed from the request path.
	Ticker string `json:"ticker"`
	// Bid is the current best bid.
	Bid *float64 `json:"bid"`
	// Ask is the current best ask.
	Ask *float64 `json:"ask"`
	// Mid is (bid + ask) / 2.
	Mid *float64 `json:"mid"`
	// LastPrice is the most recent trade price (camelCase on the wire).
	LastPrice *float64 `json:"lastPrice"`
	// LastUpdate is the wire-format last-update timestamp (camelCase).
	LastUpdate *string `json:"lastUpdate"`
}
