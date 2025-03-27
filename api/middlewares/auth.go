package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/itsHenry35/canteen-management-system/services"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// ContextKey 用于存储在上下文中的键的类型
type ContextKey string

// 上下文键
const (
	UserIDKey   ContextKey = "user_id"
	UsernameKey ContextKey = "username"
	RoleKey     ContextKey = "role"
)

// AuthMiddleware 身份验证中间件
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取令牌
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.ResponseError(w, http.StatusUnauthorized, "authorization header is required")
			return
		}

		// 解析令牌
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ResponseError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}
		tokenString := parts[1]

		// 验证令牌
		claims, err := services.ValidateToken(tokenString)
		if err != nil {
			utils.ResponseError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// 将用户信息存储在上下文中
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		// 调用下一个处理程序
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware 角色验证中间件
func RoleMiddleware(roles ...services.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从上下文获取用户角色
			userRole, ok := r.Context().Value(RoleKey).(services.UserRole)
			if !ok {
				utils.ResponseError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			// 检查用户是否有权限
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				utils.ResponseError(w, http.StatusForbidden, "forbidden")
				return
			}

			// 调用下一个处理程序
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}

// GetRoleFromContext 从上下文获取用户角色
func GetRoleFromContext(r *http.Request) (services.UserRole, bool) {
	role, ok := r.Context().Value(RoleKey).(services.UserRole)
	return role, ok
}
