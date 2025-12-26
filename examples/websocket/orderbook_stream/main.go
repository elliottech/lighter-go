// Example: Streaming order book updates via WebSocket
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

	// Subscribe to ETH-USD order book
	marketIndex := int16(0)
	if err := client.SubscribeOrderBook(marketIndex); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	fmt.Printf("Subscribed to market %d order book\n", marketIndex)

	// Process updates
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-client.OrderBookUpdates():
			if update.IsSnapshot {
				fmt.Printf("Received snapshot for market %d: %d bids, %d asks\n",
					update.MarketIndex,
					len(update.Snapshot.Bids),
					len(update.Snapshot.Asks))
			} else {
				fmt.Printf("Received delta for market %d: seq=%d, %d bid updates, %d ask updates\n",
					update.MarketIndex,
					update.Delta.Sequence,
					len(update.Delta.BidUpdates),
					len(update.Delta.AskUpdates))
			}

			// Print best bid/ask
			if update.State != nil {
				bestBid := update.State.GetBestBid()
				bestAsk := update.State.GetBestAsk()
				if bestBid != nil && bestAsk != nil {
					fmt.Printf("  Best Bid: %s @ %s | Best Ask: %s @ %s\n",
						bestBid.Size, bestBid.Price,
						bestAsk.Size, bestAsk.Price)
				}
			}
		case err := <-client.Errors():
			fmt.Printf("Error: %v\n", err)
		}
	}
}
