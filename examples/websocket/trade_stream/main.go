// Example: Streaming trades via WebSocket
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
	defer client.Close()

	// Subscribe to ETH-USD trades
	marketIndex := int16(0)
	if err := client.SubscribeTrades(marketIndex); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	fmt.Printf("Subscribed to market %d trades\n", marketIndex)
	fmt.Println("Waiting for trades...")
	fmt.Println()

	// Process updates
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-client.TradeUpdates():
			for _, trade := range update.Trades {
				side := "BUY"
				if trade.Side == "sell" {
					side = "SELL"
				}
				fmt.Printf("Trade: %s %s @ %s (market %d)\n",
					side,
					trade.Size,
					trade.Price,
					trade.MarketIndex)
			}
		case err := <-client.Errors():
			fmt.Printf("Error: %v\n", err)
		}
	}
}
