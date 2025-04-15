package scheduler

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/models"
	"github.com/robfig/cron/v3"
)

var (
	scheduler *cron.Cron
	// 添加全局变量以存储日志
	schedulerLogs     []string
	schedulerLogMutex sync.Mutex
)

// 添加日志函数
func addLog(message string) {
	schedulerLogMutex.Lock()
	defer schedulerLogMutex.Unlock()

	// 添加时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)

	// 限制日志数量，保留最新的1000条
	if len(schedulerLogs) >= 1000 {
		schedulerLogs = schedulerLogs[1:]
	}

	// 添加到日志列表
	schedulerLogs = append(schedulerLogs, logEntry)

	// 同时输出到标准日志
	log.Println(message)
}

// GetLogs 获取定时任务日志
func GetLogs() []string {
	schedulerLogMutex.Lock()
	defer schedulerLogMutex.Unlock()

	// 返回日志副本，而不是直接返回引用
	logs := make([]string, len(schedulerLogs))
	copy(logs, schedulerLogs)
	return logs
}

// Initialize 初始化定时任务
func Initialize() error {
	cfg := config.Get()

	// 如果定时任务未启用，直接返回
	if !cfg.Scheduler.Enabled {
		addLog("定时任务未启用")
		return nil
	}

	// 创建一个新的定时任务调度器
	scheduler = cron.New(cron.WithSeconds())

	// 添加清理过期餐食的定时任务
	// 时间格式为 HH:MM，转换为 cron 表达式 "0 MM HH * * *"
	timeParts := strings.Split(cfg.Scheduler.CleanupTime, ":")
	if len(timeParts) != 2 {
		return fmt.Errorf("无效的时间格式：%s，应为 HH:MM", cfg.Scheduler.CleanupTime)
	}

	cleanupCron := fmt.Sprintf("0 %s %s * * *", timeParts[1], timeParts[0])
	_, err := scheduler.AddFunc(cleanupCron, cleanupExpiredMeals)
	if err != nil {
		return fmt.Errorf("添加清理过期餐食的定时任务失败：%v", err)
	}

	addLog(fmt.Sprintf("已添加清理过期餐食的定时任务，执行时间：%s", cfg.Scheduler.CleanupTime))

	// 添加选餐提醒的定时任务（每小时检查一次）
	_, err = scheduler.AddFunc("0 0 * * * *", checkAndSendReminders)
	if err != nil {
		return fmt.Errorf("添加选餐提醒的定时任务失败：%v", err)
	}

	addLog(fmt.Sprintf("已添加选餐提醒的定时任务，选餐截止前 %d 小时发送提醒", cfg.Scheduler.ReminderBeforeEndHours))

	// 启动定时任务
	scheduler.Start()
	addLog("定时任务管理器已启动")

	return nil
}

// cleanupExpiredMeals 清理过期餐食的任务
func cleanupExpiredMeals() {
	addLog("开始执行清理过期餐食的定时任务...")
	err := models.CleanupExpiredMeals()
	if err != nil {
		addLog(fmt.Sprintf("清理过期餐食失败：%v", err))
	} else {
		addLog("清理过期餐食成功")
	}
}

// checkAndSendReminders 检查并发送选餐提醒的任务
func checkAndSendReminders() {
	addLog("开始检查需要发送提醒的选餐...")

	// 获取配置
	cfg := config.Get()
	reminderHours := cfg.Scheduler.ReminderBeforeEndHours

	// 获取所有餐
	meals, err := models.GetAllMeals()
	if err != nil {
		addLog(fmt.Sprintf("获取餐列表失败：%v", err))
		return
	}

	// 当前时间
	now := time.Now()

	// 计算提醒时间窗口
	reminderWindow := now.Add(time.Duration(reminderHours) * time.Hour)

	// 遍历所有餐，检查是否需要发送提醒
	for _, meal := range meals {
		// 检查是否在提醒窗口：
		// 1. 当前时间小于选餐截止时间（选餐还未结束）
		// 2. 选餐截止时间在提醒窗口内（当前时间 + reminderHours 小时内）
		// 3. 选餐已经开始（当前时间大于选餐开始时间）
		if meal.SelectionEndTime.After(now) &&
			meal.SelectionEndTime.Before(reminderWindow) &&
			meal.SelectionStartTime.Before(now) {

			addLog(fmt.Sprintf("餐 ID=%d 符合发送提醒条件，开始发送...", meal.ID))

			// 调用模型层的通知函数
			err := models.NotifyUnselectedStudentsByMealId(meal.ID)
			if err != nil {
				addLog(fmt.Sprintf("为餐 ID=%d 发送提醒失败：%v", meal.ID, err))
			} else {
				addLog(fmt.Sprintf("已成功为餐 ID=%d 的未选餐学生发送提醒", meal.ID))
			}
		}
	}

	addLog("选餐提醒检查完成")
}

// Stop 停止定时任务管理器
func Stop() {
	if scheduler != nil {
		scheduler.Stop()
		addLog("定时任务管理器已停止")
	}
}
