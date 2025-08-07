package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Parse required environment variables
	config, err := parseEnvConfig()
	if err != nil {
		log.Fatalf("Failed to parse environment configuration: %v", err)
	}

	fmt.Printf("🔄 Config: %+v\n", config)

	// Create HTTP client
	httpClient := client.NewHTTPClient(config.BaseURL)
	if httpClient == nil {
		log.Fatalf("Failed to create HTTP client")
	}

	// Create transaction client
	txClient, err := client.NewTxClient(httpClient, config.PrivateKey, config.AccountIndex, config.ApiKeyIndex, config.ChainID)
	if err != nil {
		log.Fatalf("Failed to create transaction client: %v", err)
	}

	// Create update leverage transaction request
	updateLeverageReq := &types.UpdateLeverageTxReq{
		MarketIndex:           config.MarketIndex,
		InitialMarginFraction: config.InitialMarginFraction,
		MarginMode:            config.MarginMode,
	}

	// Get the signed transaction
	txInfo, err := txClient.GetUpdateLeverageTransaction(updateLeverageReq, nil)
	if err != nil {
		log.Fatalf("Failed to create update leverage transaction: %v", err)
	}

	fmt.Printf("🔄 Transaction Info: %+v\n", txInfo)

	// Send the transaction
	txHash, err := httpClient.SendRawTx(txInfo)
	if err != nil {
		log.Fatalf("Failed to send update leverage transaction: %v", err)
	}

	fmt.Printf("✅ Update leverage transaction sent successfully!\n")
	fmt.Printf("📝 Transaction Hash: %s\n", txHash)
	fmt.Printf("🎯 Market Index: %d\n", config.MarketIndex)
	fmt.Printf("📊 Initial Margin Fraction: %d\n", config.InitialMarginFraction)
	fmt.Printf("🔄 Margin Mode: %s\n", getMarginModeString(config.MarginMode))
	fmt.Printf("🏦 Account Index: %d\n", config.AccountIndex)
	fmt.Printf("🔑 API Key Index: %d\n", config.ApiKeyIndex)
}

type Config struct {
	BaseURL               string
	PrivateKey            string
	ChainID               uint32
	AccountIndex          int64
	ApiKeyIndex           uint8
	MarketIndex           uint8
	InitialMarginFraction uint16
	MarginMode            uint8
}

func parseEnvConfig() (*Config, error) {
	config := &Config{}

	// Required: Base URL for the Lighter API
	config.BaseURL = os.Getenv("LIGHTER_BASE_URL")
	if config.BaseURL == "" {
		return nil, fmt.Errorf("LIGHTER_BASE_URL environment variable is required")
	}

	// Required: Private key for signing transactions
	config.PrivateKey = os.Getenv("LIGHTER_PRIVATE_KEY")
	if config.PrivateKey == "" {
		return nil, fmt.Errorf("LIGHTER_PRIVATE_KEY environment variable is required")
	}

	// Required: Chain ID
	chainIDStr := os.Getenv("LIGHTER_CHAIN_ID")
	if chainIDStr == "" {
		return nil, fmt.Errorf("LIGHTER_CHAIN_ID environment variable is required")
	}
	chainID, err := strconv.ParseUint(chainIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_CHAIN_ID: %v", err)
	}
	config.ChainID = uint32(chainID)

	// Required: Account index
	accountIndexStr := os.Getenv("LIGHTER_ACCOUNT_INDEX")
	if accountIndexStr == "" {
		return nil, fmt.Errorf("LIGHTER_ACCOUNT_INDEX environment variable is required")
	}
	accountIndex, err := strconv.ParseInt(accountIndexStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_ACCOUNT_INDEX: %v", err)
	}
	config.AccountIndex = accountIndex

	// Required: API key index
	apiKeyIndexStr := os.Getenv("LIGHTER_API_KEY_INDEX")
	if apiKeyIndexStr == "" {
		return nil, fmt.Errorf("LIGHTER_API_KEY_INDEX environment variable is required")
	}
	apiKeyIndex, err := strconv.ParseUint(apiKeyIndexStr, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_API_KEY_INDEX: %v", err)
	}
	config.ApiKeyIndex = uint8(apiKeyIndex)

	// Required: Market index (coin/market to update leverage for)
	marketIndexStr := os.Getenv("LIGHTER_MARKET_INDEX")
	if marketIndexStr == "" {
		return nil, fmt.Errorf("LIGHTER_MARKET_INDEX environment variable is required")
	}
	marketIndex, err := strconv.ParseUint(marketIndexStr, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_MARKET_INDEX: %v", err)
	}
	config.MarketIndex = uint8(marketIndex)

	// Required: Initial margin fraction (leverage setting)
	initialMarginFractionStr := os.Getenv("LIGHTER_INITIAL_MARGIN_FRACTION")
	if initialMarginFractionStr == "" {
		return nil, fmt.Errorf("LIGHTER_INITIAL_MARGIN_FRACTION environment variable is required")
	}
	initialMarginFraction, err := strconv.ParseUint(initialMarginFractionStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_INITIAL_MARGIN_FRACTION: %v", err)
	}
	config.InitialMarginFraction = uint16(initialMarginFraction)

	// Optional: Margin mode (defaults to CrossMargin)
	marginModeStr := os.Getenv("LIGHTER_MARGIN_MODE")
	if marginModeStr == "" {
		config.MarginMode = txtypes.CrossMargin // Default to cross margin
	} else {
		marginMode, err := strconv.ParseUint(marginModeStr, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid LIGHTER_MARGIN_MODE: %v", err)
		}
		if uint8(marginMode) != txtypes.CrossMargin && uint8(marginMode) != txtypes.IsolatedMargin {
			return nil, fmt.Errorf("invalid LIGHTER_MARGIN_MODE: must be %d (CrossMargin) or %d (IsolatedMargin)", txtypes.CrossMargin, txtypes.IsolatedMargin)
		}
		config.MarginMode = uint8(marginMode)
	}

	return config, nil
}

func getMarginModeString(marginMode uint8) string {
	switch marginMode {
	case txtypes.CrossMargin:
		return "Cross Margin"
	case txtypes.IsolatedMargin:
		return "Isolated Margin"
	default:
		return "Unknown"
	}
}
