# SimplyBook to Google Calendar 同步服務

這是一個基於 Go 的服務，用於監聽 SimplyBook 的預約更新，並將其同步到 Google 日曆中。

## 功能

- 接收 SimplyBook 的 webhook 通知
- 根據通知類型（創建/更新/刪除）查詢預約詳情
- 在 Google 日曆中同步創建/更新/刪除相應的事件

## 架構

服務主要由以下幾個部分組成：

1. **SimplyBook API 客戶端**：與 SimplyBook API 通信，獲取預約信息
2. **Google 日曆 API 客戶端**：管理 Google 日曆事件
3. **Webhook 處理器**：處理 SimplyBook 發送的通知
4. **配置管理**：管理服務配置和憑證

## 安裝與設置

### 前置條件

- Go 1.19 或更高版本
- SimplyBook API 身份驗證（公司登錄名和 API 金鑰）
- Google 日曆 API 服務帳號憑證

### 安裝步驟

1. 克隆倉庫

```bash
git clone https://github.com/yourusername/async-booking.git
cd async-booking
```

2. 安裝依賴

```bash
go mod download
```

3. 創建配置文件 `config.json`

```json
{
  "server": {
    "port": 8080,
    "webhook_path": "/webhook"
  },
  "simplybook": {
    "company_login": "your-company-login",
    "api_key": "your-api-key"
  },
  "google_calendar": {
    "credentials_file": "./google-credentials.json",
    "calendar_id": "your-calendar-id@group.calendar.google.com"
  }
}
```

4. 準備 Google 服務帳號憑證

- 將您的 Google 服務帳號憑證 JSON 文件保存為 `google-credentials.json`
- 確保服務帳號已被授予適當的權限來訪問目標日曆

### 運行服務

```bash
go run cmd/server/main.go -config=./config.json
```

或者使用環境變數：

```bash
export SERVER_PORT=8080
export WEBHOOK_PATH="/webhook"
export SIMPLYBOOK_COMPANY_LOGIN="your-company-login"
export SIMPLYBOOK_API_KEY="your-api-key"
export GOOGLE_CALENDAR_CREDENTIALS_FILE="./google-credentials.json"
export GOOGLE_CALENDAR_ID="your-calendar-id@group.calendar.google.com"

go run cmd/server/main.go
```

## 在 SimplyBook 配置 Webhook

1. 登錄 SimplyBook 管理面板
2. 設置 webhook 指向您的服務 URL（例如：`https://your-domain.com/webhook`）

## 部署到 GCP

### 使用 Cloud Run 部署

1. 構建 Docker 映像

```bash
docker build -t gcr.io/your-project-id/simplybook-gcal-sync .
```

2. 推送映像到 Google Container Registry

```bash
docker push gcr.io/your-project-id/simplybook-gcal-sync
```

3. 部署到 Cloud Run

```bash
gcloud run deploy simplybook-gcal-sync \
  --image gcr.io/your-project-id/simplybook-gcal-sync \
  --platform managed \
  --allow-unauthenticated \
  --set-env-vars="SIMPLYBOOK_COMPANY_LOGIN=your-company-login,SIMPLYBOOK_API_KEY=your-api-key,GOOGLE_CALENDAR_ID=your-calendar-id"
```

注意：對於 Google 服務帳號憑證，建議使用 GCP 的 Secret Manager。

## 許可證

MIT 