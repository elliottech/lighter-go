// Example: Creating a take-profit order
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/examples"
	"github.com/elliottech/lighter-go/types"
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

	// Create a take-profit order
	// This is typically used to lock in profits on an existing position
	marketIndex := int16(0)          // ETH-USD perp
	size := int64(1000000)           // 0.01 ETH (scaled)
	triggerPrice := uint32(3500_000000) // Trigger when price rises to $3500 (6 decimals)
	isBuy := false                   // Sell when triggered (closing a long position)
	expiry := time.Now().Add(7 * 24 * time.Hour).UnixMilli()

	nonce := int64(-1)
	opts := &types.TransactOpts{
		Nonce: &nonce,
	}

	txInfo, err := signerClient.CreateTakeProfitOrder(marketIndex, size, triggerPrice, isBuy, expiry, opts)
	if err != nil {
		log.Fatalf("Failed to create take-profit order: %v", err)
	}

	fmt.Printf("Take-profit order created!\n")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Market: %d\n", marketIndex)
	fmt.Printf("  Size: %d\n", size)
	fmt.Printf("  Trigger Price: %d (triggers when price >= this)\n", triggerPrice)
	fmt.Printf("  Side when triggered: SELL\n")
	fmt.Printf("  Expiry: %s\n", time.UnixMilli(expiry).Format(time.RFC3339))

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit order: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
