package main

import (
	"backend/internal/config"
	admincontroller "backend/internal/controller/admin"
	appcontroller "backend/internal/controller/app"
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	"backend/internal/router"
	adminservice "backend/internal/service/admin"
	appservice "backend/internal/service/app"
	commonservice "backend/internal/service/common"
	"backend/internal/utils/jwt"

	"go.uber.org/zap"
)

// buildDependencies 组装当前应用实例的服务、Handler 和中间件。
func buildDependencies(cfg config.Config, dao *mysql.DAO, logger *zap.Logger) router.Dependencies {
	tokens := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpirationDays)
	commonOptionsService := commonservice.NewOptionsService(dao, logger)
	appVersionService := appservice.NewAppVersionService(dao, logger)
	appAvatarService := appservice.NewAvatarService(dao, cfg.Storage, logger)
	appFeedbackService := appservice.NewFeedbackService(dao, cfg.Storage, logger)
	appOptionsService := appservice.NewOptionsService(dao, commonOptionsService, logger)
	appProfileService := appservice.NewProfileService(dao, cfg.Storage, logger)
	appRecordService := appservice.NewRecordService(dao, appProfileService, logger)
	appUserService := appservice.NewUserService(dao, tokens, cfg.Storage, logger)
	adminApartmentService := adminservice.NewApartmentService(dao, logger)
	adminAppVersionService := adminservice.NewAppVersionService(dao, logger)
	adminAuthService := adminservice.NewAuthService(dao, tokens, logger)
	adminClassService := adminservice.NewClassService(dao, logger)
	adminFeedbackService := adminservice.NewFeedbackService(dao, cfg.Storage, logger)
	adminLeaveTypeService := adminservice.NewLeaveTypeService(dao, logger)
	adminOverviewService := adminservice.NewOverviewService(dao, logger)
	adminRecordService := adminservice.NewRecordService(dao, logger)
	adminStudentService := adminservice.NewStudentService(dao, cfg.Storage, logger)
	appVersionHandler := appcontroller.NewAppVersionHandler(appVersionService, cfg.App.PackageDir, logger)
	appAvatarHandler := appcontroller.NewAvatarHandler(appAvatarService, logger)
	appFeedbackHandler := appcontroller.NewFeedbackHandler(appFeedbackService, logger)
	appOptionsHandler := appcontroller.NewOptionsHandler(appOptionsService, commonOptionsService, logger)
	appProfileHandler := appcontroller.NewProfileHandler(appProfileService, logger)
	appRecordHandler := appcontroller.NewRecordHandler(appRecordService, logger)
	appUserHandler := appcontroller.NewUserHandler(appUserService, logger)
	adminApartmentHandler := admincontroller.NewApartmentHandler(adminApartmentService, logger)
	adminAppVersionHandler := admincontroller.NewAppVersionHandler(adminAppVersionService, logger)
	adminAuthHandler := admincontroller.NewAuthHandler(adminAuthService, logger)
	adminClassHandler := admincontroller.NewClassHandler(adminClassService, logger)
	adminFeedbackHandler := admincontroller.NewFeedbackHandler(adminFeedbackService, logger)
	adminLeaveTypeHandler := admincontroller.NewLeaveTypeHandler(adminLeaveTypeService, logger)
	adminOptionsHandler := admincontroller.NewOptionsHandler(commonOptionsService, logger)
	adminOverviewHandler := admincontroller.NewOverviewHandler(adminOverviewService, logger)
	adminRecordHandler := admincontroller.NewRecordHandler(adminRecordService, logger)
	adminStudentHandler := admincontroller.NewStudentHandler(adminStudentService, logger)
	return router.Dependencies{
		Logger:               logger,
		AppVersions:          appVersionHandler,
		AppAvatar:            appAvatarHandler,
		AppFeedback:          appFeedbackHandler,
		AppOptions:           appOptionsHandler,
		AppProfile:           appProfileHandler,
		AppRecord:            appRecordHandler,
		AppUser:              appUserHandler,
		AdminApartment:       adminApartmentHandler,
		AdminAppVersion:      adminAppVersionHandler,
		AdminAuth:            adminAuthHandler,
		AdminClass:           adminClassHandler,
		AdminFeedback:        adminFeedbackHandler,
		AdminLeaveType:       adminLeaveTypeHandler,
		AdminOptions:         adminOptionsHandler,
		AdminOverview:        adminOverviewHandler,
		AdminRecord:          adminRecordHandler,
		AdminStudent:         adminStudentHandler,
		AuthMiddleware:       middleware.JWTAuthMiddleware(tokens, dao, logger),
		AppVersionMiddleware: middleware.AppVersionMiddleware(appVersionService, logger),
	}
}
