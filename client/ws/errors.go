package ws

import (
	"errors"
	"fmt"
)

// Common WebSocket errors
var (
	ErrNotConnected                 = errors.New("websocket not connected")
	ErrAlreadyConnected             = errors.New("websocket already connected")
	ErrConnectionClosed             = errors.New("websocket connection closed")
	ErrConnectionTimeout            = errors.New("connection timeout waiting for server acknowledgment")
	ErrMaxReconnectAttemptsExceeded = errors.New("max reconnect attempts exceeded")
	ErrSubscriptionFailed           = errors.New("subscription failed")
	ErrUnsubscribeFailed            = errors.New("unsubscribe failed")
	ErrInvalidMessage               = errors.New("invalid message format")
	ErrSequenceGap                  = errors.New("order book sequence gap detected")
	ErrOrderBookNotFound            = errors.New("order book state not found")
	ErrAuthTokenRequired            = errors.New("auth token required for private channel subscription")
	ErrSubscriptionTimeout          = errors.New("subscription confirmation timeout")
	ErrAlreadySubscribed            = errors.New("already subscribed")
	ErrNotSubscribed                = errors.New("not subscribed")
	ErrBatchTooLarge                = errors.New("transaction batch exceeds maximum of 50")
)

// WsError represents an error from the WebSocket server
type WsError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *WsError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ws error %d: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("ws error %d: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *WsError) Unwrap() error {
	return e.Err
}

// NewWsError creates a new WsError
func NewWsError(code int, message string) *WsError {
	return &WsError{
		Code:    code,
		Message: message,
	}
}

// NewWsErrorWithCause creates a new WsError with an underlying cause
func NewWsErrorWithCause(code int, message string, err error) *WsError {
	return &WsError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ConnectionError represents a connection error
type ConnectionError struct {
	Err error
}

// Error implements the error interface
func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.Err)
}

// Unwrap returns the underlying error
func (e *ConnectionError) Unwrap() error {
	return e.Err
}
