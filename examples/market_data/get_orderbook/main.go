// Example: Getting order book data
package main

import (
	"fmt"
	"log"

	"github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/examples"
	"github.com/elliottech/lighter-go/types/api"
)

func main() {
	apiURL := examples.GetAPIURL()
	httpClient := http.NewFullClient(apiURL)

	// Get order book for market 0 (ETH-USD)
	marketIndex := int16(0)
	orderBooks, err := httpClient.Order().GetOrderBooks(&marketIndex, api.MarketFilterAll)
	if err != nil {
		log.Fatalf("Failed to get order books: %v", err)
	}

	for _, ob := range orderBooks.OrderBooks {
		fmt.Printf("Order Book for Market %d:\n", ob.MarketIndex)
		fmt.Printf("  Timestamp: %d\n", ob.Timestamp)

		fmt.Printf("\n  Top 5 Bids:\n")
		for i, bid := range ob.Bids {
			if i >= 5 {
				break
			}
			fmt.Printf("    %s @ %s\n", bid.Size, bid.Price)
		}

		fmt.Printf("\n  Top 5 Asks:\n")
		for i, ask := range ob.Asks {
			if i >= 5 {
				break
			}
			fmt.Printf("    %s @ %s\n", ask.Size, ask.Price)
		}
	}
}
