package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// Market Data Streams (DEPRECATED - use LighterWebsocketClient instead)

// StreamOrderBook subscribes to order book updates for a market ID
// DEPRECATED: Use LighterWebsocketClient.Public().SubscribeOrderBook() instead
func (ws *WSClient) StreamOrderBook(ctx context.Context, marketId uint8, callback func(*WSOrderBookUpdate) error) error {
	handler := func(data []byte) error {
		var response struct {
			Type string             `json:"type"`
			Data *WSOrderBookUpdate `json:"data"`
		}
		
		if err := json.Unmarshal(data, &response); err != nil {
			return fmt.Errorf("failed to unmarshal order book update: %v", err)
		}
		
		if response.Type != ChannelOrderBook || response.Data == nil {
			return nil // Not an order book message
		}
		
		return callback(response.Data)
	}
	
	ws.AddHandler(MessageTypeOrderBook, handler)
	return ws.Subscribe(fmt.Sprintf("%s/%d", ChannelOrderBook, marketId), "")
}

// StreamTicker subscribes to ticker updates for a market ID
func (ws *WSClient) StreamTicker(ctx context.Context, marketId uint8, callback func(*WSTickerUpdate) error) error {
	handler := func(data []byte) error {
		var response struct {
			Type string           `json:"type"`
			Data *WSTickerUpdate  `json:"data"`
		}
		
		if err := json.Unmarshal(data, &response); err != nil {
			return fmt.Errorf("failed to unmarshal ticker update: %v", err)
		}
		
		if response.Type != ChannelTicker || response.Data == nil {
			return nil
		}
		
		return callback(response.Data)
	}
	
	ws.AddHandler(ChannelTicker, handler)  
	return ws.Subscribe(fmt.Sprintf("%s/%d", ChannelTicker, marketId), "")
}

// StreamTrades subscribes to trade updates for a market ID
func (ws *WSClient) StreamTrades(ctx context.Context, marketId uint8, callback func(*WSTradeUpdate) error) error {
	handler := func(data []byte) error {
		var response struct {
			Type string         `json:"type"`
			Data *WSTradeUpdate `json:"data"`
		}
		
		if err := json.Unmarshal(data, &response); err != nil {
			return fmt.Errorf("failed to unmarshal trade update: %v", err)
		}
		
		if response.Type != ChannelTrades || response.Data == nil {
			return nil
		}
		
		return callback(response.Data)
	}
	
	ws.AddHandler(ChannelTrades, handler)
	return ws.Subscribe(fmt.Sprintf("%s/%d", ChannelTrades, marketId), "")
}

// StreamMarkPrice subscribes to mark price updates for a market ID
func (ws *WSClient) StreamMarkPrice(ctx context.Context, marketId uint8, callback func(*WSTickerUpdate) error) error {
	handler := func(data []byte) error {
		var response struct {
			Type string           `json:"type"`
			Data *WSTickerUpdate  `json:"data"`
		}
		
		if err := json.Unmarshal(data, &response); err != nil {
			return fmt.Errorf("failed to unmarshal mark price update: %v", err)
		}
		
		if response.Type != ChannelMarkPrice || response.Data == nil {
			return nil
		}
		
		return callback(response.Data)
	}
	
	ws.AddHandler(ChannelMarkPrice, handler)
	return ws.Subscribe(fmt.Sprintf("%s/%d", ChannelMarkPrice, marketId), "")
}

// Account Data Streams

// Deprecated: StreamAccount is deprecated in favor of the new Bybit-style architecture
// Use LighterWebsocketPrivateService.SubscribeAccount() instead
func (ws *WSClient) StreamAccount(ctx context.Context, accountId int64, callback func(*WSAccountUpdate) error) error {
	// Note: Based on Python implementation, account streams don't require authentication
	// The account_id itself serves as the access control
	
	handler := func(data []byte) error {
		var accountUpdate WSAccountUpdate
		
		if err := json.Unmarshal(data, &accountUpdate); err != nil {
			return fmt.Errorf("failed to unmarshal account update: %v", err)
		}
		
		// Check if it's an account message (either snapshot or update)
		if accountUpdate.Type != MessageTypeAccount && accountUpdate.Type != MessageTypeAccountSubscribed {
			return nil
		}
		
		return callback(&accountUpdate)
	}
	
	// Register handlers for both snapshot and update messages
	ws.AddHandler(MessageTypeAccount, handler)
	ws.AddHandler(MessageTypeAccountSubscribed, handler)
	return ws.Subscribe(fmt.Sprintf("%s/%d", ChannelAccount, accountId), "")
}

// Note: StreamOrders is deprecated - orders are handled through StreamAccount
// Use StreamAccount instead, which includes order updates, positions, and balance changes

// Unsubscribe methods

// UnsubscribeOrderBook unsubscribes from order book updates
func (ws *WSClient) UnsubscribeOrderBook(marketId uint8) error {
	ws.RemoveHandler(MessageTypeOrderBook)
	return ws.Unsubscribe(fmt.Sprintf("%s/%d", ChannelOrderBook, marketId), "")
}

// UnsubscribeTicker unsubscribes from ticker updates
func (ws *WSClient) UnsubscribeTicker(marketId uint8) error {
	ws.RemoveHandler(ChannelTicker)
	return ws.Unsubscribe(fmt.Sprintf("%s/%d", ChannelTicker, marketId), "")
}

// UnsubscribeTrades unsubscribes from trade updates
func (ws *WSClient) UnsubscribeTrades(marketId uint8) error {
	ws.RemoveHandler(ChannelTrades)
	return ws.Unsubscribe(fmt.Sprintf("%s/%d", ChannelTrades, marketId), "")
}

// UnsubscribeMarkPrice unsubscribes from mark price updates
func (ws *WSClient) UnsubscribeMarkPrice(marketId uint8) error {
	ws.RemoveHandler(ChannelMarkPrice)
	return ws.Unsubscribe(fmt.Sprintf("%s/%d", ChannelMarkPrice, marketId), "")
}

// UnsubscribeAccount unsubscribes from account updates
func (ws *WSClient) UnsubscribeAccount(accountId int64) error {
	ws.RemoveHandler(MessageTypeAccount)
	return ws.Unsubscribe(fmt.Sprintf("%s/%d", ChannelAccount, accountId), "")
}

// UnsubscribeOrders unsubscribes from order updates
func (ws *WSClient) UnsubscribeOrders() error {
	ws.RemoveHandler(ChannelOrders)
	return ws.Unsubscribe(ChannelOrders, "")
}

// Batch subscription convenience methods

// StreamMultiple subscribes to multiple streams at once using batch subscription
// This is more efficient than calling individual Stream methods
// Example usage:
//   subscriptions := []SubscriptionRequest{
//       NewOrdersSubscription(),
//       NewAccountSubscription(accountId),
//       NewOrderBookSubscription(1),
//       NewTickerSubscription(2),
//   }
//   ws.StreamMultiple(ctx, subscriptions, handlers)
func (ws *WSClient) StreamMultiple(ctx context.Context, requests []SubscriptionRequest, handlers map[string]WSHandler) error {
	// Add all handlers first
	for channel, handler := range handlers {
		ws.AddHandler(channel, handler)
	}
	
	// Batch subscribe to all channels
	return ws.SubscribeMultiple(requests)
}

// Utility methods for data conversion

// ConvertWSOrderBookToKubi converts WebSocket order book data to kubi format
func ConvertWSOrderBookToKubi(wsData *WSOrderBookUpdate, marketId uint8) (*OrderBookDataResponse, error) {
	orderBooks := make([]OrderBookData, 0, 1)
	
	bids := make([]PriceLevel, 0, len(wsData.Bids))
	for i, bid := range wsData.Bids {
		if len(bid) < 2 {
			log.Printf("[WSClient] Warning: bid[%d] has insufficient length %d, skipping", i, len(bid))
			continue
		}
		bids = append(bids, PriceLevel{
			Price:    bid[0],
			Quantity: bid[1],
		})
	}
	
	asks := make([]PriceLevel, 0, len(wsData.Asks))
	for i, ask := range wsData.Asks {
		if len(ask) < 2 {
			log.Printf("[WSClient] Warning: ask[%d] has insufficient length %d, skipping", i, len(ask))
			continue
		}
		asks = append(asks, PriceLevel{
			Price:    ask[0],
			Quantity: ask[1],
		})
	}
	
	orderBook := OrderBookData{
		MarketId: marketId, // Use provided market ID directly
		Bids:     bids,
		Asks:     asks,
	}
	
	orderBooks = append(orderBooks, orderBook)
	
	return &OrderBookDataResponse{
		ResultCode: ResultCode{Code: CodeOK},
		OrderBooks: orderBooks,
	}, nil
}

// ConvertWSTickerToKubi converts WebSocket ticker data to kubi format
func ConvertWSTickerToKubi(wsData *WSTickerUpdate) (*WSTickerUpdate, error) {
	// For now, just return the same structure
	// This might need conversion based on the actual kubi PriceUpdate structure
	return wsData, nil
}

// Note: GetMarketIdFromSymbol function removed as WebSocket streams now use market_id directly