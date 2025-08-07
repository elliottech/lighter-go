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
	fmt.Printf("🔍 Debugging Signature Issues for Update Leverage\n")
	fmt.Printf("=================================================\n\n")

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

	fmt.Printf("📋 Configuration Check:\n")
	fmt.Printf("  🌐 Base URL: %s\n", config.BaseURL)
	fmt.Printf("  🏦 Account Index: %d\n", config.AccountIndex)
	fmt.Printf("  🔑 API Key Index: %d\n", config.ApiKeyIndex)
	fmt.Printf("  🔗 Chain ID: %d\n", config.ChainID)
	fmt.Printf("  🎯 Market Index: %d\n", config.MarketIndex)
	fmt.Printf("  📊 Initial Margin Fraction: %d\n", config.InitialMarginFraction)
	fmt.Printf("  🔄 Margin Mode: %s\n", getMarginModeString(config.MarginMode))
	
	// Check private key format
	fmt.Printf("\n🔐 Private Key Check:\n")
	privateKeyLength := len(config.PrivateKey)
	fmt.Printf("  📏 Private Key Length: %d characters\n", privateKeyLength)
	
	// Remove 0x prefix if present for length calculation
	cleanPrivateKey := config.PrivateKey
	if len(cleanPrivateKey) >= 2 && cleanPrivateKey[:2] == "0x" {
		cleanPrivateKey = cleanPrivateKey[2:]
		fmt.Printf("  ✅ Has 0x prefix (will be removed)\n")
	} else {
		fmt.Printf("  ⚠️  No 0x prefix\n")
	}
	
	fmt.Printf("  📏 Clean Key Length: %d characters\n", len(cleanPrivateKey))
	fmt.Printf("  🎯 Expected Length: 80 characters (40 bytes hex)\n")
	
	if len(cleanPrivateKey) != 80 {
		fmt.Printf("  ❌ ISSUE: Private key should be exactly 80 hex characters (40 bytes)\n")
		fmt.Printf("  💡 Your key is %d characters, expected 80\n", len(cleanPrivateKey))
		if len(cleanPrivateKey) < 80 {
			fmt.Printf("  💡 Key might be truncated or missing characters\n")
		} else {
			fmt.Printf("  💡 Key might have extra characters\n")
		}
		return
	} else {
		fmt.Printf("  ✅ Private key length is correct\n")
	}

	// Test hex decoding
	fmt.Printf("\n🧮 Hex Decoding Test:\n")
	for i, char := range cleanPrivateKey {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			fmt.Printf("  ❌ Invalid hex character '%c' at position %d\n", char, i)
			fmt.Printf("  💡 Private key must contain only hex characters (0-9, a-f, A-F)\n")
			return
		}
	}
	fmt.Printf("  ✅ All characters are valid hex\n")

	// Create HTTP client
	httpClient := client.NewHTTPClient(config.BaseURL)
	if httpClient == nil {
		log.Fatalf("Failed to create HTTP client")
	}

	// Try to create transaction client - this is where signature issues often start
	fmt.Printf("\n🔧 Transaction Client Creation:\n")
	txClient, err := client.NewTxClient(httpClient, config.PrivateKey, config.AccountIndex, config.ApiKeyIndex, config.ChainID)
	if err != nil {
		fmt.Printf("  ❌ ISSUE: Failed to create transaction client: %v\n", err)
		fmt.Printf("  💡 This suggests the private key format is incorrect\n")
		fmt.Printf("  💡 Common fixes:\n")
		fmt.Printf("     - Ensure private key is exactly 80 hex characters\n")
		fmt.Printf("     - Remove any extra spaces or newlines\n")
		fmt.Printf("     - Check if key is from the correct network/format\n")
		return
	}
	fmt.Printf("  ✅ Transaction client created successfully\n")

	// Test transaction creation (this is where most signature issues occur)
	fmt.Printf("\n🔨 Transaction Creation Test:\n")
	updateLeverageReq := &types.UpdateLeverageTxReq{
		MarketIndex:           config.MarketIndex,
		InitialMarginFraction: config.InitialMarginFraction,
		MarginMode:            config.MarginMode,
	}

	// Get the signed transaction
	txInfo, err := txClient.GetUpdateLeverageTransaction(updateLeverageReq, nil)
	if err != nil {
		fmt.Printf("  ❌ ISSUE: Failed to create update leverage transaction: %v\n", err)
		fmt.Printf("  💡 This could be:\n")
		fmt.Printf("     - Nonce retrieval failure (check network connectivity)\n")
		fmt.Printf("     - Parameter validation failure\n")
		fmt.Printf("     - Internal signing error\n")
		return
	}
	fmt.Printf("  ✅ Transaction created successfully\n")
	
	// Display transaction details for verification
	fmt.Printf("\n📝 Transaction Details:\n")
	fmt.Printf("  🏦 Account Index: %d\n", txInfo.AccountIndex)
	fmt.Printf("  🔑 API Key Index: %d\n", txInfo.ApiKeyIndex)
	fmt.Printf("  🎯 Market Index: %d\n", txInfo.MarketIndex)
	fmt.Printf("  📊 Initial Margin Fraction: %d\n", txInfo.InitialMarginFraction)
	fmt.Printf("  🔄 Margin Mode: %d\n", txInfo.MarginMode)
	fmt.Printf("  ⏰ Expired At: %d\n", txInfo.ExpiredAt)
	fmt.Printf("  🔢 Nonce: %d\n", txInfo.Nonce)
	fmt.Printf("  📝 Signed Hash: %s\n", txInfo.SignedHash)
	fmt.Printf("  ✍️  Signature Length: %d bytes\n", len(txInfo.Sig))

	// Test sending the transaction
	fmt.Printf("\n📤 Transaction Sending Test:\n")
	fmt.Printf("  🔄 Attempting to send transaction...\n")
	
	txHash, err := httpClient.SendRawTx(txInfo)
	if err != nil {
		fmt.Printf("  ❌ ISSUE: Failed to send transaction: %v\n", err)
		
		// Provide specific guidance based on error
		if err.Error() == `{"code":21120,"message":"invalid signature"}` {
			fmt.Printf("\n🩺 Signature Error Diagnosis:\n")
			fmt.Printf("  🔍 This specific error (21120) suggests:\n")
			fmt.Printf("     1. Private key doesn't match the account index\n")
			fmt.Printf("     2. Wrong chain ID for the network\n")
			fmt.Printf("     3. Account index or API key index mismatch\n")
			fmt.Printf("     4. Clock synchronization issues\n")
			fmt.Printf("\n💡 Troubleshooting Steps:\n")
			fmt.Printf("  1. ✅ Verify account index matches your private key\n")
			fmt.Printf("  2. ✅ Check chain ID matches the network (current: %d)\n", config.ChainID)
			fmt.Printf("  3. ✅ Ensure API key index is correct (current: %d)\n", config.ApiKeyIndex)
			fmt.Printf("  4. ✅ Check if account exists on this network\n")
			fmt.Printf("  5. ✅ Verify private key corresponds to this account\n")
			fmt.Printf("  6. ✅ Ensure system clock is synchronized\n")
			fmt.Printf("\n🔧 Quick Fixes to Try:\n")
			fmt.Printf("  • Try chain ID 1 (mainnet) or 421614 (testnet)\n")
			fmt.Printf("  • Try API key index 0 if you're unsure\n")
			fmt.Printf("  • Double-check account index from your wallet\n")
			fmt.Printf("  • Verify private key is from the same network\n")
		}
		return
	}
	
	fmt.Printf("  ✅ Transaction sent successfully!\n")
	fmt.Printf("  📝 Transaction Hash: %s\n", txHash)
	fmt.Printf("\n🎉 All signature checks passed! Your configuration is correct.\n")
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