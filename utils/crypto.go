package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/itsHenry35/canteen-management-system/config"
)

// EncryptData 使用AES-GCM加密数据
func EncryptData(plaintext string) (string, error) {
	// 获取配置中的加密密钥
	cfg := config.Get()
	key := []byte(cfg.Security.EncryptionKey)

	// 创建新的加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 创建一个新的随机数
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密数据
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// 将结果进行base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptData 使用AES-GCM解密数据
func DecryptData(encryptedData string) (string, error) {
	// 获取配置中的加密密钥
	cfg := config.Get()
	key := []byte(cfg.Security.EncryptionKey)

	// 解码base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	// 创建新的加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 获取nonce大小
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// 提取nonce
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密数据
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateQRCodeData 生成学生二维码数据
func GenerateQRCodeData(studentID int, fullName string) (string, error) {
	// 组合数据
	data := fmt.Sprintf("student:%d:%s", studentID, fullName)

	// 加密数据
	return EncryptData(data)
}

// ValidateQRCodeData 验证学生二维码数据
func ValidateQRCodeData(encryptedData string) (int, string, error) {
	// 解密数据
	data, err := DecryptData(encryptedData)
	if err != nil {
		return 0, "", err
	}

	// 解析数据
	var studentID int
	var fullName string
	_, err = fmt.Sscanf(data, "student:%d:%s", &studentID, &fullName)
	if err != nil {
		return 0, "", errors.New("invalid QR code data")
	}

	return studentID, fullName, nil
}
