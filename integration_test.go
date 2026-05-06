//go:build integration

package flashalpha_test

import (
	"context"
	"encoding/json"
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

// ── Zero-DTE ─────────────────────────────────────────────────────────────────

// Comprehensive end-to-end test of ZeroDteTyped. Every typed pointer field
// is asserted non-nil so a renamed json struct tag would surface immediately.
// The full untyped-shape coverage lives in TestIntegrationZeroDte_AllNewFields;
// this is the typed mirror.
func TestIntegrationZeroDteTyped_DeserializesAllFields(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.ZeroDteTyped(ctx, "SPX")
	if err != nil {
		t.Fatalf("ZeroDteTyped SPX: %v", err)
	}
	if r.Symbol != "SPX" {
		t.Errorf("Symbol = %q, want SPX", r.Symbol)
	}
	if r.NoZeroDte {
		if r.NextZeroDteExpiry == nil {
			t.Error("NoZeroDte=true but NextZeroDteExpiry nil")
		}
		return
	}

	checkNotNil := func(field string, v interface{}) {
		t.Helper()
		// Use reflect-free check — pointer-typed fields are typed nil
		// but compare equal to nil only when the interface is empty.
		switch p := v.(type) {
		case *float64:
			if p == nil {
				t.Errorf("%s nil", field)
			}
		case *int:
			if p == nil {
				t.Errorf("%s nil", field)
			}
		case *int64:
			if p == nil {
				t.Errorf("%s nil", field)
			}
		case *string:
			if p == nil {
				t.Errorf("%s nil", field)
			}
		case string:
			if p == "" {
				t.Errorf("%s empty", field)
			}
		}
	}

	// top-level
	checkNotNil("UnderlyingPrice", r.UnderlyingPrice)
	checkNotNil("AsOf", r.AsOf)

	// regime
	if r.Regime == nil {
		t.Fatal("Regime nil")
	}
	checkNotNil("Regime.Label", r.Regime.Label)
	checkNotNil("Regime.GammaFlip", r.Regime.GammaFlip)
	checkNotNil("Regime.SpotVsFlip", r.Regime.SpotVsFlip)
	checkNotNil("Regime.SpotToFlipPct", r.Regime.SpotToFlipPct)
	checkNotNil("Regime.DistanceToFlipDollars", r.Regime.DistanceToFlipDollars)
	checkNotNil("Regime.DistanceToFlipSigmas", r.Regime.DistanceToFlipSigmas)

	// exposures
	if r.Exposures == nil {
		t.Fatal("Exposures nil")
	}
	checkNotNil("Exposures.NetGex", r.Exposures.NetGex)
	checkNotNil("Exposures.NetDex", r.Exposures.NetDex)
	checkNotNil("Exposures.NetVex", r.Exposures.NetVex)
	checkNotNil("Exposures.NetChex", r.Exposures.NetChex)
	checkNotNil("Exposures.PctOfTotalGex", r.Exposures.PctOfTotalGex)
	checkNotNil("Exposures.TotalChainNetGex", r.Exposures.TotalChainNetGex)

	// expected_move
	if r.ExpectedMove == nil {
		t.Fatal("ExpectedMove nil")
	}
	checkNotNil("ExpectedMove.Implied1SdDollars", r.ExpectedMove.Implied1SdDollars)
	checkNotNil("ExpectedMove.Implied1SdPct", r.ExpectedMove.Implied1SdPct)
	checkNotNil("ExpectedMove.UpperBound", r.ExpectedMove.UpperBound)
	checkNotNil("ExpectedMove.LowerBound", r.ExpectedMove.LowerBound)
	checkNotNil("ExpectedMove.StraddlePrice", r.ExpectedMove.StraddlePrice)
	checkNotNil("ExpectedMove.AtmIv", r.ExpectedMove.AtmIv)

	// pin_risk
	if r.PinRisk == nil {
		t.Fatal("PinRisk nil")
	}
	checkNotNil("PinRisk.MagnetStrike", r.PinRisk.MagnetStrike)
	checkNotNil("PinRisk.MagnetGex", r.PinRisk.MagnetGex)
	checkNotNil("PinRisk.DistanceToMagnetPct", r.PinRisk.DistanceToMagnetPct)
	checkNotNil("PinRisk.PinScore", r.PinRisk.PinScore)
	if r.PinRisk.Components == nil {
		t.Fatal("PinRisk.Components nil")
	}
	checkNotNil("PinRisk.Components.OiScore", r.PinRisk.Components.OiScore)
	checkNotNil("PinRisk.Components.ProximityScore", r.PinRisk.Components.ProximityScore)
	checkNotNil("PinRisk.Components.TimeScore", r.PinRisk.Components.TimeScore)
	checkNotNil("PinRisk.Components.GammaScore", r.PinRisk.Components.GammaScore)
	checkNotNil("PinRisk.MaxPain", r.PinRisk.MaxPain)
	checkNotNil("PinRisk.OiConcentrationTop3Pct", r.PinRisk.OiConcentrationTop3Pct)

	// hedging — all 8 buckets + convexity_at_spot
	if r.Hedging == nil {
		t.Fatal("Hedging nil")
	}
	type bucketCase struct {
		name string
		b    *flashalpha.ZeroDteHedgingBucket
	}
	for _, bc := range []bucketCase{
		{"SpotUp10Bp", r.Hedging.SpotUp10Bp},
		{"SpotDown10Bp", r.Hedging.SpotDown10Bp},
		{"SpotUp25Bp", r.Hedging.SpotUp25Bp},
		{"SpotDown25Bp", r.Hedging.SpotDown25Bp},
		{"SpotUpHalfPct", r.Hedging.SpotUpHalfPct},
		{"SpotDownHalfPct", r.Hedging.SpotDownHalfPct},
		{"SpotUp1Pct", r.Hedging.SpotUp1Pct},
		{"SpotDown1Pct", r.Hedging.SpotDown1Pct},
	} {
		if bc.b == nil {
			t.Errorf("Hedging.%s nil", bc.name)
			continue
		}
		checkNotNil("Hedging."+bc.name+".DealerSharesToTrade", bc.b.DealerSharesToTrade)
		checkNotNil("Hedging."+bc.name+".Direction", bc.b.Direction)
		checkNotNil("Hedging."+bc.name+".NotionalUsd", bc.b.NotionalUsd)
	}
	checkNotNil("Hedging.ConvexityAtSpot", r.Hedging.ConvexityAtSpot)

	// decay
	if r.Decay == nil {
		t.Fatal("Decay nil")
	}
	checkNotNil("Decay.NetThetaDollars", r.Decay.NetThetaDollars)
	checkNotNil("Decay.CharmRegime", r.Decay.CharmRegime)
	checkNotNil("Decay.CharmDescription", r.Decay.CharmDescription)
	checkNotNil("Decay.GammaAcceleration", r.Decay.GammaAcceleration)

	// vol_context
	if r.VolContext == nil {
		t.Fatal("VolContext nil")
	}
	checkNotNil("VolContext.ZeroDteAtmIv", r.VolContext.ZeroDteAtmIv)
	checkNotNil("VolContext.SevenDteAtmIv", r.VolContext.SevenDteAtmIv)
	checkNotNil("VolContext.IvRatio0Dte7Dte", r.VolContext.IvRatio0Dte7Dte)
	checkNotNil("VolContext.Vix", r.VolContext.Vix)
	checkNotNil("VolContext.VannaExposure", r.VolContext.VannaExposure)
	checkNotNil("VolContext.VannaInterpretation", r.VolContext.VannaInterpretation)

	// flow
	if r.Flow == nil {
		t.Fatal("Flow nil")
	}
	checkNotNil("Flow.TotalVolume", r.Flow.TotalVolume)
	checkNotNil("Flow.CallVolume", r.Flow.CallVolume)
	checkNotNil("Flow.PutVolume", r.Flow.PutVolume)
	checkNotNil("Flow.NetCallMinusPutVolume", r.Flow.NetCallMinusPutVolume)
	checkNotNil("Flow.TotalOi", r.Flow.TotalOi)
	checkNotNil("Flow.CallOi", r.Flow.CallOi)
	checkNotNil("Flow.PutOi", r.Flow.PutOi)
	checkNotNil("Flow.PcRatioVolume", r.Flow.PcRatioVolume)
	checkNotNil("Flow.PcRatioOi", r.Flow.PcRatioOi)
	checkNotNil("Flow.VolumeToOiRatio", r.Flow.VolumeToOiRatio)
	checkNotNil("Flow.AtmVolumeSharePct", r.Flow.AtmVolumeSharePct)
	checkNotNil("Flow.Top3StrikeVolumePct", r.Flow.Top3StrikeVolumePct)

	// levels
	if r.Levels == nil {
		t.Fatal("Levels nil")
	}
	checkNotNil("Levels.CallWall", r.Levels.CallWall)
	checkNotNil("Levels.CallWallGex", r.Levels.CallWallGex)
	checkNotNil("Levels.CallWallStrength", r.Levels.CallWallStrength)
	checkNotNil("Levels.DistanceToCallWallPct", r.Levels.DistanceToCallWallPct)
	checkNotNil("Levels.PutWall", r.Levels.PutWall)
	checkNotNil("Levels.PutWallGex", r.Levels.PutWallGex)
	checkNotNil("Levels.PutWallStrength", r.Levels.PutWallStrength)
	checkNotNil("Levels.DistanceToPutWallPct", r.Levels.DistanceToPutWallPct)
	checkNotNil("Levels.DistanceToMagnetDollars", r.Levels.DistanceToMagnetDollars)
	checkNotNil("Levels.HighestOiStrike", r.Levels.HighestOiStrike)
	checkNotNil("Levels.HighestOiTotal", r.Levels.HighestOiTotal)
	checkNotNil("Levels.MaxPositiveGamma", r.Levels.MaxPositiveGamma)
	checkNotNil("Levels.MaxNegativeGamma", r.Levels.MaxNegativeGamma)
	checkNotNil("Levels.LevelClusterScore", r.Levels.LevelClusterScore)

	// liquidity
	if r.Liquidity == nil {
		t.Fatal("Liquidity nil")
	}
	checkNotNil("Liquidity.AtmSpreadPct", r.Liquidity.AtmSpreadPct)
	checkNotNil("Liquidity.WeightedSpreadPct", r.Liquidity.WeightedSpreadPct)
	checkNotNil("Liquidity.ExecutionScore", r.Liquidity.ExecutionScore)

	// metadata
	if r.Metadata == nil {
		t.Fatal("Metadata nil")
	}
	checkNotNil("Metadata.SnapshotAgeSeconds", r.Metadata.SnapshotAgeSeconds)
	checkNotNil("Metadata.ChainContractCount", r.Metadata.ChainContractCount)
	checkNotNil("Metadata.DataQualityScore", r.Metadata.DataQualityScore)
	checkNotNil("Metadata.GreekSmoothnessScore", r.Metadata.GreekSmoothnessScore)

	// strikes[0] — every per-strike field
	if len(r.Strikes) > 0 {
		s := r.Strikes[0]
		if s.Strike <= 0 {
			t.Error("strikes[0].Strike unpopulated")
		}
		checkNotNil("strikes[0].DistanceFromSpotPct", s.DistanceFromSpotPct)
		checkNotNil("strikes[0].CallSymbol", s.CallSymbol)
		checkNotNil("strikes[0].PutSymbol", s.PutSymbol)
		checkNotNil("strikes[0].CallGex", s.CallGex)
		checkNotNil("strikes[0].PutGex", s.PutGex)
		checkNotNil("strikes[0].NetGex", s.NetGex)
		checkNotNil("strikes[0].CallDex", s.CallDex)
		checkNotNil("strikes[0].PutDex", s.PutDex)
		checkNotNil("strikes[0].NetDex", s.NetDex)
		checkNotNil("strikes[0].NetVex", s.NetVex)
		checkNotNil("strikes[0].NetChex", s.NetChex)
		checkNotNil("strikes[0].CallOi", s.CallOi)
		checkNotNil("strikes[0].PutOi", s.PutOi)
		checkNotNil("strikes[0].CallVolume", s.CallVolume)
		checkNotNil("strikes[0].PutVolume", s.PutVolume)
		checkNotNil("strikes[0].GexSharePct", s.GexSharePct)
		checkNotNil("strikes[0].OiSharePct", s.OiSharePct)
		checkNotNil("strikes[0].VolumeSharePct", s.VolumeSharePct)
		checkNotNil("strikes[0].CallIv", s.CallIv)
		checkNotNil("strikes[0].PutIv", s.PutIv)
		checkNotNil("strikes[0].CallDelta", s.CallDelta)
		checkNotNil("strikes[0].PutDelta", s.PutDelta)
		checkNotNil("strikes[0].CallGamma", s.CallGamma)
		checkNotNil("strikes[0].PutGamma", s.PutGamma)
		checkNotNil("strikes[0].CallTheta", s.CallTheta)
		checkNotNil("strikes[0].PutTheta", s.PutTheta)
		checkNotNil("strikes[0].CallMid", s.CallMid)
		checkNotNil("strikes[0].PutMid", s.PutMid)
		checkNotNil("strikes[0].CallSpreadPct", s.CallSpreadPct)
		checkNotNil("strikes[0].PutSpreadPct", s.PutSpreadPct)
	}
	// Raw fallback should always be populated for forward-compatibility.
	if r.Raw == nil {
		t.Error("Raw map nil — should hold full decoded payload")
	}
}

// Validate the full 0DTE response shape — fine-grained hedging buckets,
// distance-to-flip in dollars/sigmas, pin sub-scores, flow concentration,
// wall strength + level cluster, the new liquidity & metadata sections,
// and per-strike greeks/quotes. Uses SPX which has daily 0DTE.
func TestIntegrationZeroDte_AllNewFields(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.ZeroDte(ctx, "SPX")
	if err != nil {
		t.Fatalf("ZeroDte SPX: %v", err)
	}
	if sym, _ := r["symbol"].(string); sym != "SPX" {
		t.Errorf("symbol = %q, want SPX", sym)
	}
	if v, ok := r["no_zero_dte"].(bool); ok && v {
		if _, ok := r["next_zero_dte_expiry"]; !ok {
			t.Error("no_zero_dte=true but next_zero_dte_expiry missing")
		}
		return
	}

	mustObj := func(parent map[string]interface{}, key string) map[string]interface{} {
		t.Helper()
		v, ok := parent[key].(map[string]interface{})
		if !ok {
			t.Fatalf("%s missing or not an object", key)
		}
		return v
	}
	checkKeys := func(obj map[string]interface{}, prefix string, keys []string) {
		t.Helper()
		for _, k := range keys {
			if _, ok := obj[k]; !ok {
				t.Errorf("%s.%s missing", prefix, k)
			}
		}
	}

	checkKeys(r, "(top-level)",
		[]string{"underlying_price", "expiration", "as_of", "market_open",
			"time_to_close_hours", "time_to_close_pct"})

	checkKeys(mustObj(r, "regime"), "regime",
		[]string{"label", "description", "gamma_flip", "spot_vs_flip", "spot_to_flip_pct",
			"distance_to_flip_dollars", "distance_to_flip_sigmas"})

	checkKeys(mustObj(r, "exposures"), "exposures",
		[]string{"net_gex", "net_dex", "net_vex", "net_chex",
			"pct_of_total_gex", "total_chain_net_gex"})

	checkKeys(mustObj(r, "expected_move"), "expected_move",
		[]string{"implied_1sd_dollars", "implied_1sd_pct", "remaining_1sd_dollars",
			"remaining_1sd_pct", "upper_bound", "lower_bound",
			"straddle_price", "atm_iv"})

	pr := mustObj(r, "pin_risk")
	checkKeys(pr, "pin_risk",
		[]string{"magnet_strike", "magnet_gex", "distance_to_magnet_pct",
			"pin_score", "components", "max_pain",
			"oi_concentration_top3_pct", "description"})
	checkKeys(mustObj(pr, "components"), "pin_risk.components",
		[]string{"oi_score", "proximity_score", "time_score", "gamma_score"})

	hedging := mustObj(r, "hedging")
	for _, bucket := range []string{
		"spot_up_10bp", "spot_down_10bp",
		"spot_up_25bp", "spot_down_25bp",
		"spot_up_half_pct", "spot_down_half_pct",
		"spot_up_1pct", "spot_down_1pct",
	} {
		b, ok := hedging[bucket].(map[string]interface{})
		if !ok {
			t.Errorf("hedging.%s missing or not an object", bucket)
			continue
		}
		checkKeys(b, "hedging."+bucket,
			[]string{"dealer_shares_to_trade", "direction", "notional_usd"})
	}
	if _, ok := hedging["convexity_at_spot"]; !ok {
		t.Error("hedging.convexity_at_spot missing")
	}

	checkKeys(mustObj(r, "decay"), "decay",
		[]string{"net_theta_dollars", "theta_per_hour_remaining", "charm_regime",
			"charm_description", "gamma_acceleration", "description"})

	checkKeys(mustObj(r, "vol_context"), "vol_context",
		[]string{"zero_dte_atm_iv", "seven_dte_atm_iv", "iv_ratio_0dte_7dte",
			"vix", "vanna_exposure", "vanna_interpretation", "description"})

	checkKeys(mustObj(r, "flow"), "flow",
		[]string{"total_volume", "call_volume", "put_volume",
			"net_call_minus_put_volume",
			"total_oi", "call_oi", "put_oi",
			"pc_ratio_volume", "pc_ratio_oi", "volume_to_oi_ratio",
			"atm_volume_share_pct", "top3_strike_volume_pct"})

	checkKeys(mustObj(r, "levels"), "levels",
		[]string{"call_wall", "call_wall_gex", "call_wall_strength",
			"distance_to_call_wall_pct",
			"put_wall", "put_wall_gex", "put_wall_strength",
			"distance_to_put_wall_pct",
			"distance_to_magnet_dollars",
			"highest_oi_strike", "highest_oi_total",
			"max_positive_gamma", "max_negative_gamma",
			"level_cluster_score"})

	checkKeys(mustObj(r, "liquidity"), "liquidity",
		[]string{"atm_spread_pct", "weighted_spread_pct", "execution_score"})

	checkKeys(mustObj(r, "metadata"), "metadata",
		[]string{"snapshot_age_seconds", "chain_contract_count",
			"data_quality_score", "greek_smoothness_score"})

	strikes, ok := r["strikes"].([]interface{})
	if !ok {
		t.Fatal("strikes missing or not an array")
	}
	if len(strikes) > 0 {
		s, ok := strikes[0].(map[string]interface{})
		if !ok {
			t.Fatal("strikes[0] not an object")
		}
		checkKeys(s, "strikes[0]",
			[]string{"strike", "distance_from_spot_pct",
				"call_symbol", "put_symbol",
				"call_gex", "put_gex", "net_gex",
				"call_dex", "put_dex", "net_dex",
				"net_vex", "net_chex",
				"call_oi", "put_oi", "call_volume", "put_volume",
				"gex_share_pct", "oi_share_pct", "volume_share_pct",
				"call_iv", "put_iv",
				"call_delta", "put_delta",
				"call_gamma", "put_gamma",
				"call_theta", "put_theta",
				"call_mid", "put_mid",
				"call_spread_pct", "put_spread_pct"})
	}
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

// TestMaxPain_EveryFieldDeclaredInPocoMustBeReferenced asserts that every
// leaf field declared in MaxPainResponse is referenced. JSON-roundtrips the
// raw map response into the typed POCO so the wire format and type
// definition stay in lockstep.
func TestMaxPain_EveryFieldDeclaredInPocoMustBeReferenced(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	raw, err := client.MaxPain(ctx, "SPY")
	if err != nil {
		t.Fatalf("MaxPain SPY: %v", err)
	}
	buf, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	r := &flashalpha.MaxPainResponse{}
	if err := json.Unmarshal(buf, r); err != nil {
		t.Fatalf("decode into MaxPainResponse: %v", err)
	}

	// ── top-level scalars ──
	if r.Symbol != "SPY" {
		t.Errorf("Symbol=%q", r.Symbol)
	}
	if r.UnderlyingPrice == nil || *r.UnderlyingPrice <= 0 {
		t.Errorf("UnderlyingPrice=%v", r.UnderlyingPrice)
	}
	if r.AsOf == "" {
		t.Error("AsOf empty")
	}
	if r.MaxPainStrike == nil {
		t.Error("MaxPainStrike nil")
	}
	if r.Signal == nil ||
		(*r.Signal != "bullish" && *r.Signal != "bearish" && *r.Signal != "neutral") {
		t.Errorf("Signal=%v", r.Signal)
	}
	if r.Expiration == nil || *r.Expiration == "" {
		t.Error("Expiration empty")
	}
	if r.PutCallOiRatio == nil {
		t.Error("PutCallOiRatio nil")
	}
	if r.Regime == nil {
		t.Error("Regime nil")
	} else {
		switch *r.Regime {
		case "positive_gamma", "negative_gamma", "neutral", "undetermined":
		default:
			t.Errorf("Regime=%q", *r.Regime)
		}
	}
	if r.PinProbability == nil || *r.PinProbability < 0 || *r.PinProbability > 100 {
		t.Errorf("PinProbability=%v", r.PinProbability)
	}

	// ── distance ──
	if r.Distance == nil {
		t.Fatal("Distance nil")
	}
	if r.Distance.Absolute == nil {
		t.Error("Distance.Absolute nil")
	}
	if r.Distance.Percent == nil {
		t.Error("Distance.Percent nil")
	}
	if r.Distance.Direction == nil ||
		(*r.Distance.Direction != "above" && *r.Distance.Direction != "below" && *r.Distance.Direction != "at") {
		t.Errorf("Distance.Direction=%v", r.Distance.Direction)
	}

	// ── pain_curve[] ──
	if len(r.PainCurve) == 0 {
		t.Fatal("PainCurve empty")
	}
	pc := r.PainCurve[0]
	if pc.Strike == nil {
		t.Error("PainCurve[0].Strike nil")
	}
	if pc.CallPain == nil {
		t.Error("PainCurve[0].CallPain nil")
	}
	if pc.PutPain == nil {
		t.Error("PainCurve[0].PutPain nil")
	}
	if pc.TotalPain == nil {
		t.Error("PainCurve[0].TotalPain nil")
	}

	// ── oi_by_strike[] ──
	if len(r.OiByStrike) == 0 {
		t.Fatal("OiByStrike empty")
	}
	oi := r.OiByStrike[0]
	if oi.Strike == nil {
		t.Error("OiByStrike[0].Strike nil")
	}
	if oi.CallOi == nil {
		t.Error("OiByStrike[0].CallOi nil")
	}
	if oi.PutOi == nil {
		t.Error("OiByStrike[0].PutOi nil")
	}
	if oi.TotalOi == nil {
		t.Error("OiByStrike[0].TotalOi nil")
	}
	if oi.CallVolume == nil {
		t.Error("OiByStrike[0].CallVolume nil")
	}
	if oi.PutVolume == nil {
		t.Error("OiByStrike[0].PutVolume nil")
	}

	// ── max_pain_by_expiration[] (no filter on this call) ──
	if len(r.MaxPainByExpiration) == 0 {
		t.Fatal("MaxPainByExpiration empty")
	}
	mr := r.MaxPainByExpiration[0]
	if mr.Expiration == nil || *mr.Expiration == "" {
		t.Error("MaxPainByExpiration[0].Expiration empty")
	}
	if mr.MaxPainStrike == nil {
		t.Error("MaxPainByExpiration[0].MaxPainStrike nil")
	}
	if mr.Dte == nil {
		t.Error("MaxPainByExpiration[0].Dte nil")
	}
	if mr.TotalOi == nil {
		t.Error("MaxPainByExpiration[0].TotalOi nil")
	}

	// ── dealer_alignment ──
	if r.DealerAlignment == nil {
		t.Fatal("DealerAlignment nil")
	}
	da := r.DealerAlignment
	if da.Alignment == nil {
		t.Error("DealerAlignment.Alignment nil")
	} else {
		switch *da.Alignment {
		case "converging", "moderate", "diverging", "unknown":
		default:
			t.Errorf("DealerAlignment.Alignment=%q", *da.Alignment)
		}
	}
	if da.Description == nil || *da.Description == "" {
		t.Error("DealerAlignment.Description empty")
	}
	if da.GammaFlip == nil {
		t.Error("DealerAlignment.GammaFlip nil")
	}
	if da.CallWall == nil {
		t.Error("DealerAlignment.CallWall nil")
	}
	if da.PutWall == nil {
		t.Error("DealerAlignment.PutWall nil")
	}

	// ── expected_move ──
	if r.ExpectedMove == nil {
		t.Fatal("ExpectedMove nil")
	}
	em := r.ExpectedMove
	if em.StraddlePrice == nil {
		t.Error("ExpectedMove.StraddlePrice nil")
	}
	if em.AtmIv == nil {
		t.Error("ExpectedMove.AtmIv nil")
	}
	if em.MaxPainWithinExpectedRange == nil {
		t.Error("ExpectedMove.MaxPainWithinExpectedRange nil")
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

// TestVrp_EveryFieldDeclaredInPocoMustBeReferenced asserts that every leaf
// field declared in VrpResponse is referenced by at least one assertion.
// 100% field-coverage discipline.
func TestVrp_EveryFieldDeclaredInPocoMustBeReferenced(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.Vrp(ctx, "SPY")
	if err != nil {
		t.Fatalf("Vrp SPY: %v", err)
	}

	// ── top-level scalars ──
	if r.Symbol != "SPY" {
		t.Errorf("Symbol=%q", r.Symbol)
	}
	if r.UnderlyingPrice <= 0 {
		t.Errorf("UnderlyingPrice=%v", r.UnderlyingPrice)
	}
	if r.AsOf == "" {
		t.Error("AsOf empty")
	}
	if r.VarianceRiskPrem == nil {
		t.Error("VarianceRiskPrem nil")
	}
	if r.ConvexityPremium == nil {
		t.Error("ConvexityPremium nil")
	}
	if r.FairVol == nil {
		t.Error("FairVol nil")
	}
	if r.NetHarvestScore == nil {
		t.Error("NetHarvestScore nil")
	}
	if r.DealerFlowRisk == nil {
		t.Error("DealerFlowRisk nil")
	}
	if r.Warnings == nil {
		t.Error("Warnings nil (should be empty slice not nil)")
	}

	// ── Vrp.* core ──
	for label, ptr := range map[string]*float64{
		"AtmIv": r.Vrp.AtmIv, "Rv5d": r.Vrp.Rv5d, "Rv10d": r.Vrp.Rv10d,
		"Rv20d": r.Vrp.Rv20d, "Rv30d": r.Vrp.Rv30d,
		"Vrp5d": r.Vrp.Vrp5d, "Vrp10d": r.Vrp.Vrp10d,
		"Vrp20d": r.Vrp.Vrp20d, "Vrp30d": r.Vrp.Vrp30d,
		"ZScore": r.Vrp.ZScore,
	} {
		if ptr == nil {
			t.Errorf("Vrp.%s nil", label)
		}
	}
	if r.Vrp.Percentile == nil {
		t.Error("Vrp.Percentile nil")
	}

	// ── Directional ──
	for label, ptr := range map[string]*float64{
		"PutWingIv25d": r.Directional.PutWingIv25d,
		"CallWingIv25d": r.Directional.CallWingIv25d,
		"DownsideRv20d": r.Directional.DownsideRv20d,
		"UpsideRv20d": r.Directional.UpsideRv20d,
		"DownsideVrp": r.Directional.DownsideVrp,
		"UpsideVrp": r.Directional.UpsideVrp,
	} {
		if ptr == nil {
			t.Errorf("Directional.%s nil", label)
		}
	}

	// ── TermVrp[] ──
	if len(r.TermVrp) == 0 {
		t.Error("TermVrp empty")
	}

	// ── GexConditioned + VannaConditioned ──
	if r.GexConditioned == nil {
		t.Error("GexConditioned nil")
	} else {
		if r.GexConditioned.Regime == "" {
			t.Error("GexConditioned.Regime empty")
		}
		if r.GexConditioned.Interpretation == "" {
			t.Error("GexConditioned.Interpretation empty")
		}
	}
	if r.VannaConditioned == nil {
		t.Error("VannaConditioned nil")
	} else {
		if r.VannaConditioned.Outlook == "" {
			t.Error("VannaConditioned.Outlook empty")
		}
		if r.VannaConditioned.Interpretation == "" {
			t.Error("VannaConditioned.Interpretation empty")
		}
	}

	// ── Regime — NetGex lives HERE ──
	if r.Regime.Gamma == "" {
		t.Error("Regime.Gamma empty")
	}
	if r.Regime.NetGex == 0 {
		t.Error("Regime.NetGex unset (zero — likely missing)")
	}
	if r.Regime.GammaFlip == nil {
		t.Error("Regime.GammaFlip nil")
	}

	// ── StrategyScores ──
	if r.StrategyScores == nil {
		t.Error("StrategyScores nil on live (expected non-nil)")
	} else {
		for label, ptr := range map[string]*int{
			"ShortPutSpread": r.StrategyScores.ShortPutSpread,
			"ShortStrangle": r.StrategyScores.ShortStrangle,
			"IronCondor": r.StrategyScores.IronCondor,
			"CalendarSpread": r.StrategyScores.CalendarSpread,
		} {
			if ptr == nil {
				t.Errorf("StrategyScores.%s nil", label)
			}
		}
	}

	// ── Macro (live includes FedFunds) ──
	if r.Macro == nil {
		t.Fatal("Macro nil on live")
	}
	for label, ptr := range map[string]*float64{
		"Vix": r.Macro.Vix, "Vix3m": r.Macro.Vix3m,
		"VixTermSlope": r.Macro.VixTermSlope, "Dgs10": r.Macro.Dgs10,
		"FedFunds": r.Macro.FedFunds,
	} {
		if ptr == nil {
			t.Errorf("Macro.%s nil", label)
		}
	}
	// HySpread may currently be null on live (data gap) — just check the
	// field exists in the shape (it's declared in VrpMacro struct).

	// Customer-trap raw-map check: top-level keys must NOT include
	// z_score / percentile / atm_iv / net_gex / put_vrp / call_vrp /
	// harvest_score (the canonical paths are nested).
	for _, trap := range []string{"z_score", "percentile", "atm_iv",
		"net_gex", "put_vrp", "call_vrp", "harvest_score"} {
		if _, ok := r.Raw[trap]; ok {
			t.Errorf("top-level %q must not exist (use nested path)", trap)
		}
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

// TestExposureSummary: every field declared in ExposureSummaryResponse must
// be referenced. Inherits the original bug-#1 check: net_gex must NOT exist
// at the top level — it lives under `exposures`.
func TestExposureSummary_NetGex_UnderExposures(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := integrationCtx()
	defer cancel()

	r, err := client.ExposureSummary(ctx, "SPY")
	if err != nil {
		t.Fatalf("ExposureSummary SPY: %v", err)
	}
	// Original bug #1
	if _, ok := r["net_gex"]; ok {
		t.Error("net_gex present at top level — must live under exposures")
	}
	// ── top-level scalars ──
	if sym, _ := r["symbol"].(string); sym != "SPY" {
		t.Errorf("symbol = %q, want SPY", sym)
	}
	if _, ok := r["underlying_price"].(float64); !ok {
		t.Errorf("underlying_price missing/non-number: %v", r["underlying_price"])
	}
	if asOf, _ := r["as_of"].(string); asOf == "" {
		t.Errorf("as_of missing/empty: %v", r["as_of"])
	}
	if _, ok := r["gamma_flip"].(float64); !ok {
		t.Errorf("gamma_flip missing/non-number: %v", r["gamma_flip"])
	}
	regime, ok := r["regime"].(string)
	if !ok {
		t.Fatal("regime missing or not a string at top level")
	}
	switch regime {
	case "positive_gamma", "negative_gamma", "neutral", "undetermined":
	default:
		t.Errorf("regime = %q unexpected", regime)
	}
	// ── exposures block (4 fields) ──
	exp, ok := r["exposures"].(map[string]interface{})
	if !ok {
		t.Fatal("exposures missing or wrong type")
	}
	for _, k := range []string{"net_gex", "net_dex", "net_vex", "net_chex"} {
		v, present := exp[k]
		if !present {
			t.Errorf("exposures.%s missing", k)
			continue
		}
		if _, isNum := v.(float64); !isNum {
			t.Errorf("exposures.%s not numeric: %T", k, v)
		}
	}
	// ── interpretation block (3 fields) ──
	interp, _ := r["interpretation"].(map[string]interface{})
	for _, k := range []string{"gamma", "vanna", "charm"} {
		v, ok := interp[k].(string)
		if !ok || v == "" {
			t.Errorf("interpretation.%s missing/empty", k)
		}
	}
	// ── hedging_estimate (every leaf on both sides) ──
	hedging, _ := r["hedging_estimate"].(map[string]interface{})
	for _, sideKey := range []string{"spot_up_1pct", "spot_down_1pct"} {
		side, _ := hedging[sideKey].(map[string]interface{})
		dir, _ := side["direction"].(string)
		if dir != "buy" && dir != "sell" {
			t.Errorf("%s.direction=%q, want buy/sell", sideKey, dir)
		}
		if _, ok := side["dealer_shares_to_trade"].(float64); !ok {
			t.Errorf("%s.dealer_shares_to_trade missing/non-number", sideKey)
		}
		notional, ok := side["notional_usd"].(float64)
		if !ok {
			t.Errorf("%s.notional_usd missing/non-number", sideKey)
		}
		if notional == 0 {
			t.Errorf("%s.notional_usd is zero", sideKey)
		}
	}
	up, _ := hedging["spot_up_1pct"].(map[string]interface{})
	dn, _ := hedging["spot_down_1pct"].(map[string]interface{})
	upShares, _ := up["dealer_shares_to_trade"].(float64)
	dnShares, _ := dn["dealer_shares_to_trade"].(float64)
	if upShares != -dnShares {
		t.Errorf("hedging not symmetric: up=%v down=%v", upShares, dnShares)
	}
	// ── zero_dte block (3 fields) ──
	z, ok := r["zero_dte"].(map[string]interface{})
	if !ok {
		t.Fatal("zero_dte block missing or wrong type")
	}
	for _, k := range []string{"net_gex", "pct_of_total_gex"} {
		if _, present := z[k]; !present {
			t.Errorf("zero_dte.%s key missing", k)
		} else if v := z[k]; v != nil {
			if _, ok := v.(float64); !ok {
				t.Errorf("zero_dte.%s non-number: %T", k, v)
			}
		}
	}
	if _, present := z["expiration"]; !present {
		t.Error("zero_dte.expiration key missing")
	} else if v := z["expiration"]; v != nil {
		if _, ok := v.(string); !ok {
			t.Errorf("zero_dte.expiration non-string: %T", v)
		}
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
