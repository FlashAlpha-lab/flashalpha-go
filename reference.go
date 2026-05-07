package flashalpha

// Typed response models for the reference-data endpoints:
//
//   - GET /v1/tickers           — all available stock tickers (Free+)
//   - GET /v1/symbols           — currently queried symbols with live data
//   - GET /v1/options/{ticker}  — option chain metadata (expirations + strikes)
//
// Each endpoint returns a flat top-level object — there are no nested
// sub-blocks beyond the option-chain expiration rows. All fields are populated
// on success.

// ── /v1/tickers ──────────────────────────────────────────────────────────────

// TickersResponse is the typed body of GET /v1/tickers.
type TickersResponse struct {
	// Tickers is the alphabetised list of all available equity tickers
	// (Polygon catalog).
	Tickers []string `json:"tickers"`
	// Count is the number of tickers returned (== len(Tickers)).
	Count int `json:"count"`
}

// ── /v1/symbols ──────────────────────────────────────────────────────────────

// SymbolsResponse is the typed body of GET /v1/symbols.
//
// Reports the symbols currently held in the live in-memory store (refreshed
// every 5–10 s). Note is a server-supplied caveat safe to surface verbatim;
// LastUpdated is the ISO timestamp of the most recent refresh.
type SymbolsResponse struct {
	// Symbols is the list of symbols with live data in the store.
	Symbols []string `json:"symbols"`
	// Count is the number of symbols returned (== len(Symbols)).
	Count int `json:"count"`
	// Note is a caveat about on-demand availability (safe to surface).
	Note string `json:"note"`
	// LastUpdated is the ISO 8601 timestamp of the most recent store refresh.
	LastUpdated string `json:"last_updated"`
}

// ── /v1/options/{ticker} ─────────────────────────────────────────────────────

// OptionsExpiration is one expiration row from the options-meta endpoint —
// an expiration date plus the list of listed strikes for that expiry.
type OptionsExpiration struct {
	// Expiration is the option expiration date (YYYY-MM-DD).
	Expiration string `json:"expiration"`
	// Strikes is the list of listed strike prices for this expiration.
	Strikes []float64 `json:"strikes"`
}

// OptionsResponse is the typed body of GET /v1/options/{ticker}.
//
// Lightweight option-chain metadata — the listed expirations and per-expiry
// strike grids. ExpirationCount equals len(Expirations); TotalContracts is
// the sum of len(Expirations[i].Strikes) over all expiries.
type OptionsResponse struct {
	// Symbol is the underlying ticker echoed from the request path.
	Symbol string `json:"symbol"`
	// Expirations is the list of expirations with their listed strike grids.
	Expirations []OptionsExpiration `json:"expirations"`
	// ExpirationCount is the number of expirations (== len(Expirations)).
	ExpirationCount int `json:"expiration_count"`
	// TotalContracts is the total count of (expiration × strike) cells across
	// the chain.
	TotalContracts int `json:"total_contracts"`
}
