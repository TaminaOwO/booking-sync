package gcalendar

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Client 代表 Google 日曆 API 客戶端
type Client struct {
	service       *calendar.Service
	calendarID    string
	calendarEmail string
}

// CalendarEvent 代表 Google 日曆事件
type CalendarEvent struct {
	ID          string
	Summary     string
	Description string
	Location    string
	StartTime   time.Time
	EndTime     time.Time
	Attendees   []string
}

// NewClient 創建新的 Google 日曆 API 客戶端
func NewClient(credentialsJSON []byte, calendarID string) (*Client, error) {
	ctx := context.Background()

	// 使用服務帳號憑證創建 OAuth2 配置
	config, err := google.JWTConfigFromJSON(credentialsJSON, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("無法解析服務帳號金鑰: %w", err)
	}

	// 創建帶有 OAuth2 客戶端的日曆服務
	client := config.Client(ctx)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("無法創建日曆服務: %w", err)
	}

	return &Client{
		service:       service,
		calendarID:    calendarID,
		calendarEmail: config.Email,
	}, nil
}

// CreateEvent 在 Google 日曆中創建事件
func (c *Client) CreateEvent(event *CalendarEvent) (string, error) {
	calEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
			TimeZone: "Asia/Taipei", // 設置為台灣時區，可根據需要調整
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
			TimeZone: "Asia/Taipei",
		},
	}

	// 加入參與者
	if len(event.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, len(event.Attendees))
		for i, email := range event.Attendees {
			attendees[i] = &calendar.EventAttendee{Email: email}
		}
		calEvent.Attendees = attendees
	}

	createdEvent, err := c.service.Events.Insert(c.calendarID, calEvent).Do()
	if err != nil {
		return "", fmt.Errorf("創建事件失敗: %w", err)
	}

	return createdEvent.Id, nil
}

// UpdateEvent 更新 Google 日曆中的事件
func (c *Client) UpdateEvent(eventID string, event *CalendarEvent) error {
	calEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
			TimeZone: "Asia/Taipei",
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
			TimeZone: "Asia/Taipei",
		},
	}

	// 加入參與者
	if len(event.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, len(event.Attendees))
		for i, email := range event.Attendees {
			attendees[i] = &calendar.EventAttendee{Email: email}
		}
		calEvent.Attendees = attendees
	}

	_, err := c.service.Events.Update(c.calendarID, eventID, calEvent).Do()
	if err != nil {
		return fmt.Errorf("更新事件失敗: %w", err)
	}

	return nil
}

// DeleteEvent 刪除 Google 日曆中的事件
func (c *Client) DeleteEvent(eventID string) error {
	err := c.service.Events.Delete(c.calendarID, eventID).Do()
	if err != nil {
		return fmt.Errorf("刪除事件失敗: %w", err)
	}

	return nil
}

// GetEvent 獲取特定 Google 日曆事件
func (c *Client) GetEvent(eventID string) (*CalendarEvent, error) {
	calEvent, err := c.service.Events.Get(c.calendarID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("獲取事件失敗: %w", err)
	}

	startTime, _ := time.Parse(time.RFC3339, calEvent.Start.DateTime)
	endTime, _ := time.Parse(time.RFC3339, calEvent.End.DateTime)

	event := &CalendarEvent{
		ID:          calEvent.Id,
		Summary:     calEvent.Summary,
		Description: calEvent.Description,
		Location:    calEvent.Location,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if calEvent.Attendees != nil {
		attendees := make([]string, len(calEvent.Attendees))
		for i, attendee := range calEvent.Attendees {
			attendees[i] = attendee.Email
		}
		event.Attendees = attendees
	}

	return event, nil
}

// FindEventByBookingCode 根據預約 ID 從描述中搜索事件
func (c *Client) FindEventByBookingCode(bookingCode string) (string, error) {
	// 搜尋描述中包含預約 ID 的事件
	query := fmt.Sprintf("%s", bookingCode)
	events, err := c.service.Events.List(c.calendarID).Q(query).Do()
	if err != nil {
		return "", fmt.Errorf("搜尋事件失敗: %w", err)
	}

	if len(events.Items) == 0 {
		return "", nil // 未找到事件
	}

	return events.Items[0].Id, nil
}
