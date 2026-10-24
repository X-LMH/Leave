package admin

import (
	"backend/internal/response"
	adminservice "backend/internal/service/admin"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OverviewHandler struct {
	logger   *zap.Logger
	overview *adminservice.OverviewService
}

func NewOverviewHandler(overview *adminservice.OverviewService, logger *zap.Logger) *OverviewHandler {
	return &OverviewHandler{
		logger:   logger,
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
