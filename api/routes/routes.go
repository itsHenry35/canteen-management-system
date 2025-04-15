package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/itsHenry35/canteen-management-system/api/handlers"
	"github.com/itsHenry35/canteen-management-system/api/middlewares"
	"github.com/itsHenry35/canteen-management-system/services"
)

// SetupRouter 设置路由
func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// 应用日志中间件
	r.Use(middlewares.LogMiddleware)

	// API 路由
	api := r.PathPrefix("/api").Subrouter()

	// 公开API路由
	api.HandleFunc("/login", handlers.Login).Methods("POST")
	api.HandleFunc("/dingtalk/login", handlers.DingTalkLogin).Methods("POST")
	api.HandleFunc("/website_info", handlers.GetWebsiteInfo).Methods("GET")

	// 需要身份验证的API路由
	secured := api.PathPrefix("").Subrouter()
	secured.Use(middlewares.AuthMiddleware)

	// 管理员API路由
	adminAPI := secured.PathPrefix("/admin").Subrouter()
	adminAPI.Use(middlewares.RoleMiddleware(services.RoleAdmin))

	// 用户管理
	adminAPI.HandleFunc("/users", handlers.GetAllUsers).Methods("GET")
	adminAPI.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	adminAPI.HandleFunc("/users/{id:[0-9]+}", handlers.GetUser).Methods("GET")
	adminAPI.HandleFunc("/users/{id:[0-9]+}", handlers.UpdateUser).Methods("PUT")
	adminAPI.HandleFunc("/users/{id:[0-9]+}", handlers.DeleteUser).Methods("DELETE")

	// 学生管理
	adminAPI.HandleFunc("/students", handlers.GetAllStudents).Methods("GET")
	adminAPI.HandleFunc("/students", handlers.CreateStudent).Methods("POST")
	adminAPI.HandleFunc("/students/{id:[0-9]+}", handlers.GetStudent).Methods("GET")
	adminAPI.HandleFunc("/students/{id:[0-9]+}", handlers.UpdateStudent).Methods("PUT")
	adminAPI.HandleFunc("/students/{id:[0-9]+}", handlers.DeleteStudent).Methods("DELETE")
	adminAPI.HandleFunc("/students/{id:[0-9]+}/qrcode-data", handlers.GetStudentQRCodeData).Methods("GET")

	// 餐管理
	adminAPI.HandleFunc("/meals", handlers.GetAllMeals).Methods("GET")
	adminAPI.HandleFunc("/meals", handlers.CreateMeal).Methods("POST")
	adminAPI.HandleFunc("/meals/{id:[0-9]+}", handlers.GetMeal).Methods("GET")
	adminAPI.HandleFunc("/meals/{id:[0-9]+}", handlers.UpdateMeal).Methods("PUT")
	adminAPI.HandleFunc("/meals/{id:[0-9]+}", handlers.DeleteMeal).Methods("DELETE")
	adminAPI.HandleFunc("/meals/{id:[0-9]+}/selections", handlers.GetMealSelections).Methods("GET")
	adminAPI.HandleFunc("/meals/cleanup", handlers.CleanupExpiredMeals).Methods("POST")

	// 选餐管理
	adminAPI.HandleFunc("/selections", handlers.GetStudentSelections).Methods("GET")
	adminAPI.HandleFunc("/selections/batch", handlers.BatchSelectMeals).Methods("POST")
	adminAPI.HandleFunc("/notify/unselected", handlers.NotifyUnselectedStudents).Methods("POST")

	// 系统设置
	adminAPI.HandleFunc("/settings", handlers.GetSettings).Methods("GET")
	adminAPI.HandleFunc("/settings", handlers.UpdateSettings).Methods("PUT")

	// 危险API
	adminAPI.HandleFunc("/rebuild-mapping", handlers.RebuildParentStudentMapping).Methods("POST")
	// 添加获取重建映射日志的API
	adminAPI.HandleFunc("/rebuild-mapping/logs", handlers.GetMappingLogs).Methods("GET")

	// 食堂工作人员API路由
	canteenAPI := secured.PathPrefix("/canteen").Subrouter()
	canteenAPI.Use(middlewares.RoleMiddleware(services.RoleCanteenA, services.RoleCanteenB, services.RoleCanteenTest))

	// 扫码取餐 (现在包含记录取餐功能)
	canteenAPI.HandleFunc("/scan", handlers.ScanStudentQRCode).Methods("POST")

	// 学生API路由
	studentAPI := secured.PathPrefix("/student").Subrouter()
	studentAPI.Use(middlewares.RoleMiddleware(services.RoleStudent))

	// 选餐
	studentAPI.HandleFunc("/meals/current", handlers.GetCurrentSelectableMeals).Methods("GET")
	studentAPI.HandleFunc("/selection", handlers.GetStudentMealSelections).Methods("GET")
	studentAPI.HandleFunc("/selection", handlers.StudentSelectMeal).Methods("POST")
	studentAPI.HandleFunc("/selection/current", handlers.GetStudentCurrentSelection).Methods("GET")

	// 静态文件服务
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// 所有其他请求都指向前端入口点
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	return r
}
