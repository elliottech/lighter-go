package client

import (
	"log"
)

// Market Data Streams - Legacy methods removed
// Use LighterWebsocketPublicService / LighterWebsocketPrivateService instead

// Note: StreamTicker was removed because Lighter WebSocket API does not support ticker streams

// Note: StreamTrades was removed because Lighter WebSocket API does not support trades streams

// Note: StreamMarkPrice was removed because Lighter WebSocket API does not support mark price streams

// Account Data Streams - Legacy methods removed
// Use LighterWebsocketPrivateService.SubscribeAccount() instead

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

// Note: ConvertWSTickerToKubi was removed because ticker streams are not supported by Lighter WebSocket API

// Note: GetMarketIdFromSymbol function removed as WebSocket streams now use market_id directly