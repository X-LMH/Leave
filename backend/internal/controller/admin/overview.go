package admin

import (
	logutil "backend/internal/logger"
	"backend/internal/request"
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
	log := o.logger.With(logutil.OperationFields(c, "admin.overview.get_overview")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	data, err := o.overview.GetAdminOverview()
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}
