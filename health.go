package flashalpha

// Typed response model for `GET /health` (public, unauthenticated).
//
// Single-field heartbeat — Status is the literal string "ok" when the API is
// operational. The endpoint is not rate-limited and does not require an API
// key.

// HealthResponse is the typed body of GET /health.
type HealthResponse struct {
	// Status is the literal "ok" on healthy responses.
	Status string `json:"status"`
}
