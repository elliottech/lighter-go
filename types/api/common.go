package api

import "fmt"

const (
	CodeOK = 200
)

// BaseResponse is embedded in all API responses
type BaseResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message,omitempty"`
}

// Error returns an error if the response indicates failure
func (r *BaseResponse) Error() error {
	if r.Code != CodeOK {
		return fmt.Errorf("API error %d: %s", r.Code, r.Message)
	}
	return nil
}

// IsSuccess returns true if the response indicates success
func (r *BaseResponse) IsSuccess() bool {
	return r.Code == CodeOK
}

// Cursor represents pagination cursor information
type Cursor struct {
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// PaginationOpts contains options for paginated requests
type PaginationOpts struct {
	Limit  int
	Cursor string
}

// TimestampRange contains timestamp range options for queries
type TimestampRange struct {
	StartTimestamp int64
	EndTimestamp   int64
}

// SortOpts contains sorting options
type SortOpts struct {
	SortBy    string
	SortOrder string // "asc" or "desc"
}

// QueryBy represents the different ways to query an entity
type QueryBy string

const (
	QueryByIndex     QueryBy = "index"
	QueryByL1Address QueryBy = "l1_address"
	QueryByHash      QueryBy = "hash"
	QueryByHeight    QueryBy = "height"
	QueryByCommit    QueryBy = "commitment"
)

// MarketFilter represents market type filters
type MarketFilter string

const (
	MarketFilterAll   MarketFilter = "all"
	MarketFilterPerps MarketFilter = "perps"
	MarketFilterSpot  MarketFilter = "spot"
)

// OrderStatusFilter represents order status filters
type OrderStatusFilter string

const (
	OrderStatusAll       OrderStatusFilter = "all"
	OrderStatusOpen      OrderStatusFilter = "open"
	OrderStatusFilled    OrderStatusFilter = "filled"
	OrderStatusCancelled OrderStatusFilter = "cancelled"
	OrderStatusExpired   OrderStatusFilter = "expired"
)

// ResultInfo provides metadata about the result
type ResultInfo struct {
	Total  int64 `json:"total,omitempty"`
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}
