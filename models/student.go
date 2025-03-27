package models

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/itsHenry35/canteen-management-system/database"
	"github.com/mozillazg/go-pinyin"
)

// MealType 餐食类型
type MealType string

const (
	MealTypeA MealType = "A"
	MealTypeB MealType = "B"
)

// Student 学生模型
type Student struct {
	ID                     int        `json:"id"`
	Username               string     `json:"username"`
	FullName               string     `json:"full_name"`
	Class                  string     `json:"class"`
	DingTalkID             string     `json:"dingtalk_id"`
	LastMealCollectionDate *time.Time `json:"last_meal_collection_date,omitempty"`
}

// 生成学生用户名：stu+姓名首字母+随机数
func generateStudentUsername(fullName string) (string, error) {
	// 初始化拼音库
	args := pinyin.NewArgs()
	args.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	// 获取姓名的每个字的拼音首字母
	var firstLetters strings.Builder
	for _, char := range fullName {
		// 处理每个字符
		result := pinyin.SinglePinyin(char, args)
		if len(result) > 0 && len(result[0]) > 0 {
			// 获取拼音的首字母并转为小写
			firstLetters.WriteString(strings.ToLower(string(result[0][0])))
		}
	}

	// 获取首字母组合
	initials := firstLetters.String()
	if initials == "" {
		// 如果没有获取到任何首字母，使用原始姓名
		initials = strings.ToLower(fullName)
	}

	// 生成8位随机数
	randNum := rand.Intn(90000000) + 10000000

	// 拼接用户名
	username := fmt.Sprintf("stu%s%d", initials, randNum)

	return username, nil
}

// 检查用户名是否已存在
func isUsernameExists(db *sql.DB, username string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM students WHERE username = ?", username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateStudent 创建新学生
func CreateStudent(fullName string, class string, dingTalkID string) (*Student, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 生成用户名
	username, err := generateStudentUsername(fullName)
	if err != nil {
		return nil, err
	}

	// 检查用户名是否已存在，如果已存在则重新生成
	exists, err := isUsernameExists(db, username)
	if err != nil {
		return nil, err
	}

	// 如果用户名已存在，尝试重新生成最多5次
	attempts := 0
	for exists && attempts < 5 {
		username, err = generateStudentUsername(fullName)
		if err != nil {
			return nil, err
		}
		exists, err = isUsernameExists(db, username)
		if err != nil {
			return nil, err
		}
		attempts++
	}

	// 如果用户名仍然存在，使用时间戳确保唯一性
	if exists {
		timestamp := time.Now().UnixNano() / 1000000
		username = fmt.Sprintf("%s%d", username, timestamp)
	}

	// 如果没有提供钉钉ID，设置为空字符串
	if dingTalkID == "" {
		dingTalkID = "0"
	}

	// 设置默认日期为2000年1月1日
	defaultDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 插入学生数据
	result, err := tx.Exec(
		"INSERT INTO students (username, full_name, class, dingtalk_id, last_meal_collection_date) VALUES (?, ?, ?, ?, ?)",
		username, fullName, class, dingTalkID, defaultDate,
	)
	if err != nil {
		return nil, err
	}

	// 获取插入的 ID
	studentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 返回创建的学生
	return &Student{
		ID:                     int(studentID),
		Username:               username,
		FullName:               fullName,
		Class:                  class,
		DingTalkID:             dingTalkID,
		LastMealCollectionDate: &defaultDate,
	}, nil
}

// GetStudentByID 通过 ID 获取学生
func GetStudentByID(id int) (*Student, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询学生
	var student Student
	var lastMealCollectionDate sql.NullTime

	err := db.QueryRow(
		`SELECT s.id, s.username, s.full_name, s.class, s.dingtalk_id, s.last_meal_collection_date
		FROM students s 
		WHERE s.id = ?`,
		id,
	).Scan(
		&student.ID, &student.Username, &student.FullName, &student.Class,
		&student.DingTalkID, &lastMealCollectionDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	if lastMealCollectionDate.Valid {
		student.LastMealCollectionDate = &lastMealCollectionDate.Time
	}

	return &student, nil
}

// GetAllStudents 获取所有学生
func GetAllStudents() ([]*Student, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询所有学生
	query := `
        SELECT s.id, s.username, s.full_name, s.class, s.dingtalk_id, s.last_meal_collection_date
        FROM students s
        ORDER BY s.class, s.full_name
    `

	// 执行查询
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var students []*Student
	for rows.Next() {
		var s Student
		var lastMealCollectionDate sql.NullTime

		err := rows.Scan(
			&s.ID, &s.Username, &s.FullName, &s.Class,
			&s.DingTalkID, &lastMealCollectionDate,
		)
		if err != nil {
			return nil, err
		}

		if lastMealCollectionDate.Valid {
			s.LastMealCollectionDate = &lastMealCollectionDate.Time
		}

		students = append(students, &s)
	}

	return students, nil
}

// GetStudentByUsername 通过用户名获取学生
func GetStudentByUsername(username string) (*Student, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询学生
	var student Student
	var lastMealCollectionDate sql.NullTime

	err := db.QueryRow(
		`SELECT s.id, s.username, s.full_name, s.class, s.dingtalk_id, s.last_meal_collection_date
		FROM students s 
		WHERE s.username = ?`,
		username,
	).Scan(
		&student.ID, &student.Username, &student.FullName, &student.Class,
		&student.DingTalkID, &lastMealCollectionDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	if lastMealCollectionDate.Valid {
		student.LastMealCollectionDate = &lastMealCollectionDate.Time
	}

	return &student, nil
}

// GetStudentByDingTalkID 通过钉钉ID获取学生
func GetStudentByDingTalkID(dingTalkID string) (*Student, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询学生
	var student Student
	var lastMealCollectionDate sql.NullTime

	err := db.QueryRow(
		`SELECT s.id, s.username, s.full_name, s.class, s.dingtalk_id, s.last_meal_collection_date
		FROM students s 
		WHERE s.dingtalk_id = ?`,
		dingTalkID,
	).Scan(
		&student.ID, &student.Username, &student.FullName, &student.Class,
		&student.DingTalkID, &lastMealCollectionDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	if lastMealCollectionDate.Valid {
		student.LastMealCollectionDate = &lastMealCollectionDate.Time
	}

	return &student, nil
}

// UpdateStudent 更新学生信息
func UpdateStudent(student *Student) error {
	// 获取数据库连接
	db := database.GetDB()

	// 更新学生数据
	_, err := db.Exec(
		`UPDATE students SET full_name = ?, class = ?, dingtalk_id = ?, last_meal_collection_date = ?
		WHERE id = ?`,
		student.FullName, student.Class, student.DingTalkID, student.LastMealCollectionDate, student.ID,
	)

	return err
}

// DeleteStudent 删除学生
func DeleteStudent(id int) error {
	// 获取数据库连接
	db := database.GetDB()

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除学生的选餐记录
	_, err = tx.Exec("DELETE FROM meal_selections WHERE student_id = ?", id)
	if err != nil {
		return err
	}

	// 删除学生
	_, err = tx.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}
