package admin

import (
	"backend/internal/dao/mysql"
	logutil "backend/internal/logger"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// managementError 仅处理通用管理接口共享的业务错误，特殊业务由各 Handler 映射。
// 未知系统错误在这里记录，调用方不再重复记录。
func managementError(c *gin.Context, err error, log *zap.Logger) {
	switch {
	case errors.Is(err, adminservice.ErrorInvalidAdminInput):
		response.Error(c, response.CodeInvalidParam)
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, response.CodeManagementNotFound)
	case mysql.IsDuplicateKeyError(err):
		response.Error(c, response.CodeManagementDuplicate)
	case errors.Is(err, adminservice.ErrorApartmentInUse):
		response.Error(c, response.CodeApartmentInUse)
	case errors.Is(err, mysql.ErrorClassInUse):
		response.Error(c, response.CodeClassInUse)
	default:
		log.Error("管理端操作失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
	}
}
func managementID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || value == 0 {
		response.Error(c, response.CodeInvalidParam)
		return 0, false
	}
	return uint(value), true
}
func enabledFilter(c *gin.Context) (*bool, bool) {
	value := c.Query("is_enabled")
	if value == "" {
		return nil, true
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		response.Error(c, response.CodeInvalidParam)
		return nil, false
	}
	return &enabled, true
}
