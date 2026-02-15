// Example: Getting account information
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

	// Get account by index
	accountIndex := "0" // Replace with your account index
	accounts, err := httpClient.Account().GetAccount(api.QueryByIndex, accountIndex)
	if err != nil {
		log.Fatalf("Failed to get account: %v", err)
	}

	for _, acc := range accounts.Accounts {
		fmt.Printf("Account %d:\n", acc.Index)
		fmt.Printf("  Collateral Value: %s\n", acc.CollateralValue)
		fmt.Printf("  Position Value: %s\n", acc.PositionValue)
		fmt.Printf("  Portfolio Value: %s\n", acc.PortfolioValue)
		fmt.Printf("  Available Balance: %s\n", acc.AvailableBalance)
		fmt.Printf("  Max Withdrawable: %s\n", acc.MaxWithdrawable)
		fmt.Printf("  Initial Margin: %s\n", acc.InitialMargin)
		fmt.Printf("  Maintenance Margin: %s\n", acc.MaintenanceMargin)
		fmt.Printf("  Unrealized PnL: %s\n", acc.UnrealizedPnl)
		fmt.Printf("  Is Liquidatable: %v\n", acc.IsLiquidatable)

		if len(acc.Positions) > 0 {
			fmt.Printf("\n  Positions:\n")
			for _, pos := range acc.Positions {
				fmt.Printf("    Market %d: %s %s @ %s (PnL: %s)\n",
					pos.MarketIndex, pos.Side, pos.Size, pos.EntryPrice, pos.UnrealizedPnl)
			}
		}

		if len(acc.Assets) > 0 {
			fmt.Printf("\n  Assets:\n")
			for _, asset := range acc.Assets {
				fmt.Printf("    Asset %d: Balance=%s, Available=%s\n",
					asset.AssetIndex, asset.Balance, asset.AvailableBalance)
			}
		}
	}
}
