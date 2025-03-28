package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/async-booking/pkg/gcalendar"
	"github.com/yourusername/async-booking/pkg/simplybook"
)

// WebhookHandler 處理 SimplyBook webhook 通知
type WebhookHandler struct {
	simplybookClient *simplybook.Client
	calendarClient   *gcalendar.Client
	secretToken      string // 可選的安全令牌，用於驗證請求
}

// NewWebhookHandler 創建新的 webhook 處理器
func NewWebhookHandler(simplybookClient *simplybook.Client, calendarClient *gcalendar.Client, secretToken string) *WebhookHandler {
	return &WebhookHandler{
		simplybookClient: simplybookClient,
		calendarClient:   calendarClient,
		secretToken:      secretToken,
	}
}

// HandleWebhook 處理傳入的 webhook 請求
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// 驗證請求方法
	if r.Method != http.MethodPost {
		http.Error(w, "僅支持 POST 請求", http.StatusMethodNotAllowed)
		return
	}

	// 驗證令牌（如果已設置）
	if h.secretToken != "" {
		token := r.Header.Get("X-Simplybook-Token")
		if token != h.secretToken {
			http.Error(w, "未授權", http.StatusUnauthorized)
			return
		}
	}

	// 讀取並解析請求體
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "讀取請求體失敗", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload simplybook.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "無效的 JSON 數據", http.StatusBadRequest)
		return
	}

	// 處理 webhook 事件（非同步處理，避免超時）
	go func() {
		if err := h.processWebhookEvent(&payload); err != nil {
			log.Printf("處理 webhook 事件失敗: %v", err)
		}
	}()

	// 立即返回成功
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("webhook 已接收"))
}

// processWebhookEvent 處理 webhook 事件並更新 Google 日曆
func (h *WebhookHandler) processWebhookEvent(payload *simplybook.WebhookPayload) error {
	log.Printf("處理 %s 操作，預約 ID: %s", payload.Action, payload.BookingID)

	switch strings.ToLower(payload.Action) {
	case "create":
		return h.handleBookingCreated(payload.BookingID)
	case "update":
		return h.handleBookingUpdated(payload.BookingID)
	case "cancel":
		return h.handleBookingDeleted(payload.BookingID)
	default:
		return fmt.Errorf("不支持的操作類型: %s", payload.Action)
	}
}

// handleBookingCreated 處理新預約創建
func (h *WebhookHandler) handleBookingCreated(bookingID string) error {
	// 獲取預約詳情
	booking, err := h.simplybookClient.GetBooking(bookingID)
	if err != nil {
		return fmt.Errorf("獲取預約詳情失敗: %w", err)
	}

	// 創建日曆事件
	calEvent := createCalendarEventFromBooking(booking)
	eventID, err := h.calendarClient.CreateEvent(calEvent)
	if err != nil {
		return fmt.Errorf("創建日曆事件失敗: %w", err)
	}

	log.Printf("為預約 %s 創建了日曆事件 %s", bookingID, eventID)
	return nil
}

// handleBookingUpdated 處理預約更新
func (h *WebhookHandler) handleBookingUpdated(bookingID string) error {
	// 獲取預約詳情
	booking, err := h.simplybookClient.GetBooking(bookingID)
	if err != nil {
		return fmt.Errorf("獲取預約詳情失敗: %w", err)
	}

	// 查找現有的日曆事件
	eventID, err := h.calendarClient.FindEventByBookingID(bookingID)
	if err != nil {
		return fmt.Errorf("查找日曆事件失敗: %w", err)
	}

	if eventID == "" {
		// 事件不存在，創建新事件
		return h.handleBookingCreated(bookingID)
	}

	// 更新日曆事件
	calEvent := createCalendarEventFromBooking(booking)
	if err := h.calendarClient.UpdateEvent(eventID, calEvent); err != nil {
		return fmt.Errorf("更新日曆事件失敗: %w", err)
	}

	log.Printf("已更新預約 %s 的日曆事件 %s", bookingID, eventID)
	return nil
}

// handleBookingDeleted 處理預約刪除
func (h *WebhookHandler) handleBookingDeleted(bookingID string) error {
	// 查找現有的日曆事件
	eventID, err := h.calendarClient.FindEventByBookingID(bookingID)
	if err != nil {
		return fmt.Errorf("查找日曆事件失敗: %w", err)
	}

	if eventID == "" {
		// 事件不存在，無需操作
		log.Printf("未找到預約 %s 的日曆事件", bookingID)
		return nil
	}

	// 刪除日曆事件
	if err := h.calendarClient.DeleteEvent(eventID); err != nil {
		return fmt.Errorf("刪除日曆事件失敗: %w", err)
	}

	log.Printf("已刪除預約 %s 的日曆事件 %s", bookingID, eventID)
	return nil
}

// createCalendarEventFromBooking 從預約信息創建日曆事件
func createCalendarEventFromBooking(booking *simplybook.Booking) *gcalendar.CalendarEvent {
	// 創建事件描述，包含預約詳情
	description := fmt.Sprintf(
		"預約編號: %s\n客戶: %s\n電話: %s\n電子郵件: %s\n備註: %s\nBookingID: %s",
		booking.Code,
		booking.ClientName,
		booking.ClientPhone,
		booking.ClientEmail,
		booking.Notes,
		booking.ID,
	)

	// 創建事件標題
	summary := fmt.Sprintf("%s - %s", booking.ServiceName, booking.ClientName)

	// 設置參與者（如果有電子郵件）
	var attendees []string
	if booking.ClientEmail != "" {
		attendees = append(attendees, booking.ClientEmail)
	}

	return &gcalendar.CalendarEvent{
		Summary:     summary,
		Description: description,
		StartTime:   booking.StartTime,
		EndTime:     booking.EndTime,
		Attendees:   attendees,
	}
} 