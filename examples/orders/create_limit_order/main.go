// Example: Creating a limit order
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

	// Create a limit buy order
	marketIndex := int16(0)            // ETH-USD perp
	size := int64(1000000)             // 0.01 ETH (scaled)
	price := uint32(3000_000000)       // $3000 (scaled) - price is in 6 decimals
	isBuy := true
	expiry := time.Now().Add(24 * time.Hour).UnixMilli()

	opts := &types.TransactOpts{
		Nonce: types.NewInt64(-1),
	}

	txInfo, err := signerClient.CreateLimitOrder(marketIndex, size, price, isBuy, expiry, opts)
	if err != nil {
		log.Fatalf("Failed to create limit order: %v", err)
	}

	fmt.Printf("Limit order created!\n")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Market: %d\n", marketIndex)
	fmt.Printf("  Side: %s\n", map[bool]string{true: "BUY", false: "SELL"}[isBuy])
	fmt.Printf("  Size: %d\n", size)
	fmt.Printf("  Price: %d\n", price)
	fmt.Printf("  Expiry: %s\n", time.UnixMilli(expiry).Format(time.RFC3339))

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit order: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
