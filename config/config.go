package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	config *Config
	once   sync.Once
)

// Config 应用配置结构
type Config struct {
	Server struct {
		Port int    `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
	Database struct {
		Path string `json:"path"`
	} `json:"database"`
	DingTalk struct {
		AppKey    string `json:"app_key"`
		AppSecret string `json:"app_secret"`
		AgentID   string `json:"agent_id"`
		CorpID    string `json:"corp_id"`
	} `json:"dingtalk"`
	Security struct {
		JWTSecret     string `json:"jwt_secret"`     // JWT 密钥
		EncryptionKey string `json:"encryption_key"` // 数据加密密钥 (必须是16, 24, 或 32字节长)
	} `json:"security"`
	Website struct {
		Name           string `json:"name"`             // 网站名称
		ICPBeian       string `json:"icp_beian"`        // ICP备案信息
		PublicSecBeian string `json:"public_sec_beian"` // 公安部备案信息
		Domain         string `json:"domain"`           // 网站域名，用于通知链接
	} `json:"website"`
	Scheduler struct {
		Enabled                bool   `json:"enabled"`                   // 是否启用定时任务
		CleanupTime            string `json:"cleanup_time"`              // 清理过期餐食的时间（格式：HH:MM）
		ReminderBeforeEndHours int    `json:"reminder_before_end_hours"` // 选餐截止前多少小时发送提醒
	} `json:"scheduler"`
}

// Load 加载配置文件
func Load() error {
	var err error
	once.Do(func() {
		config = &Config{}

		// 默认配置
		config.Server.Port = 8080
		config.Server.Host = "localhost"
		config.Database.Path = "./data/canteen.db"
		config.Security.JWTSecret = "default-jwt-secret-please-change-in-production" // 默认JWT密钥
		config.Security.EncryptionKey = "default-encryption-key-needs-change"        // 默认加密密钥
		config.Website.Name = "饭卡管理系统"                                               // 默认网站名称
		config.Website.ICPBeian = ""                                                 // 默认空ICP备案信息
		config.Website.PublicSecBeian = ""                                           // 默认空公安部备案信息
		config.Website.Domain = ""                                                   // 默认域名
		config.Scheduler.Enabled = false                                             // 默认关闭定时任务
		config.Scheduler.CleanupTime = "02:00"                                       // 默认凌晨2点清理过期餐食
		config.Scheduler.ReminderBeforeEndHours = 6                                  // 默认选餐截止前6小时发送提醒

		// 检查配置文件是否存在
		if _, statErr := os.Stat("config.json"); os.IsNotExist(statErr) {
			// 配置文件不存在，创建默认配置
			data, marshalErr := json.MarshalIndent(config, "", "  ")
			if marshalErr != nil {
				err = fmt.Errorf("error creating default config: %v", marshalErr)
				return
			}

			if writeErr := os.WriteFile("config.json", data, 0644); writeErr != nil {
				err = fmt.Errorf("error writing default config: %v", writeErr)
				return
			}

			fmt.Println("Created default config.json")
		} else {
			// 配置文件存在，读取配置
			data, readErr := os.ReadFile("config.json")
			if readErr != nil {
				err = fmt.Errorf("error reading config: %v", readErr)
				return
			}

			if unmarshalErr := json.Unmarshal(data, config); unmarshalErr != nil {
				err = fmt.Errorf("error parsing config: %v", unmarshalErr)
				return
			}
		}
	})

	return err
}

// Get 获取配置实例
func Get() *Config {
	if config == nil {
		Load()
	}
	return config
}

// Save 保存当前配置到文件
func Save() error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	return os.WriteFile("config.json", data, 0644)
}
