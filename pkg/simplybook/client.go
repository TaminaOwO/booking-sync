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
	APIKey       string
	Token        string
	BaseURL      string
	HTTPClient   *http.Client
}

// 認證響應
type TokenResponse struct {
	Result string `json:"result"`
}

// NewClient 創建新的 SimplyBook API 客戶端
func NewClient(companyLogin, apiKey string) (*Client, error) {
	client := &Client{
		CompanyLogin: companyLogin,
		APIKey:       apiKey,
		BaseURL:      "https://user-api.simplybook.me",
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
	url := "https://user-api.simplybook.me/login"
	
	// 準備 JSON-RPC 請求
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getToken",
		"params":  []interface{}{c.CompanyLogin, c.APIKey},
		"id":      1,
	}
	
	requestData, err := json.Marshal(requestBody)
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
	
	var response struct {
		Result string `json:"result"`
		Error  struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析認證響應失敗: %w", err)
	}
	
	if response.Error.Message != "" {
		return fmt.Errorf("認證錯誤: %s", response.Error.Message)
	}
	
	c.Token = response.Result
	return nil
}

// 執行通用 API 請求
func (c *Client) doRequest(method string, params []interface{}) (json.RawMessage, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}
	
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化請求失敗: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Company-Login", c.CompanyLogin)
	req.Header.Set("X-Token", c.Token)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("執行請求失敗: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取響應失敗: %w", err)
	}
	
	var response struct {
		Result json.RawMessage `json:"result"`
		Error  struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析響應失敗: %w", err)
	}
	
	if response.Error.Message != "" {
		return nil, fmt.Errorf("API 錯誤: %s", response.Error.Message)
	}
	
	return response.Result, nil
}

// GetBooking 獲取預約詳情
func (c *Client) GetBooking(bookingID string) (*Booking, error) {
	result, err := c.doRequest("getBooking", []interface{}{bookingID})
	if err != nil {
		return nil, fmt.Errorf("獲取預約失敗: %w", err)
	}
	
	var booking Booking
	if err := json.Unmarshal(result, &booking); err != nil {
		return nil, fmt.Errorf("解析預約數據失敗: %w", err)
	}
	
	return &booking, nil
}

// GetServiceList 獲取服務列表
func (c *Client) GetServiceList() (map[string]Service, error) {
	result, err := c.doRequest("getEventList", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("獲取服務列表失敗: %w", err)
	}
	
	var services map[string]Service
	if err := json.Unmarshal(result, &services); err != nil {
		return nil, fmt.Errorf("解析服務列表失敗: %w", err)
	}
	
	return services, nil
}

// GetProviderList 獲取服務提供者列表
func (c *Client) GetProviderList() (map[string]Provider, error) {
	result, err := c.doRequest("getUnitList", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("獲取服務提供者列表失敗: %w", err)
	}
	
	var providers map[string]Provider
	if err := json.Unmarshal(result, &providers); err != nil {
		return nil, fmt.Errorf("解析服務提供者列表失敗: %w", err)
	}
	
	return providers, nil
} 