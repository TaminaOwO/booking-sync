package simplybook

import (
	"strings"
	"time"
)

// BookingClient 結構體用於表示客戶
type BookingClient struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// customTime 自定義時間類型，用於解析 SimplyBook API 返回的日期時間格式
type customTime struct {
	time.Time
}

// UnmarshalJSON 自定義時間解析方法
func (ct *customTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return nil
	}

	// 使用適合 SimplyBook API 返回格式的時間解析
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return err
	}

	// 設定台灣時區 (GMT+8)
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		// 如果無法載入台灣時區，使用固定偏移
		loc = time.FixedZone("GMT+8", 8*60*60)
	}

	// 將時間設為台灣時區
	ct.Time = time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		loc,
	)
	return nil
}

// Booking 表示預約資訊，根據提供的 API 響應格式修改
type Booking struct {
	ID           int           `json:"id"`
	Code         string        `json:"code"`
	StartTime    customTime    `json:"start_datetime"`
	EndTime      customTime    `json:"end_datetime"`
	Client       BookingClient `json:"client"`
	ServiceID    int           `json:"service_id,omitempty"`
	ServiceName  string        `json:"service_name,omitempty"`
	ProviderID   int           `json:"provider_id,omitempty"`
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
	Action      string `json:"notification_type"` // 'create', 'change', 'cancel', 'notify'
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

{"booking_id":"2360","booking_hash":"a59127ec2727c4a30b3a1e1f10867e61","company":"choice","notification_type":"change","webhook_timestamp":1743224826,"signature_algo":"sha256"}

**/
