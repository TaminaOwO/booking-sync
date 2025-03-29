package simplybook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 代表 SimplyBook API 客戶端
type Client struct {
	CompanyLogin string
	Username     string // 用於 REST API 的用戶名
	Password     string // 用於 REST API 的密碼
	Token        string
	BaseURL      string
	HTTPClient   *http.Client
}

// TokenResponse 認證響應
type TokenResponse struct {
	Token string `json:"token"`
}

// NewClient 創建新的 SimplyBook API 客戶端
func NewClient(companyLogin, username, password string) (*Client, error) {
	client := &Client{
		CompanyLogin: companyLogin,
		Username:     username,
		Password:     password,
		BaseURL:      "https://user-api-v2.simplybook.me",
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}

	// 獲取認證令牌
	if err := client.authenticate(); err != nil {
		return nil, err
	}

	return client, nil
}

// 進行 API 認證並獲取令牌
func (c *Client) authenticate() error {
	url := fmt.Sprintf("%s/admin/auth", c.BaseURL)

	// 根據 CURL 範例準備認證請求
	authRequest := map[string]string{
		"company":  c.CompanyLogin,
		"login":    c.Username,
		"password": c.Password,
	}

	requestData, err := json.Marshal(authRequest)
	if err != nil {
		return fmt.Errorf("序列化認證請求失敗: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("創建認證請求失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("執行認證請求失敗: %w", err)
	}
	defer resp.Body.Close()

	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("讀取認證響應失敗: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("認證失敗，狀態碼: %d, 響應: %s", resp.StatusCode, string(body))
	}

	var response TokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析認證響應失敗: %w", err)
	}

	if response.Token == "" {
		return fmt.Errorf("認證失敗: 未收到令牌")
	}

	c.Token = response.Token
	return nil
}

// doRequest 執行 REST API 請求
func (c *Client) doRequest(method, endpoint string, requestBody interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	var body io.Reader
	if requestBody != nil {
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("序列化請求失敗: %w", err)
		}
		body = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 根據 CURL 範例設置請求頭
	req.Header.Set("X-Token", c.Token)
	req.Header.Set("X-Company-Login", c.CompanyLogin)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("執行請求失敗: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取響應失敗: %w", err)
	}

	// 檢查是否是未授權錯誤（令牌可能過期）
	if resp.StatusCode == http.StatusUnauthorized {
		// 嘗試重新認證
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("令牌過期，重新認證失敗: %w", err)
		}

		// 使用新令牌重試請求
		return c.retryRequest(method, endpoint, requestBody)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API請求失敗，狀態碼: %d, 響應: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// retryRequest 使用新令牌重試請求
func (c *Client) retryRequest(method, endpoint string, requestBody interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	var body io.Reader
	if requestBody != nil {
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("序列化請求失敗: %w", err)
		}
		body = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("重試時創建請求失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 根據 CURL 範例設置請求頭
	req.Header.Set("X-Token", c.Token)
	req.Header.Set("X-Company-Login", c.CompanyLogin)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("重試請求執行失敗: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取重試響應失敗: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("重試API請求失敗，狀態碼: %d, 響應: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetBooking 獲取預約詳情
func (c *Client) GetBooking(bookingID string) (*Booking, error) {
	endpoint := fmt.Sprintf("/admin/bookings/%s", bookingID)

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("獲取預約失敗: %w", err)
	}

	var booking Booking
	if err := json.Unmarshal(respBody, &booking); err != nil {
		return nil, fmt.Errorf("解析預約數據失敗: %w", err)
	}

	return &booking, nil
}

// GetServiceList 獲取服務列表
func (c *Client) GetServiceList() (map[string]Service, error) {
	endpoint := "/admin/services"

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("獲取服務列表失敗: %w", err)
	}

	var services map[string]Service
	if err := json.Unmarshal(respBody, &services); err != nil {
		return nil, fmt.Errorf("解析服務列表失敗: %w", err)
	}

	return services, nil
}

// GetProviderList 獲取服務提供者列表
func (c *Client) GetProviderList() (map[string]Provider, error) {
	endpoint := "/admin/providers"

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("獲取服務提供者列表失敗: %w", err)
	}

	var providers map[string]Provider
	if err := json.Unmarshal(respBody, &providers); err != nil {
		return nil, fmt.Errorf("解析服務提供者列表失敗: %w", err)
	}

	return providers, nil
}
