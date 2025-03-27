package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response HTTP响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseOK 成功响应
func ResponseOK(w http.ResponseWriter, data interface{}) {
	response := Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
	JSON(w, http.StatusOK, response)
}

// ResponseError 错误响应
func ResponseError(w http.ResponseWriter, httpStatus int, message string) {
	response := Response{
		Code:    httpStatus,
		Message: message,
	}
	// 修改：始终返回200状态码，但在响应体中保留原始错误码
	JSON(w, http.StatusOK, response)
}

// JSON 返回JSON格式响应
func JSON(w http.ResponseWriter, httpStatus int, data interface{}) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// 编码响应数据
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetCurrentWeekDay 获取当前的周数和星期几
func GetCurrentWeekDay() (week, day int) {
	now := time.Now()

	// 计算当前是一年中的第几周
	year, weekNum := now.ISOWeek()
	week = year*100 + weekNum // 例如: 202510

	// 获取当前是星期几 (1-7，代表周一到周日)
	day = int(now.Weekday())
	if day == 0 {
		day = 7 // 将周日从0改为7
	}

	return week, day
}
