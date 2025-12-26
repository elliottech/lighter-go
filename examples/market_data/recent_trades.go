// Example: Getting recent trades
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/examples"
)

func main() {
	apiURL := examples.GetAPIURL()
	httpClient := http.NewFullClient(apiURL)

	// Get recent trades for market 0 (ETH-USD)
	marketIndex := int16(0)
	limit := 20

	fmt.Printf("Fetching last %d trades for market %d...\n\n", limit, marketIndex)

	trades, err := httpClient.Order().GetRecentTrades(marketIndex, limit)
	if err != nil {
		log.Fatalf("Failed to get recent trades: %v", err)
	}

	fmt.Printf("Retrieved %d trades\n", len(trades.Trades))
	fmt.Println()

	fmt.Printf("%-20s %-6s %-15s %-15s\n", "Time", "Side", "Price", "Size")
	fmt.Println("--------------------------------------------------------")

	for _, trade := range trades.Trades {
		ts := time.UnixMilli(trade.Timestamp)
		fmt.Printf("%-20s %-6s %-15s %-15s\n",
			ts.Format("2006-01-02 15:04:05"),
			trade.Side,
			trade.Price,
			trade.Size)
	}

	fmt.Println()
	fmt.Printf("Total trades shown: %d\n", len(trades.Trades))
}
