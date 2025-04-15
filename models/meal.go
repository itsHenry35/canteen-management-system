package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/itsHenry35/canteen-management-system/config"
	"github.com/itsHenry35/canteen-management-system/database"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// Meal 餐模型
type Meal struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`                 // 餐名
	SelectionStartTime time.Time `json:"selection_start_time"` // 选餐开始时间
	SelectionEndTime   time.Time `json:"selection_end_time"`   // 选餐结束时间
	EffectiveStartDate time.Time `json:"effective_start_date"` // 领餐开始生效日期
	EffectiveEndDate   time.Time `json:"effective_end_date"`   // 领餐结束生效日期
	ImagePath          string    `json:"image_path"`           // 餐的图片地址
}

// CreateMeal 创建新餐
func CreateMeal(name string, selectionStartTime, selectionEndTime, effectiveStartDate, effectiveEndDate time.Time, imagePath string) (*Meal, error) {
	// 校验时间
	if err := validateMealTimes(0, selectionStartTime, selectionEndTime, effectiveStartDate, effectiveEndDate); err != nil {
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

	// 插入餐数据
	result, err := tx.Exec(
		"INSERT INTO meals (name, selection_start_time, selection_end_time, effective_start_date, effective_end_date, image_path) VALUES (?, ?, ?, ?, ?, ?)",
		name, selectionStartTime, selectionEndTime, effectiveStartDate, effectiveEndDate, imagePath,
	)
	if err != nil {
		return nil, err
	}

	// 获取插入的 ID
	mealID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 返回创建的餐
	return &Meal{
		ID:                 int(mealID),
		Name:               name,
		SelectionStartTime: selectionStartTime,
		SelectionEndTime:   selectionEndTime,
		EffectiveStartDate: effectiveStartDate,
		EffectiveEndDate:   effectiveEndDate,
		ImagePath:          imagePath,
	}, nil
}

// GetMealByID 通过ID获取餐
func GetMealByID(id int) (*Meal, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询餐
	var meal Meal
	err := db.QueryRow(
		"SELECT id, name, selection_start_time, selection_end_time, effective_start_date, effective_end_date, image_path FROM meals WHERE id = ?",
		id,
	).Scan(
		&meal.ID, &meal.Name, &meal.SelectionStartTime, &meal.SelectionEndTime, &meal.EffectiveStartDate, &meal.EffectiveEndDate, &meal.ImagePath,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("餐不存在")
		}
		return nil, err
	}

	return &meal, nil
}

// GetAllMeals 获取所有餐
func GetAllMeals() ([]*Meal, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询所有餐
	rows, err := db.Query(
		"SELECT id, name, selection_start_time, selection_end_time, effective_start_date, effective_end_date, image_path FROM meals ORDER BY effective_start_date",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var meals []*Meal
	for rows.Next() {
		var meal Meal
		err := rows.Scan(
			&meal.ID, &meal.Name, &meal.SelectionStartTime, &meal.SelectionEndTime, &meal.EffectiveStartDate, &meal.EffectiveEndDate, &meal.ImagePath,
		)
		if err != nil {
			return nil, err
		}
		meals = append(meals, &meal)
	}

	return meals, nil
}

// GetCurrentSelectableMeals 获取当前可以选择的餐
func GetCurrentSelectableMeals() ([]*Meal, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询当前可以选择的餐
	now := time.Now()
	rows, err := db.Query(
		"SELECT id, name, selection_start_time, selection_end_time, effective_start_date, effective_end_date, image_path FROM meals WHERE selection_start_time <= ? AND selection_end_time >= ? ORDER BY effective_start_date",
		now, now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var meals []*Meal
	for rows.Next() {
		var meal Meal
		err := rows.Scan(
			&meal.ID, &meal.Name, &meal.SelectionStartTime, &meal.SelectionEndTime, &meal.EffectiveStartDate, &meal.EffectiveEndDate, &meal.ImagePath,
		)
		if err != nil {
			return nil, err
		}
		meals = append(meals, &meal)
	}

	return meals, nil
}

// UpdateMeal 更新餐
func UpdateMeal(meal *Meal) error {
	// 校验时间
	if err := validateMealTimes(meal.ID, meal.SelectionStartTime, meal.SelectionEndTime, meal.EffectiveStartDate, meal.EffectiveEndDate); err != nil {
		return err
	}

	// 获取数据库连接
	db := database.GetDB()

	// 更新餐数据
	_, err := db.Exec(
		"UPDATE meals SET name = ?, selection_start_time = ?, selection_end_time = ?, effective_start_date = ?, effective_end_date = ?, image_path = ? WHERE id = ?",
		meal.Name, meal.SelectionStartTime, meal.SelectionEndTime, meal.EffectiveStartDate, meal.EffectiveEndDate, meal.ImagePath, meal.ID,
	)

	return err
}

// DeleteMeal 删除餐及相关数据
func DeleteMeal(id int) error {
	// 获取数据库连接
	db := database.GetDB()

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 获取餐信息（为了获取图片路径）
	var imagePath string
	err = tx.QueryRow("SELECT image_path FROM meals WHERE id = ?", id).Scan(&imagePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("餐不存在")
		}
		return err
	}

	// 删除学生选餐记录
	_, err = tx.Exec("DELETE FROM meal_selections WHERE meal_id = ?", id)
	if err != nil {
		return err
	}

	// 删除餐记录
	_, err = tx.Exec("DELETE FROM meals WHERE id = ?", id)
	if err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	// 如果有图片，则删除图片文件
	if imagePath != "" {
		// 获取图片的物理路径
		physicalPath := "." + imagePath // 将URL路径转换为文件系统路径
		// 删除图片文件
		os.Remove(physicalPath)
	}

	return nil
}

// CleanupExpiredMeals 清理过期的餐
func CleanupExpiredMeals() error {
	// 获取数据库连接
	db := database.GetDB()

	// 查询已过期的餐
	now := time.Now()
	rows, err := db.Query("SELECT id FROM meals WHERE effective_end_date < ?", now)
	if err != nil {
		return err
	}
	defer rows.Close()

	// 处理结果
	var expiredMealIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		expiredMealIDs = append(expiredMealIDs, id)
	}

	// 删除过期的餐
	for _, id := range expiredMealIDs {
		if err := DeleteMeal(id); err != nil {
			return err
		}
	}

	return nil
}

// validateMealTimes 校验餐的时间
func validateMealTimes(mealID int, selectionStartTime, selectionEndTime, effectiveStartDate, effectiveEndDate time.Time) error {
	// 1. 所有开始时间应早于结束时间
	if !selectionStartTime.Before(selectionEndTime) {
		return errors.New("选餐开始时间必须早于选餐结束时间")
	}

	// 2. 开始日期应早于等于结束日期
	if effectiveStartDate.After(effectiveEndDate) {
		return errors.New("领餐开始日期必须早于等于领餐结束日期")
	}

	// 3. 领餐结束日期不能在过去
	now := time.Now()
	if effectiveEndDate.Before(now) {
		return errors.New("领餐结束日期不能在过去")
	}

	// 4. 领餐开始时间必须晚于最后选餐日期
	if !effectiveStartDate.After(selectionEndTime) {
		return errors.New("领餐开始时间必须晚于选餐结束时间")
	}

	// 5. 领餐开始结束区间不能与其他餐重叠
	db := database.GetDB()
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM meals 
		WHERE id != ? AND 
		((effective_start_date <= ? AND effective_end_date >= ?) OR 
		(effective_start_date <= ? AND effective_end_date >= ?) OR 
		(effective_start_date >= ? AND effective_end_date <= ?))`,
		mealID, effectiveStartDate, effectiveStartDate, effectiveEndDate, effectiveEndDate, effectiveStartDate, effectiveEndDate,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("领餐时间区间与其他餐重叠")
	}

	return nil
}

// NotifyUnselectedStudentsByMealId 根据餐ID发送提醒给未选餐的学生
func NotifyUnselectedStudentsByMealId(mealID int) error {
	// 验证餐ID是否存在
	meal, err := GetMealByID(mealID)
	if err != nil {
		return fmt.Errorf("未找到指定的餐: %v", err)
	}

	// 验证选餐时间
	now := time.Now()
	if now.Before(meal.SelectionStartTime) {
		return fmt.Errorf("选餐尚未开始，不能发送提醒")
	}
	if now.After(meal.SelectionEndTime) {
		return fmt.Errorf("选餐已结束，不能发送提醒")
	}

	// 获取所有学生
	allStudents, err := GetAllStudents()
	if err != nil {
		return fmt.Errorf("获取学生列表失败: %v", err)
	}

	// 获取该餐的选餐记录
	selections, err := GetMealSelectionsByMeal(mealID)
	if err != nil {
		return fmt.Errorf("获取选餐记录失败: %v", err)
	}

	// 创建已选餐学生ID的集合
	selectedStudentIDs := make(map[int]bool)
	for _, selection := range selections {
		selectedStudentIDs[selection.StudentID] = true
	}

	// 找出未选餐的学生
	var unselectedStudents []*Student
	for _, student := range allStudents {
		if !selectedStudentIDs[student.ID] {
			unselectedStudents = append(unselectedStudents, student)
		}
	}

	// 如果没有未选餐的学生，直接返回
	if len(unselectedStudents) == 0 {
		return nil
	}

	// 获取配置的域名
	domain := config.Get().Website.Domain

	// 构建钉钉通知消息
	title := "选餐提醒"
	markdown := fmt.Sprintf("## 选餐提醒\n\n# 亲爱的家长/同学，您尚未完成%s的选餐，请及时完成选餐。\n\n# 选餐截止时间为: %s",
		meal.Name, meal.SelectionEndTime.Format("2006-01-02 15:04:05"))

	card := utils.ActionCardMessage{
		Title:       title,
		Markdown:    markdown,
		SingleTitle: "查看详情",
		SingleURL:   fmt.Sprintf("%s/dingtalk_auth", domain),
	}

	// 收集所有钉钉ID（学生和家长）
	dingTalkIDs := make([]string, 0)

	// 处理学生钉钉ID和家长钉钉ID
	for _, student := range unselectedStudents {
		// 添加学生钉钉ID
		if student.DingTalkID != "" && student.DingTalkID != "0" {
			dingTalkIDs = append(dingTalkIDs, student.DingTalkID)
		}

		// 获取并添加家长钉钉ID
		parents, err := GetParentsByStudentID(student.ID)
		if err != nil {
			utils.LogError(fmt.Sprintf("获取学生ID=%d的家长信息失败: %v", student.ID, err))
			continue
		}

		for _, parent := range parents {
			if parent != "" && parent != "0" {
				dingTalkIDs = append(dingTalkIDs, parent)
			}
		}
	}

	// 如果有需要通知的人
	if len(dingTalkIDs) > 0 {
		// 发送通知
		err := utils.SendDingTalkActionCard(dingTalkIDs, card)
		if err != nil {
			return fmt.Errorf("发送未选餐提醒失败: %v", err)
		}
	} else {
		utils.LogError("没有找到需要通知的学生或家长")
	}

	return nil
}
