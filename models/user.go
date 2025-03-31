package models

import (
	"database/sql"
	"errors"

	"github.com/itsHenry35/canteen-management-system/database"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin       Role = "admin"
	RoleCanteenA    Role = "canteen_a"
	RoleCanteenB    Role = "canteen_b"
	RoleCanteenTest Role = "canteen_test"
)

// User 用户模型
type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"-"` // 不暴露密码
	FullName   string `json:"full_name"`
	Role       Role   `json:"role"`
	DingTalkID string `json:"dingtalk_id"`
}

// CreateUser 创建新用户
func CreateUser(username, password, fullName string, Role Role, dingtalkId string) (*User, error) {
	// 对密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 获取数据库连接
	db := database.GetDB()

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 插入用户数据
	result, err := tx.Exec(
		"INSERT INTO users (username, password, full_name, role, dingtalk_id) VALUES (?, ?, ?, ?, ?)",
		username, string(hashedPassword), fullName, Role, dingtalkId,
	)
	if err != nil {
		return nil, err
	}

	// 获取插入的 ID
	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 返回创建的用户
	return &User{
		ID:         int(userID),
		Username:   username,
		FullName:   fullName,
		Role:       Role,
		DingTalkID: dingtalkId,
	}, nil
}

// GetUserByID 通过 ID 获取用户
func GetUserByID(id int) (*User, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询用户
	var user User
	var dingTalkID sql.NullString

	err := db.QueryRow(
		"SELECT id, username, password, full_name, role, dingtalk_id FROM users WHERE id = ?",
		id,
	).Scan(
		&user.ID, &user.Username, &user.Password, &user.FullName, &user.Role, &dingTalkID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("系统中未找到用户")
		}
		return nil, err
	}

	if dingTalkID.Valid {
		user.DingTalkID = dingTalkID.String
	}

	return &user, nil
}

// GetUserByUsername 通过用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询用户
	var user User
	var dingTalkID sql.NullString

	err := db.QueryRow(
		"SELECT id, username, password, full_name, role, dingtalk_id FROM users WHERE username = ?",
		username,
	).Scan(
		&user.ID, &user.Username, &user.Password, &user.FullName, &user.Role, &dingTalkID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("系统中未找到用户")
		}
		return nil, err
	}

	if dingTalkID.Valid {
		user.DingTalkID = dingTalkID.String
	}

	return &user, nil
}

// GetUserByDingTalkID 通过钉钉ID获取用户
func GetUserByDingTalkID(dingTalkID string) (*User, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询用户
	var user User
	var dbDingTalkID sql.NullString

	err := db.QueryRow(
		"SELECT id, username, password, full_name, role, dingtalk_id FROM users WHERE dingtalk_id = ?",
		dingTalkID,
	).Scan(
		&user.ID, &user.Username, &user.Password, &user.FullName, &user.Role, &dbDingTalkID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("系统中未找到用户")
		}
		return nil, err
	}

	if dbDingTalkID.Valid {
		user.DingTalkID = dbDingTalkID.String
	}

	return &user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(user *User) error {
	// 获取数据库连接
	db := database.GetDB()

	// 更新用户数据
	_, err := db.Exec(
		"UPDATE users SET full_name = ?, dingtalk_id = ? WHERE id = ?",
		user.FullName, user.DingTalkID, user.ID,
	)

	return err
}

// UpdatePassword 更新用户密码
func UpdatePassword(userID int, newPassword string) error {
	// 对新密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 获取数据库连接
	db := database.GetDB()

	// 更新密码
	_, err = db.Exec(
		"UPDATE users SET password = ? WHERE id = ?",
		string(hashedPassword), userID,
	)

	return err
}

// DeleteUser 删除用户
func DeleteUser(id int) error {
	// 获取数据库连接
	db := database.GetDB()

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除用户
	_, err = tx.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}

// VerifyPassword 验证用户密码
func VerifyPassword(username, password string) (*User, error) {
	// 获取用户
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}

	return user, nil
}

// GetAllUsers 获取所有用户
func GetAllUsers(Role Role) ([]*User, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询条件
	var query string
	var args []interface{}

	if Role != "" {
		query = "SELECT id, username, password, full_name, role, dingtalk_id FROM users WHERE role = ?"
		args = append(args, Role)
	} else {
		query = "SELECT id, username, password, full_name, role, dingtalk_id FROM users"
	}

	// 执行查询
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var users []*User
	for rows.Next() {
		var user User
		var dingTalkID sql.NullString

		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.FullName, &user.Role, &dingTalkID,
		)
		if err != nil {
			return nil, err
		}

		if dingTalkID.Valid {
			user.DingTalkID = dingTalkID.String
		}

		users = append(users, &user)
	}

	return users, nil
}
