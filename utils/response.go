package utils

import (
	"encoding/json"
	"net/http"
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
