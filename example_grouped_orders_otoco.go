package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Parse required environment variables
	cfg, err := parseEnvConfig()
	if err != nil {
		log.Fatalf("Failed to parse environment configuration: %v", err)
	}

	fmt.Printf("🔄 Config: %+v\n", cfg.publicSummary())

	// Create HTTP client
	httpClient := client.NewHTTPClient(cfg.BaseURL)
	if httpClient == nil {
		log.Fatalf("Failed to create HTTP client")
	}

	// Create transaction client
	txClient, err := client.NewTxClient(httpClient, cfg.PrivateKey, cfg.AccountIndex, cfg.ApiKeyIndex, cfg.ChainID)
	if err != nil {
		log.Fatalf("Failed to create transaction client: %v", err)
	}

	// Compute expiry for orders (in ms since epoch)
	var expiry int64 = 0
	if cfg.MainOrderType == "LIMIT" { // GoodTillTime requires non-nil expiry
		expiry = time.Now().Add(time.Duration(cfg.OrderTTLMinutes) * time.Minute).UnixMilli()
	}
	childExpiry := time.Now().Add(time.Duration(cfg.OrderTTLMinutes) * time.Minute).UnixMilli()

	// Build primary (entry) order
	entryOrder := &types.CreateOrderTxReq{
		MarketIndex:      cfg.MarketIndex,
		ClientOrderIndex: txtypes.NilClientOrderIndex,
		BaseAmount:       cfg.MainAmount,
		Price:            cfg.MainPrice,
		IsAsk:            boolToUint8(cfg.MainIsAsk),
		Type:             cfg.mainOrderTypeConst(),
		TimeInForce:      cfg.mainTIFConst(),
		ReduceOnly:       0,
		TriggerPrice:     txtypes.NilOrderTriggerPrice,
		OrderExpiry:      expiry,
	}

	// Child orders must be opposite side, IOC, non-nil expiry, and base amount must be Nil for OTOCO
	oppositeIsAsk := boolToUint8(!cfg.MainIsAsk)

	// Take Profit (market on trigger)
	tpOrder := &types.CreateOrderTxReq{
		MarketIndex:      cfg.MarketIndex,
		ClientOrderIndex: txtypes.NilClientOrderIndex,
		BaseAmount:       txtypes.NilOrderBaseAmount,
		Price:            nonZeroPrice(cfg.TPTrigger),
		IsAsk:            oppositeIsAsk,
		Type:             txtypes.TakeProfitOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     cfg.TPTrigger,
		OrderExpiry:      childExpiry,
	}

	// Stop Loss (market on trigger)
	slOrder := &types.CreateOrderTxReq{
		MarketIndex:      cfg.MarketIndex,
		ClientOrderIndex: txtypes.NilClientOrderIndex,
		BaseAmount:       txtypes.NilOrderBaseAmount,
		Price:            nonZeroPrice(cfg.SLTrigger),
		IsAsk:            oppositeIsAsk,
		Type:             txtypes.StopLossOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     cfg.SLTrigger,
		OrderExpiry:      childExpiry,
	}

	// Create grouped request (OTOCO)
	groupedReq := &types.CreateGroupedOrdersTxReq{
		GroupingType: txtypes.GroupingTypeOneTriggersAOneCancelsTheOther,
		Orders:       []*types.CreateOrderTxReq{entryOrder, tpOrder, slOrder},
	}

	// Orders:       []*types.CreateOrderTxReq{entryOrder, tpOrder, slOrder},
	// Get the signed transaction
	txInfo, err := txClient.GetCreateGroupedOrdersTransaction(groupedReq, nil)
	if err != nil {
		log.Fatalf("Failed to create grouped orders transaction: %v", err)
	}

	fmt.Printf("🔄 Transaction Info prepared.\n")

	// Send the transaction
	txHash, err := httpClient.SendRawTx(txInfo)
	if err != nil {
		log.Fatalf("Failed to send grouped orders transaction: %v", err)
	}

	fmt.Printf("✅ Grouped orders (OTOCO) submitted!\n")
	fmt.Printf("📝 Tx Hash: %s\n", txHash)
	fmt.Printf("🎯 Market: %d | Side: %s | Amount: %d | Price: %d | Type: %s\n", cfg.MarketIndex, sideString(cfg.MainIsAsk), cfg.MainAmount, cfg.MainPrice, cfg.MainOrderType)
	fmt.Printf("📈 TP trigger: %d | 📉 SL trigger: %d | Expiry (min): %d\n", cfg.TPTrigger, cfg.SLTrigger, cfg.OrderTTLMinutes)
}

// --- helpers & config ---

type config struct {
	BaseURL      string
	PrivateKey   string
	ChainID      uint32
	AccountIndex int64
	ApiKeyIndex  uint8

	MarketIndex   uint8
	MainPrice     uint32
	MainAmount    int64
	MainIsAsk     bool
	MainOrderType string // LIMIT or MARKET

	TPTrigger uint32
	SLTrigger uint32

	OrderTTLMinutes int64 // for GTT and child orders
}

func (c *config) publicSummary() map[string]any {
	return map[string]any{
		"BaseURL":         c.BaseURL,
		"ChainID":         c.ChainID,
		"AccountIndex":    c.AccountIndex,
		"ApiKeyIndex":     c.ApiKeyIndex,
		"MarketIndex":     c.MarketIndex,
		"MainPrice":       c.MainPrice,
		"MainAmount":      c.MainAmount,
		"MainSide":        sideString(c.MainIsAsk),
		"MainOrderType":   c.MainOrderType,
		"TPTrigger":       c.TPTrigger,
		"SLTrigger":       c.SLTrigger,
		"OrderTTLMinutes": c.OrderTTLMinutes,
	}
}

func (c *config) mainOrderTypeConst() uint8 {
	if strings.EqualFold(c.MainOrderType, "MARKET") {
		return txtypes.MarketOrder
	}
	return txtypes.LimitOrder
}

func (c *config) mainTIFConst() uint8 {
	if strings.EqualFold(c.MainOrderType, "MARKET") {
		return txtypes.ImmediateOrCancel
	}
	return txtypes.GoodTillTime
}

func sideString(isAsk bool) string {
	if isAsk {
		return "SELL"
	}
	return "BUY"
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func nonZeroPrice(trigger uint32) uint32 {
	if trigger == 0 {
		return 1 // satisfy min price validation if somehow trigger is 0
	}
	return trigger
}

func parseEnvConfig() (*config, error) {
	cfg := &config{}

	// Endpoint
	cfg.BaseURL = os.Getenv("LIGHTER_BASE_URL")
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("LIGHTER_BASE_URL is required")
	}

	// Private key
	cfg.PrivateKey = os.Getenv("LIGHTER_PRIVATE_KEY")
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("LIGHTER_PRIVATE_KEY is required")
	}

	// Chain ID
	chainIDStr := os.Getenv("LIGHTER_CHAIN_ID")
	if chainIDStr == "" {
		return nil, fmt.Errorf("LIGHTER_CHAIN_ID is required")
	}
	chainID, err := strconv.ParseUint(chainIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_CHAIN_ID: %v", err)
	}
	cfg.ChainID = uint32(chainID)

	// Account index
	accStr := os.Getenv("LIGHTER_ACCOUNT_INDEX")
	if accStr == "" {
		return nil, fmt.Errorf("LIGHTER_ACCOUNT_INDEX is required")
	}
	acc, err := strconv.ParseInt(accStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_ACCOUNT_INDEX: %v", err)
	}
	cfg.AccountIndex = acc

	// API key index
	apiIdxStr := os.Getenv("LIGHTER_API_KEY_INDEX")
	if apiIdxStr == "" {
		return nil, fmt.Errorf("LIGHTER_API_KEY_INDEX is required")
	}
	apiIdx, err := strconv.ParseUint(apiIdxStr, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_API_KEY_INDEX: %v", err)
	}
	cfg.ApiKeyIndex = uint8(apiIdx)

	// Market index
	marketStr := os.Getenv("LIGHTER_MARKET_INDEX")
	if marketStr == "" {
		return nil, fmt.Errorf("LIGHTER_MARKET_INDEX is required")
	}
	market, err := strconv.ParseUint(marketStr, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_MARKET_INDEX: %v", err)
	}
	cfg.MarketIndex = uint8(market)

	// Main price
	mainPriceStr := os.Getenv("LIGHTER_MAIN_PRICE")
	if mainPriceStr == "" {
		return nil, fmt.Errorf("LIGHTER_MAIN_PRICE is required")
	}
	mainPrice, err := strconv.ParseUint(mainPriceStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_MAIN_PRICE: %v", err)
	}
	cfg.MainPrice = uint32(mainPrice)

	// Main amount
	amountStr := os.Getenv("LIGHTER_MAIN_AMOUNT")
	if amountStr == "" {
		return nil, fmt.Errorf("LIGHTER_MAIN_AMOUNT is required")
	}
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_MAIN_AMOUNT: %v", err)
	}
	cfg.MainAmount = amount

	// Main side (BUY/SELL)
	sideStr := os.Getenv("LIGHTER_MAIN_SIDE")
	if sideStr == "" {
		return nil, fmt.Errorf("LIGHTER_MAIN_SIDE is required (BUY or SELL)")
	}
	switch strings.ToUpper(sideStr) {
	case "BUY":
		cfg.MainIsAsk = false
	case "SELL":
		cfg.MainIsAsk = true
	default:
		return nil, fmt.Errorf("invalid LIGHTER_MAIN_SIDE: must be BUY or SELL")
	}

	// Main order type (optional; default LIMIT)
	orderType := strings.ToUpper(strings.TrimSpace(os.Getenv("LIGHTER_MAIN_ORDER_TYPE")))
	if orderType == "" {
		orderType = "LIMIT"
	}
	if orderType != "LIMIT" && orderType != "MARKET" {
		return nil, fmt.Errorf("invalid LIGHTER_MAIN_ORDER_TYPE: must be LIMIT or MARKET")
	}
	cfg.MainOrderType = orderType

	// TP/SL trigger prices
	tpStr := os.Getenv("LIGHTER_TP_TRIGGER")
	if tpStr == "" {
		return nil, fmt.Errorf("LIGHTER_TP_TRIGGER is required")
	}
	tp, err := strconv.ParseUint(tpStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_TP_TRIGGER: %v", err)
	}
	cfg.TPTrigger = uint32(tp)

	slStr := os.Getenv("LIGHTER_SL_TRIGGER")
	if slStr == "" {
		return nil, fmt.Errorf("LIGHTER_SL_TRIGGER is required")
	}
	sl, err := strconv.ParseUint(slStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid LIGHTER_SL_TRIGGER: %v", err)
	}
	cfg.SLTrigger = uint32(sl)

	// TTL minutes (optional; default 7 days)
	ttlStr := os.Getenv("LIGHTER_ORDER_TTL_MINUTES")
	if ttlStr == "" {
		cfg.OrderTTLMinutes = 60 * 24 * 7 // 7 days
	} else {
		ttl, err := strconv.ParseInt(ttlStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid LIGHTER_ORDER_TTL_MINUTES: %v", err)
		}
		if ttl < 5 { // adhere to MinOrderExpiryPeriod (5 minutes)
			ttl = 5
		}
		cfg.OrderTTLMinutes = ttl
	}

	return cfg, nil
}
