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
	calEvent, err := c.prepareCalendarEvent(event)
	if err != nil {
		return "", fmt.Errorf("準備日曆事件失敗: %w", err)
	}

	createdEvent, err := c.service.Events.Insert(c.calendarID, calEvent).Do()
	if err != nil {
		return "", fmt.Errorf("創建事件失敗: %w", err)
	}

	return createdEvent.Id, nil
}

// UpdateEvent 更新 Google 日曆中的事件
func (c *Client) UpdateEvent(eventID string, event *CalendarEvent) error {
	calEvent, err := c.prepareCalendarEvent(event)
	if err != nil {
		return fmt.Errorf("準備日曆事件失敗: %w", err)
	}

	_, err = c.service.Events.Update(c.calendarID, eventID, calEvent).Do()
	if err != nil {
		return fmt.Errorf("更新事件失敗: %w", err)
	}

	return nil
}

// prepareCalendarEvent 準備要發送給 Google Calendar API 的事件物件
func (c *Client) prepareCalendarEvent(event *CalendarEvent) (*calendar.Event, error) {
	// 獲取台灣時區
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		loc = time.FixedZone("GMT+8", 8*60*60)
	}

	// 確保時間是台灣時區的
	startTime := event.StartTime.In(loc)
	endTime := event.EndTime.In(loc)

	// 格式化為不帶時區信息的時間格式
	startDateTime := startTime.Format("2006-01-02T15:04:05")
	endDateTime := endTime.Format("2006-01-02T15:04:05")

	calEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Start: &calendar.EventDateTime{
			DateTime: startDateTime,
			TimeZone: "Asia/Taipei", // 明確指定台灣時區
		},
		End: &calendar.EventDateTime{
			DateTime: endDateTime,
			TimeZone: "Asia/Taipei", // 明確指定台灣時區
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

	return calEvent, nil
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

// FindEventByBookingCode 根據預約編號從描述中搜索事件
func (c *Client) FindEventByBookingCode(bookingCode string) (string, error) {
	// 搜尋描述中包含預約 Code 的事件
	query := bookingCode
	events, err := c.service.Events.List(c.calendarID).Q(query).Do()
	if err != nil {
		return "", fmt.Errorf("搜尋事件失敗: %w", err)
	}

	if len(events.Items) == 0 {
		return "", nil // 未找到事件
	}

	return events.Items[0].Id, nil
}
