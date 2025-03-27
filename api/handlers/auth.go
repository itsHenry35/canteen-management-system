package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/itsHenry35/canteen-management-system/models"
	"github.com/itsHenry35/canteen-management-system/services"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	} `json:"user"`
}

// DingTalkLoginResponse 钉钉登录响应
type DingTalkLoginResponse struct {
	// 单个学生情况
	Token string `json:"token,omitempty"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	} `json:"user,omitempty"`

	// 多个学生情况
	Students []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Class    string `json:"class"`
		Token    string `json:"token"`
	} `json:"students,omitempty"`
}

// Login 用户登录
func Login(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// 验证用户凭据（仅支持管理员和食堂用户）
	token, userObj, err := services.Login(req.Username, req.Password)
	if err != nil {
		utils.ResponseError(w, http.StatusUnauthorized, "账号或密码错误")
		return
	}

	// 构建响应
	resp := LoginResponse{
		Token: token,
		User: struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			FullName string `json:"full_name"`
			Role     string `json:"role"`
		}{},
	}

	// 设置用户信息
	user := userObj.(*models.User)
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	resp.User.FullName = user.FullName
	if user.Role == models.RoleAdmin {
		resp.User.Role = "admin"
	} else if user.Role == models.RoleCanteenA {
		resp.User.Role = "canteen_a"
	} else if user.Role == models.RoleCanteenB {
		resp.User.Role = "canteen_b"
	}

	// 返回响应
	utils.ResponseOK(w, resp)
}

// DingTalkLoginRequest 钉钉登录请求
type DingTalkLoginRequest struct {
	Code string `json:"code"`
}

// DingTalkLogin 钉钉登录
func DingTalkLogin(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req DingTalkLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// 进行钉钉免登录
	token, userObj, err := services.DingTalkLogin(req.Code)
	if err != nil {
		utils.ResponseError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// 判断返回的是单个用户还是多个学生
	switch u := userObj.(type) {
	case *models.Student:
		// 单个学生
		resp := DingTalkLoginResponse{
			Token: token,
			User: struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
				FullName string `json:"full_name"`
				Role     string `json:"role"`
			}{
				ID:       u.ID,
				Username: u.Username,
				FullName: u.FullName,
				Role:     "student",
			},
		}
		utils.ResponseOK(w, resp)

	case *models.User:
		// 管理员或食堂用户
		resp := DingTalkLoginResponse{
			Token: token,
			User: struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
				FullName string `json:"full_name"`
				Role     string `json:"role"`
			}{
				ID:       u.ID,
				Username: u.Username,
				FullName: u.FullName,
			},
		}

		if u.Role == models.RoleAdmin {
			resp.User.Role = "admin"
		} else if u.Role == models.RoleCanteenA {
			resp.User.Role = "canteen_a"
		} else if u.Role == models.RoleCanteenB {
			resp.User.Role = "canteen_b"
		}

		utils.ResponseOK(w, resp)

	case []services.StudentData:
		// 多个学生（家长登录）
		resp := DingTalkLoginResponse{
			Students: make([]struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
				FullName string `json:"full_name"`
				Class    string `json:"class"`
				Token    string `json:"token"`
			}, len(u)),
		}

		// 添加所有学生信息和各自的token
		for i, student := range u {
			resp.Students[i] = struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
				FullName string `json:"full_name"`
				Class    string `json:"class"`
				Token    string `json:"token"`
			}{
				ID:       student.ID,
				Username: student.Username,
				FullName: student.FullName,
				Class:    student.Class,
				Token:    student.Token,
			}
		}

		utils.ResponseOK(w, resp)
	}
}
