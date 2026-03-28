package flashalpha

import "fmt"

// APIError is the base error type returned by the FlashAlpha API client.
// It carries the HTTP status code, a human-readable message, and the raw
// response body (if any) parsed as a map.
type APIError struct {
	StatusCode int
	Message    string
	Response   map[string]interface{}
}

func (e *APIError) Error() string {
	return fmt.Sprintf("flashalpha: HTTP %d: %s", e.StatusCode, e.Message)
}

// AuthenticationError is returned when the API key is missing or invalid (HTTP 401).
type AuthenticationError struct {
	*APIError
}

func (e *AuthenticationError) Error() string { return e.APIError.Error() }

// TierRestrictedError is returned when the current subscription plan does not
// include the requested endpoint (HTTP 403).
type TierRestrictedError struct {
	*APIError
	CurrentPlan  string
	RequiredPlan string
}

func (e *TierRestrictedError) Error() string {
	base := e.APIError.Error()
	if e.CurrentPlan != "" || e.RequiredPlan != "" {
		return fmt.Sprintf("%s (current_plan=%s, required_plan=%s)", base, e.CurrentPlan, e.RequiredPlan)
	}
	return base
}

// NotFoundError is returned when the requested resource does not exist (HTTP 404).
type NotFoundError struct {
	*APIError
}

func (e *NotFoundError) Error() string { return e.APIError.Error() }

// RateLimitError is returned when the request rate limit is exceeded (HTTP 429).
// RetryAfter contains the value of the Retry-After response header in seconds,
// or 0 if the header was absent.
type RateLimitError struct {
	*APIError
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	base := e.APIError.Error()
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry_after=%ds)", base, e.RetryAfter)
	}
	return base
}

// ServerError is returned for HTTP 5xx responses.
type ServerError struct {
	*APIError
}

func (e *ServerError) Error() string { return e.APIError.Error() }
