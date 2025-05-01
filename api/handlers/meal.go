package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/itsHenry35/canteen-management-system/api/middlewares"
	"github.com/itsHenry35/canteen-management-system/models"
	"github.com/itsHenry35/canteen-management-system/scheduler"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// CreateMealRequest 创建餐请求
type CreateMealRequest struct {
	Name               string    `json:"name"`                 // 餐名
	SelectionStartTime time.Time `json:"selection_start_time"` // 选餐开始时间
	SelectionEndTime   time.Time `json:"selection_end_time"`   // 选餐结束时间
	EffectiveStartDate time.Time `json:"effective_start_date"` // 领餐开始生效日期
	EffectiveEndDate   time.Time `json:"effective_end_date"`   // 领餐结束生效日期
	Image              string    `json:"image"`                // Base64编码的图片
}

// UpdateMealRequest 更新餐请求
type UpdateMealRequest struct {
	Name               string    `json:"name,omitempty"`       // 餐名（可选）
	SelectionStartTime time.Time `json:"selection_start_time"` // 选餐开始时间
	SelectionEndTime   time.Time `json:"selection_end_time"`   // 选餐结束时间
	EffectiveStartDate time.Time `json:"effective_start_date"` // 领餐开始生效日期
	EffectiveEndDate   time.Time `json:"effective_end_date"`   // 领餐结束生效日期
	Image              string    `json:"image,omitempty"`      // Base64编码的图片（可选）
}

// MealSelectionRequest 选餐请求
type MealSelectionRequest struct {
	MealID   int             `json:"meal_id"`
	MealType models.MealType `json:"meal_type"`
}

// BatchMealSelectionRequest 批量选餐请求
type BatchMealSelectionRequest struct {
	StudentIDs []int           `json:"student_ids,omitempty"`
	MealID     int             `json:"meal_id"`
	MealType   models.MealType `json:"meal_type"`
}

// GetAllMeals 获取所有餐
func GetAllMeals(w http.ResponseWriter, r *http.Request) {
	// 获取餐列表
	meals, err := models.GetAllMeals()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取餐列表失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, meals)
}

// GetMeal 获取餐信息
func GetMeal(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的餐ID")
		return
	}

	// 获取餐信息
	meal, err := models.GetMealByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "未找到餐")
		return
	}

	// 返回响应
	utils.ResponseOK(w, meal)
}

// CreateMeal 创建餐
func CreateMeal(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req CreateMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 确保图片目录存在
	mealImgDir := "./data/images"
	if err := os.MkdirAll(mealImgDir, 0755); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "创建目录失败")
		return
	}

	// 保存图片
	timestamp := time.Now().Unix()
	imgFileName := utils.SaveBase64Image(req.Image, mealImgDir, "meal", timestamp)
	if imgFileName == "" {
		utils.ResponseError(w, http.StatusBadRequest, "保存图片失败")
		return
	}
	imgPath := filepath.Join("/static/images", imgFileName)

	// 创建餐
	meal, err := models.CreateMeal(req.Name, req.SelectionStartTime, req.SelectionEndTime, req.EffectiveStartDate, req.EffectiveEndDate, imgPath)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "创建餐失败: "+err.Error())
		return
	}

	// 更新自动选餐任务
	if err := scheduler.CheckAndUpdateTasks(); err != nil {
		log.Printf("更新自动选餐任务失败: %v", err)
		// 不要因为更新任务失败而中断请求
	}

	// 返回响应
	utils.ResponseOK(w, meal)
}

// UpdateMeal 更新餐
func UpdateMeal(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的餐ID")
		return
	}

	// 获取餐信息
	meal, err := models.GetMealByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "未找到餐")
		return
	}

	// 解析请求
	var req UpdateMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 更新餐的基本信息
	meal.SelectionStartTime = req.SelectionStartTime
	meal.SelectionEndTime = req.SelectionEndTime
	meal.EffectiveStartDate = req.EffectiveStartDate
	meal.EffectiveEndDate = req.EffectiveEndDate

	// 如果提供了新图片，则更新图片
	if req.Image != "" {
		// 确保图片目录存在
		mealImgDir := "./data/images"
		if err := os.MkdirAll(mealImgDir, 0755); err != nil {
			utils.ResponseError(w, http.StatusInternalServerError, "创建目录失败")
			return
		}

		// 保存新图片
		timestamp := time.Now().Unix()
		imgFileName := utils.SaveBase64Image(req.Image, mealImgDir, "meal", timestamp)
		if imgFileName == "" {
			utils.ResponseError(w, http.StatusBadRequest, "保存图片失败")
			return
		}
		newImgPath := filepath.Join("/static/images", imgFileName)

		// 删除旧图片
		if meal.ImagePath != "" {
			oldPhysicalPath := "." + meal.ImagePath // 将URL路径转换为文件系统路径
			os.Remove(oldPhysicalPath)
		}

		// 更新图片路径
		meal.ImagePath = newImgPath
	}

	if req.Name != "" {
		meal.Name = req.Name
	}

	// 更新餐
	if err := models.UpdateMeal(meal); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "更新餐失败: "+err.Error())
		return
	}

	// 更新自动选餐任务
	if err := scheduler.CheckAndUpdateTasks(); err != nil {
		log.Printf("更新自动选餐任务失败: %v", err)
		// 不要因为更新任务失败而中断请求
	}

	// 返回响应
	utils.ResponseOK(w, meal)
}

// DeleteMeal 删除餐
func DeleteMeal(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的餐ID")
		return
	}

	// 删除餐
	if err := models.DeleteMeal(id); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "删除餐失败")
		return
	}

	// 更新自动选餐任务
	if err := scheduler.CheckAndUpdateTasks(); err != nil {
		log.Printf("更新自动选餐任务失败: %v", err)
		// 不要因为更新任务失败而中断请求
	}

	// 返回响应
	utils.ResponseOK(w, map[string]bool{"success": true})
}

// GetMealSelections 获取餐的选餐情况
func GetMealSelections(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的餐ID")
		return
	}

	// 获取餐的所有选餐记录
	selections, err := models.GetMealSelectionsByMeal(id)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取选餐记录失败")
		return
	}

	// 获取所有学生
	allStudents, err := models.GetAllStudents()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取学生列表失败")
		return
	}

	// 创建已选餐学生ID的集合，以便快速检查学生是否已选餐
	selectedStudentIDs := make(map[int]bool)

	// 创建A餐和B餐学生ID列表
	var typeAStudentIDs []int
	var typeBStudentIDs []int

	// 遍历选餐记录，分类学生
	for _, selection := range selections {
		selectedStudentIDs[selection.StudentID] = true

		if selection.MealType == models.MealTypeA {
			typeAStudentIDs = append(typeAStudentIDs, selection.StudentID)
		} else if selection.MealType == models.MealTypeB {
			typeBStudentIDs = append(typeBStudentIDs, selection.StudentID)
		}
	}

	// 创建未选餐学生ID列表
	var unselectedStudentIDs []int
	for _, student := range allStudents {
		if !selectedStudentIDs[student.ID] {
			unselectedStudentIDs = append(unselectedStudentIDs, student.ID)
		}
	}

	// 返回响应
	utils.ResponseOK(w, map[string]interface{}{
		"a":          typeAStudentIDs,
		"b":          typeBStudentIDs,
		"unselected": unselectedStudentIDs,
	})
}

// StudentSelectMeal 学生选餐
func StudentSelectMeal(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req MealSelectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 验证餐食类型
	if req.MealType != models.MealTypeA && req.MealType != models.MealTypeB {
		utils.ResponseError(w, http.StatusBadRequest, "无效的餐食类型")
		return
	}

	// 从上下文获取学生ID
	studentID, ok := middlewares.GetUserIDFromContext(r)
	if !ok {
		utils.ResponseError(w, http.StatusUnauthorized, "未授权")
		return
	}

	// 创建选餐记录
	selection, err := models.CreateMealSelection(studentID, req.MealID, req.MealType, true)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "选餐失败: "+err.Error())
		return
	}

	// 返回响应
	utils.ResponseOK(w, selection)
}

// BatchSelectMeals 批量选餐
func BatchSelectMeals(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req BatchMealSelectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 验证餐食类型
	if req.MealType != models.MealTypeA && req.MealType != models.MealTypeB {
		utils.ResponseError(w, http.StatusBadRequest, "无效的餐食类型")
		return
	}

	// 批量选餐
	count, err := models.BatchSelectMeals(req.StudentIDs, req.MealID, req.MealType)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "批量选餐失败: "+err.Error())
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]interface{}{
		"success": true,
		"count":   count,
	})
}

// ImportSelectionRequest 导入选餐请求
type ImportSelectionRequest struct {
	Method   string          `json:"method"` // student_id 或 dingtalk_id
	ID       string          `json:"id"`
	MealType models.MealType `json:"meal_type"`
	MealID   int             `json:"meal_id"`
}

// ImportSelection 导入选餐
func ImportSelection(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req ImportSelectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 验证餐食类型
	if req.MealType != models.MealTypeA && req.MealType != models.MealTypeB {
		utils.ResponseError(w, http.StatusBadRequest, "无效的餐食类型")
		return
	}

	// 验证请求方法
	if req.Method != "student_id" && req.Method != "dingtalk_id" {
		utils.ResponseError(w, http.StatusBadRequest, "无效的方法，必须为 student_id 或 dingtalk_id")
		return
	}

	// 根据方法确定目标学生ID
	var studentID int
	var err error

	if req.Method == "student_id" {
		// 直接从请求获取学生ID
		studentID, err = strconv.Atoi(req.ID)
		if err != nil {
			utils.ResponseError(w, http.StatusBadRequest, "无效的学生ID")
			return
		}

		// 验证学生是否存在
		student, err := models.GetStudentByID(studentID)
		if err != nil {
			utils.ResponseError(w, http.StatusNotFound, "未找到学生")
			return
		}

		// 使用验证过的学生ID
		studentID = student.ID
	} else {
		// 使用钉钉ID查找学生
		student, err := models.GetStudentByDingTalkID(req.ID)
		if err != nil {
			// 如果学生未找到，可能是家长的钉钉ID，尝试查找关联学生
			parents, err := models.GetStudentsByParentID(req.ID)
			if err != nil || len(parents) == 0 {
				utils.ResponseError(w, http.StatusNotFound, "未找到与该钉钉ID关联的学生")
				return
			}

			// 使用第一个关联学生的ID
			student, err := models.GetStudentByDingTalkID(parents[0].StudentID)
			if err != nil {
				utils.ResponseError(w, http.StatusNotFound, "未找到学生")
				return
			}
			studentID = student.ID
		} else {
			studentID = student.ID
		}
	}

	// 创建选餐记录
	_, err = models.CreateMealSelection(studentID, req.MealID, req.MealType, false)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "选餐失败: "+err.Error())
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]interface{}{
		"success": true,
		"message": "导入选餐成功",
	})
}

// GetCurrentSelectableMeals 获取当前可选餐
func GetCurrentSelectableMeals(w http.ResponseWriter, r *http.Request) {
	// 获取当前可选餐
	meals, err := models.GetCurrentSelectableMeals()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取可选餐失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, meals)
}

// CleanupExpiredMeals 清理过期的餐
func CleanupExpiredMeals(w http.ResponseWriter, r *http.Request) {
	// 清理过期的餐
	err := models.CleanupExpiredMeals()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "清理过期餐失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]bool{"success": true})
}

// GetStudentCurrentSelection 获取学生当前日期的选餐
func GetStudentCurrentSelection(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取学生ID
	studentID, ok := middlewares.GetUserIDFromContext(r)
	if !ok {
		utils.ResponseError(w, http.StatusUnauthorized, "未授权")
		return
	}

	// 获取学生当前日期的选餐
	selection, err := models.GetStudentCurrentSelection(studentID)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取当前选餐失败")
		return
	}

	// 判断是否有选餐
	if selection == nil {
		utils.ResponseOK(w, map[string]interface{}{
			"has_selection": false,
		})
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]interface{}{
		"has_selection": true,
		"selection":     selection,
	})
}

// NotifyUnselectedStudents 手动提醒未选餐学生
func NotifyUnselectedStudents(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req NotifyUnselectedStudentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 调用模型层的通知函数
	err := models.NotifyUnselectedStudentsByMealId(req.MealID)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]interface{}{
		"success": true,
		"message": "未选餐提醒已开始发送",
	})
}
