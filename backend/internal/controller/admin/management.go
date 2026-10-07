package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func managementError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, adminservice.ErrorInvalidAdminInput):
		response.Error(c, response.CodeInvalidParam)
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, response.CodeManagementNotFound)
	case mysql.IsDuplicateKeyError(err):
		response.Error(c, response.CodeManagementDuplicate)
	case errors.Is(err, mysql.ErrorClassInUse):
		response.Error(c, response.CodeClassInUse)
	default:
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
