// Example: strategy signals, earnings analytics, and multi-leg structure pricing.
//
// Run with:
//
//	FLASHALPHA_API_KEY=your_key go run ./examples/strategy_earnings_structures
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

func main() {
	apiKey := os.Getenv("FLASHALPHA_API_KEY")
	if apiKey == "" {
		log.Fatal("set FLASHALPHA_API_KEY")
	}
	client := flashalpha.NewClient(apiKey)
	ctx := context.Background()

	// 1) Strategy signal — one typed StrategyDecisionResponse for all 10 strategies.
	sig, err := client.StrategyVolCarry(ctx, "SPY",
		flashalpha.WithStrategyExpiry("2026-06-19"),
		flashalpha.WithStrategyMinCredit(0.50),
	)
	if err != nil {
		log.Fatalf("StrategyVolCarry: %v", err)
	}
	fmt.Printf("[strategy] %s decision=%s score=%d confidence=%.2f regime=%s\n",
		sig.Strategy, sig.Decision, sig.Score, sig.Confidence, sig.Regime)
	for _, why := range sig.Why {
		fmt.Printf("           why: %s\n", why)
	}

	// 2) Earnings — straddle-implied expected move into the next print.
	em, err := client.EarningsExpectedMoveTyped(ctx, "NVDA")
	if err != nil {
		log.Fatalf("EarningsExpectedMoveTyped: %v", err)
	}
	fmt.Printf("[earnings] %s earnings %s (in %d days)\n",
		em.Symbol, em.EarningsDate, em.DaysToEvent)
	if em.ExpectedMove != nil && em.ExpectedMove.EarningsImpliedPct != nil {
		fmt.Printf("           earnings-implied move: %.2f%%\n",
			*em.ExpectedMove.EarningsImpliedPct)
	}

	// 3) Multi-leg structure — at-expiry P&L for a call vertical (pure math).
	pnl, err := client.StructurePnlTyped(ctx, flashalpha.StructurePnlRequest{
		Legs: []flashalpha.StructurePnlLeg{
			{Action: "buy", Type: "call", Strike: 455, Premium: 6.20, Quantity: 1},
			{Action: "sell", Type: "call", Strike: 465, Premium: 2.10, Quantity: 1},
		},
	})
	if err != nil {
		log.Fatalf("StructurePnlTyped: %v", err)
	}
	fmt.Printf("[structure] breakevens=%v max_profit=%v max_loss=%v\n",
		pnl.Breakevens, deref(pnl.MaxProfit), deref(pnl.MaxLoss))

	// 4) Macro context — current VIX regime.
	vix, err := client.VixStateTyped(ctx)
	if err != nil {
		log.Fatalf("VixStateTyped: %v", err)
	}
	fmt.Printf("[macro] vix-state: %+v\n", vix)
}

func deref(p *float64) interface{} {
	if p == nil {
		return "unbounded"
	}
	return *p
}
