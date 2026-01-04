package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	config      *WSConfig
	conn        *websocket.Conn
	mu          sync.RWMutex
	writeMu     sync.Mutex // Separate mutex for write operations
	handlers    map[string][]WSHandler
	isConnected bool
	stopCh      chan struct{}
	authToken   string
	stopped     bool // Flag to track if stopCh is closed

	// For managing subscriptions
	subscriptions map[string]bool

	// Order book state management - like Python version
	orderBookStates map[uint16]*WSOrderBookState

	// Connection state callbacks - like Python version
	onConnected    func()
	onDisconnected func()
}

type WSHandler func(data []byte) error

// NewWSClient creates a new WebSocket client
func NewWSClient(config *WSConfig) *WSClient {
	if config == nil {
		config = DefaultWSConfig()
	}

	return &WSClient{
		config:          config,
		handlers:        make(map[string][]WSHandler),
		subscriptions:   make(map[string]bool),
		orderBookStates: make(map[uint16]*WSOrderBookState),
		stopCh:          make(chan struct{}),
	}
}

// SetAuthToken sets the authentication token for authenticated streams
func (ws *WSClient) SetAuthToken(token string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.authToken = token
}

// SetOnConnected sets callback for when connection is established
func (ws *WSClient) SetOnConnected(callback func()) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.onConnected = callback
}

// SetOnDisconnected sets callback for when connection is lost
func (ws *WSClient) SetOnDisconnected(callback func()) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.onDisconnected = callback
}

// Connect establishes WebSocket connection
func (ws *WSClient) Connect(ctx context.Context) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.isConnected {
		return nil
	}

	// Reset the channel if it was previously closed
	if ws.stopped {
		ws.stopCh = make(chan struct{})
		ws.stopped = false
	}

	log.Println("[WSClient] Connecting to Lighter WebSocket...", ws.config.URL)
	u, err := url.Parse(ws.config.URL)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %v", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}

	headers := http.Header{}
	if ws.authToken != "" {
		headers.Set("Authorization", "Bearer "+ws.authToken)
	}

	conn, resp, err := dialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("WS handshake failed, status=%d, body=%s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("failed to connect to WebSocket: %v", err)
	}

	ws.conn = conn
	ws.isConnected = true

	// Start message handler goroutines
	go ws.readMessages(ctx)
	go ws.ping(ctx)

	log.Println("[WSClient] Connected to Lighter WebSocket")
	return nil
}

// Disconnect closes the WebSocket connection
func (ws *WSClient) Disconnect() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.isConnected {
		return nil
	}

	// Only close stopCh if it hasn't been closed already
	if !ws.stopped {
		close(ws.stopCh)
		ws.stopped = true
	}
	ws.isConnected = false

	if ws.conn != nil {
		// Ensure no writes are in progress before closing
		ws.writeMu.Lock()
		err := ws.conn.Close()
		ws.conn = nil
		ws.writeMu.Unlock()
		return err
	}

	log.Println("[WSClient] Disconnected from Lighter WebSocket")
	return nil
}

// IsConnected returns connection status
func (ws *WSClient) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.isConnected
}

// Subscribe subscribes to a WebSocket channel
func (ws *WSClient) Subscribe(channel, symbol string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	subscriptionKey := channel
	if symbol != "" {
		subscriptionKey = fmt.Sprintf("%s:%s", channel, symbol)
	}

	ws.subscriptions[subscriptionKey] = true

	if !ws.isConnected {
		return fmt.Errorf("WebSocket not connected")
	}

	msg := WSSubscribeMessage{
		Type:    MessageTypeSubscribe,
		Channel: channel,
		Symbol:  symbol,
	}

	log.Printf("[WSClient] Subscribing to channel: %s (symbol: %s, key: %s)", channel, symbol, subscriptionKey)
	return ws.sendMessage(msg)
}

// Unsubscribe unsubscribes from a WebSocket channel
func (ws *WSClient) Unsubscribe(channel, symbol string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	subscriptionKey := channel
	if symbol != "" {
		subscriptionKey = fmt.Sprintf("%s:%s", channel, symbol)
	}

	delete(ws.subscriptions, subscriptionKey)

	if !ws.isConnected {
		return nil // Already disconnected
	}

	msg := WSUnsubscribeMessage{
		Type:    MessageTypeUnsubscribe,
		Channel: channel,
		Symbol:  symbol,
	}

	return ws.sendMessage(msg)
}

// Note: SubscribeMultiple and UnsubscribeMultiple methods removed as they were unused

// AddHandler adds a message handler for a specific channel
func (ws *WSClient) AddHandler(channel string, handler WSHandler) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.handlers[channel] == nil {
		ws.handlers[channel] = make([]WSHandler, 0)
	}
	ws.handlers[channel] = append(ws.handlers[channel], handler)
}

// RemoveHandler removes all handlers for a channel
func (ws *WSClient) RemoveHandler(channel string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	delete(ws.handlers, channel)
}

func (ws *WSClient) sendMessage(msg interface{}) error {
	// Protect write operations with a mutex
	ws.writeMu.Lock()
	defer ws.writeMu.Unlock()

	if ws.conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	ws.conn.SetWriteDeadline(time.Now().Add(ws.config.WriteTimeout))
	return ws.conn.WriteMessage(websocket.TextMessage, data)
}

func (ws *WSClient) readMessages(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WSClient] Panic in readMessages: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.stopCh:
			return
		default:
			if ws.conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			ws.conn.SetReadDeadline(time.Now().Add(ws.config.ReadTimeout))
			_, data, err := ws.conn.ReadMessage()
			if err != nil {
				log.Printf("[WSClient] Read error: %v", err)
				ws.handleDisconnect(ctx)
				return
			}

			ws.handleMessage(data)
		}
	}
}

func (ws *WSClient) handleMessage(data []byte) {
	// First check if this is an error message
	var errorMsg struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(data, &errorMsg); err == nil && errorMsg.Error.Code != 0 {
		log.Printf("[WSClient] WebSocket Error %d: %s", errorMsg.Error.Code, errorMsg.Error.Message)
		return
	}

	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[WSClient] Failed to unmarshal message: %v", err)
		return
	}

	// Debug: Print parsed message type (skip frequent messages)
	if msg.Type != "update/order_book" && msg.Type != MessageTypePong && msg.Type != MessageTypePing {
		log.Printf("[WSClient] Received message: %s", string(data))
		log.Printf("[WSClient] Parsed message type: %s", msg.Type)
	}

	// Handle specific message types like Python version
	switch msg.Type {
	case MessageTypePing:
		// Server sent ping, respond with pong
		pong := WSMessage{Type: MessageTypePong}
		if err := ws.sendMessage(pong); err != nil {
			log.Printf("[WSClient] Failed to send pong: %v", err)
		}
		return
	case MessageTypePong:
		// Server responded to our ping, nothing to do
		return
	case MessageTypeConnected:
		log.Println("[WSClient] Connected to Lighter WebSocket")
		if ws.onConnected != nil {
			go ws.onConnected()
		}
		return
	case MessageTypeSubscribed:
		log.Printf("[WSClient] Successfully subscribed")
		return
	case MessageTypeUnsubscribed:
		log.Printf("[WSClient] Successfully unsubscribed")
		return
	case MessageTypeOrderBookSubscribed:
		log.Printf("[WSClient] Order book subscription confirmed - processing snapshot")
		// Remove built-in snapshot handling - let registered handlers in new architecture handle this
	case MessageTypeAccountSubscribed:
		log.Printf("[WSClient] Account subscription confirmed - processing snapshot")
		// ws.handleAccountSnapshot(data)
		// return
	case MessageTypeOrderBookUpdate:
		// Remove built-in handling - let registered handlers in new architecture handle this
	case MessageTypeAccountUpdate:
		// ws.handleAccountMessage(data)
	}

	// Route message to appropriate handlers
	ws.mu.RLock()
	handlers := ws.handlers[msg.Type]
	ws.mu.RUnlock()

	for _, handler := range handlers {
		func(h WSHandler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[WSClient] Handler panic: %v", r)
				}
			}()

			if err := h(data); err != nil {
				log.Printf("[WSClient] Handler error: %v", err)
			}
		}(handler)
	}
}

func (ws *WSClient) ping(ctx context.Context) {
	ticker := time.NewTicker(ws.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.stopCh:
			return
		case <-ticker.C:
			if ws.isConnected && ws.conn != nil {
				ping := WSMessage{Type: MessageTypePing}
				if err := ws.sendMessage(ping); err != nil {
					log.Printf("[WSClient] Failed to send ping: %v", err)
				}
			}
		}
	}
}

func (ws *WSClient) handleDisconnect(ctx context.Context) {
	ws.mu.Lock()
	ws.isConnected = false
	if ws.conn != nil {
		// Ensure no writes are in progress before closing
		ws.writeMu.Lock()
		ws.conn.Close()
		ws.conn = nil
		ws.writeMu.Unlock()
	}
	onDisconnected := ws.onDisconnected
	ws.mu.Unlock()

	log.Println("[WSClient] Connection lost")

	// Call disconnect callback to notify external
	if onDisconnected != nil {
		go onDisconnected()
	}

	// Do not attempt to reconnect - let external system handle it
}

// handleAccountSnapshot handles complete account snapshot (subscribed/account_all)
func (ws *WSClient) handleAccountSnapshot(data []byte) {
	var accountSnapshot struct {
		Type               string                 `json:"type"`
		Channel            string                 `json:"channel"`
		Account            int64                  `json:"account"`
		DailyTradesCount   int                    `json:"daily_trades_count"`
		DailyVolume        float64                `json:"daily_volume"`
		MonthlyTradesCount int                    `json:"monthly_trades_count"`
		MonthlyVolume      float64                `json:"monthly_volume"`
		TotalTradesCount   int                    `json:"total_trades_count"`
		TotalVolume        float64                `json:"total_volume"`
		WeeklyTradesCount  int                    `json:"weekly_trades_count"`
		WeeklyVolume       float64                `json:"weekly_volume"`
		Positions          map[string]interface{} `json:"positions"`
		Shares             []interface{}          `json:"shares"`
		Trades             map[string]interface{} `json:"trades"`
		FundingHistories   map[string]interface{} `json:"funding_histories"`
	}

	if err := json.Unmarshal(data, &accountSnapshot); err != nil {
		log.Printf("[WSClient] Failed to unmarshal account snapshot: %v", err)
		return
	}

	log.Printf("[WSClient] Loaded account snapshot for account %d: %d positions, %d total trades",
		accountSnapshot.Account, len(accountSnapshot.Positions), accountSnapshot.TotalTradesCount)
}

// handleAccountMessage handles incremental account updates (update/account_all)
func (ws *WSClient) handleAccountMessage(data []byte) {
	var accountUpdate WSAccountUpdate
	if err := json.Unmarshal(data, &accountUpdate); err != nil {
		log.Printf("[WSClient] Failed to unmarshal account update: %v", err)
		return
	}

	log.Printf("[WSClient] Account update for account %d, %d positions, type: %s",
		accountUpdate.Account, len(accountUpdate.Positions), accountUpdate.Type)
}

// GetOrderBookState returns current order book state for a market (like Python version)
func (ws *WSClient) GetOrderBookState(marketId uint16) *WSOrderBookState {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.orderBookStates[marketId]
}
