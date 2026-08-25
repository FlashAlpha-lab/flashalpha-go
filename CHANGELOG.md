# Changelog

All notable changes to `flashalpha-go` are documented here. This project follows
[Keep a Changelog](https://keepachangelog.com/) and [Semantic Versioning](https://semver.org/).

## [1.3.0] - 2026-08-25

### Added
- **`data_as_of` response envelope.** Every successful response now carries
  `data_as_of`, reporting when each upstream feed last delivered to the node that
  answered: equity and index spot, their option chains, futures and futures
  options, the classified trade tape, settled open interest, and the macro series,
  each reported separately because they arrive over different pipes and fail
  independently. `endpoint_version` identifies the deployment that produced the
  response.
- **`DataAsOf`** type and a `ResponseEnvelope` struct embedded in all 83 response
  types. Because the embed is anonymous, `encoding/json` flattens it: the wire
  shape is unchanged and both fields are promoted, so they read as
  `gex.DataAsOf` and `gex.EndpointVersion`. Previously `json.Unmarshal` bound only
  declared fields, so the envelope arrived and was discarded.
- Feed fields are `*string`, so a feed the node has not seen is `nil` rather than
  an empty string. Absent envelopes leave `DataAsOf` nil, so existing code is
  unaffected.

### Notes
- Read each feed against its own cadence rather than against `as_of`. Settled open
  interest dated to the previous session's close is correct, since it is published
  once per session; an options feed an hour behind during the regular session is
  not.
- A `nil` means that node has not seen that feed, not that it is broken.
- The field evidences that a feed delivered recently. It does not assert that every
  contract in a chain is equally current.
- Endpoints returning a bare JSON array carry the same information in the
  `X-Data-As-Of` and `X-Endpoint-Version` response headers.

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
