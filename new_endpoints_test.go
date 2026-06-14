package flashalpha_test

import (
	"context"
	"strings"
	"testing"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

// ── Exposure extras ──────────────────────────────────────────────────────────

func TestExposureSheet(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol": "SPY",
		"totals": map[string]interface{}{"net_gex": float64(2850000000)},
		"strikes": []interface{}{
			map[string]interface{}{"strike": float64(595), "net_gex": float64(56000000)},
		},
	})
	resp, err := client.ExposureSheetTyped(context.Background(), "SPY",
		flashalpha.WithSheetExpiration("2026-06-18"), flashalpha.WithSheetMinOI(100))
	if err != nil {
		t.Fatalf("ExposureSheetTyped: %v", err)
	}
	if cap.Path != "/v1/exposure/sheet/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "expiration=2026-06-18") || !strings.Contains(cap.Query, "min_oi=100") {
		t.Errorf("query = %q, want expiration + min_oi", cap.Query)
	}
	if resp.Totals.NetGex != 2850000000 {
		t.Errorf("Totals.NetGex = %v", resp.Totals.NetGex)
	}
	if len(resp.Strikes) != 1 || resp.Strikes[0].Strike != 595 {
		t.Errorf("Strikes = %+v", resp.Strikes)
	}
}

func TestExposureTermStructure(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":        "SPY",
		"by_dte_bucket": []interface{}{map[string]interface{}{"bucket": "0-7d", "net_gex": float64(920000000)}},
	})
	resp, err := client.ExposureTermStructureTyped(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("ExposureTermStructureTyped: %v", err)
	}
	if cap.Path != "/v1/exposure/term-structure/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if len(resp.ByDteBucket) != 1 || resp.ByDteBucket[0].Bucket != "0-7d" {
		t.Errorf("ByDteBucket = %+v", resp.ByDteBucket)
	}
}

func TestExposureBasket(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"constituent_count": float64(3),
		"aggregate":         map[string]interface{}{"net_gex": float64(1640000000)},
	})
	resp, err := client.ExposureBasketTyped(context.Background(), "AAPL,MSFT,NVDA",
		flashalpha.WithBasketWeights("0.4,0.3,0.3"))
	if err != nil {
		t.Fatalf("ExposureBasketTyped: %v", err)
	}
	if cap.Path != "/v1/exposure/basket" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "symbols=AAPL%2CMSFT%2CNVDA") || !strings.Contains(cap.Query, "weights=") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.ConstituentCount != 3 || resp.Aggregate.NetGex != 1640000000 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestExposureOiDiff(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":                   "SPY",
		"prior_snapshot_available": true,
		"total_call_oi_change":     float64(124000),
	})
	resp, err := client.ExposureOiDiffTyped(context.Background(), "SPY", flashalpha.WithOiDiffTopN(5))
	if err != nil {
		t.Fatalf("ExposureOiDiffTyped: %v", err)
	}
	if cap.Path != "/v1/exposure/oi-diff/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "topN=5") {
		t.Errorf("query = %q", cap.Query)
	}
	if !resp.PriorSnapshotAvailable || resp.TotalCallOiChange != 124000 {
		t.Errorf("resp = %+v", resp)
	}
}

// ── Volatility / macro / universe / surface / expected-move ──────────────────

func TestLiquidity(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":                "SPY",
		"chain_execution_score": float64(78),
		"expiries":              []interface{}{map[string]interface{}{"expiration": "2026-05-30", "label": "tight"}},
	})
	resp, err := client.LiquidityTyped(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("LiquidityTyped: %v", err)
	}
	if cap.Path != "/v1/liquidity/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.ChainExecutionScore != 78 || resp.Expiries[0].Label != "tight" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSkewTerm(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":   "SPY",
		"expiries": []interface{}{map[string]interface{}{"expiry": "2026-05-30", "skew_25d": float64(4.7)}},
	})
	resp, err := client.SkewTermTyped(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("SkewTermTyped: %v", err)
	}
	if cap.Path != "/v1/volatility/skew-term/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.Expiries[0].Skew25d == nil || *resp.Expiries[0].Skew25d != 4.7 {
		t.Errorf("Skew25d = %+v", resp.Expiries[0].Skew25d)
	}
}

func TestSpotVolCorrelation(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":                   "SPY",
		"spot_vol_correlation_20d": float64(-0.62),
	})
	resp, err := client.SpotVolCorrelationTyped(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("SpotVolCorrelationTyped: %v", err)
	}
	if cap.Path != "/v1/volatility/spot-vol-correlation/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.SpotVolCorrelation20d == nil || *resp.SpotVolCorrelation20d != -0.62 {
		t.Errorf("corr = %+v", resp.SpotVolCorrelation20d)
	}
}

func TestDispersion(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"index":               "SPX",
		"implied_correlation": float64(0.412),
	})
	resp, err := client.DispersionTyped(context.Background(), "SPX", "AAPL,MSFT,NVDA",
		flashalpha.WithDispersionWeights("0.5,0.3,0.2"), flashalpha.WithDispersionHorizonDays(20))
	if err != nil {
		t.Fatalf("DispersionTyped: %v", err)
	}
	if cap.Path != "/v1/dispersion" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "index=SPX") || !strings.Contains(cap.Query, "horizon_days=20") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.ImpliedCorrelation == nil || *resp.ImpliedCorrelation != 0.412 {
		t.Errorf("ImpliedCorrelation = %+v", resp.ImpliedCorrelation)
	}
}

func TestVixState(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"vix":   float64(18.4),
		"state": "overvixing",
	})
	resp, err := client.VixStateTyped(context.Background())
	if err != nil {
		t.Fatalf("VixStateTyped: %v", err)
	}
	if cap.Path != "/v1/macro/vix-state" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.State != "overvixing" {
		t.Errorf("State = %q", resp.State)
	}
}

func TestUniverse(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"count":   float64(252),
		"symbols": []interface{}{map[string]interface{}{"symbol": "AAPL", "tier": float64(1)}},
	})
	resp, err := client.UniverseTyped(context.Background(),
		flashalpha.WithUniverseSort("symbol"), flashalpha.WithUniverseLimit(20))
	if err != nil {
		t.Fatalf("UniverseTyped: %v", err)
	}
	if cap.Path != "/v1/universe" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "sort=symbol") || !strings.Contains(cap.Query, "limit=20") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Count != 252 || resp.Symbols[0].Symbol != "AAPL" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSurfaceSvi(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":         "SPY",
		"svi_parameters": []interface{}{map[string]interface{}{"expiry": "2026-06-20", "atm_iv": float64(16.42)}},
	})
	resp, err := client.SurfaceSviTyped(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("SurfaceSviTyped: %v", err)
	}
	if cap.Path != "/v1/surface/svi/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.SviParameters[0].AtmIv != 16.42 {
		t.Errorf("AtmIv = %v", resp.SviParameters[0].AtmIv)
	}
}

func TestExpectedMove(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol": "SPY",
		"expected_moves": []interface{}{
			map[string]interface{}{"expiry": "2026-06-20", "expectedMovePct": float64(3.25)},
		},
	})
	resp, err := client.ExpectedMoveTyped(context.Background(), "SPY",
		flashalpha.WithExpectedMoveExpiry("2026-06-20"))
	if err != nil {
		t.Fatalf("ExpectedMoveTyped: %v", err)
	}
	if cap.Path != "/v1/expected-move/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "expiry=2026-06-20") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.ExpectedMoves[0].ExpectedMovePct != 3.25 {
		t.Errorf("ExpectedMovePct = %v", resp.ExpectedMoves[0].ExpectedMovePct)
	}
}

// ── VRP param-change + history ───────────────────────────────────────────────

func TestVrpWithDate(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"symbol": "SPY"})
	_, err := client.Vrp(context.Background(), "SPY", flashalpha.WithVrpDate("2026-06-01"))
	if err != nil {
		t.Fatalf("Vrp: %v", err)
	}
	if cap.Path != "/v1/vrp/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "date=2026-06-01") {
		t.Errorf("query = %q, want date param", cap.Query)
	}
}

func TestVrpWithoutDateBackCompat(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"symbol": "SPY"})
	_, err := client.Vrp(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("Vrp: %v", err)
	}
	if cap.Query != "" {
		t.Errorf("query = %q, want empty", cap.Query)
	}
}

func TestZeroDteTypedWithExpiry(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"symbol": "SPY"})
	_, err := client.ZeroDteTyped(context.Background(), "SPY",
		flashalpha.WithStrikeRange(0.05), flashalpha.WithZeroDteExpiry("2026-06-08"))
	if err != nil {
		t.Fatalf("ZeroDteTyped: %v", err)
	}
	if cap.Path != "/v1/exposure/zero-dte/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "expiry=2026-06-08") || !strings.Contains(cap.Query, "strike_range=") {
		t.Errorf("query = %q, want expiry + strike_range params", cap.Query)
	}
}

func TestVrpHistory(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":      "SPY",
		"data_points": float64(2),
		"history":     []interface{}{map[string]interface{}{"date": "2026-06-01", "vrp_20d": float64(3.1)}},
	})
	resp, err := client.VrpHistoryTyped(context.Background(), "SPY", flashalpha.WithVrpHistoryDays(60))
	if err != nil {
		t.Fatalf("VrpHistoryTyped: %v", err)
	}
	if cap.Path != "/v1/vrp/SPY/history" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "days=60") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.DataPoints != 2 || resp.History[0].Date != "2026-06-01" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestRealizedVolatility(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":           "AAPL",
		"underlying_price": float64(201.5),
		"estimators": map[string]interface{}{
			"close_to_close": map[string]interface{}{"rv10": float64(18.4), "rv20": float64(21.2), "rv30": float64(22.7)},
			"yang_zhang":     map[string]interface{}{"rv10": float64(17.0), "rv20": float64(19.7), "rv30": float64(20.3)},
		},
	})
	resp, err := client.RealizedVolatility(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("RealizedVolatility: %v", err)
	}
	if cap.Path != "/v1/volatility/realized/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.Estimators.CloseToClose.Rv10 == nil || *resp.Estimators.CloseToClose.Rv10 != 18.4 {
		t.Errorf("CloseToClose.Rv10 = %+v", resp.Estimators.CloseToClose.Rv10)
	}
	if resp.Estimators.YangZhang.Rv30 == nil || *resp.Estimators.YangZhang.Rv30 != 20.3 {
		t.Errorf("YangZhang.Rv30 = %+v", resp.Estimators.YangZhang.Rv30)
	}
}

func TestVolatilityForecast(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol": "AAPL",
		"ewma":   map[string]interface{}{"lambda": float64(0.94), "vol_annualized": float64(19.6), "next_day_forecast": float64(19.6)},
		"har_rv": map[string]interface{}{
			"vol_annualized":    float64(20.3),
			"components":        map[string]interface{}{"daily": float64(18.4), "weekly": float64(21.2), "monthly": float64(22.7)},
			"next_day_forecast": float64(20.1),
		},
		"garch": map[string]interface{}{
			"model":        "garch_1_1",
			"distribution": "student_t",
			"params":       map[string]interface{}{"omega": float64(0.000003), "alpha": float64(0.078), "beta": float64(0.905), "dof": float64(6.2)},
			"persistence":  float64(0.983),
			"converged":    true,
			"forecast": []interface{}{
				map[string]interface{}{"horizon_days": float64(1), "vol_annualized": float64(20.1)},
				map[string]interface{}{"horizon_days": float64(21), "vol_annualized": float64(21.0)},
			},
		},
	})
	resp, err := client.VolatilityForecast(context.Background(), "AAPL",
		flashalpha.WithForecastDist("student_t"))
	if err != nil {
		t.Fatalf("VolatilityForecast: %v", err)
	}
	if cap.Path != "/v1/volatility/forecast/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "dist=student_t") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Ewma.Lambda != 0.94 || resp.HarRv.Components.Monthly != 22.7 {
		t.Errorf("resp = %+v", resp)
	}
	if resp.Garch.Params.Dof == nil || *resp.Garch.Params.Dof != 6.2 {
		t.Errorf("Garch.Params.Dof = %+v", resp.Garch.Params.Dof)
	}
	if !resp.Garch.Converged || len(resp.Garch.Forecast) != 2 || resp.Garch.Forecast[1].HorizonDays != 21 {
		t.Errorf("Garch.Forecast = %+v", resp.Garch.Forecast)
	}
}

// ── Flow extras ──────────────────────────────────────────────────────────────

func TestFlowDealerPremium(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":             "SPY",
		"net_dealer_premium": float64(-5750000),
	})
	resp, err := client.FlowDealerPremiumTyped(context.Background(), "SPY",
		flashalpha.WithFlowWindowMinutes(240))
	if err != nil {
		t.Fatalf("FlowDealerPremiumTyped: %v", err)
	}
	if cap.Path != "/v1/flow/options/SPY/dealer-premium" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "windowMinutes=240") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.NetDealerPremium != -5750000 {
		t.Errorf("NetDealerPremium = %v", resp.NetDealerPremium)
	}
}

func TestFlowStockBars(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":     "SPY",
		"resolution": "5m",
		"count":      float64(1),
		"bars":       []interface{}{map[string]interface{}{"ts": "2026-06-05T18:00:00Z", "closed": true}},
	})
	resp, err := client.FlowStockBarsTyped(context.Background(), "SPY", "5m",
		flashalpha.WithBarsMinutes(60))
	if err != nil {
		t.Fatalf("FlowStockBarsTyped: %v", err)
	}
	if cap.Path != "/v1/flow/stocks/SPY/bars" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "resolution=5m") || !strings.Contains(cap.Query, "minutes=60") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Resolution != "5m" || len(resp.Bars) != 1 || !resp.Bars[0].Closed {
		t.Errorf("resp = %+v", resp)
	}
}

func TestFlowZeroDteSnapshot(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":         "SPY",
		"flow_direction": map[string]interface{}{"label": "amplifying"},
	})
	resp, err := client.FlowZeroDteSnapshotTyped(context.Background(), "SPY",
		flashalpha.WithZeroDteFlowExpiry("2026-06-08"))
	if err != nil {
		t.Fatalf("FlowZeroDteSnapshotTyped: %v", err)
	}
	if cap.Path != "/v1/flow/zero-dte/snapshot/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "expiry=2026-06-08") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.FlowDirection == nil || resp.FlowDirection.Label != "amplifying" {
		t.Errorf("FlowDirection = %+v", resp.FlowDirection)
	}
	if resp.Raw["symbol"] != "SPY" {
		t.Errorf("Raw not populated: %+v", resp.Raw)
	}
}

func TestFlowZeroDteSeries(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":   "SPY",
		"bar_size": "1m",
		"bars":     []interface{}{map[string]interface{}{"t": "2026-06-05T18:44:00Z", "pin_score": float64(82)}},
	})
	resp, err := client.FlowZeroDteSeriesTyped(context.Background(), "SPY",
		flashalpha.WithZeroDteFlowBar("1m"), flashalpha.WithZeroDteFlowMinutes(120))
	if err != nil {
		t.Fatalf("FlowZeroDteSeriesTyped: %v", err)
	}
	if cap.Path != "/v1/flow/zero-dte/series/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "bar=1m") || !strings.Contains(cap.Query, "minutes=120") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.BarSize != "1m" || resp.Bars[0].PinScore == nil || *resp.Bars[0].PinScore != 82 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestFlowZeroDteHedgeFlow(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol": "SPY",
		"side":   "calls",
		"bars":   []interface{}{map[string]interface{}{"t": "x", "bar": float64(1240000), "cumulative": float64(18432000)}},
	})
	resp, err := client.FlowZeroDteHedgeFlowTyped(context.Background(), "SPY",
		flashalpha.WithZeroDteFlowSide("calls"), flashalpha.WithZeroDteFlowBar("5m"))
	if err != nil {
		t.Fatalf("FlowZeroDteHedgeFlowTyped: %v", err)
	}
	if cap.Path != "/v1/flow/zero-dte/hedge-flow/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "side=calls") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Side != "calls" || resp.Bars[0].Cumulative != 18432000 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestFlowZeroDteHeatmap(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":       "SPY",
		"metric":       "gex",
		"strikes_grid": []interface{}{float64(585), float64(586)},
		"bars":         []interface{}{map[string]interface{}{"t": "x", "values": []interface{}{float64(-1.2e8), float64(8.1e8)}}},
	})
	resp, err := client.FlowZeroDteHeatmapTyped(context.Background(), "SPY",
		flashalpha.WithZeroDteFlowMetric("gex"), flashalpha.WithZeroDteFlowMode("raw"), flashalpha.WithZeroDteFlowBar("1m"))
	if err != nil {
		t.Fatalf("FlowZeroDteHeatmapTyped: %v", err)
	}
	if cap.Path != "/v1/flow/zero-dte/heatmap/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "metric=gex") || !strings.Contains(cap.Query, "mode=raw") {
		t.Errorf("query = %q", cap.Query)
	}
	if len(resp.StrikesGrid) != 2 || len(resp.Bars[0].Values) != 2 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestFlowZeroDteStrikeFlow(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":       "SPY",
		"strikes_grid": []interface{}{float64(588), float64(589)},
		"bars":         []interface{}{map[string]interface{}{"t": "x", "contracts": []interface{}{float64(1200), float64(1850)}}},
	})
	resp, err := client.FlowZeroDteStrikeFlowTyped(context.Background(), "SPY",
		flashalpha.WithZeroDteFlowBar("1m"), flashalpha.WithZeroDteFlowMinutes(120))
	if err != nil {
		t.Fatalf("FlowZeroDteStrikeFlowTyped: %v", err)
	}
	if cap.Path != "/v1/flow/zero-dte/strike-flow/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if len(resp.Bars[0].Contracts) != 2 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestFlowZeroDteLeaderboard(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"metric":      "heat",
		"n":           float64(2),
		"market_open": true,
		"entries": []interface{}{
			map[string]interface{}{"rank": float64(1), "symbol": "SPY", "value": float64(98.5)},
			map[string]interface{}{"rank": float64(2), "symbol": "QQQ", "value": float64(91.2)},
		},
	})
	resp, err := client.FlowZeroDteLeaderboardTyped(context.Background(),
		flashalpha.WithZeroDteFlowMetric("heat"), flashalpha.WithZeroDteFlowN(2))
	if err != nil {
		t.Fatalf("FlowZeroDteLeaderboardTyped: %v", err)
	}
	if cap.Path != "/v1/flow/zero-dte/leaderboard" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "metric=heat") || !strings.Contains(cap.Query, "n=2") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Metric != "heat" || resp.N != 2 || !resp.MarketOpen {
		t.Errorf("resp = %+v", resp)
	}
	if len(resp.Entries) != 2 || resp.Entries[0].Symbol != "SPY" || resp.Entries[1].Rank != 2 {
		t.Errorf("entries = %+v", resp.Entries)
	}
}

// ── Strategy signals (shared envelope) ───────────────────────────────────────

func TestStrategyFlowAnomaly(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"strategy": "flow_anomaly",
		"symbol":   "SPY",
		"decision": "candidate",
		"score":    float64(72),
		"regime":   "bullish_flow_imbalance",
		"metrics":  map[string]interface{}{"call_put_volume_ratio": float64(1.84), "underlying_price": float64(589.12)},
		"best_structures": []interface{}{
			map[string]interface{}{"rank": float64(1), "structure": "short_put_spread"},
		},
	})
	resp, err := client.StrategyFlowAnomaly(context.Background(), "SPY", flashalpha.WithStrategyExpiry("2026-06-19"))
	if err != nil {
		t.Fatalf("StrategyFlowAnomaly: %v", err)
	}
	if cap.Path != "/v1/strategies/flow-anomaly/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "expiry=2026-06-19") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Decision != "candidate" || resp.Score != 72 || resp.Regime != "bullish_flow_imbalance" {
		t.Errorf("resp = %+v", resp)
	}
	if resp.Metrics["call_put_volume_ratio"] != 1.84 {
		t.Errorf("metrics = %+v", resp.Metrics)
	}
	if len(resp.BestStructures) != 1 || resp.BestStructures[0].Structure != "short_put_spread" {
		t.Errorf("best_structures = %+v", resp.BestStructures)
	}
}

func TestStrategyVolCarryParams(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"strategy": "vol_carry", "decision": "neutral"})
	_, err := client.StrategyVolCarry(context.Background(), "SPY",
		flashalpha.WithStrategyTargetShortDelta(0.20),
		flashalpha.WithStrategyMaxWidth(10),
		flashalpha.WithStrategyMinCredit(0.15))
	if err != nil {
		t.Fatalf("StrategyVolCarry: %v", err)
	}
	if cap.Path != "/v1/strategies/vol-carry/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "targetShortDelta=0.2") || !strings.Contains(cap.Query, "maxWidth=10") {
		t.Errorf("query = %q", cap.Query)
	}
}

func TestStrategyTermStructureNoParams(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"strategy": "term_structure", "decision": "neutral"})
	_, err := client.StrategyTermStructure(context.Background(), "SPY")
	if err != nil {
		t.Fatalf("StrategyTermStructure: %v", err)
	}
	if cap.Path != "/v1/strategies/term-structure/SPY" {
		t.Errorf("path = %q", cap.Path)
	}
}

func TestStrategyYieldEnhancement(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{"strategy": "yield_enhancement", "decision": "candidate"})
	_, err := client.StrategyYieldEnhancement(context.Background(), "SPY",
		flashalpha.WithStrategyStructure("cash_secured_put"),
		flashalpha.WithStrategyTargetDelta(0.20),
		flashalpha.WithStrategyExcludeEarningsBeforeExpiry(true))
	if err != nil {
		t.Fatalf("StrategyYieldEnhancement: %v", err)
	}
	if !strings.Contains(cap.Query, "structure=cash_secured_put") || !strings.Contains(cap.Query, "excludeEarningsBeforeExpiry=true") {
		t.Errorf("query = %q", cap.Query)
	}
}

// ── Earnings family ──────────────────────────────────────────────────────────

func TestEarningsCalendar(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"events": []interface{}{map[string]interface{}{"symbol": "AAPL", "earnings_date": "2026-06-09", "days_to_event": float64(4)}},
		"count":  float64(1),
	})
	resp, err := client.EarningsCalendarTyped(context.Background(),
		flashalpha.WithEarningsCalendarDays(14),
		flashalpha.WithEarningsCalendarSymbols("AAPL,MSFT"),
		flashalpha.WithEarningsCalendarImportance(3))
	if err != nil {
		t.Fatalf("EarningsCalendarTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/calendar" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "days=14") || !strings.Contains(cap.Query, "importance=3") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Count != 1 || resp.Events[0].Symbol != "AAPL" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEarningsExpectedMove(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":        "AAPL",
		"earnings_date": "2026-06-09",
		"expected_move": map[string]interface{}{"earnings_implied_pct": float64(4.6)},
	})
	resp, err := client.EarningsExpectedMoveTyped(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("EarningsExpectedMoveTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/expected-move/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.ExpectedMove == nil || resp.ExpectedMove.EarningsImpliedPct == nil || *resp.ExpectedMove.EarningsImpliedPct != 4.6 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEarningsHistory(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":  "AAPL",
		"count":   float64(1),
		"history": []interface{}{map[string]interface{}{"date": "2026-04-30", "iv_crush_pct": float64(38.5)}},
	})
	resp, err := client.EarningsHistoryTyped(context.Background(), "AAPL", flashalpha.WithEarningsHistoryLimit(12))
	if err != nil {
		t.Fatalf("EarningsHistoryTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/history/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "limit=12") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.History[0].IvCrushPct == nil || *resp.History[0].IvCrushPct != 38.5 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEarningsIvCrush(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":           "AAPL",
		"current_estimate": map[string]interface{}{"expected_crush_pct": float64(41.23)},
		"distribution":     map[string]interface{}{"median": float64(38.5), "count": float64(12)},
	})
	resp, err := client.EarningsIvCrushTyped(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("EarningsIvCrushTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/iv-crush/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.CurrentEstimate == nil || resp.Distribution == nil || resp.Distribution.Count != 12 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEarningsVrp(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol":       "AAPL",
		"earnings_vrp": map[string]interface{}{"assessment": "slightly_rich", "premium_ratio": float64(1.48)},
	})
	resp, err := client.EarningsVrpTyped(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("EarningsVrpTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/vrp/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.EarningsVrp == nil || resp.EarningsVrp.Assessment != "slightly_rich" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEarningsDealerPositioning(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol": "AAPL",
		"regime": "negative_gamma",
		"levels": map[string]interface{}{"gamma_flip": float64(210)},
	})
	resp, err := client.EarningsDealerPositioningTyped(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("EarningsDealerPositioningTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/dealer-positioning/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.Regime != "negative_gamma" || resp.Levels == nil {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEarningsStrategies(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"symbol": "AAPL",
		"scores": map[string]interface{}{"short_strangle": float64(72)},
	})
	resp, err := client.EarningsStrategiesTyped(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("EarningsStrategiesTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/strategies/AAPL" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.Scores == nil || resp.Scores.ShortStrangle == nil || *resp.Scores.ShortStrangle != 72 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEarningsScreener(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"events": []interface{}{map[string]interface{}{"symbol": "AAPL", "premium_ratio": float64(1.48)}},
		"count":  float64(1),
	})
	resp, err := client.EarningsScreenerTyped(context.Background(),
		flashalpha.WithEarningsScreenerSort("vrp_richest"),
		flashalpha.WithEarningsScreenerLimit(20),
		flashalpha.WithEarningsScreenerDays(14),
		flashalpha.WithEarningsScreenerMinImportance(3))
	if err != nil {
		t.Fatalf("EarningsScreenerTyped: %v", err)
	}
	if cap.Path != "/v1/earnings/screener" {
		t.Errorf("path = %q", cap.Path)
	}
	if !strings.Contains(cap.Query, "sort=vrp_richest") || !strings.Contains(cap.Query, "min_importance=3") {
		t.Errorf("query = %q", cap.Query)
	}
	if resp.Count != 1 || resp.Events[0].Symbol != "AAPL" {
		t.Errorf("resp = %+v", resp)
	}
}

// ── Structures (POST) ────────────────────────────────────────────────────────

func TestStructurePnl(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"breakevens": []interface{}{float64(102.1)},
		"max_profit": float64(7.9),
		"max_loss":   float64(-2.1),
		"pnl_curve":  []interface{}{map[string]interface{}{"underlying": float64(100), "pnl": float64(-2.1)}},
	})
	qty := 1
	_ = qty
	resp, err := client.StructurePnlTyped(context.Background(), flashalpha.StructurePnlRequest{
		Legs: []flashalpha.StructurePnlLeg{
			{Action: "buy", Type: "call", Strike: 100, Premium: 3.20, Quantity: 1},
			{Action: "sell", Type: "call", Strike: 110, Premium: 1.10, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("StructurePnlTyped: %v", err)
	}
	if cap.Path != "/v1/structures/pnl" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.MaxProfit == nil || *resp.MaxProfit != 7.9 {
		t.Errorf("MaxProfit = %+v", resp.MaxProfit)
	}
	if len(resp.Breakevens) != 1 || resp.Breakevens[0] != 102.1 {
		t.Errorf("Breakevens = %+v", resp.Breakevens)
	}
}

func TestStructureGreeks(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"spot":            float64(102.5),
		"position_greeks": map[string]interface{}{"delta": float64(0.3142), "vega": float64(0.1163)},
	})
	resp, err := client.StructureGreeksTyped(context.Background(), flashalpha.StructureGreeksRequest{
		Legs: []flashalpha.StructureGreeksLeg{
			{Action: "buy", Type: "call", Strike: 100, Expiry: "2026-07-17", ImpliedVol: 0.28, Quantity: 1},
		},
		Spot: 102.5,
	})
	if err != nil {
		t.Fatalf("StructureGreeksTyped: %v", err)
	}
	if cap.Path != "/v1/structures/greeks" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.PositionGreeks == nil || resp.PositionGreeks.Delta != 0.3142 {
		t.Errorf("PositionGreeks = %+v", resp.PositionGreeks)
	}
}

// ── Screener fields ──────────────────────────────────────────────────────────

func TestScreenerFields(t *testing.T) {
	_, client, cap := newCapturingServer(t, 200, map[string]interface{}{
		"fields": []interface{}{map[string]interface{}{"name": "atm_iv", "type": "number"}},
		"count":  float64(1),
	})
	resp, err := client.ScreenerFieldsTyped(context.Background())
	if err != nil {
		t.Fatalf("ScreenerFieldsTyped: %v", err)
	}
	if cap.Path != "/v1/screener/fields" {
		t.Errorf("path = %q", cap.Path)
	}
	if resp.Count != 1 || resp.Fields[0].Name != "atm_iv" {
		t.Errorf("resp = %+v", resp)
	}
}
