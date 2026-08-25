package flashalpha

// Typed models + client methods for the range-based realized-volatility and
// conditional-vol-forecast surfaces in the Volatility Analytics family:
//
//   - RealizedVolatility  — GET /v1/volatility/realized/{symbol}  (Alpha+)
//   - VolatilityForecast  — GET /v1/volatility/forecast/{symbol}  (Alpha+)
//
// Both decode directly into the structs below.

import (
	"context"
	"net/url"
)

// ── volatility/realized ──────────────────────────────────────────────────────

// RealizedVolWindows is one estimator's annualized realized vol (percent) over
// the 10/20/30-day windows. Reused for all five estimator fields. Each window
// is nullable.
type RealizedVolWindows struct {
	// Rv10 is the annualized realized vol over the 10-day window (percent).
	Rv10 *float64 `json:"rv10"`
	// Rv20 is the annualized realized vol over the 20-day window (percent).
	Rv20 *float64 `json:"rv20"`
	// Rv30 is the annualized realized vol over the 30-day window (percent).
	Rv30 *float64 `json:"rv30"`
}

// RealizedVolEstimators bundles the five fixed range-based realized-vol
// estimators, each computed over the 10/20/30-day windows.
type RealizedVolEstimators struct {
	// CloseToClose is the classic close-to-close estimator.
	CloseToClose RealizedVolWindows `json:"close_to_close"`
	// Parkinson is the high-low range estimator.
	Parkinson RealizedVolWindows `json:"parkinson"`
	// GarmanKlass is the OHLC estimator.
	GarmanKlass RealizedVolWindows `json:"garman_klass"`
	// RogersSatchell is the drift-independent OHLC estimator.
	RogersSatchell RealizedVolWindows `json:"rogers_satchell"`
	// YangZhang is the overnight-jump-aware estimator.
	YangZhang RealizedVolWindows `json:"yang_zhang"`
}

// RealizedVolatilityResponse is the body of GET /v1/volatility/realized/{symbol}.
type RealizedVolatilityResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol          string                `json:"symbol"`
	AsOf            string                `json:"as_of"`
	UnderlyingPrice *float64              `json:"underlying_price"`
	Estimators      RealizedVolEstimators `json:"estimators"`
}

// RealizedVolatility returns range-based realized (historical) vol estimators —
// close-to-close, Parkinson, Garman-Klass, Rogers-Satchell, and Yang-Zhang —
// each annualized over 10/20/30-day windows. Requires Alpha+.
func (c *Client) RealizedVolatility(ctx context.Context, symbol string) (*RealizedVolatilityResponse, error) {
	raw, err := c.get(ctx, "/v1/volatility/realized/"+seg(symbol), nil)
	if err != nil {
		return nil, err
	}
	out := &RealizedVolatilityResponse{}
	if err := decodeTyped("realized-volatility", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── volatility/forecast ──────────────────────────────────────────────────────

// VolatilityForecastEwma is the EWMA (RiskMetrics) conditional-vol forecast.
type VolatilityForecastEwma struct {
	// Lambda is the EWMA decay factor (RiskMetrics default 0.94).
	Lambda float64 `json:"lambda"`
	// VolAnnualized is the current EWMA vol (annualized percent).
	VolAnnualized float64 `json:"vol_annualized"`
	// NextDayForecast is the 1-day-ahead EWMA vol forecast (annualized percent).
	NextDayForecast float64 `json:"next_day_forecast"`
}

// VolatilityForecastHarComponents is the HAR-RV daily/weekly/monthly component
// breakdown.
type VolatilityForecastHarComponents struct {
	// Daily is the 1-day realized-vol component (annualized percent).
	Daily float64 `json:"daily"`
	// Weekly is the 5-day realized-vol component (annualized percent).
	Weekly float64 `json:"weekly"`
	// Monthly is the 22-day realized-vol component (annualized percent).
	Monthly float64 `json:"monthly"`
}

// VolatilityForecastHarRv is the HAR-RV (heterogeneous autoregressive realized
// vol) forecast.
type VolatilityForecastHarRv struct {
	// VolAnnualized is the current HAR-RV fitted vol (annualized percent).
	VolAnnualized float64 `json:"vol_annualized"`
	// Components is the daily/weekly/monthly component breakdown.
	Components VolatilityForecastHarComponents `json:"components"`
	// NextDayForecast is the 1-day-ahead HAR-RV forecast (annualized percent).
	NextDayForecast float64 `json:"next_day_forecast"`
}

// VolatilityForecastGarchParams is the GARCH(1,1) MLE parameter estimate. Dof
// is populated only for the student_t distribution.
type VolatilityForecastGarchParams struct {
	// Omega is the GARCH constant (long-run variance scale).
	Omega float64 `json:"omega"`
	// Alpha is the ARCH (recent-shock) coefficient.
	Alpha float64 `json:"alpha"`
	// Beta is the GARCH (persistence) coefficient.
	Beta float64 `json:"beta"`
	// Dof is the Student-t degrees of freedom (nil for the gaussian fit).
	Dof *float64 `json:"dof"`
}

// VolatilityForecastGarchHorizon is one GARCH forecast horizon point.
type VolatilityForecastGarchHorizon struct {
	// HorizonDays is the forecast horizon in trading days.
	HorizonDays int `json:"horizon_days"`
	// VolAnnualized is the forecast vol at the horizon (annualized percent).
	VolAnnualized float64 `json:"vol_annualized"`
}

// VolatilityForecastGarch is the GARCH(1,1) MLE conditional-vol forecast.
type VolatilityForecastGarch struct {
	// Model is the model identifier (e.g. "garch_1_1").
	Model string `json:"model"`
	// Distribution is the innovation distribution ("student_t" or "gaussian").
	Distribution string `json:"distribution"`
	// Params is the fitted GARCH(1,1) parameter estimate.
	Params VolatilityForecastGarchParams `json:"params"`
	// Persistence is alpha + beta (closeness to a unit root).
	Persistence float64 `json:"persistence"`
	// LongRunVolAnnualized is the unconditional vol implied by the fit
	// (annualized percent); nil when undefined.
	LongRunVolAnnualized *float64 `json:"long_run_vol_annualized"`
	// HalfLifeDays is the shock half-life in trading days.
	HalfLifeDays float64 `json:"half_life_days"`
	// Converged is true if the MLE optimiser converged.
	Converged bool `json:"converged"`
	// Forecast is the per-horizon vol forecast term structure; nil when the
	// fit did not converge.
	Forecast []VolatilityForecastGarchHorizon `json:"forecast"`
}

// VolatilityForecastResponse is the body of GET /v1/volatility/forecast/{symbol}.
type VolatilityForecastResponse struct {
	// ResponseEnvelope carries data_as_of and endpoint_version.
	ResponseEnvelope
	Symbol string                  `json:"symbol"`
	AsOf   string                  `json:"as_of"`
	Ewma   VolatilityForecastEwma  `json:"ewma"`
	HarRv  VolatilityForecastHarRv `json:"har_rv"`
	Garch  VolatilityForecastGarch `json:"garch"`
}

// VolatilityForecastOption configures a VolatilityForecast request.
type VolatilityForecastOption func(*volatilityForecastConfig)

type volatilityForecastConfig struct {
	dist string
}

// WithForecastDist selects the GARCH innovation distribution: "student_t"
// (default) or "gaussian".
func WithForecastDist(dist string) VolatilityForecastOption {
	return func(c *volatilityForecastConfig) { c.dist = dist }
}

// VolatilityForecast returns conditional-vol forecasts from three models —
// EWMA (RiskMetrics, lambda 0.94), HAR-RV, and a GARCH(1,1) MLE fit with a
// per-horizon forecast term structure. Requires Alpha+. Optional:
// WithForecastDist.
func (c *Client) VolatilityForecast(ctx context.Context, symbol string, opts ...VolatilityForecastOption) (*VolatilityForecastResponse, error) {
	cfg := &volatilityForecastConfig{}
	for _, o := range opts {
		o(cfg)
	}
	params := url.Values{}
	if cfg.dist != "" {
		params.Set("dist", cfg.dist)
	}
	raw, err := c.get(ctx, "/v1/volatility/forecast/"+seg(symbol), params)
	if err != nil {
		return nil, err
	}
	out := &VolatilityForecastResponse{}
	if err := decodeTyped("volatility-forecast", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
