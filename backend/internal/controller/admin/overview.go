package admin

import (
	"backend/internal/response"
	adminservice "backend/internal/service/admin"

	"github.com/gin-gonic/gin"
)

type OverviewHandler struct {
	overview *adminservice.OverviewService
}

func NewOverviewHandler(overview *adminservice.OverviewService) *OverviewHandler {
	return &OverviewHandler{
		overview: overview,
	}
}

func (o *OverviewHandler) GetOverviewHandler(c *gin.Context) {
	data, err := o.overview.GetAdminOverview()
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}
