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
	scheduler         *cron.Cron
	schedulerLogs     []string
	schedulerLogMutex sync.Mutex
	taskIDs           map[string]cron.EntryID // 存储任务ID，用于管理
	taskIDsMutex      sync.RWMutex
)

// 任务类型常量
const (
	TaskCleanup    = "cleanup"      // 清理过期餐食任务
	TaskReminder   = "reminder_"    // 选餐提醒任务
	TaskAutoSelect = "auto_select_" // 自动选餐任务
)

// 初始化
func init() {
	taskIDs = make(map[string]cron.EntryID)
}

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
	// 创建一个新的定时任务调度器
	scheduler = cron.New(cron.WithSeconds())
	scheduler.Start()
	addLog("定时任务管理器已启动")

	// 初始化任务
	return ReloadTasks()
}

// ReloadTasks 重新加载所有定时任务
func ReloadTasks() error {
	cfg := config.Get()

	// 如果定时任务总开关未启用，清除所有任务并返回
	if !cfg.Scheduler.Enabled {
		clearAllTasks()
		addLog("定时任务总开关未启用，已清除所有任务")
		return nil
	}

	// 重新加载各个任务
	var errors []string

	// 1. 清理过期餐食任务
	if err := reloadCleanupTask(); err != nil {
		errors = append(errors, fmt.Sprintf("加载清理过期餐食任务失败: %v", err))
	}

	// 2. 选餐提醒任务
	if err := reloadReminderTasks(); err != nil {
		errors = append(errors, fmt.Sprintf("加载选餐提醒任务失败: %v", err))
	}

	// 3. 检查当前菜单中是否有需要添加自动选餐的任务
	if err := reloadAutoSelectTasks(); err != nil {
		errors = append(errors, fmt.Sprintf("加载自动选餐任务失败: %v", err))
	}

	// 如果有错误，合并返回
	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// reloadCleanupTask 重新加载清理过期餐食任务
func reloadCleanupTask() error {
	cfg := config.Get()

	// 移除旧任务
	removeTask(TaskCleanup)

	// 如果任务未启用，直接返回
	if !cfg.Scheduler.CleanupEnabled {
		addLog("清理过期餐食任务未启用")
		return nil
	}

	// 时间格式为 HH:MM，转换为 cron 表达式 "0 MM HH * * *"
	timeParts := strings.Split(cfg.Scheduler.CleanupTime, ":")
	if len(timeParts) != 2 {
		return fmt.Errorf("无效的时间格式：%s，应为 HH:MM", cfg.Scheduler.CleanupTime)
	}

	cleanupCron := fmt.Sprintf("0 %s %s * * *", timeParts[1], timeParts[0])
	entryID, err := scheduler.AddFunc(cleanupCron, cleanupExpiredMeals)
	if err != nil {
		return fmt.Errorf("添加清理过期餐食的定时任务失败：%v", err)
	}

	// 保存任务ID
	saveTaskID(TaskCleanup, entryID)
	addLog(fmt.Sprintf("已添加清理过期餐食的定时任务，执行时间：%s", cfg.Scheduler.CleanupTime))

	return nil
}

// reloadAutoSelectTasks 重新加载自动选餐任务
func reloadAutoSelectTasks() error {
	cfg := config.Get()

	// 移除所有以'auto_select_'开头的任务
	removeTasksWithPrefix(TaskAutoSelect)

	// 如果任务未启用，直接返回
	if !cfg.Scheduler.AutoSelectEnabled {
		addLog("自动选餐任务未启用")
		return nil
	}

	// 获取所有餐
	meals, err := models.GetAllMeals()
	if err != nil {
		return fmt.Errorf("获取餐列表失败：%v", err)
	}

	// 当前时间
	now := time.Now()

	// 为每个未过期且选餐结束时间在未来的餐添加自动选餐任务
	for _, meal := range meals {
		// 只为选餐结束时间在当前时间之后的餐添加自动选餐任务
		if meal.SelectionEndTime.After(now) {
			// 为该餐创建一个在选餐截止时执行的定时任务
			// 计算从现在到选餐截止时间还有多长时间
			duration := meal.SelectionEndTime.Sub(now)
			seconds := int(duration.Seconds())

			// 只有当选餐截止时间在未来时才添加任务
			if seconds > 0 {
				// 创建一个cron表达式，在指定的时间点运行一次
				// 格式：秒 分 时 日 月 星期
				cronExpr := fmt.Sprintf("%d %d %d %d %d *",
					meal.SelectionEndTime.Second(),
					meal.SelectionEndTime.Minute(),
					meal.SelectionEndTime.Hour(),
					meal.SelectionEndTime.Day(),
					int(meal.SelectionEndTime.Month()))

				// 创建一个闭包，捕获当前的mealID
				autoSelectFunc := func(mealID int) func() {
					return func() {
						autoSelectMeals(mealID)
					}
				}(meal.ID)

				// 添加定时任务
				entryID, err := scheduler.AddFunc(cronExpr, autoSelectFunc)
				if err != nil {
					addLog(fmt.Sprintf("为餐ID=%d添加自动选餐任务失败：%v", meal.ID, err))
					continue
				}

				// 保存任务ID
				taskKey := fmt.Sprintf("%s%d", TaskAutoSelect, meal.ID)
				saveTaskID(taskKey, entryID)
				addLog(fmt.Sprintf("已为餐ID=%d添加自动选餐任务，执行时间：%s",
					meal.ID, meal.SelectionEndTime.Format("2006-01-02 15:04:05")))
			}
		}
	}

	return nil
}

// CheckAndUpdateTasks 检查并更新任务（当餐菜单变更时调用）
func CheckAndUpdateTasks() error {
	// 更新自动选餐任务
	if err := reloadAutoSelectTasks(); err != nil {
		return fmt.Errorf("更新自动选餐任务失败: %v", err)
	}

	// 更新提醒任务
	if err := reloadReminderTasks(); err != nil {
		return fmt.Errorf("更新提醒任务失败: %v", err)
	}

	return nil
}

// autoSelectMeals 自动为未选餐学生选餐
func autoSelectMeals(mealID int) {
	addLog(fmt.Sprintf("开始为餐ID=%d的未选餐学生自动选餐...", mealID))

	// 调用模型层的自动选餐函数
	count, err := models.BatchSelectMealsRandomly(mealID)
	if err != nil {
		addLog(fmt.Sprintf("为餐ID=%d自动选餐失败：%v", mealID, err))
		return
	}

	addLog(fmt.Sprintf("已成功为餐ID=%d的%d名未选餐学生完成自动选餐", mealID, count))
}

func cleanupExpiredMeals() {
	addLog("开始执行清理过期餐食的定时任务...")
	err := models.CleanupExpiredMeals()
	if err != nil {
		addLog(fmt.Sprintf("清理过期餐食失败：%v", err))
	} else {
		addLog("清理过期餐食成功")
	}
}

// reloadReminderTasks 重新加载所有提醒任务
func reloadReminderTasks() error {
	cfg := config.Get()

	// 移除所有以'reminder_'开头的任务
	removeTasksWithPrefix(TaskReminder)

	// 如果任务未启用，直接返回
	if !cfg.Scheduler.ReminderEnabled {
		addLog("选餐提醒任务未启用")
		return nil
	}

	// 获取所有餐
	meals, err := models.GetAllMeals()
	if err != nil {
		return fmt.Errorf("获取餐列表失败：%v", err)
	}

	// 当前时间
	now := time.Now()

	// 为每个餐在选餐截止前特定小时添加提醒任务
	for _, meal := range meals {
		reminderTime := meal.SelectionEndTime.Add(-time.Duration(cfg.Scheduler.ReminderBeforeEndHours) * time.Hour)

		// 只为提醒时间在未来的餐添加提醒任务
		if reminderTime.After(now) {
			cronExpr := fmt.Sprintf("%d %d %d %d %d *",
				reminderTime.Second(),
				reminderTime.Minute(),
				reminderTime.Hour(),
				reminderTime.Day(),
				int(reminderTime.Month()))

			// 创建一个闭包捕获当前的mealID
			reminderFunc := func(mealID int) func() {
				return func() {
					sendReminderForMeal(mealID)
				}
			}(meal.ID)

			// 添加定时任务
			entryID, err := scheduler.AddFunc(cronExpr, reminderFunc)
			if err != nil {
				addLog(fmt.Sprintf("为餐ID=%d添加提醒任务失败：%v", meal.ID, err))
				continue
			}

			// 保存任务ID
			taskKey := fmt.Sprintf("%s%d", TaskReminder, meal.ID)
			saveTaskID(taskKey, entryID)
			addLog(fmt.Sprintf("已为餐ID=%d添加提醒任务，执行时间：%s",
				meal.ID, reminderTime.Format("2006-01-02 15:04:05")))
		}
	}

	return nil
}

// sendReminderForMeal 为特定餐发送提醒
func sendReminderForMeal(mealID int) {
	addLog(fmt.Sprintf("开始为餐ID=%d发送未选餐提醒...", mealID))
	err := models.NotifyUnselectedStudentsByMealId(mealID)
	if err != nil {
		addLog(fmt.Sprintf("为餐ID=%d发送提醒失败：%v", mealID, err))
	} else {
		addLog(fmt.Sprintf("已成功为餐ID=%d的未选餐学生发送提醒", mealID))
	}
}

// 任务ID管理函数
func saveTaskID(key string, id cron.EntryID) {
	taskIDsMutex.Lock()
	defer taskIDsMutex.Unlock()
	taskIDs[key] = id
}

func removeTask(key string) {
	taskIDsMutex.Lock()
	defer taskIDsMutex.Unlock()

	if id, exists := taskIDs[key]; exists {
		scheduler.Remove(id)
		delete(taskIDs, key)
		addLog(fmt.Sprintf("已移除任务：%s", key))
	}
}

func removeTasksWithPrefix(prefix string) {
	taskIDsMutex.Lock()
	defer taskIDsMutex.Unlock()

	for key, id := range taskIDs {
		if strings.HasPrefix(key, prefix) {
			scheduler.Remove(id)
			delete(taskIDs, key)
			addLog(fmt.Sprintf("已移除任务：%s", key))
		}
	}
}

func clearAllTasks() {
	taskIDsMutex.Lock()
	defer taskIDsMutex.Unlock()

	for key, id := range taskIDs {
		scheduler.Remove(id)
		delete(taskIDs, key)
	}
	addLog("已清除所有定时任务")
}

// Stop 停止定时任务管理器
func Stop() {
	if scheduler != nil {
		scheduler.Stop()
		addLog("定时任务管理器已停止")
	}
}
