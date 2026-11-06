// Should you sell options today? Print the sell-window answer for one symbol.
//
//	FLASHALPHA_API_KEY=... go run ./examples/sell_window SPY
//
// Docs: https://flashalpha.com/docs/lab-api-strategies-sell-window
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	flashalpha "github.com/FlashAlpha-lab/flashalpha-go"
)

func main() {
	symbol := "SPY"
	if len(os.Args) > 1 {
		symbol = os.Args[1]
	}

	client := flashalpha.NewClient(os.Getenv("FLASHALPHA_API_KEY"))
	answer, err := client.SellWindow(context.Background(), symbol)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s session %s\n", answer.Symbol, answer.Session)
	if answer.LegsIndicative != nil && *answer.LegsIndicative {
		fmt.Println("Quotes are from before the 09:35 ET entry. Call again at or after 09:35 ET for the strikes to trade.")
	}
	for _, window := range answer.Windows {
		expiry := "no listed expiry"
		if window.Expiry != nil {
			expiry = *window.Expiry
		}
		fmt.Printf("%s: %s\n", expiry, window.Action)
		for _, trade := range window.Candidates {
			fmt.Printf("  %s, exit: %s\n", trade.Structure, trade.Exit)
			for _, leg := range trade.Legs {
				fmt.Printf("    %s %d %g %s\n", leg.Side, leg.Quantity, leg.Strike, leg.Type)
			}
			if trade.MidCredit != nil {
				fmt.Printf("    $%.2f credit at mid\n", *trade.MidCredit*100)
			}
		}
	}
}
