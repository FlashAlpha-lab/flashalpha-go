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

// ── Customer regression tests ────────────────────────────────────────────────
//
// Each test below replays one of the bugs an Alpha-tier user hit during
// integration of an automated trading daemon. All tests call PUBLIC SDK
// methods only — no raw HTTP, no unexported access, no mocks. The goal is
// to lock in the SDK's exposed surface against regressions.

// Issue #5 — SDK was missing Vrp(). The method now exists on the client.

func TestVrp_ReturnsFullPayload(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}

	// Top-level scalars
	if r.Symbol != "SPY" {
		t.Errorf("Symbol = %q, want SPY", r.Symbol)
	}
	if r.UnderlyingPrice <= 0 {
		t.Errorf("UnderlyingPrice = %v, want > 0", r.UnderlyingPrice)
	}
	if r.AsOf == "" {
		t.Error("AsOf is empty")
	}
	if r.NetHarvestScore == nil {
		t.Error("NetHarvestScore is nil at top level")
	}
	if r.DealerFlowRisk == nil {
		t.Error("DealerFlowRisk is nil at top level")
	}

	// Vrp core block
	core := r.Vrp
	if core.ZScore == nil {
		t.Error("Vrp.ZScore is nil")
	}
	if core.Percentile == nil {
		t.Error("Vrp.Percentile is nil")
	} else if *core.Percentile < 0 || *core.Percentile > 100 {
		t.Errorf("Vrp.Percentile = %d out of [0,100]", *core.Percentile)
	}
	if core.AtmIv == nil {
		t.Error("Vrp.AtmIv is nil")
	}
	if core.Rv20d == nil {
		t.Error("Vrp.Rv20d is nil")
	}
	if core.Vrp20d == nil {
		t.Error("Vrp.Vrp20d is nil")
	}

	// Directional skew — canonical names; assert raw JSON tag presence
	dir, ok := r.Raw["directional"].(map[string]interface{})
	if !ok {
		t.Fatal("Raw.directional missing or wrong type")
	}
	for _, k := range []string{
		"put_wing_iv_25d", "call_wing_iv_25d",
		"downside_rv_20d", "upside_rv_20d",
		"downside_vrp", "upside_vrp",
	} {
		if _, ok := dir[k]; !ok {
			t.Errorf("directional.%s missing in raw payload", k)
		}
	}

	// Regime block
	switch r.Regime.Gamma {
	case "positive_gamma", "negative_gamma", "neutral":
	default:
		t.Errorf("Regime.Gamma = %q unexpected", r.Regime.Gamma)
	}

	// Term VRP
	if r.TermVrp == nil {
		t.Error("TermVrp slice is nil")
	}

	// Gex-conditioned (nullable)
	if r.GexConditioned != nil && r.GexConditioned.Regime == "" {
		t.Error("GexConditioned.Regime empty")
	}

	// Strategy scores (nullable)
	if r.StrategyScores != nil {
		ss := r.StrategyScores
		check := func(name string, v *int) {
			if v != nil && (*v < 0 || *v > 100) {
				t.Errorf("StrategyScores.%s = %d out of [0,100]", name, *v)
			}
		}
		check("ShortPutSpread", ss.ShortPutSpread)
		check("ShortStrangle", ss.ShortStrangle)
		check("IronCondor", ss.IronCondor)
		check("CalendarSpread", ss.CalendarSpread)
	}
}

// Issue #1 — Nested response structures. The customer accessed
// top-level keys that don't exist; data lives under sub-objects.

func TestVrp_CoreMetrics_AreNested(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}
	// Raw payload must NOT have these at the top level (customer trap).
	for _, k := range []string{"z_score", "percentile", "atm_iv", "rv_20d", "vrp_20d"} {
		if _, ok := r.Raw[k]; ok {
			t.Errorf("raw[%q] present at top level — must be nested under 'vrp'", k)
		}
	}
	// Typed access via the canonical nested path.
	if r.Vrp.ZScore == nil {
		t.Error("r.Vrp.ZScore is nil — should be populated for SPY")
	}
}

func TestVrp_HarvestScore_UnderGexConditioned(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}
	if _, ok := r.Raw["harvest_score"]; ok {
		t.Error("raw.harvest_score present at top level — must live under gex_conditioned")
	}
	if r.GexConditioned != nil {
		gc, ok := r.Raw["gex_conditioned"].(map[string]interface{})
		if !ok {
			t.Fatal("raw.gex_conditioned missing or wrong type")
		}
		for _, k := range []string{"harvest_score", "regime", "interpretation"} {
			if _, ok := gc[k]; !ok {
				t.Errorf("gex_conditioned.%s missing", k)
			}
		}
	}
}

func TestVrp_NetGex_UnderRegime(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}
	if _, ok := r.Raw["net_gex"]; ok {
		t.Error("raw.net_gex present at top level on vrp payload — must be under regime")
	}
	if _, ok := r.Raw["gamma_flip"]; ok {
		t.Error("raw.gamma_flip present at top level on vrp payload — must be under regime")
	}
	reg, ok := r.Raw["regime"].(map[string]interface{})
	if !ok {
		t.Fatal("raw.regime missing")
	}
	if _, ok := reg["net_gex"]; !ok {
		t.Error("regime.net_gex missing")
	}
	if _, ok := reg["gamma"]; !ok {
		t.Error("regime.gamma missing")
	}
}

func TestVrp_CompositeScores_TopLevel(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}
	if _, ok := r.Raw["net_harvest_score"]; !ok {
		t.Error("raw.net_harvest_score missing — should be top-level")
	}
	if _, ok := r.Raw["dealer_flow_risk"]; !ok {
		t.Error("raw.dealer_flow_risk missing — should be top-level")
	}
	if r.NetHarvestScore == nil {
		t.Error("typed NetHarvestScore is nil")
	}
	if r.DealerFlowRisk == nil {
		t.Error("typed DealerFlowRisk is nil")
	}
}

func TestExposureSummary_NetGex_UnderExposures(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.ExposureSummary(ctx, "SPY")
	if err != nil {
		t.Fatalf("ExposureSummary SPY: %v", err)
	}
	if sym, _ := r["symbol"].(string); sym != "SPY" {
		t.Errorf("symbol = %q, want SPY", sym)
	}
	if _, ok := r["net_gex"]; ok {
		t.Error("net_gex present at top level — must live under exposures")
	}
	exp, ok := r["exposures"].(map[string]interface{})
	if !ok {
		t.Fatal("exposures missing or wrong type")
	}
	if v, ok := exp["net_gex"]; !ok {
		t.Error("exposures.net_gex missing")
	} else if _, isNum := v.(float64); !isNum {
		t.Errorf("exposures.net_gex not numeric: %T", v)
	}
	for _, k := range []string{"net_dex", "net_vex", "net_chex"} {
		if v, ok := exp[k]; ok {
			if _, isNum := v.(float64); !isNum {
				t.Errorf("exposures.%s not numeric: %T", k, v)
			}
		}
	}
	regime, ok := r["regime"].(string)
	if !ok {
		t.Fatal("regime missing or not a string at top level")
	}
	switch regime {
	case "positive_gamma", "negative_gamma", "neutral":
	default:
		t.Errorf("regime = %q unexpected", regime)
	}
}

// Issue #2 — Field naming. downside_vrp / upside_vrp are canonical;
// put_vrp / call_vrp are silent-null traps.

func TestVrp_Directional_UsesDownsideUpsideNames(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}
	d, ok := r.Raw["directional"].(map[string]interface{})
	if !ok {
		t.Fatal("raw.directional missing")
	}
	for _, k := range []string{
		"downside_vrp", "upside_vrp",
		"put_wing_iv_25d", "call_wing_iv_25d",
	} {
		if _, ok := d[k]; !ok {
			t.Errorf("directional.%s missing", k)
		}
	}
	for _, k := range []string{"put_vrp", "call_vrp"} {
		if _, ok := d[k]; ok {
			t.Errorf("directional.%s present — should NOT exist (use downside_vrp/upside_vrp)", k)
		}
	}
}

// Issue #3 — URL pattern mix. SDK methods route to canonical paths.

func TestStockSummary_RoutesCorrectly(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.StockSummary(ctx, "SPY")
	if err != nil {
		t.Fatalf("StockSummary SPY: %v", err)
	}
	if sym, _ := r["symbol"].(string); sym != "SPY" {
		t.Errorf("symbol = %q, want SPY", sym)
	}
	if _, ok := r["price"]; !ok {
		t.Error("price missing")
	}
	if len(r) < 4 {
		t.Errorf("payload looks empty: %d keys", len(r))
	}
}

func TestStockQuote_RoutesCorrectly(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.StockQuote(ctx, "SPY")
	if err != nil {
		t.Fatalf("StockQuote SPY: %v", err)
	}
	if tk, _ := r["ticker"].(string); tk != "SPY" {
		t.Errorf("ticker = %q, want SPY", tk)
	}
}

func TestAllExposureMethods_RouteCorrectly(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	calls := map[string]func() (map[string]interface{}, error){
		"Gex":             func() (map[string]interface{}, error) { return client.Gex(ctx, "SPY") },
		"Dex":             func() (map[string]interface{}, error) { return client.Dex(ctx, "SPY") },
		"Vex":             func() (map[string]interface{}, error) { return client.Vex(ctx, "SPY") },
		"Chex":            func() (map[string]interface{}, error) { return client.Chex(ctx, "SPY") },
		"ExposureSummary": func() (map[string]interface{}, error) { return client.ExposureSummary(ctx, "SPY") },
		"ExposureLevels":  func() (map[string]interface{}, error) { return client.ExposureLevels(ctx, "SPY") },
	}
	for name, fn := range calls {
		got, err := fn()
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		sym, _ := got["symbol"].(string)
		if sym != "SPY" {
			t.Errorf("%s: symbol = %q, want SPY", name, sym)
		}
	}
}

func TestVrp_MethodRoutesCorrectly(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}
	if r.Symbol != "SPY" {
		t.Errorf("symbol = %q, want SPY", r.Symbol)
	}
}

// Issue #4 — Screener URL is /v1/screener (renamed v0.3.1).

func TestScreener_ReturnsValidEnvelope(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	limit := 5
	r, err := client.Screener(ctx, flashalpha.ScreenerRequest{Limit: &limit})
	if err != nil {
		t.Fatalf("Screener: %v", err)
	}
	meta, ok := r["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta missing or wrong type")
	}
	for _, k := range []string{"total_count", "returned_count", "universe_size", "tier", "as_of"} {
		if _, ok := meta[k]; !ok {
			t.Errorf("meta.%s missing", k)
		}
	}
	returned, ok := meta["returned_count"].(float64)
	if !ok {
		t.Fatal("meta.returned_count not numeric")
	}
	if int(returned) > limit {
		t.Errorf("returned_count = %d > limit %d", int(returned), limit)
	}
	tier, _ := meta["tier"].(string)
	if tier != "growth" && tier != "alpha" {
		t.Errorf("tier = %q, want growth or alpha", tier)
	}
	if _, ok := r["data"]; !ok {
		t.Error("data missing")
	}
}

func TestScreener_FullRow_Readable(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	limit := 1
	r, err := client.Screener(ctx, flashalpha.ScreenerRequest{
		Select: []string{"*"},
		Limit:  &limit,
	})
	if err != nil {
		t.Fatalf("Screener select *: %v", err)
	}
	data, ok := r["data"].([]interface{})
	if !ok {
		t.Fatal("data missing or wrong type")
	}
	if len(data) == 0 {
		t.Skip("no rows returned for universe")
	}
	row, ok := data[0].(map[string]interface{})
	if !ok {
		t.Fatal("row not an object")
	}
	for _, k := range []string{"symbol", "price", "regime"} {
		if _, ok := row[k]; !ok {
			t.Errorf("row missing %q", k)
		}
	}
}
