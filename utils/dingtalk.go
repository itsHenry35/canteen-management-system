package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/itsHenry35/canteen-management-system/config"
)

// DingTalkToken 钉钉访问令牌结构
type DingTalkToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
}

var (
	dingTalkToken *DingTalkToken
)

// GetDingTalkToken 获取钉钉访问令牌
func GetDingTalkToken() (string, error) {
	// 检查令牌是否存在且有效
	if dingTalkToken != nil && time.Now().Before(dingTalkToken.ExpiresAt) {
		return dingTalkToken.AccessToken, nil
	}

	// 获取配置
	cfg := config.Get()
	appKey := cfg.DingTalk.AppKey
	appSecret := cfg.DingTalk.AppSecret

	// 请求URL
	url := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s", appKey, appSecret)

	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to request DingTalk token: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查响应是否成功
	if result.ErrCode != 0 {
		return "", fmt.Errorf("DingTalk API error: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}

	// 保存令牌
	dingTalkToken = &DingTalkToken{
		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Second * time.Duration(result.ExpiresIn-60)), // 提前60秒过期
	}

	return dingTalkToken.AccessToken, nil
}

// GetDingTalkUserInfo 获取钉钉用户信息
func GetDingTalkUserInfo(code string) (*DingTalkUserInfo, error) {
	// 最大重试次数
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 如果不是第一次尝试，等待一段时间再重试
		// 等待时间随着重试次数增加而增加 (500ms, 1000ms, 2000ms)
		if attempt > 0 {
			backoffTime := time.Duration(500*1<<uint(attempt-1)) * time.Millisecond
			time.Sleep(backoffTime)
			log.Printf("重试获取钉钉用户信息，第 %d 次尝试, 等待时间: %v", attempt+1, backoffTime)
		}

		// 获取访问令牌
		accessToken, err := GetDingTalkToken()
		if err != nil {
			lastErr = err
			continue
		}

		// 请求URL
		url := fmt.Sprintf("https://oapi.dingtalk.com/user/getuserinfo?access_token=%s&code=%s", accessToken, code)

		// 发送请求
		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("failed to request DingTalk user info: %v", err)
			continue
		}

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %v", err)
			continue
		}

		// 解析响应
		var result struct {
			ErrCode  int    `json:"errcode"`
			ErrMsg   string `json:"errmsg"`
			UserID   string `json:"userid"`
			Name     string `json:"name"`
			DeviceID string `json:"deviceId"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %v", err)
			continue
		}

		// 如果是QPS超限错误，进行重试
		if result.ErrCode == 88 || result.ErrCode == -1 {
			lastErr = fmt.Errorf("DingTalk API QPS limit: %s (code: %d)", result.ErrMsg, result.ErrCode)
			continue
		}

		// 检查响应是否成功
		if result.ErrCode != 0 {
			return nil, fmt.Errorf("DingTalk API error: %s (code: %d)", result.ErrMsg, result.ErrCode)
		}

		// 返回用户信息
		return &DingTalkUserInfo{
			UserID:   result.UserID,
			Name:     result.Name,
			DeviceID: result.DeviceID,
		}, nil
	}

	return nil, fmt.Errorf("获取钉钉用户信息失败，已重试 %d 次: %v", maxRetries, lastErr)
}

// DingTalkUserInfo 钉钉用户信息结构
type DingTalkUserInfo struct {
	UserID   string `json:"userid"`
	Name     string `json:"name"`
	DeviceID string `json:"deviceId"`
}

// DingTalkGuardianStudentRel 家长学生关系结构
type DingTalkGuardianStudentRel struct {
	GuardianUserID string `json:"guardian_userid"`
	Relation       string `json:"relation"`
	StudentUserId  string `json:"student_userid"`
	StudentName    string `json:"student_name"`
}

// GetAllClassIDs 获取所有班级的ID
func GetAllClassIDs() ([]string, error) {
	// 获取访问令牌
	accessToken, err := GetDingTalkToken()
	if err != nil {
		return nil, err
	}

	// 存储所有班级ID
	var classIDs []string

	// 从根部门开始递归查找班级
	err = findClassDepartments(accessToken, 0, &classIDs)
	if err != nil {
		return nil, err
	}

	fmt.Printf("总共获取到 %d 个班级ID\n", len(classIDs))
	return classIDs, nil
}

// findClassDepartments 递归查找班级部门
func findClassDepartments(accessToken string, superID int, classIDs *[]string) error {
	// 请求URL
	url := "https://oapi.dingtalk.com/topapi/edu/dept/list"

	// 构建请求数据
	data := map[string]interface{}{
		"page_size": 30,
		"page_no":   1,
	}

	// 只有非0的superID才需要添加
	if superID > 0 {
		data["super_id"] = superID
	}

	// 处理分页
	hasMore := true
	pageNo := 1

	for hasMore {
		data["page_no"] = pageNo
		jsonData, err := json.Marshal(data)
		fmt.Println("请求数据:", string(jsonData))
		if err != nil {
			return fmt.Errorf("编码请求失败: %v", err)
		}

		// 发送请求
		time.Sleep(500 * time.Millisecond) // 避免触发QPS限制
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		q := req.URL.Query()
		q.Add("access_token", accessToken)
		fmt.Println("请求URL:", req.URL.String())
		fmt.Println("access_token:", accessToken)
		req.URL.RawQuery = q.Encode()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("发送请求失败: %v", err)
		}

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取响应失败: %v", err)
		}

		// 打印响应，用于调试
		if superID == 0 {
			fmt.Println("根部门列表响应:", string(body))
		}

		// 解析响应
		var result struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
			Success bool   `json:"success"`
			Result  struct {
				HasMore bool `json:"has_more"`
				SuperID int  `json:"super_id"`
				Details []struct {
					DeptID   int    `json:"dept_id"`
					DeptType string `json:"dept_type"`
					Name     string `json:"name"`
				} `json:"details"`
			} `json:"result"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("解析响应失败: %v", err)
		}

		// 检查响应是否成功
		if !result.Success || result.ErrCode != 0 {
			// 如果是无数据，直接返回
			if result.ErrCode == 60123 {
				return nil
			}
			return fmt.Errorf("钉钉API错误: %s (代码: %d)", result.ErrMsg, result.ErrCode)
		}

		// 处理当前页的部门
		for _, dept := range result.Result.Details {
			if dept.DeptType == "class" {
				// 如果是班级，添加到班级列表
				*classIDs = append(*classIDs, strconv.Itoa(dept.DeptID))
				fmt.Printf("发现班级: ID=%d, 名称=%s\n", dept.DeptID, dept.Name)
			} else {
				// 如果不是班级，递归查找子部门
				fmt.Printf("发现部门: ID=%d, 名称=%s, 类型=%s\n", dept.DeptID, dept.Name, dept.DeptType)
				err = findClassDepartments(accessToken, dept.DeptID, classIDs)
				if err != nil {
					fmt.Printf("获取部门ID=%d的子部门失败: %v\n", dept.DeptID, err)
					// 继续处理其他部门，不中断整个过程
				}
			}
		}

		// 检查是否有更多页
		hasMore = result.Result.HasMore
		pageNo++
	}

	return nil
}

// GetClassParentStudentRelations 获取指定班级的所有家长-学生关系
func GetClassParentStudentRelations(classID string) ([]DingTalkGuardianStudentRel, error) {
	// 获取访问令牌
	accessToken, err := GetDingTalkToken()
	if err != nil {
		return nil, err
	}

	// 存储所有找到的关系
	var allRelations []DingTalkGuardianStudentRel

	// 分页获取数据
	pageNo := 1
	hasMore := true

	for hasMore {
		// 请求URL
		url := "https://oapi.dingtalk.com/topapi/edu/user/relation/list"

		// 构建请求数据
		data := map[string]interface{}{
			"class_id":  classID,
			"page_size": 30,
			"page_no":   pageNo,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("编码请求失败: %v", err)
		}

		// 发送请求
		time.Sleep(500 * time.Millisecond) // 避免触发QPS限制
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		q := req.URL.Query()
		q.Add("access_token", accessToken)
		req.URL.RawQuery = q.Encode()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("发送请求失败: %v", err)
		}

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}

		// 解析响应
		var result struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
			Success bool   `json:"success"`
			Result  struct {
				HasMore   bool `json:"has_more"`
				Relations []struct {
					ClassID      int    `json:"class_id"`
					FromUserid   string `json:"from_userid"`
					RelationCode string `json:"relation_code"`
					RelationName string `json:"relation_name"`
					ToUserid     string `json:"to_userid"`
				} `json:"relations"`
			} `json:"result"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		// 检查响应是否成功
		if !result.Success || result.ErrCode != 0 {
			// 如果是无数据，直接返回空结果
			if result.ErrCode == 60123 {
				return []DingTalkGuardianStudentRel{}, nil
			}
			return nil, fmt.Errorf("钉钉API错误: %s (代码: %d)", result.ErrMsg, result.ErrCode)
		}

		// 添加获取到的关系
		for _, rel := range result.Result.Relations {
			relation := DingTalkGuardianStudentRel{
				GuardianUserID: rel.FromUserid,
				Relation:       rel.RelationName,
				StudentUserId:  rel.ToUserid,
				StudentName:    "", // 这里没有学生名称，可能需要额外调用API获取
			}
			allRelations = append(allRelations, relation)
		}

		// 检查是否有更多数据
		hasMore = result.Result.HasMore
		if hasMore {
			pageNo++
		}
	}

	return allRelations, nil
}

// ActionCardMessage 定义钉钉卡片消息结构
type ActionCardMessage struct {
	Title       string `json:"title"`
	Markdown    string `json:"markdown"`
	SingleTitle string `json:"single_title"`
	SingleURL   string `json:"single_url"`
}

// SendDingTalkActionCard 发送钉钉卡片消息
func SendDingTalkActionCard(userIDs []string, card ActionCardMessage) error {
	// 获取访问令牌
	accessToken, err := GetDingTalkToken()
	if err != nil {
		return err
	}

	// 获取配置
	cfg := config.Get()
	agentID := cfg.DingTalk.AgentID

	// 如果userIDs长度超过100，需要分批发送
	if len(userIDs) > 100 {
		var batches [][]string
		for i := 0; i < len(userIDs); i += 100 {
			end := i + 100
			if end > len(userIDs) {
				end = len(userIDs)
			}
			batches = append(batches, userIDs[i:end])
		}

		// 分批发送
		for _, batch := range batches {
			err := sendDingTalkActionCardBatch(accessToken, agentID, batch, card)
			time.Sleep(1000 * time.Millisecond) // 避免触发QPS限制
			if err != nil {
				return err
			}
		}
		return nil
	}

	// 单批发送
	return sendDingTalkActionCardBatch(accessToken, agentID, userIDs, card)
}

// sendDingTalkActionCardBatch 按批次发送钉钉卡片消息
func sendDingTalkActionCardBatch(accessToken, agentID string, userIDs []string, card ActionCardMessage) error {
	// 构建请求数据
	data := map[string]interface{}{
		"agent_id":    agentID,
		"userid_list": strings.Join(userIDs, ","),
		"msg": map[string]interface{}{
			"msgtype": "action_card",
			"action_card": map[string]string{
				"title":        card.Title,
				"markdown":     card.Markdown,
				"single_title": card.SingleTitle,
				"single_url":   card.SingleURL,
			},
		},
	}

	// 编码请求数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	// 请求URL
	url := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", accessToken)

	// 发送请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send DingTalk action card: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查响应是否成功
	if result.ErrCode != 0 {
		return fmt.Errorf("DingTalk API error: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}

	return nil
}

// LogError 记录错误信息到日志
func LogError(message string) {
	log.Printf("ERROR: %s", message)
}
