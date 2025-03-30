package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/itsHenry35/canteen-management-system/api/middlewares"
	"github.com/itsHenry35/canteen-management-system/models"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// CreateStudentRequest 创建学生请求
type CreateStudentRequest struct {
	FullName   string `json:"full_name"`
	Class      string `json:"class"`
	DingTalkID string `json:"dingtalk_id,omitempty"`
}

// UpdateStudentRequest 更新学生请求
type UpdateStudentRequest struct {
	FullName   string `json:"full_name,omitempty"`
	Class      string `json:"class,omitempty"`
	DingTalkID string `json:"dingtalk_id,omitempty"`
}

// GetAllStudents 获取所有学生
func GetAllStudents(w http.ResponseWriter, r *http.Request) {
	// 获取学生列表
	students, err := models.GetAllStudents()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取学生列表失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, students)
}

// GetStudent 获取学生信息
func GetStudent(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的学生ID")
		return
	}

	// 获取学生信息
	student, err := models.GetStudentByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "未找到学生")
		return
	}

	// 返回响应
	utils.ResponseOK(w, student)
}

// CreateStudent 创建学生
func CreateStudent(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req CreateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 创建学生
	student, err := models.CreateStudent(req.FullName, req.Class, req.DingTalkID)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "创建学生失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, student)
}

// UpdateStudent 更新学生信息
func UpdateStudent(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的学生ID")
		return
	}

	// 解析请求
	var req UpdateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 获取学生信息
	student, err := models.GetStudentByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "未找到学生")
		return
	}

	// 更新学生信息
	if req.FullName != "" {
		student.FullName = req.FullName
	}
	if req.Class != "" {
		student.Class = req.Class
	}
	if req.DingTalkID != "" {
		student.DingTalkID = req.DingTalkID
	}

	if err := models.UpdateStudent(student); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "更新学生失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, student)
}

// DeleteStudent 删除学生
func DeleteStudent(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的学生ID")
		return
	}

	// 删除学生
	if err := models.DeleteStudent(id); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "删除学生失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]bool{"success": true})
}

// GetStudentQRCodeData 获取学生二维码数据
func GetStudentQRCodeData(w http.ResponseWriter, r *http.Request) {
	// 解析路径参数
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的学生ID")
		return
	}

	// 获取学生信息
	student, err := models.GetStudentByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "未找到学生")
		return
	}

	// 生成加密的二维码数据
	qrData, err := utils.GenerateQRCodeData(student.ID, student.FullName)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "生成二维码数据失败")
		return
	}

	// 返回响应
	utils.ResponseOK(w, map[string]string{
		"qr_data": qrData,
	})
}

// GetStudentMealSelections 获取学生选餐记录
func GetStudentMealSelections(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取学生ID
	studentID, ok := middlewares.GetUserIDFromContext(r)
	if !ok {
		utils.ResponseError(w, http.StatusUnauthorized, "未授权")
		return
	}

	// 获取学生的所有选餐记录
	selections, err := models.GetMealSelectionsByStudent(studentID)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取选餐记录失败")
		return
	}

	// 获取当前可选餐
	selectableMeals, err := models.GetCurrentSelectableMeals()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取可选餐失败")
		return
	}

	// 构建符合API文档的响应格式
	var responseSelections []map[string]interface{}

	// 处理已选餐记录
	for _, selection := range selections {
		meal := selection.Meal
		if meal == nil {
			continue
		}

		// 检查这个餐是否在当前可选餐列表中
		selectable := false
		for _, sMeal := range selectableMeals {
			if sMeal.ID == meal.ID {
				selectable = true
				break
			}
		}

		responseSelections = append(responseSelections, map[string]interface{}{
			"meal_id":              meal.ID,
			"meal_type":            selection.MealType,
			"selectable":           selectable,
			"id":                   meal.ID,
			"name":                 meal.Name,
			"selection_start_time": meal.SelectionStartTime,
			"selection_end_time":   meal.SelectionEndTime,
			"effective_start_date": meal.EffectiveStartDate,
			"effective_end_date":   meal.EffectiveEndDate,
			"image_path":           meal.ImagePath,
		})
	}

	// 添加可选但尚未选择的餐
	for _, meal := range selectableMeals {
		// 检查这个餐是否已经在响应中
		exists := false
		for _, rs := range responseSelections {
			if rs["id"].(int) == meal.ID {
				exists = true
				break
			}
		}

		if !exists {
			responseSelections = append(responseSelections, map[string]interface{}{
				"meal_id":              meal.ID,
				"meal_type":            nil,
				"selectable":           true,
				"id":                   meal.ID,
				"name":                 meal.Name,
				"selection_start_time": meal.SelectionStartTime,
				"selection_end_time":   meal.SelectionEndTime,
				"effective_start_date": meal.EffectiveStartDate,
				"effective_end_date":   meal.EffectiveEndDate,
				"image_path":           meal.ImagePath,
			})
		}
	}

	// 返回响应
	utils.ResponseOK(w, map[string]interface{}{
		"selections": responseSelections,
	})
}

// GetStudentSelections 获取所有学生选餐统计
func GetStudentSelections(w http.ResponseWriter, r *http.Request) {
	// 获取所有学生
	students, err := models.GetAllStudents()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取学生列表失败")
		return
	}
	totalStudents := len(students)

	// 获取所有餐
	meals, err := models.GetAllMeals()
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取餐列表失败")
		return
	}

	// 构建选餐数据列表
	var selectionsData []map[string]interface{}

	for _, meal := range meals {
		// 获取该餐的所有选餐记录
		selections, err := models.GetMealSelectionsByMeal(meal.ID)
		if err != nil {
			utils.ResponseError(w, http.StatusInternalServerError, "获取选餐记录失败")
			return
		}

		// 统计该餐的选餐情况
		var typeACount, typeBCount int
		for _, selection := range selections {
			if selection.MealType == models.MealTypeA {
				typeACount++
			} else if selection.MealType == models.MealTypeB {
				typeBCount++
			}
		}

		// 构建该餐的选餐数据
		mealData := map[string]interface{}{
			"meal_id":          meal.ID,
			"name":             meal.Name,
			"image_path":       meal.ImagePath,
			"effective_start":  meal.EffectiveStartDate,
			"effective_end":    meal.EffectiveEndDate,
			"selection_start":  meal.SelectionStartTime,
			"selection_end":    meal.SelectionEndTime,
			"total":            totalStudents,
			"total_a":          typeACount,
			"total_b":          typeBCount,
			"total_unselected": totalStudents - len(selections),
		}

		selectionsData = append(selectionsData, mealData)
	}

	// 返回响应
	utils.ResponseOK(w, selectionsData)
}
