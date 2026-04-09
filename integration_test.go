//go:build integration

package flashalpha_test

import (
	"context"
	"os"
	"testing"
	"time"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

// newIntegrationClient returns a Client built from the FLASHALPHA_API_KEY
// environment variable, or skips the test if the variable is not set.
func newIntegrationClient(t *testing.T) *flashalpha.Client {
	t.Helper()
	key := os.Getenv("FLASHALPHA_API_KEY")
	if key == "" {
		t.Skip("FLASHALPHA_API_KEY not set — skipping integration test")
	}
	return flashalpha.NewClient(key)
}

func integrationCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func TestIntegrationHealth(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	t.Logf("Health response: %v", got)
}

func TestIntegrationAccount(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Account(ctx)
	if err != nil {
		t.Fatalf("Account: %v", err)
	}
	t.Logf("Account response: %v", got)
}

func TestIntegrationTickers(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Tickers(ctx)
	if err != nil {
		t.Fatalf("Tickers: %v", err)
	}
	t.Logf("Tickers response keys: %v", keys(got))
}

func TestIntegrationSymbols(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Symbols(ctx)
	if err != nil {
		t.Fatalf("Symbols: %v", err)
	}
	t.Logf("Symbols response: %v", got)
}

func TestIntegrationGex(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Gex(ctx, "SPY")
	if err != nil {
		t.Fatalf("Gex SPY: %v", err)
	}
	t.Logf("Gex SPY response keys: %v", keys(got))
}

func TestIntegrationGexWithMinOI(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Gex(ctx, "SPY", flashalpha.WithMinOI(1000))
	if err != nil {
		t.Fatalf("Gex SPY min_oi: %v", err)
	}
	t.Logf("Gex SPY min_oi response keys: %v", keys(got))
}

func TestIntegrationDex(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Dex(ctx, "SPY")
	if err != nil {
		t.Fatalf("Dex SPY: %v", err)
	}
	t.Logf("Dex SPY response keys: %v", keys(got))
}

func TestIntegrationExposureLevels(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.ExposureLevels(ctx, "SPY")
	if err != nil {
		t.Fatalf("ExposureLevels SPY: %v", err)
	}
	t.Logf("ExposureLevels response keys: %v", keys(got))
}

func TestIntegrationSurface(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Surface(ctx, "SPY")
	if err != nil {
		t.Fatalf("Surface SPY: %v", err)
	}
	t.Logf("Surface response keys: %v", keys(got))
}

func TestIntegrationStockQuote(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.StockQuote(ctx, "SPY")
	if err != nil {
		t.Fatalf("StockQuote SPY: %v", err)
	}
	t.Logf("StockQuote SPY response: %v", got)
}

func TestIntegrationGreeks(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Greeks(ctx, flashalpha.GreeksParams{
		Spot:   450,
		Strike: 455,
		DTE:    30,
		Sigma:  0.2,
		Type:   "call",
	})
	if err != nil {
		t.Fatalf("Greeks: %v", err)
	}
	t.Logf("Greeks response: %v", got)
}

func TestIntegrationIV(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.IV(ctx, flashalpha.IVParams{
		Spot:   450,
		Strike: 450,
		DTE:    30,
		Price:  10.5,
		Type:   "call",
	})
	if err != nil {
		t.Fatalf("IV: %v", err)
	}
	t.Logf("IV response: %v", got)
}

func TestIntegrationVolatility(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Volatility(ctx, "SPY")
	if err != nil {
		t.Fatalf("Volatility SPY: %v", err)
	}
	t.Logf("Volatility response keys: %v", keys(got))
}

// ── Max Pain ─────────────────────────────────────────────────────────────────

func TestIntegrationMaxPain(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.MaxPain(ctx, "SPY")
	if err != nil {
		t.Fatalf("MaxPain SPY: %v", err)
	}
	if _, ok := got["max_pain_strike"]; !ok {
		t.Error("missing max_pain_strike")
	}
	if _, ok := got["pain_curve"]; !ok {
		t.Error("missing pain_curve")
	}
	if _, ok := got["dealer_alignment"]; !ok {
		t.Error("missing dealer_alignment")
	}
}

func TestIntegrationMaxPainFieldStructure(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.MaxPain(ctx, "SPY")
	if err != nil {
		t.Fatalf("MaxPain SPY: %v", err)
	}
	dist := got["distance"].(map[string]interface{})
	dir := dist["direction"].(string)
	if dir != "above" && dir != "below" && dir != "at" {
		t.Errorf("direction = %q", dir)
	}
	signal := got["signal"].(string)
	if signal != "bullish" && signal != "bearish" && signal != "neutral" {
		t.Errorf("signal = %q", signal)
	}
}

func TestIntegrationMaxPainMultiExpiry(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.MaxPain(ctx, "SPY")
	if err != nil {
		t.Fatalf("MaxPain SPY: %v", err)
	}
	if cal, ok := got["max_pain_by_expiration"].([]interface{}); ok {
		if len(cal) == 0 {
			t.Error("max_pain_by_expiration is empty")
		}
	}
}

// ── Screener ─────────────────────────────────────────────────────────────────

func TestIntegrationScreenerEmpty(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	got, err := client.Screener(ctx, flashalpha.ScreenerRequest{})
	if err != nil {
		t.Fatalf("Screener empty: %v", err)
	}
	if _, ok := got["meta"]; !ok {
		t.Errorf("missing meta key")
	}
	if _, ok := got["data"]; !ok {
		t.Errorf("missing data key")
	}
	meta := got["meta"].(map[string]interface{})
	tier := meta["tier"].(string)
	if tier != "growth" && tier != "alpha" {
		t.Errorf("tier = %q, want growth or alpha", tier)
	}
}

func TestIntegrationScreenerSimpleFilter(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	limit := 5
	got, err := client.Screener(ctx, flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{
			Field:    "regime",
			Operator: "in",
			Value:    []string{"positive_gamma", "negative_gamma"},
		},
		Select: []string{"symbol", "regime", "price"},
		Limit:  &limit,
	})
	if err != nil {
		t.Fatalf("Screener filter: %v", err)
	}
	data := got["data"].([]interface{})
	for _, row := range data {
		r := row.(map[string]interface{})
		regime := r["regime"].(string)
		if regime != "positive_gamma" && regime != "negative_gamma" {
			t.Errorf("unexpected regime: %s", regime)
		}
	}
}

func TestIntegrationScreenerAndGroupWithSort(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	limit := 5
	got, err := client.Screener(ctx, flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerGroup{
			Op: "and",
			Conditions: []interface{}{
				flashalpha.ScreenerLeaf{Field: "atm_iv", Operator: "gte", Value: 0},
				flashalpha.ScreenerLeaf{Field: "atm_iv", Operator: "lte", Value: 500},
			},
		},
		Sort:   []flashalpha.ScreenerSort{{Field: "atm_iv", Direction: "desc"}},
		Select: []string{"symbol", "atm_iv"},
		Limit:  &limit,
	})
	if err != nil {
		t.Fatalf("Screener AND group: %v", err)
	}
	data := got["data"].([]interface{})
	var prev *float64
	for _, row := range data {
		r := row.(map[string]interface{})
		iv, ok := r["atm_iv"].(float64)
		if !ok {
			continue // nil
		}
		if prev != nil && iv > *prev {
			t.Errorf("not sorted desc: %v > %v", iv, *prev)
		}
		prev = &iv
	}
}

func TestIntegrationScreenerSelectStar(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	limit := 1
	got, err := client.Screener(ctx, flashalpha.ScreenerRequest{
		Select: []string{"*"},
		Limit:  &limit,
	})
	if err != nil {
		t.Fatalf("Screener select *: %v", err)
	}
	data := got["data"].([]interface{})
	if len(data) > 0 {
		row := data[0].(map[string]interface{})
		if _, ok := row["symbol"]; !ok {
			t.Errorf("missing symbol in full flat object")
		}
		if _, ok := row["price"]; !ok {
			t.Errorf("missing price in full flat object")
		}
	}
}

func TestIntegrationScreenerLimitRespected(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	limit := 3
	got, err := client.Screener(ctx, flashalpha.ScreenerRequest{Limit: &limit})
	if err != nil {
		t.Fatalf("Screener limit: %v", err)
	}
	meta := got["meta"].(map[string]interface{})
	returned := int(meta["returned_count"].(float64))
	if returned > 3 {
		t.Errorf("returned_count = %d, want <= 3", returned)
	}
	if len(got["data"].([]interface{})) > 3 {
		t.Errorf("data len > 3")
	}
}

func TestIntegrationScreenerInvalidField(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	_, err := client.Screener(ctx, flashalpha.ScreenerRequest{
		Filters: flashalpha.ScreenerLeaf{
			Field: "not_a_real_field_xyz", Operator: "eq", Value: 1,
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
}

// keys returns the top-level keys of a map for logging.
func keys(m map[string]interface{}) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
