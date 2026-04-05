package flashalpha_test

import (
	"context"
	"encoding/json"
	"errors"
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

// ── Screener ──────────────────────────────────────────────────────────────────

func TestScreenerEmpty(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		buf := new(strings.Builder)
		_, _ = buf.ReadFrom(r.Body)
		gotBody = buf.String()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"meta":{"tier":"growth"},"data":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := flashalpha.NewClientWithURL("test-key", srv.URL)

	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{})
	if err != nil {
		t.Fatalf("Screener: unexpected error: %v", err)
	}
	if gotPath != "/v1/screener/live" {
		t.Errorf("path = %q, want /v1/screener/live", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody != "{}" {
		t.Errorf("body = %q, want {}", gotBody)
	}
}

func TestScreenerWithFilters(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		_, _ = buf.ReadFrom(r.Body)
		gotBody = buf.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"meta":{},"data":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := flashalpha.NewClientWithURL("test-key", srv.URL)

	limit := 20
	req := flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerGroup{
			Op: "and",
			Conditions: []interface{}{
				flashalpha.ScreenerLeaf{Field: "regime", Operator: "eq", Value: "positive_gamma"},
				flashalpha.ScreenerLeaf{Field: "harvest_score", Operator: "gte", Value: 65},
			},
		},
		Sort:   []flashalpha.ScreenerSort{{Field: "harvest_score", Direction: "desc"}},
		Select: []string{"symbol", "price", "harvest_score"},
		Limit:  &limit,
	}
	_, err := client.Screener(context.Background(), req)
	if err != nil {
		t.Fatalf("Screener: unexpected error: %v", err)
	}
	if !strings.Contains(gotBody, `"op":"and"`) {
		t.Errorf("body missing op=and: %s", gotBody)
	}
	if !strings.Contains(gotBody, "positive_gamma") {
		t.Errorf("body missing positive_gamma: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"limit":20`) {
		t.Errorf("body missing limit=20: %s", gotBody)
	}
}

// newBodyCapturingServer records the body, method, and headers; returns a
// configurable response status + body.
func newBodyCapturingServer(t *testing.T, status int, respBody string) (*flashalpha.Client, *struct {
	Method string
	Path   string
	Body   string
	Header http.Header
}) {
	t.Helper()
	cap := &struct {
		Method string
		Path   string
		Body   string
		Header http.Header
	}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.Method = r.Method
		cap.Path = r.URL.Path
		cap.Header = r.Header.Clone()
		buf := new(strings.Builder)
		_, _ = buf.ReadFrom(r.Body)
		cap.Body = buf.String()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(srv.Close)
	client := flashalpha.NewClientWithURL("test-key", srv.URL)
	return client, cap
}

func TestScreenerContentTypeHeader(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", cap.Header.Get("Content-Type"))
	}
	if cap.Header.Get("X-Api-Key") != "test-key" {
		t.Errorf("X-Api-Key header missing")
	}
}

func TestScreenerLeafFilter(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{Field: "regime", Operator: "eq", Value: "positive_gamma"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"field":"regime"`) {
		t.Errorf("body missing field: %s", cap.Body)
	}
	if !strings.Contains(cap.Body, `"operator":"eq"`) {
		t.Errorf("body missing operator: %s", cap.Body)
	}
}

func TestScreenerOrGroup(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerGroup{
			Op: "or",
			Conditions: []interface{}{
				flashalpha.ScreenerLeaf{Field: "vrp_regime", Operator: "eq", Value: "toxic_short_vol"},
				flashalpha.ScreenerLeaf{Field: "vrp_regime", Operator: "eq", Value: "event_only"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"op":"or"`) {
		t.Errorf("body missing op=or: %s", cap.Body)
	}
	if !strings.Contains(cap.Body, "toxic_short_vol") {
		t.Errorf("body missing toxic_short_vol: %s", cap.Body)
	}
}

func TestScreenerNestedAndInsideOr(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerGroup{
			Op: "or",
			Conditions: []interface{}{
				flashalpha.ScreenerGroup{
					Op: "and",
					Conditions: []interface{}{
						flashalpha.ScreenerLeaf{Field: "regime", Operator: "eq", Value: "positive_gamma"},
						flashalpha.ScreenerLeaf{Field: "harvest_score", Operator: "gte", Value: 70},
					},
				},
				flashalpha.ScreenerGroup{
					Op: "and",
					Conditions: []interface{}{
						flashalpha.ScreenerLeaf{Field: "regime", Operator: "eq", Value: "negative_gamma"},
						flashalpha.ScreenerLeaf{Field: "atm_iv", Operator: "gte", Value: 50},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"op":"or"`) || !strings.Contains(cap.Body, `"op":"and"`) {
		t.Errorf("body missing nested ops: %s", cap.Body)
	}
}

func TestScreenerBetweenOperator(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{
			Field: "atm_iv", Operator: "between", Value: []int{15, 25},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"operator":"between"`) {
		t.Errorf("body missing operator=between: %s", cap.Body)
	}
	if !strings.Contains(cap.Body, "[15,25]") {
		t.Errorf("body missing [15,25]: %s", cap.Body)
	}
}

func TestScreenerInOperator(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{
			Field: "term_state", Operator: "in", Value: []string{"contango", "mixed"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"operator":"in"`) || !strings.Contains(cap.Body, "contango") {
		t.Errorf("body missing operator=in/contango: %s", cap.Body)
	}
}

func TestScreenerIsNotNullOperator(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{Field: "vrp_regime", Operator: "is_not_null"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"operator":"is_not_null"`) {
		t.Errorf("body missing is_not_null: %s", cap.Body)
	}
	// Value should be omitted via omitempty
	if strings.Contains(cap.Body, `"value"`) {
		t.Errorf("body should not have value field for is_not_null: %s", cap.Body)
	}
}

func TestScreenerCascadingFilters(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerGroup{
			Op: "and",
			Conditions: []interface{}{
				flashalpha.ScreenerLeaf{Field: "regime", Operator: "eq", Value: "positive_gamma"},
				flashalpha.ScreenerLeaf{Field: "expiries.days_to_expiry", Operator: "lte", Value: 14},
				flashalpha.ScreenerLeaf{Field: "strikes.call_oi", Operator: "gte", Value: 2000},
				flashalpha.ScreenerLeaf{Field: "contracts.type", Operator: "eq", Value: "C"},
			},
		},
		Select: []string{"*"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, need := range []string{"expiries.days_to_expiry", "strikes.call_oi", "contracts.type"} {
		if !strings.Contains(cap.Body, need) {
			t.Errorf("body missing %s: %s", need, cap.Body)
		}
	}
}

func TestScreenerFormulas(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Formulas: []flashalpha.ScreenerFormula{
			{Alias: "vrp_ratio", Expression: "atm_iv / rv_20d"},
		},
		Filters: flashalpha.ScreenerLeaf{Formula: "vrp_ratio", Operator: "gte", Value: 1.2},
		Sort:    []flashalpha.ScreenerSort{{Formula: "vrp_ratio", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"alias":"vrp_ratio"`) {
		t.Errorf("body missing alias: %s", cap.Body)
	}
	if !strings.Contains(cap.Body, `"formula":"vrp_ratio"`) {
		t.Errorf("body missing formula ref: %s", cap.Body)
	}
}

func TestScreenerMultiSort(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Sort: []flashalpha.ScreenerSort{
			{Field: "dealer_flow_risk", Direction: "asc"},
			{Field: "harvest_score", Direction: "desc"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"direction":"asc"`) || !strings.Contains(cap.Body, `"direction":"desc"`) {
		t.Errorf("body missing multi-sort: %s", cap.Body)
	}
}

func TestScreenerPagination(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	limit, offset := 10, 10
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Limit: &limit, Offset: &offset,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"limit":10`) || !strings.Contains(cap.Body, `"offset":10`) {
		t.Errorf("body missing pagination: %s", cap.Body)
	}
}

func TestScreenerNegativeNumber(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{Field: "net_gex", Operator: "lt", Value: -500000},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, "-500000") {
		t.Errorf("body missing -500000: %s", cap.Body)
	}
}

func TestScreenerParsesResponse(t *testing.T) {
	client, _ := newBodyCapturingServer(t, 200,
		`{"meta":{"total_count":7,"tier":"alpha","universe_size":250},"data":[{"symbol":"SPY","price":656.01}]}`)
	result, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta := result["meta"].(map[string]interface{})
	if meta["tier"] != "alpha" {
		t.Errorf("tier = %v, want alpha", meta["tier"])
	}
	data := result["data"].([]interface{})
	row := data[0].(map[string]interface{})
	if row["symbol"] != "SPY" || row["price"].(float64) != 656.01 {
		t.Errorf("row = %v", row)
	}
}

func TestScreenerRaw(t *testing.T) {
	client, cap := newBodyCapturingServer(t, 200, `{"meta":{},"data":[]}`)
	_, err := client.ScreenerRaw(context.Background(), map[string]interface{}{
		"limit":  5,
		"select": []string{"symbol", "price"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cap.Body, `"limit":5`) {
		t.Errorf("body missing limit: %s", cap.Body)
	}
}

func TestScreenerThrows400ValidationError(t *testing.T) {
	client, _ := newBodyCapturingServer(t, 400,
		`{"status":"ERROR","error":"validation_error","message":"Field 'harvest_score' requires the Alpha plan or higher."}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{Field: "harvest_score", Operator: "gte", Value: 65},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Alpha") {
		t.Errorf("expected error to mention Alpha: %v", err)
	}
}

func TestScreenerThrows403TierRestricted(t *testing.T) {
	client, _ := newBodyCapturingServer(t, 403,
		`{"status":"ERROR","error":"tier_restricted","message":"Screener requires Growth plan.","current_plan":"Free","required_plan":"Growth"}`)
	_, err := client.Screener(context.Background(), flashalpha.ScreenerRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	var tierErr *flashalpha.TierRestrictedError
	if !errors.As(err, &tierErr) {
		t.Errorf("expected TierRestrictedError, got %T: %v", err, err)
		return
	}
	if tierErr.CurrentPlan != "Free" || tierErr.RequiredPlan != "Growth" {
		t.Errorf("plans = %q / %q", tierErr.CurrentPlan, tierErr.RequiredPlan)
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
