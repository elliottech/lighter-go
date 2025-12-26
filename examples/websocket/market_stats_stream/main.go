// Example: Streaming market statistics via WebSocket
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elliottech/lighter-go/client/ws"
	"github.com/elliottech/lighter-go/examples"
)

func main() {
	wsURL := examples.GetWSURL()

	// Create WebSocket client with default options
	opts := ws.DefaultOptions().
		WithOnConnect(func() {
			fmt.Println("Connected to WebSocket!")
		}).
		WithOnDisconnect(func(err error) {
			fmt.Printf("Disconnected: %v\n", err)
		})

	client := ws.NewClient(wsURL, opts)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Connect
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close() //nolint:errcheck // Cleanup on exit

	// Subscribe to all market stats
	if err := client.SubscribeAllMarketStats(); err != nil {
		log.Fatalf("Failed to subscribe to all market stats: %v", err)
	}
	fmt.Println("Subscribed to all market stats")
	fmt.Println("Waiting for updates...")
	fmt.Println()

	// Process updates
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-client.MarketStatsUpdates():
			if update.AllStats != nil {
				// All markets update
				for _, stats := range update.AllStats {
					printStats(&stats)
				}
			} else if update.Stats != nil {
				// Single market update
				printStats(update.Stats)
			}
		case err := <-client.Errors():
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func printStats(stats *ws.MarketStats) {
	fmt.Printf("Market %d:\n", stats.MarketIndex)
	fmt.Printf("  Last Price:    %s\n", stats.LastPrice)
	fmt.Printf("  Mark Price:    %s\n", stats.MarkPrice)
	fmt.Printf("  Index Price:   %s\n", stats.IndexPrice)
	fmt.Printf("  24h High:      %s\n", stats.High24h)
	fmt.Printf("  24h Low:       %s\n", stats.Low24h)
	fmt.Printf("  24h Volume:    %s\n", stats.Volume24h)
	fmt.Printf("  24h Change:    %s%%\n", stats.PriceChangePct)
	fmt.Printf("  Open Interest: %s\n", stats.OpenInterest)
	fmt.Printf("  Funding Rate:  %s\n", stats.FundingRate)
	fmt.Println()
}
