package models

import (
	"database/sql"
	"errors"

	"github.com/itsHenry35/canteen-management-system/database"
)

// ParentStudentRelation 家长学生关系模型
type ParentStudentRelation struct {
	ID          int    `json:"id"`
	ParentID    string `json:"parent_id"`    // 家长钉钉ID
	StudentID   string `json:"student_id"`   // 学生ID
	StudentName string `json:"student_name"` // 学生姓名
}

// GetStudentsByParentID 根据家长ID获取所有关联的学生
func GetStudentsByParentID(parentID string) ([]*ParentStudentRelation, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 执行查询
	rows, err := db.Query(`
        SELECT psr.id, psr.parent_id, psr.student_id, s.full_name
        FROM parent_student_relations psr
        JOIN students s ON psr.student_id = s.dingtalk_id
        WHERE psr.parent_id = ?
    `, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var relations []*ParentStudentRelation
	for rows.Next() {
		var relation ParentStudentRelation
		err := rows.Scan(&relation.ID, &relation.ParentID, &relation.StudentID, &relation.StudentName)
		if err != nil {
			return nil, err
		}
		relations = append(relations, &relation)
	}

	return relations, nil
}

// GetStudentIDsByParentID 根据家长ID获取所有关联学生的ID
func GetStudentIDsByParentID(parentID string) ([]int, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 执行查询
	rows, err := db.Query("SELECT student_id FROM parent_student_relations WHERE parent_id = ?", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	var studentIDs []int
	for rows.Next() {
		var studentID int
		err := rows.Scan(&studentID)
		if err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, studentID)
	}

	return studentIDs, nil
}

// SaveParentStudentRelation 保存家长学生关系
func SaveParentStudentRelation(parentID string, studentID string) error {
	// 获取数据库连接
	db := database.GetDB()

	// 检查关系是否已存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM parent_student_relations WHERE parent_id = ? AND student_id = ?",
		parentID, studentID).Scan(&count)
	if err != nil {
		return err
	}

	// 如果关系已存在，不需要再次添加
	if count > 0 {
		return nil
	}

	// 插入新的关系
	_, err = db.Exec(
		"INSERT INTO parent_student_relations (parent_id, student_id) VALUES (?, ?)",
		parentID, studentID,
	)

	return err
}

// DeleteParentStudentRelation 删除家长学生关系
func DeleteParentStudentRelation(parentID string, studentID int) error {
	// 获取数据库连接
	db := database.GetDB()

	// 删除关系
	_, err := db.Exec(
		"DELETE FROM parent_student_relations WHERE parent_id = ? AND student_id = ?",
		parentID, studentID,
	)

	return err
}

// GetParentStudentRelationByID 通过ID获取家长学生关系
func GetParentStudentRelationByID(id int) (*ParentStudentRelation, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 查询关系
	var relation ParentStudentRelation
	err := db.QueryRow(
		`SELECT psr.id, psr.parent_id, psr.student_id, s.full_name
		FROM parent_student_relations psr
		JOIN students s ON psr.student_id = s.id
		WHERE psr.id = ?`,
		id,
	).Scan(&relation.ID, &relation.ParentID, &relation.StudentID, &relation.StudentName)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("relation not found")
		}
		return nil, err
	}

	return &relation, nil
}

// ClearAllParentStudentRelations 清空所有家长-学生关系
func ClearAllParentStudentRelations() error {
	// 获取数据库连接
	db := database.GetDB()

	// 删除所有关系
	_, err := db.Exec("DELETE FROM parent_student_relations")
	return err
}

// GetParentsByStudentID 根据学生ID获取所有家长的钉钉ID
func GetParentsByStudentID(studentID int) ([]string, error) {
	// 获取数据库连接
	db := database.GetDB()

	// 首先获取学生的钉钉ID
	var studentDingTalkID string
	err := db.QueryRow("SELECT dingtalk_id FROM students WHERE id = ?", studentID).Scan(&studentDingTalkID)
	if err != nil {
		return nil, err
	}

	// 如果学生没有钉钉ID或钉钉ID为0，则返回空列表
	if studentDingTalkID == "" || studentDingTalkID == "0" {
		return []string{}, nil
	}

	// 执行查询获取关联的家长ID
	rows, err := db.Query("SELECT parent_id FROM parent_student_relations WHERE student_id = ?", studentDingTalkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 收集家长ID
	var parentIDs []string
	for rows.Next() {
		var parentID string
		if err := rows.Scan(&parentID); err != nil {
			return nil, err
		}
		parentIDs = append(parentIDs, parentID)
	}

	return parentIDs, nil
}
