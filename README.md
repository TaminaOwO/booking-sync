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

## 配置說明

### 本地開發配置

1. 從範例檔案創建您的配置：
   ```bash
   cp config.json.example config.json
   cp .env.example .env
   ```

2. 編輯這些文件，填入您的實際配置值

3. 獲取 Google Calendar API 憑證：
   - 進入 Google Cloud Console > API 和服務 > 憑證
   - 創建服務帳號並下載 JSON 密鑰
   - 將下載的 JSON 文件保存為 `google-credentials.json`

### 敏感資料處理

所有敏感配置都應使用環境變數或 Secret Manager 進行管理：

- `SIMPLYBOOK_COMPANY_LOGIN` - SimplyBook 公司登錄名
- `SIMPLYBOOK_API_KEY` - SimplyBook API 金鑰
- `GOOGLE_CALENDAR_CREDENTIALS_FILE` - Google 憑證文件路徑
- `GOOGLE_CALENDAR_ID` - Google 日曆 ID

**注意**：請勿將敏感配置提交到版本控制系統。檔案 `config.json`、`google-credentials.json` 和 `.env` 已加入 `.gitignore`。

## 部署到 Google Cloud

### 準備工作

1. 創建必要的 Secrets：
   ```bash
   # 為 SimplyBook API 金鑰建立 Secret
   gcloud secrets create simplybook-api-key --data-file=<(echo -n "YOUR_API_KEY")

   # 為 SimplyBook 公司登錄名建立 Secret
   gcloud secrets create simplybook-company-login --data-file=<(echo -n "YOUR_COMPANY_LOGIN")

   # 為 Google Calendar ID 建立 Secret
   gcloud secrets create google-calendar-id --data-file=<(echo -n "YOUR_CALENDAR_ID")

   # 為 Google 憑證建立 Secret
   gcloud secrets create google-calendar-creds --data-file=google-credentials.json
   ```

2. 授予 Cloud Run 服務帳號訪問權限：
   ```bash
   # 獲取服務帳號
   SERVICE_ACCOUNT=$(gcloud iam service-accounts list --filter="displayName:Cloud Run Service Agent" --format="value(email)")

   # 授予訪問權限
   for SECRET in simplybook-api-key simplybook-company-login google-calendar-id google-calendar-creds; do
     gcloud secrets add-iam-policy-binding $SECRET \
       --member="serviceAccount:$SERVICE_ACCOUNT" \
       --role="roles/secretmanager.secretAccessor"
   done
   ```

### 構建和部署

1. 構建容器：
   ```bash
   gcloud builds submit --tag gcr.io/booking-sync-455103/booking-sync .
   ```

2. 部署到 Cloud Run：
   ```bash
   # 設置您的專案 ID
   PROJECT_ID="your-project-id"  # 替換為您的實際專案 ID

   gcloud run deploy booking-sync \
     --image=gcr.io/${PROJECT_ID}/booking-sync \
     --platform=managed \
     --region=asia-east1 \
     --service-account=booking-sync-service@${PROJECT_ID}.iam.gserviceaccount.com \
     --set-env-vars="GOOGLE_CALENDAR_CREDENTIALS_FILE=/secrets/google-calendar-creds" \
     --update-secrets="\
/secrets/google-calendar-creds=google-calendar-creds:latest,\
SIMPLYBOOK_COMPANY_LOGIN=simplybook-company-login:latest,\
SIMPLYBOOK_API_KEY=simplybook-api-key:latest,\
GOOGLE_CALENDAR_ID=google-calendar-id:latest" \
     --project=${PROJECT_ID}
   ```

## 許可證

MIT 