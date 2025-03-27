package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/itsHenry35/canteen-management-system/database"
)

// MealSelection 学生选餐记录
type MealSelection struct {
	ID         int       `json:"id"`
	StudentID  int       `json:"student_id"`
	MealID     int       `json:"meal_id"`
	MealType   MealType  `json:"meal_type"`
	SelectedAt time.Time `json:"selected_at"`
	Student    *Student  `json:"student,omitempty"`
	Meal       *Meal     `json:"meal,omitempty"`
}

// CreateMealSelection 创建学生选餐记录
func CreateMealSelection(studentID, mealID int, mealType MealType) (*MealSelection, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 验证餐ID是否存在
	meal, err := GetMealByID(mealID)
	if err != nil {
		return nil, err
	}

	// 验证学生ID是否存在
	student, err := GetStudentByID(studentID)
	if err != nil {
		return nil, err
	}

	// 验证是否在选餐时间范围内
	now := time.Now()
	if now.Before(meal.SelectionStartTime) || now.After(meal.SelectionEndTime) {
		return nil, errors.New("不在选餐时间范围内")
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 检查是否已经有选餐记录
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM meal_selections WHERE student_id = ? AND meal_id = ?", studentID, mealID).Scan(&count)
	if err != nil {
		return nil, err
	}

	var result sql.Result
	if count > 0 {
		// 更新已有记录
		result, err = tx.Exec(
			"UPDATE meal_selections SET meal_type = ?, selected_at = ? WHERE student_id = ? AND meal_id = ?",
			mealType, now, studentID, mealID,
		)
	} else {
		// 插入新记录
		result, err = tx.Exec(
			"INSERT INTO meal_selections (student_id, meal_id, meal_type, selected_at) VALUES (?, ?, ?, ?)",
			studentID, mealID, mealType, now,
		)
	}

	if err != nil {
		return nil, err
	}

	// 如果是插入新记录，获取插入的ID
	var selectionID int64
	if count == 0 {
		selectionID, err = result.LastInsertId()
		if err != nil {
			return nil, err
		}
	} else {
		// 如果是更新记录，获取现有记录的ID
		err = tx.QueryRow("SELECT id FROM meal_selections WHERE student_id = ? AND meal_id = ?", studentID, mealID).Scan(&selectionID)
		if err != nil {
			return nil, err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 返回选餐记录
	return &MealSelection{
		ID:         int(selectionID),
		StudentID:  studentID,
		MealID:     mealID,
		MealType:   mealType,
		SelectedAt: now,
		Student:    student,
		Meal:       meal,
	}, nil
}

// GetMealSelectionByStudentAndMeal 根据学生ID和餐ID获取选餐记录
func GetMealSelectionByStudentAndMeal(studentID, mealID int) (*MealSelection, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询选餐记录
	var selection MealSelection
	err := db.QueryRow(
		"SELECT id, student_id, meal_id, meal_type, selected_at FROM meal_selections WHERE student_id = ? AND meal_id = ?",
		studentID, mealID,
	).Scan(
		&selection.ID, &selection.StudentID, &selection.MealID, &selection.MealType, &selection.SelectedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 返回nil表示没有记录
		}
		return nil, err
	}

	// 加载学生信息
	student, err := GetStudentByID(studentID)
	if err == nil {
		selection.Student = student
	}

	// 加载餐信息
	meal, err := GetMealByID(mealID)
	if err == nil {
		selection.Meal = meal
	}

	return &selection, nil
}

// GetMealSelectionsByStudent 获取学生的所有选餐记录
func GetMealSelectionsByStudent(studentID int) ([]*MealSelection, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询学生的所有选餐记录
	rows, err := db.Query(
		"SELECT id, student_id, meal_id, meal_type, selected_at FROM meal_selections WHERE student_id = ? ORDER BY selected_at DESC",
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var selections []*MealSelection
	for rows.Next() {
		var selection MealSelection
		err := rows.Scan(
			&selection.ID, &selection.StudentID, &selection.MealID, &selection.MealType, &selection.SelectedAt,
		)
		if err != nil {
			return nil, err
		}
		selections = append(selections, &selection)
	}

	// 加载每个选餐记录的餐信息
	for _, selection := range selections {
		meal, err := GetMealByID(selection.MealID)
		if err == nil {
			selection.Meal = meal
		}
	}

	return selections, nil
}

// GetMealSelectionsByMeal 获取餐的所有选餐记录
func GetMealSelectionsByMeal(mealID int) ([]*MealSelection, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询餐的所有选餐记录
	rows, err := db.Query(
		"SELECT id, student_id, meal_id, meal_type, selected_at FROM meal_selections WHERE meal_id = ? ORDER BY selected_at",
		mealID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var selections []*MealSelection
	for rows.Next() {
		var selection MealSelection
		err := rows.Scan(
			&selection.ID, &selection.StudentID, &selection.MealID, &selection.MealType, &selection.SelectedAt,
		)
		if err != nil {
			return nil, err
		}
		selections = append(selections, &selection)
	}

	// 加载每个选餐记录的学生信息
	for _, selection := range selections {
		student, err := GetStudentByID(selection.StudentID)
		if err == nil {
			selection.Student = student
		}
	}

	return selections, nil
}

// BatchSelectMeals 批量为学生选餐
func BatchSelectMeals(studentIDs []int, mealID int, mealType MealType) (int, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 验证餐ID是否存在
	_, err := GetMealByID(mealID)
	if err != nil {
		return 0, err
	}

	// 验证是否在选餐时间范围内
	now := time.Now()

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 计数器
	var count int

	// 处理每个学生
	for _, studentID := range studentIDs {
		// 检查学生是否存在
		_, err := GetStudentByID(studentID)
		if err != nil {
			continue // 跳过不存在的学生
		}

		// 检查是否已有选餐记录
		var existCount int
		err = tx.QueryRow("SELECT COUNT(*) FROM meal_selections WHERE student_id = ? AND meal_id = ?", studentID, mealID).Scan(&existCount)
		if err != nil {
			continue
		}

		if existCount > 0 {
			// 更新已有记录
			_, err = tx.Exec(
				"UPDATE meal_selections SET meal_type = ?, selected_at = ? WHERE student_id = ? AND meal_id = ?",
				mealType, now, studentID, mealID,
			)
		} else {
			// 插入新记录
			_, err = tx.Exec(
				"INSERT INTO meal_selections (student_id, meal_id, meal_type, selected_at) VALUES (?, ?, ?, ?)",
				studentID, mealID, mealType, now,
			)
		}

		if err == nil {
			count++
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteMealSelection 删除选餐记录
func DeleteMealSelection(id int) error {
	// 获取数据库连接
	db := database.GetDB()

	// 删除选餐记录
	_, err := db.Exec("DELETE FROM meal_selections WHERE id = ?", id)
	return err
}

// DeleteMealSelectionsByMeal 删除餐的所有选餐记录
func DeleteMealSelectionsByMeal(mealID int) error {
	// 获取数据库连接
	db := database.GetDB()

	// 删除餐的所有选餐记录
	_, err := db.Exec("DELETE FROM meal_selections WHERE meal_id = ?", mealID)
	return err
}

// GetStudentCurrentSelection 获取学生当前日期的选餐
func GetStudentCurrentSelection(studentID int) (*MealSelection, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 获取当前日期
	now := time.Now()

	// 查询当前时间有效的餐
	var mealID int
	err := db.QueryRow(
		"SELECT id FROM meals WHERE effective_start_date <= ? AND effective_end_date >= ?",
		now, now,
	).Scan(&mealID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 返回nil表示今天没有有效的餐
		}
		return nil, err
	}

	// 查询学生对该餐的选餐记录
	return GetMealSelectionByStudentAndMeal(studentID, mealID)
}
