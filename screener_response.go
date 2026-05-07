package flashalpha

// Typed response model for `POST /v1/screener` (Growth+ live, Alpha+ for
// formulas / harvest scores).
//
// The response body is split into two halves:
//
//   - Meta — pagination, universe sizing, tier, and the snapshot timestamp.
//   - Data — the result rows. Each row's columns depend on the request's
//     `select` argument (and any computed `formulas`), so the shape is open;
//     rows are typed as map[string]any here. Inspect ScreenerRequest.Select
//     for the columns you asked for.

// ScreenerResponseMeta is the meta block returned by POST /v1/screener.
//
// TotalCount is the size of the pre-paginated result set; ReturnedCount is
// the row count actually returned in this page (==len(Data)). UniverseSize is
// the number of symbols visible to the caller's tier (10 on Growth, ~250 on
// Alpha).
type ScreenerResponseMeta struct {
	// TotalCount is the number of rows that matched the filter before paging.
	TotalCount int `json:"total_count"`
	// ReturnedCount is the number of rows in this page (== len(Data)).
	ReturnedCount int `json:"returned_count"`
	// UniverseSize is the size of the visible symbol universe at the
	// caller's tier.
	UniverseSize int `json:"universe_size"`
	// Offset is the request's paging offset (echoed).
	Offset int `json:"offset"`
	// Limit is the request's paging limit (echoed).
	Limit int `json:"limit"`
	// Tier is the caller's plan tier — "growth" or "alpha".
	Tier string `json:"tier"`
	// AsOf is the ET wall-clock timestamp of the underlying snapshot.
	AsOf string `json:"as_of"`
}

// ScreenerResponse is the typed body of POST /v1/screener.
//
// Data is intentionally typed as a slice of generic maps because the row
// shape is request-driven (it depends on `select` and any defined
// `formulas`). For typed access, inspect the column names you supplied in
// ScreenerRequest.Select.
type ScreenerResponse struct {
	// Meta is the pagination + universe + tier + as-of metadata.
	Meta *ScreenerResponseMeta `json:"meta"`
	// Data is the result rows. Columns depend on the request's `select` and
	// any defined `formulas` — typed as raw maps for forward compatibility.
	Data []map[string]any `json:"data"`
}
