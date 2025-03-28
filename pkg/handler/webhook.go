package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/booking-sync-455103/booking-sync/pkg/gcalendar"
	"github.com/booking-sync-455103/booking-sync/pkg/simplybook"
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

	// 記錄原始的請求數據，以便查看資料格式
	log.Printf("收到 webhook 請求，原始數據: %s", string(body))

	var payload simplybook.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "無效的 JSON 數據", http.StatusBadRequest)
		return
	}

	// 記錄解析後的資料結構
	log.Printf("解析後的資料: Action=%s, BookingID=%s", payload.Action, payload.BookingID)

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
	
	// 先獲取預約詳情和對應的日曆事件ID
	booking, eventID, err := h.getBookingAndEvent(payload.BookingID)
	if err != nil {
		return err
	}
	
	action := strings.ToLower(payload.Action)
	
	// 根據操作類型處理
	switch action {
	case "create":
		return h.handleBookingCreated(booking, eventID, payload.BookingID)
	case "update":
		return h.handleBookingUpdated(booking, eventID, payload.BookingID)
	case "cancel":
		return h.handleBookingDeleted(eventID, payload.BookingID)
	default:
		return fmt.Errorf("不支持的操作類型: %s", payload.Action)
	}
}

// getBookingAndEvent 獲取預約詳情和對應的日曆事件ID（如存在）
func (h *WebhookHandler) getBookingAndEvent(bookingID string) (*simplybook.Booking, string, error) {
	// 獲取預約詳情
	booking, err := h.simplybookClient.GetBooking(bookingID)
	if err != nil {
		return nil, "", fmt.Errorf("獲取預約詳情失敗: %w", err)
	}

	// 查找現有的日曆事件
	eventID, err := h.calendarClient.FindEventByBookingCode(booking.Code)
	if err != nil {
		return booking, "", fmt.Errorf("查找日曆事件失敗: %w", err)
	}

	return booking, eventID, nil
}

// handleBookingCreated 處理新預約創建
func (h *WebhookHandler) handleBookingCreated(booking *simplybook.Booking, eventID, bookingID string) error {
	// 如果已經存在事件，則不需要再創建
	if eventID != "" {
		log.Printf("預約 %s 的日曆事件已存在 %s", bookingID, eventID)
		return nil
	}

	// 創建日曆事件
	calEvent := createCalendarEventFromBooking(booking)
	newEventID, err := h.calendarClient.CreateEvent(calEvent)
	if err != nil {
		return fmt.Errorf("創建日曆事件失敗: %w", err)
	}

	log.Printf("為預約 %s 創建了日曆事件 %s", bookingID, newEventID)
	return nil
}

// handleBookingUpdated 處理預約更新
func (h *WebhookHandler) handleBookingUpdated(booking *simplybook.Booking, eventID, bookingID string) error {
	if eventID == "" {
		// 事件不存在，創建新事件
		calEvent := createCalendarEventFromBooking(booking)
		newEventID, err := h.calendarClient.CreateEvent(calEvent)
		if err != nil {
			return fmt.Errorf("創建日曆事件失敗: %w", err)
		}
		log.Printf("為更新的預約 %s 創建了新的日曆事件 %s", bookingID, newEventID)
		return nil
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
func (h *WebhookHandler) handleBookingDeleted(eventID, bookingID string) error {
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
	description := booking.Code

	// 創建事件標題
	summary := booking.ClientName

	// 設置參與者（如果有電子郵件）
	// var attendees []string
	// if booking.ClientEmail != "" {
	// 	attendees = append(attendees, booking.ClientEmail)
	// }

	return &gcalendar.CalendarEvent{
		Summary:     summary,
		Description: description,
		StartTime:   booking.StartTime,
		EndTime:     booking.EndTime,
		// Attendees:   attendees,
	}
} 