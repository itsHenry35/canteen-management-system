package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/itsHenry35/canteen-management-system/api/middlewares"
	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/models"
	"github.com/itsHenry35/canteen-management-system/scheduler"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	FullName   string      `json:"full_name"`
	Role       models.Role `json:"role"`
	DingTalkID string      `json:"dingtalk_id,omitempty"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	FullName   string `json:"full_name,omitempty"`
	Password   string `json:"password,omitempty"`
	DingTalkID string `json:"dingtalk_id,omitempty"`
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	DingTalk struct {
		AppKey    string `json:"app_key"`
		AppSecret string `json:"app_secret"`
		AgentID   string `json:"agent_id"`
		CorpID    string `json:"corp_id"`
	} `json:"dingtalk"`
	Website struct {
		Name           string `json:"name"`
		ICPBeian       string `json:"icp_beian"`
		PublicSecBeian string `json:"public_sec_beian"`
		Domain         string `json:"domain"`
	} `json:"website"`
	Scheduler struct {
		Enabled                bool   `json:"enabled"`
		CleanupTime            string `json:"cleanup_time"`
		ReminderBeforeEndHours int    `json:"reminder_before_end_hours"`
	} `json:"scheduler"`
}

// NotifyUnselectedStudentsRequest 提醒未选餐学生请求
type NotifyUnselectedStudentsRequest struct {
	MealID int `json:"meal_id"`
}

// GetAllUsers 获取所有用户
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	RoleStr := r.URL.Query().Get("role")
	var Role models.Role
	if RoleStr != "" {
		Role = models.Role(RoleStr)
	}

	// 获取用户列表
	users, err := models.GetAllUsers(Role)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "failed to get users")
		return
	}

	// 返回响应
	utils.ResponseOK(w, users)
}

// GetUser 获取用户信息
func GetUser(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// 获取用户信息
	user, err := models.GetUserByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "系统中未找到用户")
		return
	}

	// 返回响应
	utils.ResponseOK(w, user)
}

// CreateUser 创建用户
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// 验证用户类型
	if req.Role != models.RoleAdmin && req.Role != models.RoleCanteenA && req.Role != models.RoleCanteenB && req.Role != models.RoleCanteenTest {
		utils.ResponseError(w, http.StatusBadRequest, "invalid user type")
		return
	}

	// 创建用户
	if req.DingTalkID == "" {
		req.DingTalkID = "0"
	}
	user, err := models.CreateUser(req.Username, req.Password, req.FullName, req.Role, req.DingTalkID)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// 返回响应
	utils.ResponseOK(w, user)
}

// UpdateUser 更新用户信息
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// 解析请求
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// 获取用户信息
	user, err := models.GetUserByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "系统中未找到用户")
		return
	}

	// 更新用户信息
	needUpdate := false
	if req.FullName != "" {
		user.FullName = req.FullName
		needUpdate = true
	}

	if req.DingTalkID != "" {
		user.DingTalkID = req.DingTalkID
		needUpdate = true
	}

	if needUpdate {
		if err := models.UpdateUser(user); err != nil {
			utils.ResponseError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
	}

	// 更新密码
	if req.Password != "" {
		if err := models.UpdatePassword(id, req.Password); err != nil {
			utils.ResponseError(w, http.StatusInternalServerError, "failed to update password")
			return
		}
	}

	// 返回响应
	utils.ResponseOK(w, user)
}

// DeleteUser 删除用户
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// 获取当前登录用户ID
	currentUserID, ok := middlewares.GetUserIDFromContext(r)
	if !ok {
		utils.ResponseError(w, http.StatusUnauthorized, "未授权")
		return
	}

	// 检查是否尝试删除自己
	if id == currentUserID {
		utils.ResponseError(w, http.StatusForbidden, "不能删除自己的账户")
		return
	}

	// 删除用户
	if err := models.DeleteUser(id); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]bool{"success": true})
}

// GetSettings 获取系统设置
func GetSettings(w http.ResponseWriter, r *http.Request) {
	// 获取配置
	cfg := config.Get()

	// 返回响应
	utils.ResponseOK(w, cfg)
}

// UpdateSettings 更新系统设置
func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// 获取配置
	cfg := config.Get()

	// 更新钉钉设置
	cfg.DingTalk.AppKey = req.DingTalk.AppKey
	cfg.DingTalk.AppSecret = req.DingTalk.AppSecret
	cfg.DingTalk.AgentID = req.DingTalk.AgentID
	cfg.DingTalk.CorpID = req.DingTalk.CorpID
	// 更新网站设置
	cfg.Website.Name = req.Website.Name
	cfg.Website.ICPBeian = req.Website.ICPBeian
	cfg.Website.PublicSecBeian = req.Website.PublicSecBeian
	cfg.Website.Domain = req.Website.Domain
	// 更新定时任务设置
	cfg.Scheduler.Enabled = req.Scheduler.Enabled
	cfg.Scheduler.CleanupTime = req.Scheduler.CleanupTime
	cfg.Scheduler.ReminderBeforeEndHours = req.Scheduler.ReminderBeforeEndHours

	// 保存配置
	if err := config.Save(); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}

	// 返回响应
	utils.ResponseOK(w, cfg)
}

// GetSchedulerLogs 获取定时任务运行日志
func GetSchedulerLogs(w http.ResponseWriter, r *http.Request) {
	// 获取日志
	logs := scheduler.GetLogs()

	// 返回响应
	utils.ResponseOK(w, map[string]interface{}{
		"logs": logs,
	})
}
