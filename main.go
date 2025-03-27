package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itsHenry35/canteen-management-system/api/routes"
	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/database"
)

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置路由
	router := routes.SetupRouter()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Get().Server.Port),
		Handler: router,
	}

	// 启动服务器（非阻塞）
	go func() {
		log.Printf("Server is running on port %d", config.Get().Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
