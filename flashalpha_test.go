package flashalpha_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

// helper: start a test server that always returns the given status and JSON body.
func newServer(t *testing.T, status int, body map[string]interface{}) (*httptest.Server, *flashalpha.Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	client := flashalpha.NewClientWithURL("test-key", srv.URL)
	return srv, client
}

// helper: start a server that records the last request.
type capturedRequest struct {
	Path   string
	Query  string
	Header http.Header
}

func newCapturingServer(t *testing.T, status int, body map[string]interface{}) (*httptest.Server, *flashalpha.Client, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Path = r.URL.Path
		captured.Query = r.URL.RawQuery
		captured.Header = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	client := flashalpha.NewClientWithURL("my-secret-key", srv.URL)
	return srv, client, captured
}

// ── API key header ────────────────────────────────────────────────────────────

func TestAPIKeyHeader(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"status": "ok"})
	ctx := context.Background()
	_, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cap.Header.Get("X-Api-Key"); got != "my-secret-key" {
		t.Errorf("X-Api-Key header = %q, want %q", got, "my-secret-key")
	}
}

// ── Health ────────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	_, client := newServer(t, 200, map[string]interface{}{"status": "ok"})
	got, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: unexpected error: %v", err)
	}
	if got["status"] != "ok" {
		t.Errorf("Health: got %v, want status=ok", got)
	}
}

// ── Account ───────────────────────────────────────────────────────────────────

func TestAccount(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"plan": "growth"})
	_, err := client.Account(context.Background())
	if err != nil {
		t.Fatalf("Account: unexpected error: %v", err)
	}
	if cap.Path != "/v1/account" {
		t.Errorf("Account path = %q, want /v1/account", cap.Path)
	}
}

// ── Gex ───────────────────────────────────────────────────────────────────────

func TestGex(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"data": []interface{}{}})
	_, err := client.Gex(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("Gex: unexpected error: %v", err)
	}
	if cap.Path != "/v1/exposure/gex/SPY" {
		t.Errorf("Gex path = %q, want /v1/exposure/gex/SPY", cap.Path)
	}
}

func TestGexWithExpiration(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"data": []interface{}{}})
	_, err := client.Gex(context.Background(), "SPY", flashalpha.WithExpiration("2025-12-19"))
	if err != nil {
		t.Fatalf("Gex with expiration: unexpected error: %v", err)
	}
	if !strings.Contains(cap.Query, "expiration=2025-12-19") {
		t.Errorf("Gex query = %q, want expiration param", cap.Query)
	}
}

func TestGexWithMinOI(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"data": []interface{}{}})
	_, err := client.Gex(context.Background(), "SPY", flashalpha.WithMinOI(500))
	if err != nil {
		t.Fatalf("Gex with minOI: unexpected error: %v", err)
	}
	if !strings.Contains(cap.Query, "min_oi=500") {
		t.Errorf("Gex query = %q, want min_oi param", cap.Query)
	}
}

// ── Dex ───────────────────────────────────────────────────────────────────────

func TestDex(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"data": []interface{}{}})
	_, err := client.Dex(context.Background(), "QQQ")
	if err != nil {
		t.Fatalf("Dex: unexpected error: %v", err)
	}
	if cap.Path != "/v1/exposure/dex/QQQ" {
		t.Errorf("Dex path = %q", cap.Path)
	}
}

func TestDexWithExpiration(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"data": []interface{}{}})
	_, err := client.Dex(context.Background(), "QQQ", flashalpha.WithDexExpiration("2025-12-19"))
	if err != nil {
		t.Fatalf("Dex with expiration: unexpected error: %v", err)
	}
	if !strings.Contains(cap.Query, "expiration=2025-12-19") {
		t.Errorf("Dex query = %q, want expiration param", cap.Query)
	}
}

// ── Vex ───────────────────────────────────────────────────────────────────────

func TestVex(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"data": []interface{}{}})
	_, err := client.Vex(context.Background(), "IWM")
	if err != nil {
		t.Fatalf("Vex: unexpected error: %v", err)
	}
	if cap.Path != "/v1/exposure/vex/IWM" {
		t.Errorf("Vex path = %q", cap.Path)
	}
}

// ── Chex ──────────────────────────────────────────────────────────────────────

func TestChex(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"data": []interface{}{}})
	_, err := client.Chex(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("Chex: unexpected error: %v", err)
	}
	if cap.Path != "/v1/exposure/chex/AAPL" {
		t.Errorf("Chex path = %q", cap.Path)
	}
}

// ── Exposure endpoints ────────────────────────────────────────────────────────

func TestExposureLevels(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"levels": []interface{}{}})
	_, err := client.ExposureLevels(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("ExposureLevels: %v", err)
	}
	if cap.Path != "/v1/exposure/levels/SPY" {
		t.Errorf("ExposureLevels path = %q", cap.Path)
	}
}

func TestExposureSummary(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"summary": map[string]interface{}{}})
	_, err := client.ExposureSummary(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("ExposureSummary: %v", err)
	}
	if cap.Path != "/v1/exposure/summary/SPY" {
		t.Errorf("ExposureSummary path = %q", cap.Path)
	}
}

func TestNarrative(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"narrative": "bullish"})
	_, err := client.Narrative(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("Narrative: %v", err)
	}
	if cap.Path != "/v1/exposure/narrative/SPY" {
		t.Errorf("Narrative path = %q", cap.Path)
	}
}

func TestZeroDte(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"regime": "high_gamma"})
	_, err := client.ZeroDte(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("ZeroDte: %v", err)
	}
	if cap.Path != "/v1/exposure/zero-dte/SPY" {
		t.Errorf("ZeroDte path = %q", cap.Path)
	}
}

func TestZeroDteWithStrikeRange(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"regime": "high_gamma"})
	_, err := client.ZeroDte(context.Background(), "SPY", flashalpha.WithStrikeRange(0.05))
	if err != nil {
		t.Fatalf("ZeroDte with strike_range: %v", err)
	}
	if !strings.Contains(cap.Query, "strike_range=") {
		t.Errorf("ZeroDte query = %q, want strike_range param", cap.Query)
	}
}

func TestExposureHistory(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"history": []interface{}{}})
	_, err := client.ExposureHistory(context.Background(), "SPY", flashalpha.WithDays(30))
	if err != nil {
		t.Fatalf("ExposureHistory: %v", err)
	}
	if cap.Path != "/v1/exposure/history/SPY" {
		t.Errorf("ExposureHistory path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "days=30") {
		t.Errorf("ExposureHistory query = %q, want days=30", cap.Query)
	}
}

// ── Market data ───────────────────────────────────────────────────────────────

func TestStockQuote(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"last": 450.0})
	_, err := client.StockQuote(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("StockQuote: %v", err)
	}
	if cap.Path != "/stockquote/SPY" {
		t.Errorf("StockQuote path = %q", cap.Path)
	}
}

func TestOptionQuote(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"quotes": []interface{}{}})
	_, err := client.OptionQuote(context.Background(), "SPY",
		flashalpha.WithOptionExpiry("2025-12-19"),
		flashalpha.WithStrike(450.0),
		flashalpha.WithOptionType("call"),
	)
	if err != nil {
		t.Fatalf("OptionQuote: %v", err)
	}
	if cap.Path != "/optionquote/SPY" {
		t.Errorf("OptionQuote path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "expiry=2025-12-19") {
		t.Errorf("OptionQuote query = %q, want expiry param", cap.Query)
	}
	if !strings.Contains(cap.Query, "strike=450") {
		t.Errorf("OptionQuote query = %q, want strike param", cap.Query)
	}
	if !strings.Contains(cap.Query, "type=call") {
		t.Errorf("OptionQuote query = %q, want type param", cap.Query)
	}
}

func TestStockSummary(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"price": 450.0})
	_, err := client.StockSummary(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("StockSummary: %v", err)
	}
	if cap.Path != "/v1/stock/SPY/summary" {
		t.Errorf("StockSummary path = %q", cap.Path)
	}
}

func TestSurface(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"surface": []interface{}{}})
	_, err := client.Surface(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("Surface: %v", err)
	}
	if cap.Path != "/v1/surface/SPY" {
		t.Errorf("Surface path = %q", cap.Path)
	}
}

// ── Historical ────────────────────────────────────────────────────────────────

func TestHistoricalStockQuote(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"quotes": []interface{}{}})
	_, err := client.HistoricalStockQuote(context.Background(), "SPY", "2025-01-15")
	if err != nil {
		t.Fatalf("HistoricalStockQuote: %v", err)
	}
	if cap.Path != "/historical/stockquote/SPY" {
		t.Errorf("HistoricalStockQuote path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "date=2025-01-15") {
		t.Errorf("HistoricalStockQuote query = %q, want date param", cap.Query)
	}
}

func TestHistoricalStockQuoteWithTime(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"quotes": []interface{}{}})
	_, err := client.HistoricalStockQuote(context.Background(), "SPY", "2025-01-15", "10:30")
	if err != nil {
		t.Fatalf("HistoricalStockQuote with time: %v", err)
	}
	if !strings.Contains(cap.Query, "time=10%3A30") && !strings.Contains(cap.Query, "time=10:30") {
		t.Errorf("HistoricalStockQuote query = %q, want time param", cap.Query)
	}
}

func TestHistoricalOptionQuote(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"quotes": []interface{}{}})
	_, err := client.HistoricalOptionQuote(context.Background(), "SPY", "2025-01-15",
		flashalpha.WithHistExpiry("2025-01-17"),
		flashalpha.WithHistStrike(450.0),
		flashalpha.WithHistType("put"),
		flashalpha.WithHistTime("14:00"),
	)
	if err != nil {
		t.Fatalf("HistoricalOptionQuote: %v", err)
	}
	if cap.Path != "/historical/optionquote/SPY" {
		t.Errorf("HistoricalOptionQuote path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "date=2025-01-15") {
		t.Errorf("HistoricalOptionQuote query = %q, want date param", cap.Query)
	}
}

// ── Pricing ───────────────────────────────────────────────────────────────────

func TestGreeks(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"delta": 0.5})
	r := 0.05
	_, err := client.Greeks(context.Background(), flashalpha.GreeksParams{
		Spot: 450, Strike: 450, DTE: 30, Sigma: 0.2, Type: "call", R: &r,
	})
	if err != nil {
		t.Fatalf("Greeks: %v", err)
	}
	if cap.Path != "/v1/pricing/greeks" {
		t.Errorf("Greeks path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "spot=450") {
		t.Errorf("Greeks query = %q, want spot param", cap.Query)
	}
	if !strings.Contains(cap.Query, "type=call") {
		t.Errorf("Greeks query = %q, want type param", cap.Query)
	}
}

func TestGreeksDefaultType(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"delta": 0.5})
	_, err := client.Greeks(context.Background(), flashalpha.GreeksParams{
		Spot: 100, Strike: 100, DTE: 30, Sigma: 0.2,
		// Type intentionally omitted — should default to "call"
	})
	if err != nil {
		t.Fatalf("Greeks default type: %v", err)
	}
	if !strings.Contains(cap.Query, "type=call") {
		t.Errorf("Greeks default type query = %q, want type=call", cap.Query)
	}
}

func TestIV(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"iv": 0.25})
	_, err := client.IV(context.Background(), flashalpha.IVParams{
		Spot: 450, Strike: 450, DTE: 30, Price: 10.5, Type: "call",
	})
	if err != nil {
		t.Fatalf("IV: %v", err)
	}
	if cap.Path != "/v1/pricing/iv" {
		t.Errorf("IV path = %q", cap.Path)
	}
}

func TestKelly(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"kelly_fraction": 0.12})
	_, err := client.Kelly(context.Background(), flashalpha.KellyParams{
		Spot: 450, Strike: 460, DTE: 14, Sigma: 0.2, Premium: 3.5, Mu: 0.08, Type: "call",
	})
	if err != nil {
		t.Fatalf("Kelly: %v", err)
	}
	if cap.Path != "/v1/pricing/kelly" {
		t.Errorf("Kelly path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "mu=0.08") {
		t.Errorf("Kelly query = %q, want mu param", cap.Query)
	}
}

// ── Volatility ────────────────────────────────────────────────────────────────

func TestVolatility(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"iv_rank": 55.0})
	_, err := client.Volatility(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("Volatility: %v", err)
	}
	if cap.Path != "/v1/volatility/SPY" {
		t.Errorf("Volatility path = %q", cap.Path)
	}
}

func TestAdvVolatility(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"svi": map[string]interface{}{}})
	_, err := client.AdvVolatility(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("AdvVolatility: %v", err)
	}
	if cap.Path != "/v1/adv_volatility/SPY" {
		t.Errorf("AdvVolatility path = %q", cap.Path)
	}
}

// ── Reference data ────────────────────────────────────────────────────────────

func TestTickers(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"tickers": []interface{}{"SPY", "QQQ"}})
	_, err := client.Tickers(context.Background())
	if err != nil {
		t.Fatalf("Tickers: %v", err)
	}
	if cap.Path != "/v1/tickers" {
		t.Errorf("Tickers path = %q", cap.Path)
	}
}

func TestOptions(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"chain": map[string]interface{}{}})
	_, err := client.Options(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("Options: %v", err)
	}
	if cap.Path != "/v1/options/SPY" {
		t.Errorf("Options path = %q", cap.Path)
	}
}

func TestSymbols(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"symbols": []interface{}{"SPY"}})
	_, err := client.Symbols(context.Background())
	if err != nil {
		t.Fatalf("Symbols: %v", err)
	}
	if cap.Path != "/v1/symbols" {
		t.Errorf("Symbols path = %q", cap.Path)
	}
}

// ── Error handling ────────────────────────────────────────────────────────────

func TestError401(t *testing.T) {
	_, client := newServer(t, 401, map[string]interface{}{"message": "invalid api key"})
	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	var authErr *flashalpha.AuthenticationError
	if !isType(err, &authErr) {
		t.Errorf("expected *AuthenticationError, got %T: %v", err, err)
	}
}

func TestError403(t *testing.T) {
	_, client := newServer(t, 403, map[string]interface{}{
		"message":       "tier restricted",
		"current_plan":  "free",
		"required_plan": "growth",
	})
	_, err := client.ExposureSummary(context.Background(), "SPY")
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
	te, ok := err.(*flashalpha.TierRestrictedError)
	if !ok {
		t.Fatalf("expected *TierRestrictedError, got %T", err)
	}
	if te.CurrentPlan != "free" {
		t.Errorf("CurrentPlan = %q, want free", te.CurrentPlan)
	}
	if te.RequiredPlan != "growth" {
		t.Errorf("RequiredPlan = %q, want growth", te.RequiredPlan)
	}
}

func TestError404(t *testing.T) {
	_, client := newServer(t, 404, map[string]interface{}{"detail": "symbol not found"})
	_, err := client.Gex(context.Background(), "UNKNOWN")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	var nfe *flashalpha.NotFoundError
	if !isType(err, &nfe) {
		t.Errorf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func TestError429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(429)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": "rate limit exceeded"})
	}))
	defer srv.Close()
	client := flashalpha.NewClientWithURL("test-key", srv.URL)

	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for 429, got nil")
	}
	rle, ok := err.(*flashalpha.RateLimitError)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rle.RetryAfter != 60 {
		t.Errorf("RetryAfter = %d, want 60", rle.RetryAfter)
	}
}

func TestError500(t *testing.T) {
	_, client := newServer(t, 500, map[string]interface{}{"message": "internal server error"})
	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	var se *flashalpha.ServerError
	if !isType(err, &se) {
		t.Errorf("expected *ServerError, got %T: %v", err, err)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// isType checks whether err can be assigned to target (a pointer-to-pointer to
// the desired concrete type). This avoids importing errors.As from Go 1.13.
func isType[T any](err error, target **T) bool {
	v, ok := err.(*T)
	if ok {
		*target = v
	}
	return ok
}
