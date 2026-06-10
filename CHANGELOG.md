# Changelog

All notable changes to `flashalpha-go` are documented here. This project follows
[Keep a Changelog](https://keepachangelog.com/) and [Semantic Versioning](https://semver.org/).

## [1.1.0] - 2026-06-08

Full parity with the expanded FlashAlpha analytics API. ~33 new endpoints added,
each with an untyped method (`map[string]interface{}`) and a strongly-typed
`*Typed` variant. All existing method signatures remain backward-compatible —
new parameters are added as optional functional options.

### Added

- **Strategy Signals (10 endpoints):** `StrategyFlowAnomaly`,
  `StrategyExpiryPositioning`, `StrategyZeroDte`, `StrategyDealerRegime`,
  `StrategyVolCarry`, `StrategyYieldEnhancement`, `StrategySurfaceAnomaly`,
  `StrategySkew`, `StrategyTermStructure`, `StrategyTailPricing` — all return the
  shared `StrategyDecisionResponse` envelope (decision / score / confidence /
  regime / best_structures / why / avoid_if / data_quality). Tunable via
  `WithStrategyExpiry`, `WithStrategyMinOpenInterest`, `WithStrategyWingWidth`,
  `WithStrategyTargetShortDelta`, `WithStrategyMaxWidth`, `WithStrategyMinCredit`,
  `WithStrategyTargetDelta`, `WithStrategyStructure`,
  `WithStrategyExcludeEarningsBeforeExpiry`.
- **Earnings analytics (8 endpoints):** `EarningsCalendar`,
  `EarningsExpectedMove`, `EarningsHistory`, `EarningsIvCrush`, `EarningsVrp`,
  `EarningsDealerPositioning`, `EarningsStrategies`, `EarningsScreener` (with
  their `*Typed` variants and `WithEarnings*` options).
- **Multi-leg structures (POST, pure-math):** `StructurePnl` /
  `StructureGreeks` with `StructurePnlRequest` / `StructureGreeksRequest` and
  typed responses (payoff curve, breakevens, max P/L; aggregate position Greeks).
- **Zero-DTE Flow:** `FlowZeroDteSnapshot`, `FlowZeroDteSeries`,
  `FlowZeroDteHedgeFlow`, `FlowZeroDteHeatmap`, `FlowZeroDteStrikeFlow`
  (`WithZeroDteFlowBar`, `WithZeroDteFlowMinutes`, `WithZeroDteFlowSide`,
  `WithZeroDteFlowMetric`, `WithZeroDteFlowMode`).
- **Flow extras:** `FlowDealerPremium`, `FlowStockBars` (`WithBarsMinutes`).
- **Exposure extras:** `ExposureSheet` (`WithSheetExpiration`/`WithSheetMinOI`),
  `ExposureTermStructure`, `ExposureBasket` (`WithBasketWeights`),
  `ExposureOiDiff` (`WithOiDiffTopN`).
- **Volatility / macro:** `Liquidity`, `SkewTerm`, `SpotVolCorrelation`,
  `ExpectedMove` (`WithExpectedMoveExpiry`), `Dispersion`
  (`WithDispersionWeights`/`WithDispersionHorizonDays`), `SurfaceSvi`,
  `VrpHistory` (`WithVrpHistoryDays`), `VixState`, `Universe`
  (`WithUniverseSort`/`WithUniverseLimit`).
- **Screener:** `ScreenerFields` — enumerate screener fields + operators.
- New `examples/strategy_earnings_structures/main.go` exercising the strategy,
  earnings, and structure families.
- Unit tests (mock HTTP) for every new method asserting path, query params, and
  typed response shape.

### Changed

- `Vrp` gains the optional `WithVrpDate(date)` option to request a historical
  session snapshot instead of the live one.
- `ZeroDte` / `ZeroDteTyped` gain the optional `WithZeroDteExpiry(expiry)` option
  to target 1DTE/2DTE/any expiry via the same 0DTE selector. (Fix: the `expiry`
  param is now forwarded by the typed variant as well as the untyped one.)

### Notes

- No breaking changes. All previously published method signatures are preserved.
- README, `llms.txt`, and `AGENTS.md` updated to document the new endpoint
  families for SEO and LLM discovery.

## [1.0.1] - 2025-05-21

- Add live Flow API tier (24 endpoints) incl. flow signals; audit fixes.

## [1.0.0] - 2025-05-15

- Initial GA release.
