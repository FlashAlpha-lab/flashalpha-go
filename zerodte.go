package flashalpha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ZeroDteRegime is the gamma-regime block of a ZeroDteResponse.
type ZeroDteRegime struct {
	Label                 string   `json:"label"`
	Description           string   `json:"description"`
	GammaFlip             *float64 `json:"gamma_flip"`
	SpotVsFlip            string   `json:"spot_vs_flip"`
	SpotToFlipPct         *float64 `json:"spot_to_flip_pct"`
	DistanceToFlipDollars *float64 `json:"distance_to_flip_dollars"`
	DistanceToFlipSigmas  *float64 `json:"distance_to_flip_sigmas"`
}

// ZeroDteExposures aggregates net 0DTE GEX/DEX/VEX/CHEX and full-chain context.
type ZeroDteExposures struct {
	NetGex           *float64 `json:"net_gex"`
	NetDex           *float64 `json:"net_dex"`
	NetVex           *float64 `json:"net_vex"`
	NetChex          *float64 `json:"net_chex"`
	PctOfTotalGex    *float64 `json:"pct_of_total_gex"`
	TotalChainNetGex *float64 `json:"total_chain_net_gex"`
}

// ZeroDteExpectedMove holds implied 1σ move + remaining-session 1σ + ATM straddle.
type ZeroDteExpectedMove struct {
	Implied1SdDollars   *float64 `json:"implied_1sd_dollars"`
	Implied1SdPct       *float64 `json:"implied_1sd_pct"`
	Remaining1SdDollars *float64 `json:"remaining_1sd_dollars"`
	Remaining1SdPct    *float64  `json:"remaining_1sd_pct"`
	UpperBound          *float64 `json:"upper_bound"`
	LowerBound          *float64 `json:"lower_bound"`
	StraddlePrice       *float64 `json:"straddle_price"`
	AtmIv               *float64 `json:"atm_iv"`
}

// ZeroDtePinComponents is the sub-score breakdown for ZeroDtePinRisk.PinScore.
type ZeroDtePinComponents struct {
	OiScore        *int `json:"oi_score"`
	ProximityScore *int `json:"proximity_score"`
	TimeScore      *int `json:"time_score"`
	GammaScore     *int `json:"gamma_score"`
}

// ZeroDtePinRisk is the pin-risk block — magnet strike + composite + sub-scores.
type ZeroDtePinRisk struct {
	MagnetStrike            *float64              `json:"magnet_strike"`
	MagnetGex               *float64              `json:"magnet_gex"`
	DistanceToMagnetPct     *float64              `json:"distance_to_magnet_pct"`
	PinScore                *int                  `json:"pin_score"`
	Components              *ZeroDtePinComponents `json:"components"`
	MaxPain                 *float64              `json:"max_pain"`
	OiConcentrationTop3Pct  *float64              `json:"oi_concentration_top3_pct"`
	Description             string                `json:"description"`
}

// ZeroDteHedgingBucket is one row of dealer hedging flow at a specific spot delta.
type ZeroDteHedgingBucket struct {
	DealerSharesToTrade *float64 `json:"dealer_shares_to_trade"`
	Direction           string   `json:"direction"`
	NotionalUsd         *float64 `json:"notional_usd"`
}

// ZeroDteHedging holds dealer hedging flow at ±10bp/25bp/50bp/100bp + GEX convexity.
type ZeroDteHedging struct {
	SpotUp10Bp       *ZeroDteHedgingBucket `json:"spot_up_10bp"`
	SpotDown10Bp     *ZeroDteHedgingBucket `json:"spot_down_10bp"`
	SpotUp25Bp       *ZeroDteHedgingBucket `json:"spot_up_25bp"`
	SpotDown25Bp     *ZeroDteHedgingBucket `json:"spot_down_25bp"`
	SpotUpHalfPct    *ZeroDteHedgingBucket `json:"spot_up_half_pct"`
	SpotDownHalfPct  *ZeroDteHedgingBucket `json:"spot_down_half_pct"`
	SpotUp1Pct       *ZeroDteHedgingBucket `json:"spot_up_1pct"`
	SpotDown1Pct     *ZeroDteHedgingBucket `json:"spot_down_1pct"`
	ConvexityAtSpot  *float64              `json:"convexity_at_spot"`
}

// ZeroDteDecay is the time-decay block — net theta + per-hour rate + acceleration.
type ZeroDteDecay struct {
	NetThetaDollars       *float64 `json:"net_theta_dollars"`
	ThetaPerHourRemaining *float64 `json:"theta_per_hour_remaining"`
	CharmRegime           string   `json:"charm_regime"`
	CharmDescription      string   `json:"charm_description"`
	GammaAcceleration     *float64 `json:"gamma_acceleration"`
	Description           string   `json:"description"`
}

// ZeroDteVolContext gives the vol-surface context — 0DTE vs 7DTE IV + vanna read.
type ZeroDteVolContext struct {
	ZeroDteAtmIv        *float64 `json:"zero_dte_atm_iv"`
	SevenDteAtmIv       *float64 `json:"seven_dte_atm_iv"`
	IvRatio0Dte7Dte     *float64 `json:"iv_ratio_0dte_7dte"`
	Vix                 *float64 `json:"vix"`
	VannaExposure       *float64 `json:"vanna_exposure"`
	VannaInterpretation string   `json:"vanna_interpretation"`
	Description         string   `json:"description"`
}

// ZeroDteFlow is the flow block — volume/OI aggregates + concentration metrics.
type ZeroDteFlow struct {
	TotalVolume            *int64   `json:"total_volume"`
	CallVolume             *int64   `json:"call_volume"`
	PutVolume              *int64   `json:"put_volume"`
	NetCallMinusPutVolume  *int64   `json:"net_call_minus_put_volume"`
	TotalOi                *int64   `json:"total_oi"`
	CallOi                 *int64   `json:"call_oi"`
	PutOi                  *int64   `json:"put_oi"`
	PcRatioVolume          *float64 `json:"pc_ratio_volume"`
	PcRatioOi              *float64 `json:"pc_ratio_oi"`
	VolumeToOiRatio        *float64 `json:"volume_to_oi_ratio"`
	AtmVolumeSharePct      *float64 `json:"atm_volume_share_pct"`
	Top3StrikeVolumePct    *float64 `json:"top3_strike_volume_pct"`
}

// ZeroDteLevels holds key strikes — call/put walls (with strength), gamma extrema,
// highest-OI strike, distance-to-magnet, and the level-cluster composite.
type ZeroDteLevels struct {
	CallWall                 *float64 `json:"call_wall"`
	CallWallGex              *float64 `json:"call_wall_gex"`
	CallWallStrength         *float64 `json:"call_wall_strength"`
	DistanceToCallWallPct    *float64 `json:"distance_to_call_wall_pct"`
	PutWall                  *float64 `json:"put_wall"`
	PutWallGex               *float64 `json:"put_wall_gex"`
	PutWallStrength          *float64 `json:"put_wall_strength"`
	DistanceToPutWallPct     *float64 `json:"distance_to_put_wall_pct"`
	DistanceToMagnetDollars  *float64 `json:"distance_to_magnet_dollars"`
	HighestOiStrike          *float64 `json:"highest_oi_strike"`
	HighestOiTotal           *int64   `json:"highest_oi_total"`
	MaxPositiveGamma         *float64 `json:"max_positive_gamma"`
	MaxNegativeGamma         *float64 `json:"max_negative_gamma"`
	LevelClusterScore        *int     `json:"level_cluster_score"`
}

// ZeroDteLiquidity is the bid-ask liquidity context for the 0DTE chain.
type ZeroDteLiquidity struct {
	AtmSpreadPct      *float64 `json:"atm_spread_pct"`
	WeightedSpreadPct *float64 `json:"weighted_spread_pct"`
	ExecutionScore    *int     `json:"execution_score"`
}

// ZeroDteMetadata is the staleness + quality metadata for the snapshot.
type ZeroDteMetadata struct {
	SnapshotAgeSeconds   *float64 `json:"snapshot_age_seconds"`
	ChainContractCount   *int     `json:"chain_contract_count"`
	DataQualityScore     *int     `json:"data_quality_score"`
	GreekSmoothnessScore *int     `json:"greek_smoothness_score"`
}

// ZeroDteStrike is one row in ZeroDteResponse.Strikes — per-strike exposure,
// flow, greeks, and quote/spread metrics.
type ZeroDteStrike struct {
	Strike              float64  `json:"strike"`
	DistanceFromSpotPct *float64 `json:"distance_from_spot_pct"`
	CallSymbol          string   `json:"call_symbol"`
	PutSymbol           string   `json:"put_symbol"`
	CallGex             *float64 `json:"call_gex"`
	PutGex              *float64 `json:"put_gex"`
	NetGex              *float64 `json:"net_gex"`
	CallDex             *float64 `json:"call_dex"`
	PutDex              *float64 `json:"put_dex"`
	NetDex              *float64 `json:"net_dex"`
	NetVex              *float64 `json:"net_vex"`
	NetChex             *float64 `json:"net_chex"`
	CallOi              *int64   `json:"call_oi"`
	PutOi               *int64   `json:"put_oi"`
	CallVolume          *int64   `json:"call_volume"`
	PutVolume           *int64   `json:"put_volume"`
	GexSharePct         *float64 `json:"gex_share_pct"`
	OiSharePct          *float64 `json:"oi_share_pct"`
	VolumeSharePct      *float64 `json:"volume_share_pct"`
	CallIv              *float64 `json:"call_iv"`
	PutIv               *float64 `json:"put_iv"`
	CallDelta           *float64 `json:"call_delta"`
	PutDelta            *float64 `json:"put_delta"`
	CallGamma           *float64 `json:"call_gamma"`
	PutGamma            *float64 `json:"put_gamma"`
	CallTheta           *float64 `json:"call_theta"`
	PutTheta            *float64 `json:"put_theta"`
	CallMid             *float64 `json:"call_mid"`
	PutMid              *float64 `json:"put_mid"`
	CallSpreadPct       *float64 `json:"call_spread_pct"`
	PutSpreadPct        *float64 `json:"put_spread_pct"`
}

// ZeroDteResponse is the full payload from GET /v1/exposure/zero-dte/{symbol}.
//
// On weekends/holidays or symbols with no 0DTE today, NoZeroDte is true and most
// fields are zero/nil — only Symbol, AsOf, Message, and NextZeroDteExpiry are
// populated.
//
// Raw holds the underlying decoded JSON for any field not modeled here.
type ZeroDteResponse struct {
	Symbol            string                 `json:"symbol"`
	UnderlyingPrice   *float64               `json:"underlying_price"`
	Expiration        *string                `json:"expiration"`
	AsOf              string                 `json:"as_of"`
	MarketOpen        bool                   `json:"market_open"`
	TimeToCloseHours  *float64               `json:"time_to_close_hours"`
	TimeToClosePct    *float64               `json:"time_to_close_pct"`
	Regime            *ZeroDteRegime         `json:"regime"`
	Exposures         *ZeroDteExposures      `json:"exposures"`
	ExpectedMove      *ZeroDteExpectedMove   `json:"expected_move"`
	PinRisk           *ZeroDtePinRisk        `json:"pin_risk"`
	Hedging           *ZeroDteHedging        `json:"hedging"`
	Decay             *ZeroDteDecay          `json:"decay"`
	VolContext        *ZeroDteVolContext     `json:"vol_context"`
	Flow              *ZeroDteFlow           `json:"flow"`
	Levels            *ZeroDteLevels         `json:"levels"`
	Liquidity         *ZeroDteLiquidity      `json:"liquidity"`
	Metadata          *ZeroDteMetadata       `json:"metadata"`
	Strikes           []ZeroDteStrike        `json:"strikes"`

	// Warnings is optional — only present near close (<5 min) when greeks may be unstable.
	Warnings []string `json:"warnings,omitempty"`

	// No-0DTE fallback fields
	NoZeroDte           bool    `json:"no_zero_dte"`
	Message             string  `json:"message"`
	NextZeroDteExpiry   *string `json:"next_zero_dte_expiry"`

	// Raw holds the unparsed JSON for forward compatibility with new fields.
	Raw map[string]interface{} `json:"-"`
}

// ZeroDteTyped is the strongly-typed variant of ZeroDte. The original
// ZeroDte continues to return map[string]interface{} unchanged.
//
// Returns a fully-populated *ZeroDteResponse with the Raw payload attached
// for any field not yet modeled.
func (c *Client) ZeroDteTyped(ctx context.Context, symbol string, opts ...ZeroDteOption) (*ZeroDteResponse, error) {
	cfg := &zeroDteConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.strikeRange != nil {
		params.Set("strike_range", strconv.FormatFloat(*cfg.strikeRange, 'f', -1, 64))
	}
	raw, err := c.get(ctx, "/v1/exposure/zero-dte/"+symbol, params)
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("flashalpha: re-encode zero-dte: %w", err)
	}
	out := &ZeroDteResponse{}
	if err := json.Unmarshal(buf, out); err != nil {
		return nil, fmt.Errorf("flashalpha: decode zero-dte: %w", err)
	}
	out.Raw = raw
	return out, nil
}
