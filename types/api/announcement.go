package api

// Announcement represents an exchange announcement
type Announcement struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Type        string `json:"type"` // "info", "warning", "maintenance", "update"
	Priority    int    `json:"priority"`
	StartTime   int64  `json:"start_time,omitempty"`
	EndTime     int64  `json:"end_time,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
	IsActive    bool   `json:"is_active"`
	URL         string `json:"url,omitempty"`
}

// Announcements is the response for announcement queries
type Announcements struct {
	BaseResponse
	Announcements []Announcement `json:"announcements"`
}

// Notification represents a user notification
type Notification struct {
	ID            int64  `json:"id"`
	AccountIndex  int64  `json:"account_index"`
	Type          string `json:"type"` // "order_filled", "liquidation", "deposit", "withdraw", etc.
	Title         string `json:"title"`
	Message       string `json:"message"`
	Data          string `json:"data,omitempty"` // JSON-encoded additional data
	IsRead        bool   `json:"is_read"`
	CreatedAt     int64  `json:"created_at"`
	ReadAt        int64  `json:"read_at,omitempty"`
}

// Notifications is the response for notification queries
type Notifications struct {
	BaseResponse
	Notifications []Notification `json:"notifications"`
	UnreadCount   int64          `json:"unread_count"`
	Cursor        Cursor         `json:"cursor,omitempty"`
}

// RespAckNotification is the response for acknowledging a notification
type RespAckNotification struct {
	BaseResponse
	NotificationID int64 `json:"notification_id"`
	Acknowledged   bool  `json:"acknowledged"`
}
