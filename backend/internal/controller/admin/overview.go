package admin

import (
	"backend/internal/response"
	adminservice "backend/internal/service/admin"

	"github.com/gin-gonic/gin"
)

func GetOverviewHandler(c *gin.Context) {
	data, err := adminservice.GetAdminOverview()
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}
