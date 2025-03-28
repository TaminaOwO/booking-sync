FROM golang:1.19-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 構建應用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o /simplybook-gcal-sync ./cmd/server

# 使用輕量級的 alpine 鏡像
FROM alpine:3.16

# 安裝 CA 證書，用於 HTTPS 請求
RUN apk --no-cache add ca-certificates tzdata

# 設置默認時區為亞洲/台北
ENV TZ=Asia/Taipei

# 從構建器複製編譯好的二進制文件
COPY --from=builder /simplybook-gcal-sync /simplybook-gcal-sync

# 設置應用程序運行的用戶
RUN adduser -D -H -h /app appuser
USER appuser

# 設置健康檢查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:${SERVER_PORT:-8080}/health || exit 1

# 暴露端口
EXPOSE 8080

# 運行應用程序
ENTRYPOINT ["/simplybook-gcal-sync"] 