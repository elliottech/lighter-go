package ws

import (
	"sync"
	"testing"
)

func TestNewOrderBookState(t *testing.T) {
	obs := NewOrderBookState(0)

	if obs.MarketIndex != 0 {
		t.Errorf("expected MarketIndex 0, got %d", obs.MarketIndex)
	}

	if len(obs.Bids) != 0 {
		t.Errorf("expected empty Bids, got %d", len(obs.Bids))
	}

	if len(obs.Asks) != 0 {
		t.Errorf("expected empty Asks, got %d", len(obs.Asks))
	}
}

func TestOrderBookState_ApplySnapshot(t *testing.T) {
	obs := NewOrderBookState(0)

	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
			{Price: "99", Size: "20"},
		},
		Asks: []OrderBookLevel{
			{Price: "101", Size: "15"},
			{Price: "102", Size: "25"},
		},
	}

	err := obs.ApplySnapshot(snapshot)
	if err != nil {
		t.Fatalf("ApplySnapshot failed: %v", err)
	}

	if obs.GetSequence() != 100 {
		t.Errorf("expected sequence 100, got %d", obs.GetSequence())
	}

	if len(obs.Bids) != 2 {
		t.Errorf("expected 2 bids, got %d", len(obs.Bids))
	}

	if len(obs.Asks) != 2 {
		t.Errorf("expected 2 asks, got %d", len(obs.Asks))
	}

	// Verify bid levels
	bid, exists := obs.Bids["100"]
	if !exists || bid.Size != "10" {
		t.Errorf("expected bid at 100 with size 10, got %+v", bid)
	}

	// Verify ask levels
	ask, exists := obs.Asks["101"]
	if !exists || ask.Size != "15" {
		t.Errorf("expected ask at 101 with size 15, got %+v", ask)
	}
}

func TestOrderBookState_ApplySnapshot_ClearsExisting(t *testing.T) {
	obs := NewOrderBookState(0)

	// Apply first snapshot
	snapshot1 := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
		},
		Asks: []OrderBookLevel{
			{Price: "101", Size: "15"},
		},
	}
	obs.ApplySnapshot(snapshot1)

	// Apply second snapshot with different prices
	snapshot2 := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    200,
		Bids: []OrderBookLevel{
			{Price: "105", Size: "5"},
		},
		Asks: []OrderBookLevel{
			{Price: "106", Size: "8"},
		},
	}
	obs.ApplySnapshot(snapshot2)

	// Old prices should be gone
	_, exists := obs.Bids["100"]
	if exists {
		t.Error("old bid at 100 should have been removed")
	}

	// New prices should be present
	bid, exists := obs.Bids["105"]
	if !exists || bid.Size != "5" {
		t.Errorf("expected bid at 105 with size 5, got %+v", bid)
	}
}

func TestOrderBookState_ApplyDelta(t *testing.T) {
	obs := NewOrderBookState(0)

	// First apply a snapshot
	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
		},
		Asks: []OrderBookLevel{
			{Price: "101", Size: "15"},
		},
	}
	obs.ApplySnapshot(snapshot)

	// Apply delta
	delta := &OrderBookDelta{
		MarketIndex: 0,
		Sequence:    101,
		BidUpdates: []OrderBookLevel{
			{Price: "100", Size: "20"}, // Update existing
			{Price: "99", Size: "30"},  // Add new
		},
		AskUpdates: []OrderBookLevel{
			{Price: "101", Size: "0"}, // Remove
		},
	}

	err := obs.ApplyDelta(delta)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if obs.GetSequence() != 101 {
		t.Errorf("expected sequence 101, got %d", obs.GetSequence())
	}

	// Check updated bid
	bid, exists := obs.Bids["100"]
	if !exists || bid.Size != "20" {
		t.Errorf("expected bid at 100 with size 20, got %+v", bid)
	}

	// Check new bid
	bid, exists = obs.Bids["99"]
	if !exists || bid.Size != "30" {
		t.Errorf("expected bid at 99 with size 30, got %+v", bid)
	}

	// Check removed ask
	_, exists = obs.Asks["101"]
	if exists {
		t.Error("ask at 101 should have been removed")
	}
}

func TestOrderBookState_ApplyDelta_SequenceGap(t *testing.T) {
	obs := NewOrderBookState(0)

	// First apply a snapshot
	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
		},
		Asks: []OrderBookLevel{},
	}
	obs.ApplySnapshot(snapshot)

	// Apply delta with sequence gap
	delta := &OrderBookDelta{
		MarketIndex: 0,
		Sequence:    103, // Gap: expected 101
		BidUpdates:  []OrderBookLevel{},
		AskUpdates:  []OrderBookLevel{},
	}

	err := obs.ApplyDelta(delta)
	if err != ErrSequenceGap {
		t.Errorf("expected ErrSequenceGap, got %v", err)
	}
}

func TestOrderBookState_ApplyDelta_EmptySize(t *testing.T) {
	obs := NewOrderBookState(0)

	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
			{Price: "99", Size: "20"},
		},
		Asks: []OrderBookLevel{},
	}
	obs.ApplySnapshot(snapshot)

	// Remove using empty string
	delta := &OrderBookDelta{
		MarketIndex: 0,
		Sequence:    101,
		BidUpdates: []OrderBookLevel{
			{Price: "99", Size: ""}, // Empty size should remove
		},
	}

	obs.ApplyDelta(delta)

	_, exists := obs.Bids["99"]
	if exists {
		t.Error("bid at 99 should have been removed with empty size")
	}
}

func TestOrderBookState_MergeUpdates(t *testing.T) {
	obs := NewOrderBookState(0)

	// Add initial data
	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
		},
		Asks: []OrderBookLevel{
			{Price: "101", Size: "15"},
		},
	}
	obs.ApplySnapshot(snapshot)

	// Merge updates
	bids := []OrderBookLevel{
		{Price: "100", Size: "25"},
		{Price: "98", Size: "50"},
	}
	asks := []OrderBookLevel{
		{Price: "101", Size: "0"},  // Remove
		{Price: "103", Size: "40"}, // Add
	}

	obs.MergeUpdates(bids, asks)

	// Check bids
	bid, exists := obs.Bids["100"]
	if !exists || bid.Size != "25" {
		t.Errorf("expected bid at 100 with size 25, got %+v", bid)
	}

	bid, exists = obs.Bids["98"]
	if !exists || bid.Size != "50" {
		t.Errorf("expected bid at 98 with size 50, got %+v", bid)
	}

	// Check asks
	_, exists = obs.Asks["101"]
	if exists {
		t.Error("ask at 101 should have been removed")
	}

	ask, exists := obs.Asks["103"]
	if !exists || ask.Size != "40" {
		t.Errorf("expected ask at 103 with size 40, got %+v", ask)
	}
}

func TestOrderBookState_GetBestBid(t *testing.T) {
	obs := NewOrderBookState(0)

	// Empty book
	best := obs.GetBestBid()
	if best != nil {
		t.Errorf("expected nil for empty book, got %+v", best)
	}

	// Add bids
	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
			{Price: "99", Size: "20"},
			{Price: "101", Size: "5"}, // Highest
		},
		Asks: []OrderBookLevel{},
	}
	obs.ApplySnapshot(snapshot)

	best = obs.GetBestBid()
	if best == nil {
		t.Fatal("expected best bid, got nil")
	}
	if best.Price != "101" || best.Size != "5" {
		t.Errorf("expected best bid at 101 with size 5, got %+v", best)
	}
}

func TestOrderBookState_GetBestAsk(t *testing.T) {
	obs := NewOrderBookState(0)

	// Empty book
	best := obs.GetBestAsk()
	if best != nil {
		t.Errorf("expected nil for empty book, got %+v", best)
	}

	// Add asks
	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids:        []OrderBookLevel{},
		Asks: []OrderBookLevel{
			{Price: "102", Size: "10"},
			{Price: "103", Size: "20"},
			{Price: "101", Size: "5"}, // Lowest
		},
	}
	obs.ApplySnapshot(snapshot)

	best = obs.GetBestAsk()
	if best == nil {
		t.Fatal("expected best ask, got nil")
	}
	if best.Price != "101" || best.Size != "5" {
		t.Errorf("expected best ask at 101 with size 5, got %+v", best)
	}
}

func TestOrderBookState_GetBids_Sorted(t *testing.T) {
	obs := NewOrderBookState(0)

	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "99", Size: "20"},
			{Price: "101", Size: "5"},
			{Price: "100", Size: "10"},
		},
		Asks: []OrderBookLevel{},
	}
	obs.ApplySnapshot(snapshot)

	bids := obs.GetBids()

	if len(bids) != 3 {
		t.Fatalf("expected 3 bids, got %d", len(bids))
	}

	// Should be sorted descending
	expectedPrices := []string{"101", "100", "99"}
	for i, bid := range bids {
		if bid.Price != expectedPrices[i] {
			t.Errorf("bid[%d]: expected price %s, got %s", i, expectedPrices[i], bid.Price)
		}
	}
}

func TestOrderBookState_GetAsks_Sorted(t *testing.T) {
	obs := NewOrderBookState(0)

	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids:        []OrderBookLevel{},
		Asks: []OrderBookLevel{
			{Price: "103", Size: "20"},
			{Price: "101", Size: "5"},
			{Price: "102", Size: "10"},
		},
	}
	obs.ApplySnapshot(snapshot)

	asks := obs.GetAsks()

	if len(asks) != 3 {
		t.Fatalf("expected 3 asks, got %d", len(asks))
	}

	// Should be sorted ascending
	expectedPrices := []string{"101", "102", "103"}
	for i, ask := range asks {
		if ask.Price != expectedPrices[i] {
			t.Errorf("ask[%d]: expected price %s, got %s", i, expectedPrices[i], ask.Price)
		}
	}
}

func TestOrderBookState_Clone(t *testing.T) {
	obs := NewOrderBookState(0)

	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    100,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
		},
		Asks: []OrderBookLevel{
			{Price: "101", Size: "15"},
		},
	}
	obs.ApplySnapshot(snapshot)

	// Clone
	clone := obs.Clone()

	// Verify clone has same data
	if clone.GetSequence() != 100 {
		t.Errorf("expected clone sequence 100, got %d", clone.GetSequence())
	}

	// Modify original
	obs.MergeUpdates([]OrderBookLevel{{Price: "100", Size: "999"}}, nil)

	// Clone should be unaffected
	bid := clone.Bids["100"]
	if bid.Size != "10" {
		t.Errorf("clone should not be affected by original changes, got size %s", bid.Size)
	}
}

func TestOrderBookState_ConcurrentAccess(t *testing.T) {
	obs := NewOrderBookState(0)

	snapshot := &OrderBookSnapshot{
		MarketIndex: 0,
		Sequence:    0,
		Bids: []OrderBookLevel{
			{Price: "100", Size: "10"},
		},
		Asks: []OrderBookLevel{
			{Price: "101", Size: "15"},
		},
	}
	obs.ApplySnapshot(snapshot)

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent reads and writes
	wg.Add(goroutines * 3)

	// Readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = obs.GetBestBid()
			_ = obs.GetBestAsk()
			_ = obs.GetBids()
			_ = obs.GetAsks()
			_ = obs.GetSequence()
		}()
	}

	// Writers - MergeUpdates
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			obs.MergeUpdates(
				[]OrderBookLevel{{Price: "100", Size: "10"}},
				[]OrderBookLevel{{Price: "101", Size: "15"}},
			)
		}(i)
	}

	// Clone
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = obs.Clone()
		}()
	}

	wg.Wait()
}

func TestComparePrices(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"100", "100", 0},
		{"100", "99", 1},
		{"99", "100", -1},
		{"1000", "100", 1},  // Longer string wins
		{"100", "1000", -1}, // Shorter string loses
	}

	for _, tt := range tests {
		result := comparePrices(tt.a, tt.b)
		// Normalize to -1, 0, 1
		if result < 0 {
			result = -1
		} else if result > 0 {
			result = 1
		}

		if result != tt.expected {
			t.Errorf("comparePrices(%s, %s) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}
