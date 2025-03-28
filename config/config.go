package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config 包含應用程式配置
type Config struct {
	Server struct {
		Port            int    `json:"port"`
		WebhookPath     string `json:"webhook_path"`
	} `json:"server"`
	
	SimplyBook struct {
		CompanyLogin string `json:"company_login"`
		APIKey       string `json:"api_key"`
	} `json:"simplybook"`
	
	GoogleCalendar struct {
		CredentialsFile string `json:"credentials_file"`
		CalendarID      string `json:"calendar_id"`
	} `json:"google_calendar"`
}

// LoadConfig 從文件或環境變量加載配置
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// 如果提供了配置文件路徑，則從文件加載
	if configPath != "" {
		file, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("讀取配置文件失敗: %w", err)
		}
		
		if err := json.Unmarshal(file, config); err != nil {
			return nil, fmt.Errorf("解析配置文件失敗: %w", err)
		}
	}

	// 從環境變數讀取配置，優先於文件配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			config.Server.Port = p
		}
	}
	
	if path := os.Getenv("WEBHOOK_PATH"); path != "" {
		config.Server.WebhookPath = path
	}
	
	if login := os.Getenv("SIMPLYBOOK_COMPANY_LOGIN"); login != "" {
		config.SimplyBook.CompanyLogin = login
	}
	
	if apiKey := os.Getenv("SIMPLYBOOK_API_KEY"); apiKey != "" {
		config.SimplyBook.APIKey = apiKey
	}
	
	if credsFile := os.Getenv("GOOGLE_CALENDAR_CREDENTIALS_FILE"); credsFile != "" {
		config.GoogleCalendar.CredentialsFile = credsFile
	}
	
	if calID := os.Getenv("GOOGLE_CALENDAR_ID"); calID != "" {
		config.GoogleCalendar.CalendarID = calID
	}

	// 設置默認值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	
	if config.Server.WebhookPath == "" {
		config.Server.WebhookPath = "/webhook"
	}

	// 驗證必要的配置項
	if config.SimplyBook.CompanyLogin == "" {
		return nil, fmt.Errorf("缺少 SimplyBook 公司登錄名")
	}
	
	if config.SimplyBook.APIKey == "" {
		return nil, fmt.Errorf("缺少 SimplyBook API 金鑰")
	}
	
	if config.GoogleCalendar.CredentialsFile == "" {
		return nil, fmt.Errorf("缺少 Google 日曆憑證文件")
	}
	
	if config.GoogleCalendar.CalendarID == "" {
		return nil, fmt.Errorf("缺少 Google 日曆 ID")
	}

	return config, nil
}

// LoadGoogleCredentials 加載 Google 服務帳號憑證
func LoadGoogleCredentials(credentialsPath string) ([]byte, error) {
	// 解析路徑
	absPath, err := filepath.Abs(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("解析憑證文件路徑失敗: %w", err)
	}
	
	// 讀取憑證文件
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("讀取 Google 服務帳號憑證失敗: %w", err)
	}
	
	return data, nil
} 