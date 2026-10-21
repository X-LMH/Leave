package router

import (
	"backend/internal/config"
	admincontroller "backend/internal/controller/admin"
	appcontroller "backend/internal/controller/app"
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	adminservice "backend/internal/service/admin"
	appservice "backend/internal/service/app"
	commonservice "backend/internal/service/common"
	"backend/internal/utils/jwt"
)

func testDependencies(dao *mysql.DAO, settings config.Config) Dependencies {
	tokens := jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays)
	commonOptionsService := commonservice.NewOptionsService(dao)
	appVersionService := appservice.NewAppVersionService(dao)
	appAvatarService := appservice.NewAvatarService(dao, settings.Storage)
	appFeedbackService := appservice.NewFeedbackService(dao, settings.Storage)
	appOptionsService := appservice.NewOptionsService(dao, commonOptionsService)
	appProfileService := appservice.NewProfileService(dao, settings.Storage)
	appRecordService := appservice.NewRecordService(dao, appProfileService)
	appUserService := appservice.NewUserService(dao, tokens, settings.Storage)
	adminApartmentService := adminservice.NewApartmentService(dao)
	adminAppVersionService := adminservice.NewAppVersionService(dao)
	adminAuthService := adminservice.NewAuthService(dao, tokens)
	adminClassService := adminservice.NewClassService(dao)
	adminFeedbackService := adminservice.NewFeedbackService(dao, settings.Storage)
	adminLeaveTypeService := adminservice.NewLeaveTypeService(dao)
	adminOverviewService := adminservice.NewOverviewService(dao)
	adminRecordService := adminservice.NewRecordService(dao)
	adminStudentService := adminservice.NewStudentService(dao, settings.Storage)
	appVersionHandler := appcontroller.NewAppVersionHandler(appVersionService, settings.App.PackageDir)
	appAvatarHandler := appcontroller.NewAvatarHandler(appAvatarService)
	appFeedbackHandler := appcontroller.NewFeedbackHandler(appFeedbackService)
	appOptionsHandler := appcontroller.NewOptionsHandler(appOptionsService, commonOptionsService)
	appProfileHandler := appcontroller.NewProfileHandler(appProfileService)
	appRecordHandler := appcontroller.NewRecordHandler(appRecordService)
	appUserHandler := appcontroller.NewUserHandler(appUserService)
	adminApartmentHandler := admincontroller.NewApartmentHandler(adminApartmentService)
	adminAppVersionHandler := admincontroller.NewAppVersionHandler(adminAppVersionService)
	adminAuthHandler := admincontroller.NewAuthHandler(adminAuthService)
	adminClassHandler := admincontroller.NewClassHandler(adminClassService)
	adminFeedbackHandler := admincontroller.NewFeedbackHandler(adminFeedbackService)
	adminLeaveTypeHandler := admincontroller.NewLeaveTypeHandler(adminLeaveTypeService)
	adminOptionsHandler := admincontroller.NewOptionsHandler(commonOptionsService)
	adminOverviewHandler := admincontroller.NewOverviewHandler(adminOverviewService)
	adminRecordHandler := admincontroller.NewRecordHandler(adminRecordService)
	adminStudentHandler := admincontroller.NewStudentHandler(adminStudentService)
	return Dependencies{
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
		AuthMiddleware:       middleware.JWTAuthMiddleware(tokens, dao),
		AppVersionMiddleware: middleware.AppVersionMiddleware(appVersionService),
	}
}
