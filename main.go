package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itsHenry35/canteen-management-system/api/routes"
	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/database"
	"github.com/itsHenry35/canteen-management-system/scheduler" // 导入新的scheduler包
)

//go:embed all:static
var staticFiles embed.FS

// 获取嵌入的静态文件系统
func getStaticFS() (fs.FS, error) {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, err
	}
	return staticFS, nil
}

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化定时任务
	if err := scheduler.Initialize(); err != nil {
		log.Fatalf("Failed to initialize scheduler: %v", err)
	}

	// 确保餐食图片目录存在
	mealImgDir := "./data/images"
	if err := os.MkdirAll(mealImgDir, 0755); err != nil {
		log.Fatalf("Failed to create images directory: %v", err)
	}

	// 获取嵌入的静态文件系统
	staticFS, err := getStaticFS()
	if err != nil {
		log.Fatalf("Failed to get static files filesystem: %v", err)
	}

	// 设置路由，传入静态文件系统
	router := routes.SetupRouter(staticFS)

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

	// 停止定时任务
	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
