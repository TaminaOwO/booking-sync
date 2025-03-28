package simplybook

import "time"

// Booking 表示預約資訊
type Booking struct {
	ID           string    `json:"id"`
	ClientName   string    `json:"client_name"`
	ClientEmail  string    `json:"client_email"`
	ClientPhone  string    `json:"client_phone"`
	StartTime    time.Time `json:"start_datetime"`
	EndTime      time.Time `json:"end_datetime"`
	Code         string    `json:"code"`
	Confirmed    bool      `json:"confirmed"`
	ServiceID    string    `json:"event_id"`
	ServiceName  string    `json:"event_name"`
	ProviderID   string    `json:"unit_id"`
	ProviderName string    `json:"provider_name"`
	Notes        string    `json:"note"`
	Status       string    `json:"status"`
}

// Service 表示服務信息
type Service struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Duration int      `json:"duration"`
	UnitMap  []string `json:"unit_map"`
}

// Provider 表示服務提供者
type Provider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// WebhookPayload 表示 SimplyBook 的 webhook 負載
type WebhookPayload struct {
	Action     string `json:"action"`     // 'create', 'update', 'delete'
	BookingID  string `json:"booking_id"` 
	ClientID   string `json:"client_id"`
	ProviderID string `json:"provider_id"`
	ServiceID  string `json:"service_id"`
	Timestamp  string `json:"timestamp"`
} 