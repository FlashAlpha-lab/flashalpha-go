# flashalpha-go

Official Go client for the [FlashAlpha](https://flashalpha.com) options analytics API.

FlashAlpha delivers institutional-grade options analytics including a **live
options screener** (filter/rank symbols by GEX, VRP, IV, greeks, harvest scores,
and custom formulas), gamma exposure (GEX), delta exposure (DEX), vanna and
charm exposure, volatility surfaces, 0DTE analytics, and Black-Scholes-Merton
pricing utilities — all via a simple REST API.

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
| `Dex(ctx, symbol, ...DexOption)` | Delta exposure by strike | Free+ |
| `Vex(ctx, symbol, ...VexOption)` | Vanna exposure by strike | Free+ |
| `Chex(ctx, symbol, ...ChexOption)` | Charm exposure by strike | Free+ |
| `ExposureLevels(ctx, symbol)` | Key support/resistance levels from options | Free+ |
| `ExposureSummary(ctx, symbol)` | Full GEX/DEX/VEX/CHEX + hedging summary | Growth+ |
| `Narrative(ctx, symbol)` | Verbal narrative analysis of exposure | Growth+ |
| `ZeroDte(ctx, symbol, ...ZeroDteOption)` | 0DTE regime, expected move, pin risk | Growth+ |
| `MaxPain(ctx, symbol, ...MaxPainOption)` | Max pain analysis with dealer alignment, pain curve, pin probability | Growth+ |
| `ExposureHistory(ctx, symbol, ...HistoryOption)` | Daily exposure trend snapshots | Growth+ |

### Market Data

| Method | Description | Plan |
|---|---|---|
| `StockQuote(ctx, ticker)` | Live stock quote (bid/ask/mid/last) | Free+ |
| `OptionQuote(ctx, ticker, ...OptionQuoteOption)` | Option quotes with greeks | Growth+ |
| `StockSummary(ctx, symbol)` | Comprehensive stock summary | Free+ |
| `Surface(ctx, symbol)` | Volatility surface grid | Public |

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

### Reference Data

| Method | Description | Plan |
|---|---|---|
| `Tickers(ctx)` | All available stock tickers | Free+ |
| `Options(ctx, ticker)` | Option chain metadata (expirations + strikes) | Free+ |
| `Symbols(ctx)` | Currently queried symbols with live data | Free+ |

### Account and System

| Method | Description | Plan |
|---|---|---|
| `Account(ctx)` | Account info and quota usage | Free+ |
| `Health(ctx)` | API health check | Public |

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

// Historical exposure with 30-day lookback
hist, err := client.ExposureHistory(ctx, "SPY",
    flashalpha.WithDays(30),
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
