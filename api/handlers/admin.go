package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/itsHenry35/canteen-management-system/api/middlewares"
	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/models"
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
	if req.Role != models.RoleAdmin && req.Role != models.RoleCanteenA && req.Role != models.RoleCanteenB {
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
	if req.DingTalk.AppKey != "" {
		cfg.DingTalk.AppKey = req.DingTalk.AppKey
	}
	if req.DingTalk.AppSecret != "" {
		cfg.DingTalk.AppSecret = req.DingTalk.AppSecret
	}
	if req.DingTalk.AgentID != "" {
		cfg.DingTalk.AgentID = req.DingTalk.AgentID
	}
	if req.DingTalk.CorpID != "" {
		cfg.DingTalk.CorpID = req.DingTalk.CorpID
	}

	// 保存配置
	if err := config.Save(); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}

	// 返回响应
	utils.ResponseOK(w, cfg)
}

// GetDingTalkCorpID 获取钉钉企业ID
func GetDingTalkCorpID(w http.ResponseWriter, r *http.Request) {
	// 获取配置
	cfg := config.Get()

	// 返回响应
	utils.ResponseOK(w, map[string]string{
		"corp_id": cfg.DingTalk.CorpID,
	})
}

// NotifyUnselectedStudents 手动提醒未选餐学生
func NotifyUnselectedStudents(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req NotifyUnselectedStudentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 验证餐ID
	meal, err := models.GetMealByID(req.MealID)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "未找到指定的餐")
		return
	}

	// 验证选餐时间
	now := time.Now()
	if now.Before(meal.SelectionStartTime) {
		utils.ResponseError(w, http.StatusBadRequest, "选餐尚未开始，不能发送提醒")
		return
	}
	if now.After(meal.SelectionEndTime) {
		utils.ResponseError(w, http.StatusBadRequest, "选餐已结束，不能发送提醒")
		return
	}

	// 获取所有学生
	allStudents, err := models.GetAllStudents()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取学生列表失败")
		return
	}

	// 获取该餐的选餐记录
	selections, err := models.GetMealSelectionsByMeal(req.MealID)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取选餐记录失败")
		return
	}

	// 创建已选餐学生ID的集合
	selectedStudentIDs := make(map[int]bool)
	for _, selection := range selections {
		selectedStudentIDs[selection.StudentID] = true
	}

	// 找出未选餐的学生
	var unselectedStudents []*models.Student
	for _, student := range allStudents {
		if !selectedStudentIDs[student.ID] {
			unselectedStudents = append(unselectedStudents, student)
		}
	}

	// 如果没有未选餐的学生，直接返回
	if len(unselectedStudents) == 0 {
		utils.ResponseOK(w, map[string]interface{}{
			"success": true,
			"message": "所有学生都已完成选餐",
			"count":   0,
		})
		return
	}

	// 构建钉钉通知消息
	title := "选餐提醒"
	markdown := fmt.Sprintf("## 选餐提醒\n\n**亲爱的家长/同学，您尚未完成%s的选餐，请及时完成选餐。**\n\n**选餐截止时间为: %s**", meal.Name, meal.SelectionEndTime.Format("2006-01-02 15:04:05"))

	card := utils.ActionCardMessage{
		Title:       title,
		Markdown:    markdown,
		SingleTitle: "查看详情",
		SingleURL:   "https://xuancan.itshenryz.com/dingtalk_auth",
	}

	// 收集所有钉钉ID（学生和家长）
	dingTalkIDs := make([]string, 0)

	// 启动goroutine异步处理，避免阻塞请求
	go func() {
		// 处理学生钉钉ID和家长钉钉ID
		for _, student := range unselectedStudents {
			// 添加学生钉钉ID
			if student.DingTalkID != "" && student.DingTalkID != "0" {
				dingTalkIDs = append(dingTalkIDs, student.DingTalkID)
			}

			// 获取并添加家长钉钉ID
			parents, err := models.GetParentsByStudentID(student.ID)
			if err != nil {
				utils.LogError(fmt.Sprintf("获取学生ID=%d的家长信息失败: %v", student.ID, err))
				continue
			}

			for _, parent := range parents {
				if parent != "" && parent != "0" {
					dingTalkIDs = append(dingTalkIDs, parent)
				}
			}
		}

		// 如果有需要通知的人
		if len(dingTalkIDs) > 0 {
			// 发送通知
			err := utils.SendDingTalkActionCard(dingTalkIDs, card)
			if err != nil {
				utils.LogError(fmt.Sprintf("发送未选餐提醒失败: %v", err))
			}
		} else {
			utils.LogError("没有找到需要通知的学生或家长")
		}
	}()

	// 立即返回成功响应
	utils.ResponseOK(w, map[string]interface{}{
		"success": true,
		"message": "未选餐提醒已发送",
		"count":   len(unselectedStudents),
	})
}
