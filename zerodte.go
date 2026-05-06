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
	// Label is the dealer-gamma regime classification for the 0DTE chain.
	// Confirmed live values: "positive_gamma" (spot above gamma flip — dealers
	// dampen moves, mean reversion likely) or "negative_gamma" (spot below
	// flip — dealers amplify moves, trend-following likely).
	Label string `json:"label"`
	// Description is a plain-English narrative for the current regime — safe
	// to surface verbatim in customer-facing UIs (newsletters, dashboards,
	// chat bots).
	Description string `json:"description"`
	// GammaFlip is the strike where 0DTE net dealer gamma exposure crosses
	// zero. The single most-watched intraday level on this endpoint: spot
	// crossing the flip flips the dealer hedging regime.
	GammaFlip *float64 `json:"gamma_flip"`
	// SpotVsFlip is "above" or "below" — convenience label matching the sign
	// of (underlying_price - gamma_flip).
	SpotVsFlip string `json:"spot_vs_flip"`
	// SpotToFlipPct is signed % distance from spot to GammaFlip
	// (positive = spot above flip).
	SpotToFlipPct *float64 `json:"spot_to_flip_pct"`
	// DistanceToFlipDollars is the unsigned dollar distance from spot to
	// GammaFlip.
	DistanceToFlipDollars *float64 `json:"distance_to_flip_dollars"`
	// DistanceToFlipSigmas is the 1σ-normalized distance to the gamma flip,
	// using ATM IV * sqrt(t_remain) as the intraday vol scaler. Lets you
	// compare flip proximity across symbols and across time-of-day:
	// |sigmas| < 0.25 is "imminent flip risk".
	DistanceToFlipSigmas *float64 `json:"distance_to_flip_sigmas"`
}

// ZeroDteExposures aggregates net 0DTE GEX/DEX/VEX/CHEX and full-chain context.
type ZeroDteExposures struct {
	// NetGex is net 0DTE gamma exposure in dollars per 1% spot move. Positive
	// = dealers long gamma (moves dampened); negative = dealers short gamma
	// (moves amplified). Only counts same-day-expiration contracts.
	NetGex *float64 `json:"net_gex"`
	// NetDex is net 0DTE delta exposure in dollars. Sign indicates the
	// direction of the dealer hedge book against 0DTE inventory.
	NetDex *float64 `json:"net_dex"`
	// NetVex is net 0DTE vanna exposure in dollars per 1-vol-point. Captures
	// how the 0DTE dealer hedge responds to an IV shock.
	NetVex *float64 `json:"net_vex"`
	// NetChex is net 0DTE charm exposure in dollars per day — the time-decay
	// drift in dealer delta on today's expiry. Positive = dealers must BUY
	// into close to stay neutral (supportive); negative = SELL into close.
	NetChex *float64 `json:"net_chex"`
	// PctOfTotalGex is 0DTE share of full-chain net GEX as a percentage.
	// >50 means today's expiry dominates the dealer book — tradable signal
	// that intraday dealer flow will be driven primarily by 0DTE gamma.
	PctOfTotalGex *float64 `json:"pct_of_total_gex"`
	// TotalChainNetGex is full-chain (all expirations) net dealer GEX. Same
	// number as ExposureSummary.Exposures.NetGex — exposed here as the
	// denominator for PctOfTotalGex.
	TotalChainNetGex *float64 `json:"total_chain_net_gex"`
}

// ZeroDteExpectedMove holds implied 1σ move + remaining-session 1σ + ATM straddle.
type ZeroDteExpectedMove struct {
	// Implied1SdDollars is the full-day 1σ implied move in dollars from ATM
	// IV (annualised → daily). Static across the session.
	Implied1SdDollars *float64 `json:"implied_1sd_dollars"`
	// Implied1SdPct is Implied1SdDollars as a % of spot.
	Implied1SdPct *float64 `json:"implied_1sd_pct"`
	// Remaining1SdDollars is the 1σ implied move for the time REMAINING in
	// the session: ATM IV * sqrt(t_remain). Shrinks intraday as the close
	// approaches — the practical bound for "where can spot still go today?".
	Remaining1SdDollars *float64 `json:"remaining_1sd_dollars"`
	// Remaining1SdPct is Remaining1SdDollars as a % of spot.
	Remaining1SdPct *float64 `json:"remaining_1sd_pct"`
	// UpperBound is spot + Remaining1SdDollars — top of the implied
	// remaining-session range.
	UpperBound *float64 `json:"upper_bound"`
	// LowerBound is spot - Remaining1SdDollars — bottom of the implied
	// remaining-session range.
	LowerBound *float64 `json:"lower_bound"`
	// StraddlePrice is the ATM 0DTE straddle mid in dollars — the
	// market-implied expected move for the rest of the session. Often more
	// reliable than Implied1SdDollars near close because it incorporates the
	// IV smile rather than just ATM IV.
	StraddlePrice *float64 `json:"straddle_price"`
	// AtmIv is the ATM implied volatility for today's 0DTE chain (annualised
	// %, e.g. 18.5 = 18.5%).
	AtmIv *float64 `json:"atm_iv"`
}

// ZeroDtePinComponents is the sub-score breakdown for ZeroDtePinRisk.PinScore.
type ZeroDtePinComponents struct {
	// OiScore is the OI-concentration sub-score (0-100). Weight: 30% of
	// PinScore. Higher when OI clusters tightly around MagnetStrike.
	OiScore *int `json:"oi_score"`
	// ProximityScore is the magnet-proximity sub-score (0-100). Weight: 25%
	// of PinScore. Higher when spot is already close to MagnetStrike.
	ProximityScore *int `json:"proximity_score"`
	// TimeScore is the time-remaining sub-score (0-100). Weight: 25% of
	// PinScore. Rises through the session as gamma compresses to a delta
	// function near expiry.
	TimeScore *int `json:"time_score"`
	// GammaScore is the gamma-magnitude sub-score (0-100). Weight: 20% of
	// PinScore. Higher when MagnetGex is large relative to the rest of the
	// 0DTE chain.
	GammaScore *int `json:"gamma_score"`
}

// ZeroDtePinRisk is the pin-risk block — magnet strike + composite + sub-scores.
type ZeroDtePinRisk struct {
	// MagnetStrike is the strike with the largest absolute 0DTE GEX — the
	// single strike most likely to "pin" spot into the close via dealer
	// hedging flow.
	MagnetStrike *float64 `json:"magnet_strike"`
	// MagnetGex is the dealer gamma exposure at MagnetStrike (in dollars per
	// 1% spot move). Larger magnitude = stronger magnet effect.
	MagnetGex *float64 `json:"magnet_gex"`
	// DistanceToMagnetPct is the signed % distance from spot to MagnetStrike.
	DistanceToMagnetPct *float64 `json:"distance_to_magnet_pct"`
	// PinScore is a 0-100 composite — likelihood spot pins to MagnetStrike
	// at the close. Weights: OI 30% + proximity 25% + time 25% + gamma 20%.
	// >70 is a strong pin setup; <30 means a pin is unlikely.
	PinScore *int `json:"pin_score"`
	// Components is the sub-score breakdown that feeds PinScore — useful
	// for explaining WHY a pin score is what it is.
	Components *ZeroDtePinComponents `json:"components"`
	// MaxPain is the strike where total option-holder intrinsic value (× OI)
	// across today's chain is minimized. Often (but not always) close to
	// MagnetStrike.
	MaxPain *float64 `json:"max_pain"`
	// OiConcentrationTop3Pct is the share of total 0DTE OI (%) sitting at
	// the top-3 strikes. >50 means OI is highly concentrated — feeds OiScore.
	OiConcentrationTop3Pct *float64 `json:"oi_concentration_top3_pct"`
	// Description is a plain-English narrative for the current pin setup —
	// safe to surface verbatim in customer-facing UIs.
	Description string `json:"description"`
}

// ZeroDteHedgingBucket is one row of dealer hedging flow at a specific spot delta.
type ZeroDteHedgingBucket struct {
	// DealerSharesToTrade is the estimated underlying shares dealers must
	// trade to remain delta-neutral if spot moves to this bucket. Positive =
	// buy, negative = sell. Linearised from net 0DTE DEX around current spot.
	DealerSharesToTrade *float64 `json:"dealer_shares_to_trade"`
	// Direction is the lowercase convenience label matching the sign of
	// DealerSharesToTrade. Confirmed live values: "buy" or "sell".
	// (The published API doc shows uppercase "BUY"/"SELL", but the live
	// response is lowercase — this model reflects the actual response.)
	Direction string `json:"direction"`
	// NotionalUsd is |DealerSharesToTrade| × current_spot. Use for
	// cross-symbol comparison: a 1M-share hedge in SPY ($600 spot) is much
	// larger than a 1M-share hedge in HOOD ($30 spot), and notional captures
	// that.
	NotionalUsd *float64 `json:"notional_usd"`
}

// ZeroDteHedging holds dealer hedging flow at ±10bp/25bp/50bp/100bp + GEX convexity.
type ZeroDteHedging struct {
	// SpotUp10Bp is the dealer hedging flow if spot rises 10 basis points
	// (0.10%). Smallest move bucket — most relevant for tick-level scalping.
	SpotUp10Bp *ZeroDteHedgingBucket `json:"spot_up_10bp"`
	// SpotDown10Bp is the dealer hedging flow if spot falls 10 basis points.
	// Equal magnitude to SpotUp10Bp at first order, but the convexity term
	// makes the two sides asymmetric in the negative-gamma regime.
	SpotDown10Bp *ZeroDteHedgingBucket `json:"spot_down_10bp"`
	// SpotUp25Bp is the dealer hedging flow if spot rises 25 basis points.
	SpotUp25Bp *ZeroDteHedgingBucket `json:"spot_up_25bp"`
	// SpotDown25Bp is the dealer hedging flow if spot falls 25 basis points.
	SpotDown25Bp *ZeroDteHedgingBucket `json:"spot_down_25bp"`
	// SpotUpHalfPct is the dealer hedging flow if spot rises 50 basis points
	// (0.50%).
	SpotUpHalfPct *ZeroDteHedgingBucket `json:"spot_up_half_pct"`
	// SpotDownHalfPct is the dealer hedging flow if spot falls 50 basis points.
	SpotDownHalfPct *ZeroDteHedgingBucket `json:"spot_down_half_pct"`
	// SpotUp1Pct is the dealer hedging flow if spot rises 1%. Largest
	// bucket — the canonical "stress test" move.
	SpotUp1Pct *ZeroDteHedgingBucket `json:"spot_up_1pct"`
	// SpotDown1Pct is the dealer hedging flow if spot falls 1%.
	SpotDown1Pct *ZeroDteHedgingBucket `json:"spot_down_1pct"`
	// ConvexityAtSpot is the 2nd finite-difference of net GEX taken across
	// the three strikes nearest spot — proxies the curvature of the dealer
	// hedge book. Large absolute values mean small spot moves trigger
	// outsized hedging adjustments (gamma squeeze risk).
	ConvexityAtSpot *float64 `json:"convexity_at_spot"`
}

// ZeroDteDecay is the time-decay block — net theta + per-hour rate + acceleration.
type ZeroDteDecay struct {
	// NetThetaDollars is the net dollar theta of the 0DTE chain — total
	// dealer-side theta P&L if spot, IV, and skew were all frozen for one
	// trading day. (Most useful as a magnitude reference for ThetaPerHourRemaining.)
	NetThetaDollars *float64 `json:"net_theta_dollars"`
	// ThetaPerHourRemaining is NetThetaDollars / hours_to_close — the
	// effective hourly theta burn from now until the close. Accelerates
	// non-linearly as the denominator shrinks: a flat-looking theta at 10am
	// becomes the dominant force in the last 30 minutes.
	ThetaPerHourRemaining *float64 `json:"theta_per_hour_remaining"`
	// CharmRegime is a label classifying the dealer-charm direction
	// (e.g. "supportive_into_close" or "pressure_into_close"). Derived from
	// the sign of NetChex.
	CharmRegime string `json:"charm_regime"`
	// CharmDescription is a plain-English narrative for CharmRegime — safe
	// to surface verbatim.
	CharmDescription string `json:"charm_description"`
	// GammaAcceleration is the ratio of 0DTE ATM gamma to 7DTE ATM gamma
	// (typically 2-5x and rising as the session progresses). A direct
	// measurement of how much "more gamma" today's expiry has vs the rest
	// of the term structure — the headline number for why 0DTE matters.
	GammaAcceleration *float64 `json:"gamma_acceleration"`
	// Description is a plain-English narrative for the overall decay setup —
	// safe to surface verbatim.
	Description string `json:"description"`
}

// ZeroDteVolContext gives the vol-surface context — 0DTE vs 7DTE IV + vanna read.
type ZeroDteVolContext struct {
	// ZeroDteAtmIv is ATM implied vol for today's 0DTE chain (annualised %).
	ZeroDteAtmIv *float64 `json:"zero_dte_atm_iv"`
	// SevenDteAtmIv is ATM implied vol for the 7DTE chain (annualised %) —
	// the term-structure reference point.
	SevenDteAtmIv *float64 `json:"seven_dte_atm_iv"`
	// IvRatio0Dte7Dte is ZeroDteAtmIv / SevenDteAtmIv. Values < 1 mean 0DTE
	// vol is CHEAP relative to the front of the term structure (often a sign
	// of complacency — short-gamma players underpricing intraday risk).
	// Values > 1 mean 0DTE is BID — fear premium for today's session.
	IvRatio0Dte7Dte *float64 `json:"iv_ratio_0dte_7dte"`
	// Vix is the CBOE VIX index level — macro vol context for the
	// single-name 0DTE read.
	Vix *float64 `json:"vix"`
	// VannaExposure is net 0DTE vanna exposure in dollars per 1-vol-point.
	// Positive = dealers buy stock when vol drops (supportive bid);
	// negative = dealers sell on a vol drop.
	VannaExposure *float64 `json:"vanna_exposure"`
	// VannaInterpretation is a label for the dealer-vanna setup
	// (e.g. "supportive", "cascade_risk").
	VannaInterpretation string `json:"vanna_interpretation"`
	// Description is a plain-English narrative for the overall vol-context
	// read — safe to surface verbatim.
	Description string `json:"description"`
}

// ZeroDteFlow is the flow block — volume/OI aggregates + concentration metrics.
type ZeroDteFlow struct {
	// TotalVolume is total 0DTE option contracts traded so far in the session
	// (calls + puts).
	TotalVolume *int64 `json:"total_volume"`
	// CallVolume is total 0DTE call contracts traded.
	CallVolume *int64 `json:"call_volume"`
	// PutVolume is total 0DTE put contracts traded.
	PutVolume *int64 `json:"put_volume"`
	// NetCallMinusPutVolume is CallVolume - PutVolume. Positive = call-heavy
	// session; negative = put-heavy session.
	NetCallMinusPutVolume *int64 `json:"net_call_minus_put_volume"`
	// TotalOi is total 0DTE open interest entering the session (calls + puts).
	TotalOi *int64 `json:"total_oi"`
	// CallOi is 0DTE call OI entering the session.
	CallOi *int64 `json:"call_oi"`
	// PutOi is 0DTE put OI entering the session.
	PutOi *int64 `json:"put_oi"`
	// PcRatioVolume is PutVolume / CallVolume. >1 = bearish flow lean; the
	// classic put-call ratio.
	PcRatioVolume *float64 `json:"pc_ratio_volume"`
	// PcRatioOi is PutOi / CallOi — positioning lean entering the session.
	PcRatioOi *float64 `json:"pc_ratio_oi"`
	// VolumeToOiRatio is TotalVolume / TotalOi. Values > 1 mean today's
	// volume already exceeds yesterday's OI — heavy day-trading / unusually
	// active 0DTE session.
	VolumeToOiRatio *float64 `json:"volume_to_oi_ratio"`
	// AtmVolumeSharePct is the share of TotalVolume (%) traded at ATM
	// strikes. Higher = volume is concentrated at-the-money (typical for
	// pin-trading sessions).
	AtmVolumeSharePct *float64 `json:"atm_volume_share_pct"`
	// Top3StrikeVolumePct is the share of TotalVolume (%) traded at the
	// three highest-volume strikes. Concentration metric — high values
	// flag specific strike-level trades dominating the tape.
	Top3StrikeVolumePct *float64 `json:"top3_strike_volume_pct"`
}

// ZeroDteLevels holds key strikes — call/put walls (with strength), gamma extrema,
// highest-OI strike, distance-to-magnet, and the level-cluster composite.
type ZeroDteLevels struct {
	// CallWall is the strike with the largest absolute call GEX (dealer-side
	// resistance — spot tends to stall here in the positive-gamma regime).
	CallWall *float64 `json:"call_wall"`
	// CallWallGex is the dealer gamma exposure at CallWall in dollars per
	// 1% spot move.
	CallWallGex *float64 `json:"call_wall_gex"`
	// CallWallStrength is |CallWallGex| / total absolute call-side GEX —
	// the share of all call-side GEX concentrated at CallWall. Higher =
	// stronger resistance (the wall is a single dominant strike rather
	// than a smear).
	CallWallStrength *float64 `json:"call_wall_strength"`
	// DistanceToCallWallPct is signed % distance from spot to CallWall
	// (positive = wall is above spot).
	DistanceToCallWallPct *float64 `json:"distance_to_call_wall_pct"`
	// PutWall is the strike with the largest absolute put GEX (dealer-side
	// support — spot tends to bounce here in the positive-gamma regime).
	PutWall *float64 `json:"put_wall"`
	// PutWallGex is the dealer gamma exposure at PutWall in dollars per
	// 1% spot move.
	PutWallGex *float64 `json:"put_wall_gex"`
	// PutWallStrength is |PutWallGex| / total absolute put-side GEX — the
	// share of all put-side GEX concentrated at PutWall. Higher = stronger
	// support.
	PutWallStrength *float64 `json:"put_wall_strength"`
	// DistanceToPutWallPct is signed % distance from spot to PutWall
	// (negative = wall is below spot).
	DistanceToPutWallPct *float64 `json:"distance_to_put_wall_pct"`
	// DistanceToMagnetDollars is unsigned dollar distance from spot to the
	// pin-risk magnet strike (same magnet as ZeroDtePinRisk.MagnetStrike).
	DistanceToMagnetDollars *float64 `json:"distance_to_magnet_dollars"`
	// HighestOiStrike is the strike with the largest total 0DTE OI
	// (calls + puts). Often coincides with MagnetStrike but not always.
	HighestOiStrike *float64 `json:"highest_oi_strike"`
	// HighestOiTotal is the total OI (calls + puts) at HighestOiStrike.
	HighestOiTotal *int64 `json:"highest_oi_total"`
	// MaxPositiveGamma is the strike with the largest positive 0DTE net
	// GEX (typically the most call-heavy strike).
	MaxPositiveGamma *float64 `json:"max_positive_gamma"`
	// MaxNegativeGamma is the strike with the largest negative 0DTE net
	// GEX (typically the most put-heavy strike).
	MaxNegativeGamma *float64 `json:"max_negative_gamma"`
	// LevelClusterScore is a 0-100 composite — how tightly the key levels
	// (call wall, put wall, magnet, gamma flip) cluster relative to the
	// remaining-session 1σ move. High values mean every level is within a
	// 1σ ring around spot (chaotic intraday tape); low values mean levels
	// are spread out (cleaner directional setups).
	LevelClusterScore *int `json:"level_cluster_score"`
}

// ZeroDteLiquidity is the bid-ask liquidity context for the 0DTE chain.
type ZeroDteLiquidity struct {
	// AtmSpreadPct is the ATM bid-ask spread as a % of mid. The headline
	// liquidity number — wider spreads mean costlier execution and less
	// reliable greeks.
	AtmSpreadPct *float64 `json:"atm_spread_pct"`
	// WeightedSpreadPct is the volume-weighted average bid-ask spread (% of
	// mid) across all 0DTE strikes — broader liquidity context.
	WeightedSpreadPct *float64 `json:"weighted_spread_pct"`
	// ExecutionScore is a 0-100 liquidity score (70% spread + 30% ATM OI
	// depth). >80 = institutional-grade liquidity; <40 = thin chain, expect
	// slippage and unstable greeks.
	ExecutionScore *int `json:"execution_score"`
}

// ZeroDteMetadata is the staleness + quality metadata for the snapshot.
type ZeroDteMetadata struct {
	// SnapshotAgeSeconds is the age of the underlying chain snapshot in
	// seconds at the time the response was assembled. Use as a staleness
	// check: values >5s suggest you should retry or treat the read as
	// out-of-date.
	SnapshotAgeSeconds *float64 `json:"snapshot_age_seconds"`
	// ChainContractCount is the number of contracts (calls + puts) in
	// today's 0DTE chain. Lower counts mean a thinner book and noisier
	// greeks/exposures.
	ChainContractCount *int `json:"chain_contract_count"`
	// DataQualityScore is a 0-100 composite quality metric blending
	// freshness, contract count, and quote validity. Use as a single-number
	// "is this snapshot trustworthy?" signal.
	DataQualityScore *int `json:"data_quality_score"`
	// GreekSmoothnessScore is a 0-100 measurement of IV smoothness across
	// strikes. Low values flag a noisy / disjoint surface where individual
	// strike greeks may be unreliable; high values mean the surface fit is
	// clean.
	GreekSmoothnessScore *int `json:"greek_smoothness_score"`
}

// ZeroDteStrike is one row in ZeroDteResponse.Strikes — per-strike exposure,
// flow, greeks, and quote/spread metrics.
type ZeroDteStrike struct {
	// Strike is the strike price (always populated).
	Strike float64 `json:"strike"`
	// DistanceFromSpotPct is signed % distance from spot to Strike
	// (positive = strike above spot).
	DistanceFromSpotPct *float64 `json:"distance_from_spot_pct"`
	// CallSymbol is the OCC option symbol for the call side at this strike.
	CallSymbol string `json:"call_symbol"`
	// PutSymbol is the OCC option symbol for the put side at this strike.
	PutSymbol string `json:"put_symbol"`
	// CallGex is the dealer-side gamma exposure ($/1% spot move) for calls
	// at this strike.
	CallGex *float64 `json:"call_gex"`
	// PutGex is the dealer-side gamma exposure ($/1% spot move) for puts
	// at this strike.
	PutGex *float64 `json:"put_gex"`
	// NetGex is CallGex + PutGex — net dealer gamma at this strike. Sum
	// across all strikes equals ZeroDteExposures.NetGex.
	NetGex *float64 `json:"net_gex"`
	// CallDex is the dealer-side delta exposure ($) for calls at this strike.
	CallDex *float64 `json:"call_dex"`
	// PutDex is the dealer-side delta exposure ($) for puts at this strike.
	PutDex *float64 `json:"put_dex"`
	// NetDex is CallDex + PutDex — net dealer delta exposure ($) at this
	// strike.
	NetDex *float64 `json:"net_dex"`
	// NetVex is the net dealer vanna exposure at this strike ($/1-vol-point).
	NetVex *float64 `json:"net_vex"`
	// NetChex is the net dealer charm exposure at this strike ($/day).
	NetChex *float64 `json:"net_chex"`
	// CallOi is open interest on the call side at this strike.
	CallOi *int64 `json:"call_oi"`
	// PutOi is open interest on the put side at this strike.
	PutOi *int64 `json:"put_oi"`
	// CallVolume is session volume on the call side at this strike.
	CallVolume *int64 `json:"call_volume"`
	// PutVolume is session volume on the put side at this strike.
	PutVolume *int64 `json:"put_volume"`
	// GexSharePct is |NetGex| as a share (%) of total absolute 0DTE GEX —
	// how much of today's gamma footprint sits at this strike.
	GexSharePct *float64 `json:"gex_share_pct"`
	// OiSharePct is (CallOi + PutOi) as a share (%) of total 0DTE OI.
	OiSharePct *float64 `json:"oi_share_pct"`
	// VolumeSharePct is (CallVolume + PutVolume) as a share (%) of total
	// 0DTE session volume.
	VolumeSharePct *float64 `json:"volume_share_pct"`
	// CallIv is the implied vol of the call leg at this strike (annualised %).
	CallIv *float64 `json:"call_iv"`
	// PutIv is the implied vol of the put leg at this strike (annualised %).
	PutIv *float64 `json:"put_iv"`
	// CallDelta is the call leg's delta (Black-Scholes, dealer-side sign
	// convention).
	CallDelta *float64 `json:"call_delta"`
	// PutDelta is the put leg's delta.
	PutDelta *float64 `json:"put_delta"`
	// CallGamma is the call leg's gamma.
	CallGamma *float64 `json:"call_gamma"`
	// PutGamma is the put leg's gamma.
	PutGamma *float64 `json:"put_gamma"`
	// CallTheta is the call leg's theta ($/day).
	CallTheta *float64 `json:"call_theta"`
	// PutTheta is the put leg's theta ($/day).
	PutTheta *float64 `json:"put_theta"`
	// CallMid is the call mid price (NBBO mid).
	CallMid *float64 `json:"call_mid"`
	// PutMid is the put mid price (NBBO mid).
	PutMid *float64 `json:"put_mid"`
	// CallSpreadPct is the call bid-ask spread as a % of mid (liquidity
	// proxy for this leg).
	CallSpreadPct *float64 `json:"call_spread_pct"`
	// PutSpreadPct is the put bid-ask spread as a % of mid.
	PutSpreadPct *float64 `json:"put_spread_pct"`
}

// ZeroDteResponse is the full payload from GET /v1/exposure/zero-dte/{symbol}.
//
// On weekends/holidays or symbols with no 0DTE today, NoZeroDte is true and most
// fields are zero/nil — only Symbol, AsOf, Message, and NextZeroDteExpiry are
// populated.
//
// Raw holds the underlying decoded JSON for any field not modeled here.
type ZeroDteResponse struct {
	// Symbol is the underlying ticker echoed from the request path
	// (e.g. "SPY", "QQQ", "AAPL").
	Symbol string `json:"symbol"`
	// UnderlyingPrice is the spot mid at AsOf — the reference price for all
	// GEX/DEX/VEX/CHEX dollarisation in this response.
	UnderlyingPrice *float64 `json:"underlying_price"`
	// Expiration is the ISO date ("yyyy-MM-dd") of today's 0DTE expiry.
	// Nil on the no-0DTE fallback path (see NoZeroDte / NextZeroDteExpiry).
	Expiration *string `json:"expiration"`
	// AsOf is the ET wall-clock timestamp this snapshot was computed for.
	AsOf string `json:"as_of"`
	// MarketOpen is true if NYSE was open at AsOf. False overnight, weekends,
	// and holidays.
	MarketOpen bool `json:"market_open"`
	// TimeToCloseHours is hours remaining until the regular-session close.
	// Drives ThetaPerHourRemaining and Remaining1SdDollars.
	TimeToCloseHours *float64 `json:"time_to_close_hours"`
	// TimeToClosePct is fraction (0-1) of the regular session still
	// remaining (1.0 at the open, 0.0 at the close).
	TimeToClosePct *float64 `json:"time_to_close_pct"`
	// Regime is the gamma-regime block (positive_gamma vs negative_gamma,
	// gamma-flip strike, distance-to-flip metrics). See ZeroDteRegime.
	Regime *ZeroDteRegime `json:"regime"`
	// Exposures is the net 0DTE Greek totals plus the 0DTE share of
	// full-chain GEX. See ZeroDteExposures.
	Exposures *ZeroDteExposures `json:"exposures"`
	// ExpectedMove is implied 1σ (full-day and remaining-session) plus the
	// ATM straddle price. See ZeroDteExpectedMove.
	ExpectedMove *ZeroDteExpectedMove `json:"expected_move"`
	// PinRisk is the magnet strike + composite pin score + sub-scores.
	// See ZeroDtePinRisk.
	PinRisk *ZeroDtePinRisk `json:"pin_risk"`
	// Hedging is the dealer hedging-flow estimates at ±10bp / ±25bp /
	// ±50bp / ±1pct spot moves plus the GEX convexity term. See ZeroDteHedging.
	Hedging *ZeroDteHedging `json:"hedging"`
	// Decay is the time-decay block — net theta, per-hour theta burn, charm
	// regime, and the 0DTE-vs-7DTE gamma acceleration ratio. See ZeroDteDecay.
	Decay *ZeroDteDecay `json:"decay"`
	// VolContext is the vol-surface context — 0DTE vs 7DTE IV ratio, VIX,
	// and the dealer-vanna read. See ZeroDteVolContext.
	VolContext *ZeroDteVolContext `json:"vol_context"`
	// Flow is the volume/OI aggregates + concentration metrics for today's
	// 0DTE session. See ZeroDteFlow.
	Flow *ZeroDteFlow `json:"flow"`
	// Levels holds the key strikes — call/put walls (with strength), gamma
	// extrema, highest-OI strike, level-cluster composite. See ZeroDteLevels.
	Levels *ZeroDteLevels `json:"levels"`
	// Liquidity is the bid-ask context — ATM spread, weighted spread, and
	// composite execution score. See ZeroDteLiquidity.
	Liquidity *ZeroDteLiquidity `json:"liquidity"`
	// Metadata is staleness + quality metadata — snapshot age, contract
	// count, data-quality and greek-smoothness scores. See ZeroDteMetadata.
	Metadata *ZeroDteMetadata `json:"metadata"`
	// Strikes is the per-strike grid — exposure, flow, greeks, and quote
	// metrics for every 0DTE strike. See ZeroDteStrike.
	Strikes []ZeroDteStrike `json:"strikes"`

	// Warnings is optional — only present near close (<5 min) when greeks may be unstable.
	Warnings []string `json:"warnings,omitempty"`

	// NoZeroDte is the fallback flag — true on weekends, holidays, or for
	// symbols without a same-day expiry today. When true, most fields above
	// are nil/zero and only Symbol, AsOf, Message, and NextZeroDteExpiry
	// are populated.
	NoZeroDte bool `json:"no_zero_dte"`
	// Message is a plain-English explanation of the no-0DTE state when
	// NoZeroDte is true (e.g. "Market closed today" or "Symbol has no
	// daily-expiry options").
	Message string `json:"message"`
	// NextZeroDteExpiry is the ISO date ("yyyy-MM-dd") of the next available
	// 0DTE expiry on this symbol when today doesn't have one.
	NextZeroDteExpiry *string `json:"next_zero_dte_expiry"`

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
