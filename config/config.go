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
