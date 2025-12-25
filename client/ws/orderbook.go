package ws

import (
	"sort"
	"sync"
	"time"
)

// OrderBookState maintains the current state of an order book
type OrderBookState struct {
	mu          sync.RWMutex
	MarketIndex int16
	Sequence    int64
	Bids        map[string]OrderBookLevel // price -> level
	Asks        map[string]OrderBookLevel // price -> level
	LastUpdate  time.Time
}

// NewOrderBookState creates a new order book state
func NewOrderBookState(marketIndex int16) *OrderBookState {
	return &OrderBookState{
		MarketIndex: marketIndex,
		Bids:        make(map[string]OrderBookLevel),
		Asks:        make(map[string]OrderBookLevel),
	}
}

// ApplySnapshot replaces the entire order book state with a snapshot
func (obs *OrderBookState) ApplySnapshot(snapshot *OrderBookSnapshot) error {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	// Clear existing state
	obs.Bids = make(map[string]OrderBookLevel)
	obs.Asks = make(map[string]OrderBookLevel)

	// Apply snapshot
	for _, level := range snapshot.Bids {
		obs.Bids[level.Price] = level
	}
	for _, level := range snapshot.Asks {
		obs.Asks[level.Price] = level
	}

	obs.Sequence = snapshot.Sequence
	obs.LastUpdate = time.Now()
	return nil
}

// ApplyDelta applies an incremental update to the order book
func (obs *OrderBookState) ApplyDelta(delta *OrderBookDelta) error {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	// Check for sequence gap
	if delta.Sequence != obs.Sequence+1 && obs.Sequence != 0 {
		return ErrSequenceGap
	}

	// Apply bid updates
	for _, level := range delta.BidUpdates {
		if level.Size == "0" || level.Size == "" {
			delete(obs.Bids, level.Price)
		} else {
			obs.Bids[level.Price] = level
		}
	}

	// Apply ask updates
	for _, level := range delta.AskUpdates {
		if level.Size == "0" || level.Size == "" {
			delete(obs.Asks, level.Price)
		} else {
			obs.Asks[level.Price] = level
		}
	}

	obs.Sequence = delta.Sequence
	obs.LastUpdate = time.Now()
	return nil
}

// GetBestBid returns the highest bid price level
func (obs *OrderBookState) GetBestBid() *OrderBookLevel {
	obs.mu.RLock()
	defer obs.mu.RUnlock()

	if len(obs.Bids) == 0 {
		return nil
	}

	var bestPrice string
	for price := range obs.Bids {
		if bestPrice == "" || comparePrices(price, bestPrice) > 0 {
			bestPrice = price
		}
	}

	if bestPrice == "" {
		return nil
	}

	level := obs.Bids[bestPrice]
	return &level
}

// GetBestAsk returns the lowest ask price level
func (obs *OrderBookState) GetBestAsk() *OrderBookLevel {
	obs.mu.RLock()
	defer obs.mu.RUnlock()

	if len(obs.Asks) == 0 {
		return nil
	}

	var bestPrice string
	for price := range obs.Asks {
		if bestPrice == "" || comparePrices(price, bestPrice) < 0 {
			bestPrice = price
		}
	}

	if bestPrice == "" {
		return nil
	}

	level := obs.Asks[bestPrice]
	return &level
}

// GetBids returns a copy of all bid levels sorted by price descending
func (obs *OrderBookState) GetBids() []OrderBookLevel {
	obs.mu.RLock()
	defer obs.mu.RUnlock()

	levels := make([]OrderBookLevel, 0, len(obs.Bids))
	for _, level := range obs.Bids {
		levels = append(levels, level)
	}

	// Sort by price descending (highest first)
	sort.Slice(levels, func(i, j int) bool {
		return comparePrices(levels[i].Price, levels[j].Price) > 0
	})

	return levels
}

// GetAsks returns a copy of all ask levels sorted by price ascending
func (obs *OrderBookState) GetAsks() []OrderBookLevel {
	obs.mu.RLock()
	defer obs.mu.RUnlock()

	levels := make([]OrderBookLevel, 0, len(obs.Asks))
	for _, level := range obs.Asks {
		levels = append(levels, level)
	}

	// Sort by price ascending (lowest first)
	sort.Slice(levels, func(i, j int) bool {
		return comparePrices(levels[i].Price, levels[j].Price) < 0
	})

	return levels
}

// GetSpread returns the bid-ask spread
func (obs *OrderBookState) GetSpread() (string, error) {
	bestBid := obs.GetBestBid()
	bestAsk := obs.GetBestAsk()

	if bestBid == nil || bestAsk == nil {
		return "", nil
	}

	// Simple string-based spread calculation
	// In a real implementation, you'd use proper decimal arithmetic
	return subtractPrices(bestAsk.Price, bestBid.Price), nil
}

// GetMidPrice returns the mid price
func (obs *OrderBookState) GetMidPrice() string {
	bestBid := obs.GetBestBid()
	bestAsk := obs.GetBestAsk()

	if bestBid == nil || bestAsk == nil {
		return ""
	}

	return averagePrices(bestBid.Price, bestAsk.Price)
}

// GetSequence returns the current sequence number
func (obs *OrderBookState) GetSequence() int64 {
	obs.mu.RLock()
	defer obs.mu.RUnlock()
	return obs.Sequence
}

// GetLastUpdate returns the time of the last update
func (obs *OrderBookState) GetLastUpdate() time.Time {
	obs.mu.RLock()
	defer obs.mu.RUnlock()
	return obs.LastUpdate
}

// Clone returns a deep copy of the order book state
func (obs *OrderBookState) Clone() *OrderBookState {
	obs.mu.RLock()
	defer obs.mu.RUnlock()

	clone := &OrderBookState{
		MarketIndex: obs.MarketIndex,
		Sequence:    obs.Sequence,
		Bids:        make(map[string]OrderBookLevel),
		Asks:        make(map[string]OrderBookLevel),
		LastUpdate:  obs.LastUpdate,
	}

	for k, v := range obs.Bids {
		clone.Bids[k] = v
	}
	for k, v := range obs.Asks {
		clone.Asks[k] = v
	}

	return clone
}

// Helper functions for price comparison
// These are simplified implementations - in production use proper decimal math

func comparePrices(a, b string) int {
	// Simple string comparison that works for numeric strings
	// In production, use a decimal library
	if len(a) != len(b) {
		return len(a) - len(b)
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func subtractPrices(a, b string) string {
	// Placeholder - in production use decimal math
	return a + "-" + b
}

func averagePrices(a, b string) string {
	// Placeholder - in production use decimal math
	return "(" + a + "+" + b + ")/2"
}
