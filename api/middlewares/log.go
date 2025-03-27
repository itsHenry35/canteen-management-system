package middlewares

import (
	"log"
	"net/http"
	"time"
)

// LogMiddleware 日志中间件
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 记录请求信息
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		// 调用下一个处理程序
		next.ServeHTTP(w, r)

		// 记录响应信息
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
