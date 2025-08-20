package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/itsHenry35/canteen-management-system/api/middlewares"
	"github.com/itsHenry35/canteen-management-system/models"
	"github.com/itsHenry35/canteen-management-system/services"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// ScanStudentQRCodeRequest 扫描学生二维码请求
type ScanStudentQRCodeRequest struct {
	QRData string `json:"qr_data"`
}

// ScanStudentQRCodeResponse 扫描学生二维码响应
type ScanStudentQRCodeResponse struct {
	StudentID    int             `json:"student_id"`
	StudentName  string          `json:"student_name"`
	HasSelected  bool            `json:"has_selected"`
	MealType     models.MealType `json:"meal_type"`
	HasCollected bool            `json:"has_collected"`
}

// ScanStudentQRCode 扫描学生二维码并在匹配时记录取餐
func ScanStudentQRCode(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var req ScanStudentQRCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的请求")
		return
	}

	// 获取当前用户角色
	role, ok := middlewares.GetRoleFromContext(r)
	if !ok {
		utils.ResponseError(w, http.StatusUnauthorized, "无法确认操作人员身份")
		return
	}

	// 根据角色确定餐食类型
	var mealType models.MealType
	switch role {
	case services.RoleCanteenA:
		mealType = models.MealTypeA
	case services.RoleCanteenB:
		mealType = models.MealTypeB
	case services.RoleCanteenTest:
	default:
		utils.ResponseError(w, http.StatusForbidden, "无效的操作人员类型")
		return
	}

	// 解密二维码数据
	studentID, err := utils.ValidateQRCodeData(req.QRData)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "无效的二维码数据")
		return
	}

	// 获取学生信息
	student, err := models.GetStudentByID(studentID)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "未找到学生")
		return
	}

	// 获取学生当前的选餐记录
	selection, err := models.GetStudentCurrentSelection(studentID)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "获取学生选餐信息失败")
		return
	}

	// 构建响应
	resp := ScanStudentQRCodeResponse{
		StudentID:    studentID,
		StudentName:  student.FullName,
		HasSelected:  selection != nil,
		HasCollected: false,
	}

	// 如果有选餐记录，设置餐食类型
	if selection != nil {
		resp.MealType = selection.MealType
	}

	// 如果是测试账号，直接返回响应
	if role == services.RoleCanteenTest {
		utils.ResponseOK(w, resp)
		return
	}

	// 检查学生是否已在今天取餐
	today := time.Now().Format("2006-01-02")
	lastCollectionDate := student.LastMealCollectionDate
	if lastCollectionDate != nil {
		if lastCollectionDate.Format("2006-01-02") == today {
			resp.HasCollected = true
		}
	}

	// 如果餐食类型匹配且学生今天未取餐，则记录取餐
	if selection != nil && selection.MealType == mealType && !resp.HasCollected {
		// 更新学生的最后取餐日期
		now := time.Now()
		student.LastMealCollectionDate = &now

		if err := models.UpdateStudent(student); err != nil {
			utils.ResponseError(w, http.StatusInternalServerError, "更新学生取餐记录失败")
			return
		}
	}

	// 返回响应
	utils.ResponseOK(w, resp)
}
