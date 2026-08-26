# flashalpha-go

Official Go client for the [FlashAlpha](https://flashalpha.com) options analytics API.

FlashAlpha delivers institutional-grade options analytics including a **live
options screener** (filter/rank symbols by GEX, VRP, IV, greeks, harvest scores,
and custom formulas), gamma exposure (GEX), delta exposure (DEX), vanna and
charm exposure, volatility surfaces, 0DTE analytics, and Black-Scholes-Merton
pricing utilities — all via a simple REST API.

> 🔑 **[Get a free API key at flashalpha.com →](https://flashalpha.com)** · 📚 [API documentation](https://flashalpha.com/docs) · 💹 [FlashAlpha options analytics API](https://flashalpha.com)

## Installation

Requires Go 1.21 or later. No external dependencies.

```sh
go get github.com/FlashAlpha-lab/flashalpha-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

func main() {
    client := flashalpha.NewClient("YOUR_API_KEY")
    ctx := context.Background()

    // Gamma exposure for SPY
    gex, err := client.Gex(ctx, "SPY")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(gex)

    // 0DTE analytics
    dte, err := client.ZeroDte(ctx, "SPY", flashalpha.WithStrikeRange(0.05))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(dte)

    // BSM greeks
    greeks, err := client.Greeks(ctx, flashalpha.GreeksParams{
        Spot:   450,
        Strike: 455,
        DTE:    30,
        Sigma:  0.20,
        Type:   "call",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(greeks)

    // Live options screener — harvestable VRP setups
    limit := 10
    screen, err := client.Screener(ctx, flashalpha.ScreenerRequest{
        Filters: flashalpha.ScreenerGroup{
            Op: "and",
            Conditions: []interface{}{
                flashalpha.ScreenerLeaf{Field: "regime", Operator: "eq", Value: "positive_gamma"},
                flashalpha.ScreenerLeaf{Field: "harvest_score", Operator: "gte", Value: 65},
            },
        },
        Sort:   []flashalpha.ScreenerSort{{Field: "harvest_score", Direction: "desc"}},
        Select: []string{"symbol", "price", "harvest_score", "dealer_flow_risk"},
        Limit:  &limit,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(screen)
}
```

## Data provenance: `data_as_of`

Every successful JSON-object response carries `data_as_of`, reporting when each upstream
feed last delivered to the node that answered, plus `endpoint_version` identifying the
deployment that produced it. That is every method on this client except the handful that
return a bare JSON array - see the note at the end of this section.

```go
gex, err := client.GexTyped(ctx, "SPY")

*gex.DataAsOf.EquityOptionsFeed // "2026-08-25T18:48:58.204Z"
*gex.DataAsOf.OiFeed            // "2026-08-24T20:00:00.000Z"  prior session's close
gex.DataAsOf.Node               // "fa2"
gex.EndpointVersion             // "2026.08.25"
```

Every response type embeds `ResponseEnvelope`, so `DataAsOf` and `EndpointVersion` are
promoted fields on all of them rather than data `json.Unmarshal` silently drops. Feed
fields are `*string`, so a feed the node has not seen is `nil` rather than `""`.

| Field | Feed | Expected cadence |
|---|---|---|
| `node` | Which node answered | Nodes hydrate independently |
| `equity_feed` | Equity and ETF spot quotes | seconds, during market hours |
| `equity_options_feed` | Equity and ETF option quotes | seconds, during market hours |
| `index_feed` | Index spot (SPX, RUT, VIX and the other index roots) | seconds, during market hours |
| `index_options_feed` | Index option quotes | seconds, during market hours |
| `futures_feed` | Futures prices | seconds, during the futures session |
| `futures_options_feed` | Futures option quotes | seconds, during the futures session |
| `flow_feed` | Classified options and stock trade tape | seconds, during market hours |
| `oi_feed` | Settled open interest | daily, dated to the prior 16:00 ET close |
| `macro_feed` | VIX, VVIX, SKEW, MOVE, SPX, Fear & Greed | minutes; reports its OLDEST component |

### How to read it

- **Check the feeds your call depends on.** A GEX call on an equity is answered from
  `equity_feed`, `equity_options_feed` and `oi_feed`. `futures_feed` being `null` in that
  response says nothing about the answer.
- **Compare against the cadence, not the clock.** `oi_feed` at the previous session's
  close is correct: settled open interest is published once per session, so on a Monday
  the newest figure that exists is Friday's. An options feed an hour behind during the
  regular session is not correct.
- **`null` means "not seen on this node", not "broken".** A node that has never been
  asked for a futures symbol has never opened that feed.
- **Spot and options are separate on purpose.** They arrive over different pipes and can
  fail independently.
- **It evidences feed activity, not per-contract freshness.** An illiquid strike may not
  have quoted for hours while its feed is healthy.
- **`data_as_of` is not `as_of`.** `as_of` is response-generation time or the newest
  contract in the payload, depending on the endpoint. `data_as_of` describes the feeds
  behind it.

### Bare-array endpoints

A few endpoints return a bare JSON array, which has nowhere to put an envelope in the
body. The API sends the same information in the `X-Data-As-Of` and `X-Endpoint-Version`
response headers instead - but this client returns the parsed body only and does not
surface response headers, so the envelope is **not reachable through those methods**.
Call the HTTP endpoint directly if you need provenance for one of them.

Full reference: <https://flashalpha.com/docs/lab-api-overview#response-envelope> and the
methodology whitepaper at <https://flashalpha.com/methodology#freshness-reporting>.
## Strategy signals, earnings, and structures

The SDK covers the full FlashAlpha analytics surface — actionable **strategy
signals**, **earnings analytics**, and multi-leg **structure** pricing:

```go
// Strategy signal — one typed decision envelope (StrategyDecisionResponse)
// shared by all 10 strategy endpoints.
sig, err := client.StrategyVolCarry(ctx, "SPY",
    flashalpha.WithStrategyExpiry("2026-06-19"),
    flashalpha.WithStrategyMinCredit(0.50),
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%s: decision=%s score=%d regime=%s\n",
    sig.Strategy, sig.Decision, sig.Score, sig.Regime)

// Earnings — straddle-implied expected move into the print
em, err := client.EarningsExpectedMoveTyped(ctx, "NVDA")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("earnings on %s, expected move block: %+v\n",
    em.EarningsDate, em.ExpectedMove)

// Multi-leg structure — payoff curve for an arbitrary spread (pure math)
pnl, err := client.StructurePnlTyped(ctx, flashalpha.StructurePnlRequest{
    Legs: []flashalpha.StructurePnlLeg{
        {Action: "buy", Type: "call", Strike: 455, Premium: 6.20, Quantity: 1},
        {Action: "sell", Type: "call", Strike: 465, Premium: 2.10, Quantity: 1},
    },
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("breakevens=%v max_profit=%v max_loss=%v\n",
    pnl.Breakevens, pnl.MaxProfit, pnl.MaxLoss)
```

## Authentication

Every request requires an API key passed via the `X-Api-Key` header. Get your
key at [flashalpha.com](https://flashalpha.com).

```go
client := flashalpha.NewClient(os.Getenv("FLASHALPHA_API_KEY"))
```

To override the base URL (for testing or staging):

```go
client := flashalpha.NewClientWithURL(apiKey, "https://staging.flashalpha.com")
```

## All Methods

All methods take `context.Context` as the first argument and return
`(map[string]interface{}, error)`.

### Exposure Analytics

| Method | Description | Plan |
|---|---|---|
| `Gex(ctx, symbol, ...GexOption)` | Gamma exposure by strike | Free+ |
| `Dex(ctx, symbol, ...DexOption)` | Delta exposure by strike | Basic+ |
| `Vex(ctx, symbol, ...VexOption)` | Vanna exposure by strike | Basic+ |
| `Chex(ctx, symbol, ...ChexOption)` | Charm exposure by strike | Basic+ |
| `ExposureLevels(ctx, symbol)` | Key support/resistance levels from options | Free+ |
| `ExposureSummary(ctx, symbol)` | Full GEX/DEX/VEX/CHEX + hedging summary | Growth+ |
| `Narrative(ctx, symbol)` | Verbal narrative analysis of exposure | Growth+ |
| `ZeroDte(ctx, symbol, ...ZeroDteOption)` | 0DTE regime, expected move, pin risk (`WithZeroDteExpiry` targets 1DTE/2DTE/any expiry) | Growth+ |
| `MaxPain(ctx, symbol, ...MaxPainOption)` | Max pain analysis with dealer alignment, pain curve, pin probability | Basic+ |
| `ExposureSheet(ctx, symbol, ...ExposureSheetOption)` | Full per-strike exposure sheet — net GEX/DEX/VEX/CHEX by strike (`WithSheetExpiration`, `WithSheetMinOI`) | Growth+ |
| `ExposureTermStructure(ctx, symbol)` | Dealer exposure bucketed by DTE — gamma/vanna/charm term structure | Growth+ |
| `ExposureBasket(ctx, symbols, ...BasketOption)` | Aggregate dealer exposure across a custom basket (`WithBasketWeights`) | Growth+ |
| `ExposureOiDiff(ctx, symbol, ...OiDiffOption)` | Day-over-day open-interest change by strike, top movers (`WithOiDiffTopN`) | Growth+ |

### Flow (live, simulation-aware) — Growth+ (raw tape, unusual-flow signals, OI simulator state & the full live bundle are Alpha)

Each method has a strongly-typed `*Typed` variant (e.g. `FlowLevelsTyped`).

| Method | Description |
|--------|-------------|
| `FlowLevels(ctx, symbol, ...FlowOption)` | Live gamma flip / call & put walls / max pain |
| `FlowPinRisk(ctx, symbol, ...FlowOption)` | 0DTE pin-risk score + component breakdown |
| `FlowSummary(ctx, symbol, ...FlowOption)` | At-a-glance flow direction + headline GEX shift |
| `FlowOi(ctx, symbol, ...FlowOption)` | Open-interest simulator state (official vs intraday) |
| `FlowGex(ctx, symbol, ...FlowOption)` | Live (flow-adjusted) GEX + per-strike profile |
| `FlowDex(ctx, symbol, ...FlowOption)` | Live (flow-adjusted) DEX + per-strike profile |
| `FlowDealerRisk(ctx, symbol, ...FlowOption)` | Settled-vs-live dealer GEX/DEX + flow adjustment |
| `FlowLive(ctx, symbol, ...FlowOption)` | Everything-at-once live flow bundle |
| `FlowSignals(ctx, symbol, ...FlowOption)` | Scored, classified unusual-flow feed (block/sweep, intent, 0-100 score) |
| `FlowSignalsSummary(ctx, symbol, ...FlowOption)` | Net bullish/bearish + opening/closing premium roll-up + top 10 signals |
| `FlowOptionRecent(ctx, symbol, ...FlowOption)` | Recent option trades, newest-first |
| `FlowOptionSummary(ctx, symbol, ...FlowOption)` | Per-underlying option-flow aggregates |
| `FlowOptionBlocks(ctx, symbol, ...FlowOption)` | Large option prints (`size >= minSize`) |
| `FlowOptionHistory(ctx, symbol, ...FlowOption)` | Per-minute option-flow buckets |
| `FlowOptionCumulative(ctx, symbol, ...FlowOption)` | Cumulative option net-flow series |
| `FlowStockRecent(ctx, symbol, ...FlowOption)` | Recent stock trades, newest-first |
| `FlowStockSummary(ctx, symbol)` | Per-symbol stock-flow aggregates |
| `FlowStockBlocks(ctx, symbol, ...FlowOption)` | Large stock prints (`size >= minSize`) |
| `FlowStockHistory(ctx, symbol, ...FlowOption)` | Per-minute stock-flow buckets w/ OHLC |
| `FlowStockCumulative(ctx, symbol, ...FlowOption)` | Cumulative stock net-flow series |
| `FlowOptionsLeaderboard(ctx, ...FlowOption)` | Cross-symbol option-flow leaderboard |
| `FlowOptionsOutliers(ctx, ...FlowOption)` | Cross-symbol option-flow outliers |
| `FlowStocksLeaderboard(ctx, ...FlowOption)` | Cross-symbol stock-flow leaderboard |
| `FlowStocksOutliers(ctx, ...FlowOption)` | Cross-symbol stock-flow outliers |
| `FlowDealerPremium(ctx, symbol, ...FlowOption)` | Net dealer option premium paid/received over a rolling window |
| `FlowStockBars(ctx, symbol, resolution, ...BarsOption)` | OHLCV-style intraday stock-flow bars (1s/1m/5m/15m/30m/1h/4h) |

### Zero-DTE Flow (intraday 0DTE, simulation-aware)

Live same-day-expiry flow analytics. Each has a typed `*Typed` variant.

| Method | Description | Plan |
|---|---|---|
| `FlowZeroDteSnapshot(ctx, symbol)` | 0DTE exposure snapshot + live flow direction | Growth+ |
| `FlowZeroDteSeries(ctx, symbol, ...ZeroDteFlowOption)` | Intraday 0DTE GEX/DEX/vex/pin time series (`WithZeroDteFlowBar`, `WithZeroDteFlowMinutes`) | Growth+ |
| `FlowZeroDteHedgeFlow(ctx, symbol, ...ZeroDteFlowOption)` | Estimated dealer hedging flow by side (`WithZeroDteFlowSide`) | Growth+ |
| `FlowZeroDteHeatmap(ctx, symbol, ...ZeroDteFlowOption)` | Strike × time heatmap of gex/dex/vex/chex/oi/signed_flow (`WithZeroDteFlowMetric`, `WithZeroDteFlowMode`) | Alpha+ |
| `FlowZeroDteStrikeFlow(ctx, symbol, ...ZeroDteFlowOption)` | Per-strike intraday 0DTE signed flow | Alpha+ |

### Strategy Signals

One decision-style endpoint per strategy. All return the shared typed
`*StrategyDecisionResponse` envelope (verdict, conviction, rationale, suggested
structure, risk). Tunable via `WithStrategyExpiry`, `WithStrategyMinOpenInterest`,
`WithStrategyWingWidth`, `WithStrategyTargetShortDelta`, `WithStrategyMaxWidth`,
`WithStrategyMinCredit`, `WithStrategyTargetDelta`, `WithStrategyStructure`,
`WithStrategyExcludeEarningsBeforeExpiry`.

| Method | Description | Plan |
|---|---|---|
| `StrategyFlowAnomaly(ctx, symbol, ...)` | Unusual options-flow anomaly read | Growth+ |
| `StrategyExpiryPositioning(ctx, symbol, ...)` | Expiry-positioning / pinning structure | Basic+ |
| `StrategyZeroDte(ctx, symbol, ...)` | 0DTE intraday strategy signal | Growth+ (+0DTE) |
| `StrategyDealerRegime(ctx, symbol, ...)` | Dealer gamma/vanna regime call | Growth+ |
| `StrategyVolCarry(ctx, symbol, ...)` | Vol-carry / short-premium harvest | Alpha+ |
| `StrategyYieldEnhancement(ctx, symbol, ...)` | Covered-call / put-write yield structure | Growth+ |
| `StrategySurfaceAnomaly(ctx, symbol, ...)` | Vol-surface mispricing / arbitrage | Alpha+ |
| `StrategySkew(ctx, symbol, ...)` | Skew steepness / risk-reversal signal | Growth+ |
| `StrategyTermStructure(ctx, symbol)` | Term-structure (contango/backwardation) signal | Growth+ |
| `StrategyTailPricing(ctx, symbol, ...)` | Tail / convexity pricing signal | Growth+ |

### Earnings Analytics

| Method | Description | Plan |
|---|---|---|
| `EarningsCalendar(ctx, ...EarningsCalendarOption)` | Upcoming earnings calendar (`WithEarningsCalendarDays`, `...Symbols`, `...Importance`) | Growth+ |
| `EarningsExpectedMove(ctx, symbol)` | Straddle-implied expected move into earnings | Growth+ |
| `EarningsHistory(ctx, symbol, ...EarningsHistoryOption)` | Historical earnings moves vs implied (`WithEarningsHistoryLimit`) | Growth+ |
| `EarningsIvCrush(ctx, symbol)` | Pre/post-earnings IV-crush analytics | Growth+ |
| `EarningsVrp(ctx, symbol)` | Earnings variance risk premium | Alpha+ |
| `EarningsDealerPositioning(ctx, symbol)` | Dealer positioning into the print | Alpha+ |
| `EarningsStrategies(ctx, symbol)` | Suggested earnings option structures | Alpha+ |
| `EarningsScreener(ctx, ...EarningsScreenerOption)` | Rank upcoming earnings by edge (`WithEarningsScreenerSort`, `...Limit`, `...Days`, `...MinImportance`) | Alpha+ |

### Multi-Leg Structures (pure-math, POST)

| Method | Description | Plan |
|---|---|---|
| `StructurePnl(ctx, StructurePnlRequest)` | Payoff/P&L curve for an arbitrary multi-leg structure | Basic+ |
| `StructureGreeks(ctx, StructureGreeksRequest)` | Aggregate BSM greeks for a multi-leg structure | Basic+ |

### Market Data

| Method | Description | Plan |
|---|---|---|
| `StockQuote(ctx, ticker)` | Live stock quote (bid/ask/mid/last) | Free+ |
| `OptionQuote(ctx, ticker, ...OptionQuoteOption)` | Option quotes with greeks | Growth+ |
| `StockSummary(ctx, symbol)` | Comprehensive stock summary | Free+ |
| `Surface(ctx, symbol)` | Volatility surface grid | Public |
| `SurfaceSvi(ctx, symbol)` | Calibrated SVI surface parameters (raw SVI a/b/rho/m/sigma per slice) | Alpha+ |

### Historical Data

| Method | Description | Plan |
|---|---|---|
| `HistoricalStockQuote(ctx, ticker, date, time...)` | Minute-by-minute stock quotes | Free+ |
| `HistoricalOptionQuote(ctx, ticker, date, ...HistOptOption)` | Minute-by-minute option quotes | Free+ |

### Pricing and Sizing

| Method | Description | Plan |
|---|---|---|
| `Greeks(ctx, GreeksParams)` | Full BSM greeks (first, second, third order) | Free+ |
| `IV(ctx, IVParams)` | Implied volatility from market price | Free+ |
| `Kelly(ctx, KellyParams)` | Kelly criterion optimal position size | Growth+ |

### Volatility Analytics

| Method | Description | Plan |
|---|---|---|
| `Volatility(ctx, symbol)` | Comprehensive volatility analysis | Growth+ |
| `AdvVolatility(ctx, symbol)` | SVI parameters, variance surface, arbitrage detection | Alpha+ |
| `Vrp(ctx, symbol, ...VrpOption)` | Variance risk premium analytics — IV vs RV spread, gamma/vanna conditioning, strategy scores. Returns typed `*VrpResponse` with nested `Vrp.ZScore`, `Regime.NetGex`, `GexConditioned.HarvestScore`, `Directional.DownsideVrp`/`UpsideVrp`. `WithVrpDate` requests a historical session. | Alpha+ |
| `VrpHistory(ctx, symbol, ...VrpHistoryOption)` | Variance-risk-premium time series (`WithVrpHistoryDays`) | Alpha+ |
| `ExpectedMove(ctx, symbol, ...ExpectedMoveOption)` | Straddle-implied expected move (`WithExpectedMoveExpiry`) | Basic+ |
| `Liquidity(ctx, symbol)` | Options liquidity profile — spreads, depth, volume/OI quality | Growth+ |
| `SkewTerm(ctx, symbol)` | Skew + term-structure grid (25-delta risk reversals across expiries) | Growth+ |
| `SpotVolCorrelation(ctx, symbol)` | Spot-vol correlation / leverage effect estimate | Growth+ |
| `Dispersion(ctx, index, symbols, ...DispersionOption)` | Index vs single-name dispersion / correlation trade analytics (`WithDispersionWeights`, `WithDispersionHorizonDays`) | Alpha+ |
| `RealizedVolatility(ctx, symbol)` | Range-based realized vol estimators (close-to-close, Parkinson, Garman-Klass, Rogers-Satchell, Yang-Zhang) over 10/20/30-day windows | Alpha+ |
| `VolatilityForecast(ctx, symbol, ...VolatilityForecastOption)` | Conditional vol forecasts — EWMA, HAR-RV, and GARCH(1,1) MLE term structure (`WithForecastDist`) | Alpha+ |

### Macro and Universe

| Method | Description | Plan |
|---|---|---|
| `VixState(ctx)` | VIX regime state — level, term structure, percentile, contango/backwardation | Growth+ |
| `Universe(ctx, ...UniverseOption)` | Ranked tradable universe snapshot (`WithUniverseSort`, `WithUniverseLimit`) | Public |

### Reference Data

| Method | Description | Plan |
|---|---|---|
| `Screener(ctx, ScreenerRequest)` | Live options screener — filter/rank by GEX, VRP, IV, greeks, harvest score, custom formulas | Growth+ |
| `ScreenerFields(ctx)` | Discoverable list of screener fields + operators (build queries dynamically) | Free+ |
| `Tickers(ctx)` | All available stock tickers | Free+ |
| `Options(ctx, ticker)` | Option chain metadata (expirations + strikes) | Free+ |
| `Symbols(ctx)` | Currently queried symbols with live data | Free+ |

### Account and System

| Method | Description | Plan |
|---|---|---|
| `Account(ctx)` | Account info and quota usage | Free+ |
| `Health(ctx)` | API health check | Public |

## Futures (CME)

FlashAlpha serves the full options-analytics stack for **CME futures** across six complexes - equity index (`ES=F`, `NQ=F`, `RTY=F`, `YM=F`, `MES=F`, `MNQ=F`), metals (`GC=F` gold, `SI=F` silver), energy (`CL=F` crude oil, `NG=F` natural gas), the Treasury curve (`ZT=F`, `ZF=F`, `ZN=F`, `TN=F`, `ZB=F`, `UB=F`), grains (`ZC=F` corn, `ZS=F` soybeans, `ZW=F` wheat) and crypto (`BTC=F` bitcoin). Options-on-futures are priced with **Black-76** (forward-priced) and each root carries its own CME contract multiplier, so notionals and dollar gamma are in real dollars. Note the quote conventions: Treasuries are quoted in points of par and grains in cents, so their multipliers are the contract size divided by 100. Everything that works for an equity works for futures: gamma exposure (GEX), DEX, VEX, CHEX, key levels, max pain, the IV surface, exposure summary, narrative, and live flow.

```go
// Gamma exposure for the E-mini S&P 500 future
gex, err := client.Gex(ctx, "ES=F")
if err != nil {
    log.Fatal(err)
}
fmt.Println(gex)
```

Use the `=F` suffix - bare `ES`/`NQ` are equities, not futures. In raw REST paths URL-encode the `=` as `%3D` (e.g. `GET /v1/exposure/gex/GC%3DF`); SDK methods take the plain string `"GC=F"`. Futures symbols require the Growth plan or higher. Historical replay for futures is coming; live analytics are available now.

## Functional Options

Optional parameters use the functional options pattern:

```go
// Gex with expiration filter and minimum open interest
gex, err := client.Gex(ctx, "SPY",
    flashalpha.WithExpiration("2025-12-19"),
    flashalpha.WithMinOI(500),
)

// Delta exposure filtered to one expiry
dex, err := client.Dex(ctx, "QQQ",
    flashalpha.WithDexExpiration("2025-12-19"),
)

// 0DTE analytics with custom strike range
dte, err := client.ZeroDte(ctx, "SPY",
    flashalpha.WithStrikeRange(0.05),
)

// Option quote filtered by expiry, strike, and type
oq, err := client.OptionQuote(ctx, "SPY",
    flashalpha.WithOptionExpiry("2025-12-19"),
    flashalpha.WithStrike(450.0),
    flashalpha.WithOptionType("call"),
)
```

## Pricing Parameters

The `Greeks`, `IV`, and `Kelly` endpoints accept parameter structs:

```go
// Greeks
result, err := client.Greeks(ctx, flashalpha.GreeksParams{
    Spot:   450.0,
    Strike: 455.0,
    DTE:    30.0,   // days to expiration
    Sigma:  0.20,   // annualized implied volatility
    Type:   "call", // "call" or "put" (default: "call")
    // R and Q are optional *float64 pointers
})

// Implied Volatility
iv, err := client.IV(ctx, flashalpha.IVParams{
    Spot:   450.0,
    Strike: 450.0,
    DTE:    30.0,
    Price:  10.5,  // market price of the option
    Type:   "call",
})

// Kelly Criterion
kelly, err := client.Kelly(ctx, flashalpha.KellyParams{
    Spot:    450.0,
    Strike:  460.0,
    DTE:     14.0,
    Sigma:   0.20,
    Premium: 3.50,
    Mu:      0.08,  // expected annual return of the underlying
    Type:    "call",
})
```

## Error Handling

All errors implement the `error` interface. Use type assertions to access
structured error data:

```go
result, err := client.Gex(ctx, "SPY")
if err != nil {
    switch e := err.(type) {
    case *flashalpha.AuthenticationError:
        // HTTP 401 — invalid or missing API key
        fmt.Println("auth error:", e.Message)
    case *flashalpha.TierRestrictedError:
        // HTTP 403 — endpoint requires a higher plan
        fmt.Printf("need %s plan, have %s\n", e.RequiredPlan, e.CurrentPlan)
    case *flashalpha.NotFoundError:
        // HTTP 404 — symbol or resource not found
        fmt.Println("not found:", e.Message)
    case *flashalpha.RateLimitError:
        // HTTP 429 — rate limit exceeded
        fmt.Printf("rate limited, retry after %d seconds\n", e.RetryAfter)
    case *flashalpha.ServerError:
        // HTTP 5xx — API-side error
        fmt.Println("server error:", e.StatusCode)
    case *flashalpha.APIError:
        // any other non-200 status
        fmt.Printf("api error %d: %s\n", e.StatusCode, e.Message)
    default:
        fmt.Println("unexpected error:", err)
    }
}
```

### Error Types

| Type | HTTP Status | Description |
|---|---|---|
| `*AuthenticationError` | 401 | Invalid or missing API key |
| `*TierRestrictedError` | 403 | Endpoint requires a higher subscription tier |
| `*NotFoundError` | 404 | Symbol or resource not found |
| `*RateLimitError` | 429 | Request rate limit exceeded |
| `*ServerError` | 5xx | Internal API error |
| `*APIError` | other | Catch-all for any other non-200 status |

## Running Tests

Unit tests use only the standard library and require no API key:

```sh
go test ./...
```

Integration tests hit the live API and require a key:

```sh
FLASHALPHA_API_KEY=your_key go test -tags integration ./...
```

## License

MIT. See [LICENSE](LICENSE).

## Other SDKs

| Language | Package | Repository |
|----------|---------|------------|
| Python | `pip install flashalpha` | [flashalpha-python](https://github.com/FlashAlpha-lab/flashalpha-python) |
| JavaScript | `npm i flashalpha` | [flashalpha-js](https://github.com/FlashAlpha-lab/flashalpha-js) |
| .NET | `dotnet add package FlashAlpha` | [flashalpha-dotnet](https://github.com/FlashAlpha-lab/flashalpha-dotnet) |
| Java | Maven Central | [flashalpha-java](https://github.com/FlashAlpha-lab/flashalpha-java) |
| MCP | Claude / LLM tool server | [flashalpha-mcp](https://github.com/FlashAlpha-lab/flashalpha-mcp) |

## Links

- [FlashAlpha](https://flashalpha.com) — API keys, docs, pricing
- [API Documentation](https://flashalpha.com/docs)
- [Examples](https://github.com/FlashAlpha-lab/flashalpha-examples) — runnable tutorials
- [GEX Explained](https://github.com/FlashAlpha-lab/gex-explained) — gamma exposure theory and code
- [0DTE Options Analytics](https://github.com/FlashAlpha-lab/0dte-options-analytics) — 0DTE pin risk, expected move, dealer hedging
- [Volatility Surface Python](https://github.com/FlashAlpha-lab/volatility-surface-python) — SVI calibration, variance swap, skew analysis
- [Awesome Options Analytics](https://github.com/FlashAlpha-lab/awesome-options-analytics) — curated resource list

## What the paid tiers unlock

The free tier covers single-expiry GEX on equities, key levels, the BSM Greeks/IV
calculator and stock quotes. Paid tiers add:

- **DEX, VEX (vanna) and CHEX (charm) exposure, plus max pain** — from the **Basic tier**
  ($79/mo), with ETF and index symbols.
- **Full-chain GEX, 0DTE and flow analytics** — from the **Growth tier** ($299/mo).
- **Point-in-time replay since 2017, SVI vol surfaces, VRP analytics, higher-order Greeks**,
  uncached and unlimited — the **Alpha tier** ($1,499/mo). FlashAlpha is one of the only
  public APIs publishing aggregate vanna and charm exposure across the full universe, with
  no look-ahead and no training-serving skew.

Built for quants, prop desks, and vol funds. See the full picture and get a key:
**[flashalpha.com/for-quant-teams](https://flashalpha.com/for-quant-teams?utm_source=github&utm_medium=readme&utm_campaign=repo-flashalpha-go)**
