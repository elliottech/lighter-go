// Example: Getting active orders
package main

import (
	"fmt"
	"log"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/examples"
)

func main() {
	privateKey := examples.GetPrivateKey()
	if privateKey == "" {
		log.Fatal("LIGHTER_PRIVATE_KEY environment variable not set")
	}

	apiURL := examples.GetAPIURL()
	httpClient := http.NewFullClient(apiURL)

	signerClient, err := client.NewSignerClient(httpClient, privateKey, 1, 0, 0, nil)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	// Get all active orders across all markets
	fmt.Println("Fetching active orders for all markets...")
	orders, err := signerClient.GetOpenOrders(nil)
	if err != nil {
		log.Fatalf("Failed to get open orders: %v", err)
	}

	fmt.Printf("Found %d active orders\n\n", len(orders.Orders))

	if len(orders.Orders) == 0 {
		fmt.Println("No active orders")
	}

	for i, order := range orders.Orders {
		fmt.Printf("Order #%d:\n", i+1)
		fmt.Printf("  Index: %d\n", order.Index)
		fmt.Printf("  Market: %d\n", order.MarketIndex)
		fmt.Printf("  Side: %s\n", order.Side)
		fmt.Printf("  Type: %s\n", order.Type)
		fmt.Printf("  Price: %s\n", order.Price)
		fmt.Printf("  Size: %s\n", order.Size)
		fmt.Printf("  Filled Size: %s\n", order.FilledSize)
		fmt.Printf("  Remaining: %s\n", order.RemainingSize)
		fmt.Printf("  Status: %s\n", order.Status)
		fmt.Printf("  Created: %d\n", order.CreatedAt)
		fmt.Println()
	}

	// Get orders for a specific market (ETH-USD, market 0)
	fmt.Println("Fetching active orders for ETH-USD market only...")
	marketIndex := int16(0)
	ethOrders, err := signerClient.GetOpenOrders(&marketIndex)
	if err != nil {
		log.Fatalf("Failed to get ETH-USD orders: %v", err)
	}

	fmt.Printf("Found %d active orders for ETH-USD\n", len(ethOrders.Orders))
}
