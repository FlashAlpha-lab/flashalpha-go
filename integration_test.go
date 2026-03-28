//go:build integration

package flashalpha_test

import (
	"context"
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

// keys returns the top-level keys of a map for logging.
func keys(m map[string]interface{}) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
