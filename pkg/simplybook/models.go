package simplybook

import "time"

// BookingClient 結構體用於表示客戶
type BookingClient struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// Booking 表示預約資訊，根據提供的 API 響應格式修改
type Booking struct {
	ID           string        `json:"id"`
	Code         string        `json:"code"`
	StartTime    time.Time     `json:"start_datetime"`
	EndTime      time.Time     `json:"end_datetime"`
	Client       BookingClient `json:"client"`
	ServiceID    string        `json:"service_id,omitempty"`
	ServiceName  string        `json:"service_name,omitempty"`
	ProviderID   string        `json:"provider_id,omitempty"`
	ProviderName string        `json:"provider_name,omitempty"`
	Confirmed    bool          `json:"confirmed,omitempty"`
	Notes        string        `json:"notes,omitempty"`
	Status       string        `json:"status,omitempty"`
}

// Service 表示服務信息
type Service struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Duration    int      `json:"duration"`
	ProvidersID []string `json:"providers_id"`
}

// Provider 表示服務提供者
type Provider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// WebhookPayload 表示 SimplyBook 的 webhook 負載
type WebhookPayload struct {
	Action      string `json:"notification_type"` // 'create', 'update', 'cancel'
	BookingID   string `json:"booking_id"`
	Company     string `json:"company"`
	BookingHash string `json:"booking_hash"`
	Timestamp   string `json:"webhook_timestamp"`
}

/** webhook example

{
	"booking_id":"2359",
	"booking_hash":"8fc073069dacec5b52775d741a9edbe8",
	"company":"choice",
	"notification_type":"notify",
	"webhook_timestamp":1743210065,
	"signature_algo":"sha256"
}

**/
