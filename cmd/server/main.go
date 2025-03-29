package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/booking-sync-455103/booking-sync/config"
	"github.com/booking-sync-455103/booking-sync/pkg/gcalendar"
	"github.com/booking-sync-455103/booking-sync/pkg/handler"
	"github.com/booking-sync-455103/booking-sync/pkg/simplybook"
)

func main() {
	// 解析命令行參數
	configPath := flag.String("config", "", "配置文件路徑")
	flag.Parse()

	// 如果沒有指定配置文件，則使用環境變數
	if *configPath == "" {
		if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
			*configPath = envPath
			log.Printf("使用環境變數中的配置文件路徑: %s", *configPath)
		}
	}

	// 加載配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加載配置失敗: %v", err)
	}

	// 初始化 SimplyBook 客戶端
	simplybookClient, err := simplybook.NewClient(
		cfg.SimplyBook.CompanyLogin,
		cfg.SimplyBook.UserName,
		cfg.SimplyBook.Password,
	)
	if err != nil {
		log.Fatalf("初始化 SimplyBook 客戶端失敗: %v", err)
	}

	// 載入 Google 服務帳號憑證
	googleCreds, err := config.LoadGoogleCredentials(cfg.GoogleCalendar.CredentialsFile)
	if err != nil {
		log.Fatalf("載入 Google 憑證失敗: %v", err)
	}

	// 初始化 Google 日曆客戶端
	calendarClient, err := gcalendar.NewClient(googleCreds, cfg.GoogleCalendar.CalendarID)
	if err != nil {
		log.Fatalf("初始化 Google 日曆客戶端失敗: %v", err)
	}

	// 創建 webhook 處理器
	webhookHandler := handler.NewWebhookHandler(
		simplybookClient,
		calendarClient,
		"",
	)

	// 設置 HTTP 路由
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.Server.WebhookPath, webhookHandler.HandleWebhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("服務正常運行中"))
	})

	// 優先使用環境變數 PORT
	port := cfg.Server.Port
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			port = p
			log.Printf("使用環境變數 PORT 設置的端口: %d", port)
		} else {
			log.Printf("無效的 PORT 環境變數值: %s，使用配置端口: %d", portEnv, port)
		}
	}

	// 設置伺服器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// 設置優雅關閉的處理
	go func() {
		// 等待中斷信號
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("關閉伺服器...")

		// 創建關閉伺服器的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("強制關閉伺服器: %v", err)
		}

		log.Println("伺服器已優雅關閉")
	}()

	// 直接啟動伺服器（不在 goroutine 中）
	log.Printf("伺服器正在監聽端口 %d...", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}
