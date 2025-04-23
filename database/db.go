package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/utils"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

//go:embed migrations/schema.sql
var schemaFS embed.FS

var db *sql.DB

// Initialize 初始化数据库连接
func Initialize() error {
	// 确保数据库目录存在
	dbPath := config.Get().Database.Path
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// 检查数据库文件是否存在，用于判断是否为首次运行
	isFirstRun := !fileExists(dbPath)

	// 打开数据库连接
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// 创建数据库表
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	// 如果是首次运行，创建管理员账户并生成安全密钥
	if isFirstRun {
		if err := setupInitialSystem(); err != nil {
			return fmt.Errorf("failed to setup initial system: %v", err)
		}
	}

	return nil
}

// setupInitialSystem 首次运行时的系统设置
func setupInitialSystem() error {
	// 生成随机管理员密码
	adminPassword, err := utils.GenerateRandomPassword(12)
	if err != nil {
		return fmt.Errorf("failed to generate admin password: %v", err)
	}

	// 生成JWT密钥
	jwtSecret, err := utils.GenerateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate JWT secret: %v", err)
	}

	// 生成加密密钥 (必须是16, 24, 或 32字节长)
	encryptionKey, err := utils.GenerateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate encryption key: %v", err)
	}

	// 更新config中的密钥
	cfg := config.Get()
	cfg.Security.JWTSecret = jwtSecret
	cfg.Security.EncryptionKey = encryptionKey
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	// 对密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %v", err)
	}

	// 删除可能存在的默认管理员用户
	_, err = db.Exec("DELETE FROM users WHERE username = 'admin'")
	if err != nil {
		return fmt.Errorf("failed to delete existing admin user: %v", err)
	}

	// 插入管理员用户
	_, err = db.Exec(
		"INSERT INTO users (username, password, full_name, role) VALUES (?, ?, ?, ?)",
		"admin", string(hashedPassword), "系统管理员", "admin",
	)
	if err != nil {
		return fmt.Errorf("failed to insert admin user: %v", err)
	}

	// 在控制台打印管理员密码
	log.Println("========================================================")
	log.Println("  首次启动系统，已创建管理员账户:")
	log.Println("  用户名: admin")
	log.Printf("  密码: %s", adminPassword)
	log.Println("  请妥善保管此密码，首次登录后请立即修改密码！")
	log.Println("========================================================")

	return nil
}

// GetDB 获取数据库连接
func GetDB() *sql.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// createTables 创建数据库表
func createTables() error {
	// 从嵌入的文件系统读取 schema.sql 文件
	schemaSQL, err := schemaFS.ReadFile("migrations/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	// 执行 SQL 语句
	_, err = db.Exec(string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}

	return nil
}

// fileExists 检查文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
