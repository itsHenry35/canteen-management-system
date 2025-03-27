package utils

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveBase64Image 保存Base64编码的图片
func SaveBase64Image(base64Data, directory, prefix string, timestamp interface{}) string {
	// 移除可能的Base64前缀
	encodedData := base64Data
	if idx := strings.Index(base64Data, ";base64,"); idx > 0 {
		encodedData = base64Data[idx+8:]
	}

	// 解码Base64数据
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return ""
	}

	// 创建目录
	if err := os.MkdirAll(directory, 0755); err != nil {
		return ""
	}

	// 创建文件名
	fileName := fmt.Sprintf("%s_%v.jpg", prefix, timestamp)
	filePath := filepath.Join(directory, fileName)

	// 保存图片
	if err := os.WriteFile(filePath, decodedData, 0644); err != nil {
		return ""
	}

	return fileName
}
