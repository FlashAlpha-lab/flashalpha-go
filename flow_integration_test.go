//go:build integration

package flashalpha_test

import (
	"testing"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

// Flow integration tests — hit the real /v1/flow/* surface (Alpha+) and
// assert every contract field is present on the live response (and on
// nested array-element shapes when the arrays are non-empty). They use the
// untyped methods so the assertion is against the raw wire JSON.

const flowSym = "SPY"

// requireKeys fails the test if any name in keys is absent from m.
func requireKeys(t *testing.T, m map[string]interface{}, keys []string, where string) {
	t.Helper()
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Errorf("%s: missing field %q", where, k)
		}
	}
}

// firstElem returns the first element of m[key] as a map, or nil + false.
func firstElem(m map[string]interface{}, key string) (map[string]interface{}, bool) {
	arr, ok := m[key].([]interface{})
	if !ok || len(arr) == 0 {
		return nil, false
	}
	el, ok := arr[0].(map[string]interface{})
	return el, ok
}

func TestIntegrationFlowLevels(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowLevels(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowLevels: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "underlying_price", "expiry",
		"live_gamma_flip", "live_call_wall", "live_put_wall", "live_max_pain"},
		"flow/levels")
}

func TestIntegrationFlowPinRisk(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowPinRisk(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowPinRisk: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "underlying_price", "expiry",
		"live_pin_risk", "magnet_strike", "distance_to_magnet_pct",
		"time_to_close_hours", "breakdown"}, "flow/pin-risk")
	if b, ok := r["breakdown"].(map[string]interface{}); ok {
		requireKeys(t, b, []string{"oi_score", "proximity_score", "time_score",
			"gamma_score"}, "flow/pin-risk.breakdown")
	} else {
		t.Errorf("flow/pin-risk: breakdown not an object")
	}
}

func TestIntegrationFlowSummary(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowSummary(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowSummary: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "underlying_price", "expiry",
		"flow_direction", "intraday_oi_delta", "contracts_with_flow",
		"contracts_total", "live_gex", "flow_gex_pct_shift"}, "flow/summary")
}

func TestIntegrationFlowOi(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOi(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowOi: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "expiry", "official_oi",
		"simulated_oi", "intraday_oi_delta", "oi_delta_confidence",
		"effective_oi", "contracts_total", "contracts_with_flow"}, "flow/oi")
}

func TestIntegrationFlowGex(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowGex(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowGex: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "underlying_price", "expiry",
		"live_net_gex", "live_net_gex_label", "live_gamma_flip", "strikes"},
		"flow/gex")
	if el, ok := firstElem(r, "strikes"); ok {
		requireKeys(t, el, []string{"strike", "call_gex", "put_gex", "net_gex",
			"call_oi", "put_oi", "call_volume", "put_volume"},
			"flow/gex.strikes[0]")
	} else {
		t.Errorf("flow/gex: expected non-empty strikes")
	}
}

func TestIntegrationFlowDex(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowDex(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowDex: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "underlying_price", "expiry",
		"live_net_dex", "strikes"}, "flow/dex")
	if el, ok := firstElem(r, "strikes"); ok {
		requireKeys(t, el, []string{"strike", "call_dex", "put_dex", "net_dex"},
			"flow/dex.strikes[0]")
	} else {
		t.Errorf("flow/dex: expected non-empty strikes")
	}
}

func TestIntegrationFlowDealerRisk(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowDealerRisk(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowDealerRisk: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "underlying_price", "expiry",
		"settled_net_gex", "live_net_gex", "flow_gex_adjustment",
		"flow_gex_pct_shift", "settled_net_dex", "live_net_dex",
		"flow_dex_adjustment", "flow_dex_pct_shift",
		"total_abs_delta_contracts", "contracts_with_flow", "flow_direction",
		"description"}, "flow/dealer-risk")
}

func TestIntegrationFlowLive(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowLive(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowLive: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "as_of", "underlying_price", "expiry",
		"contracts", "contracts_with_flow", "official_oi", "simulated_oi",
		"intraday_oi_delta", "oi_delta_confidence", "effective_oi", "live_gex",
		"live_gex_delta", "live_gamma_flip", "live_call_wall", "live_put_wall",
		"live_max_pain", "live_pin_risk", "flow_adjusted_dealer_risk"},
		"flow/live")
	if b, ok := r["flow_adjusted_dealer_risk"].(map[string]interface{}); ok {
		requireKeys(t, b, []string{"settled_net_gex", "live_net_gex",
			"flow_gex_adjustment", "flow_gex_pct_shift", "settled_net_dex",
			"live_net_dex", "flow_dex_adjustment", "flow_dex_pct_shift",
			"total_abs_delta_contracts", "flow_direction", "description"},
			"flow/live.flow_adjusted_dealer_risk")
	} else {
		t.Errorf("flow/live: flow_adjusted_dealer_risk not an object")
	}
}

func TestIntegrationFlowOptionRecent(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOptionRecent(ctx, flowSym, flashalpha.WithFlowLimit(5))
	if err != nil {
		t.Fatalf("FlowOptionRecent: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "count", "totalAvailable", "trades"},
		"flow/options/recent")
	if el, ok := firstElem(r, "trades"); ok {
		requireKeys(t, el, []string{"ts", "instrumentId", "expiry", "strike",
			"right", "price", "size", "side", "isBlock", "bid", "ask"},
			"flow/options/recent.trades[0]")
	}
}

func TestIntegrationFlowOptionSummary(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOptionSummary(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowOptionSummary: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "contractsWithTrades", "totalTrades",
		"buyVolume", "sellVolume", "midVolume", "netVolume",
		"biggestSingleTrade"}, "flow/options/summary")
}

func TestIntegrationFlowOptionBlocks(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOptionBlocks(ctx, flowSym, flashalpha.WithFlowMinSize(50))
	if err != nil {
		t.Fatalf("FlowOptionBlocks: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "minSize", "count", "blocks"},
		"flow/options/blocks")
	if el, ok := firstElem(r, "blocks"); ok {
		requireKeys(t, el, []string{"ts", "expiry", "strike", "right", "price",
			"size", "side"}, "flow/options/blocks.blocks[0]")
	}
}

func TestIntegrationFlowOptionHistory(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOptionHistory(ctx, flowSym, flashalpha.WithFlowMinutes(30))
	if err != nil {
		t.Fatalf("FlowOptionHistory: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "minutes", "count", "buckets"},
		"flow/options/history")
	if el, ok := firstElem(r, "buckets"); ok {
		requireKeys(t, el, []string{"ts", "buyVolume", "sellVolume",
			"midVolume", "netVolume", "tradeCount", "biggestTrade", "vwap",
			"high", "low"}, "flow/options/history.buckets[0]")
	}
}

func TestIntegrationFlowOptionCumulative(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOptionCumulative(ctx, flowSym, flashalpha.WithFlowMinutes(60))
	if err != nil {
		t.Fatalf("FlowOptionCumulative: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "minutes", "count", "points"},
		"flow/options/cumulative")
	if el, ok := firstElem(r, "points"); ok {
		requireKeys(t, el, []string{"ts", "netVolume", "cumulative", "vwap",
			"tradeCount"}, "flow/options/cumulative.points[0]")
	}
}

func TestIntegrationFlowStockRecent(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowStockRecent(ctx, flowSym, flashalpha.WithFlowLimit(5))
	if err != nil {
		t.Fatalf("FlowStockRecent: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "count", "totalAvailable", "trades"},
		"flow/stocks/recent")
	if el, ok := firstElem(r, "trades"); ok {
		requireKeys(t, el, []string{"ts", "price", "size", "side", "isBlock",
			"bid", "ask"}, "flow/stocks/recent.trades[0]")
	}
}

func TestIntegrationFlowStockSummary(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowStockSummary(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowStockSummary: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "totalTrades", "buyVolume",
		"sellVolume", "midVolume", "netVolume", "biggestSingleTrade"},
		"flow/stocks/summary")
}

func TestIntegrationFlowStockBlocks(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowStockBlocks(ctx, flowSym, flashalpha.WithFlowMinSize(1000))
	if err != nil {
		t.Fatalf("FlowStockBlocks: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "minSize", "count", "blocks"},
		"flow/stocks/blocks")
	if el, ok := firstElem(r, "blocks"); ok {
		requireKeys(t, el, []string{"ts", "price", "size", "side", "bid",
			"ask"}, "flow/stocks/blocks.blocks[0]")
	}
}

func TestIntegrationFlowStockHistory(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowStockHistory(ctx, flowSym, flashalpha.WithFlowMinutes(30))
	if err != nil {
		t.Fatalf("FlowStockHistory: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "minutes", "count", "buckets"},
		"flow/stocks/history")
	if el, ok := firstElem(r, "buckets"); ok {
		requireKeys(t, el, []string{"ts", "buyVolume", "sellVolume",
			"midVolume", "netVolume", "tradeCount", "biggestTrade", "vwap",
			"open", "close", "high", "low"}, "flow/stocks/history.buckets[0]")
	}
}

func TestIntegrationFlowStockCumulative(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowStockCumulative(ctx, flowSym, flashalpha.WithFlowMinutes(60))
	if err != nil {
		t.Fatalf("FlowStockCumulative: %v", err)
	}
	requireKeys(t, r, []string{"symbol", "minutes", "count", "points"},
		"flow/stocks/cumulative")
	if el, ok := firstElem(r, "points"); ok {
		requireKeys(t, el, []string{"ts", "netVolume", "cumulative", "vwap",
			"tradeCount"}, "flow/stocks/cumulative.points[0]")
	}
}

func TestIntegrationFlowOptionsLeaderboard(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOptionsLeaderboard(ctx, flashalpha.WithFlowN(3))
	if err != nil {
		t.Fatalf("FlowOptionsLeaderboard: %v", err)
	}
	requireKeys(t, r, []string{"generatedUtc", "n", "windowMinutes", "buyers",
		"sellers"}, "flow/options/leaderboard")
	for _, side := range []string{"buyers", "sellers"} {
		if el, ok := firstElem(r, side); ok {
			requireKeys(t, el, []string{"symbol", "netVolume", "netNotional",
				"buyVolume", "sellVolume", "avgPremium", "tradeCount",
				"lastTradeUtc"}, "flow/options/leaderboard."+side+"[0]")
			break
		}
	}
}

func TestIntegrationFlowOptionsOutliers(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowOptionsOutliers(ctx, flashalpha.WithFlowLimit(3))
	if err != nil {
		t.Fatalf("FlowOptionsOutliers: %v", err)
	}
	requireKeys(t, r, []string{"generatedUtc", "windowMinutes", "tracked",
		"qualified", "limit", "outliers"}, "flow/options/outliers")
	if el, ok := firstElem(r, "outliers"); ok {
		requireKeys(t, el, []string{"symbol", "tradeCount", "buyVolume",
			"sellVolume", "midVolume", "netVolume", "imbalancePct", "skew",
			"notional", "netNotional", "biggestTrade", "biggestTradeUtc",
			"biggestAgeSec", "lastVwap", "lastTradeUtc", "lastTradeAgeSec"},
			"flow/options/outliers.outliers[0]")
	}
}

func TestIntegrationFlowStocksLeaderboard(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowStocksLeaderboard(ctx, flashalpha.WithFlowN(3))
	if err != nil {
		t.Fatalf("FlowStocksLeaderboard: %v", err)
	}
	requireKeys(t, r, []string{"generatedUtc", "n", "windowMinutes", "buyers",
		"sellers"}, "flow/stocks/leaderboard")
	for _, side := range []string{"buyers", "sellers"} {
		if el, ok := firstElem(r, side); ok {
			requireKeys(t, el, []string{"symbol", "netVolume", "netNotional",
				"buyVolume", "sellVolume", "vwap", "tradeCount",
				"lastTradeUtc"}, "flow/stocks/leaderboard."+side+"[0]")
			break
		}
	}
}

func TestIntegrationFlowStocksOutliers(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()
	r, err := client.FlowStocksOutliers(ctx, flashalpha.WithFlowLimit(3))
	if err != nil {
		t.Fatalf("FlowStocksOutliers: %v", err)
	}
	requireKeys(t, r, []string{"generatedUtc", "windowMinutes", "tracked",
		"qualified", "limit", "outliers"}, "flow/stocks/outliers")
	if el, ok := firstElem(r, "outliers"); ok {
		requireKeys(t, el, []string{"symbol", "tradeCount", "buyVolume",
			"sellVolume", "midVolume", "netVolume", "imbalancePct", "skew",
			"notional", "netNotional", "biggestTrade", "biggestTradeUtc",
			"biggestAgeSec", "lastVwap", "lastTradeUtc", "lastTradeAgeSec"},
			"flow/stocks/outliers.outliers[0]")
	}
}

// Ensure the typed wrappers decode without error against the live API.
func TestIntegrationFlowTypedSmoke(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	lv, err := client.FlowLevelsTyped(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowLevelsTyped: %v", err)
	}
	if lv.Symbol != flowSym {
		t.Errorf("FlowLevelsTyped symbol = %q, want %q", lv.Symbol, flowSym)
	}
	gx, err := client.FlowGexTyped(ctx, flowSym)
	if err != nil {
		t.Fatalf("FlowGexTyped: %v", err)
	}
	if len(gx.Strikes) == 0 {
		t.Errorf("FlowGexTyped: expected non-empty strikes")
	}
	if _, err := client.FlowLiveTyped(ctx, flowSym); err != nil {
		t.Fatalf("FlowLiveTyped: %v", err)
	}
}
