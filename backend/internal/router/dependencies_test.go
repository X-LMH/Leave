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
	"backend/internal/testutil"
	"backend/internal/utils/jwt"
	"testing"

	"go.uber.org/zap"
)

func testDependencies(t *testing.T, dao *mysql.DAO, settings config.Config) Dependencies {
	limiter := testutil.NewRateLimiter(t)
	tokens := jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays)
	commonOptionsService := commonservice.NewOptionsService(dao, zap.NewNop())
	appVersionService := appservice.NewAppVersionService(dao, zap.NewNop())
	appAvatarService := appservice.NewAvatarService(dao, settings.Storage, zap.NewNop())
	appFeedbackService := appservice.NewFeedbackService(dao, settings.Storage, zap.NewNop())
	appOptionsService := appservice.NewOptionsService(dao, commonOptionsService, zap.NewNop())
	appProfileService := appservice.NewProfileService(dao, settings.Storage, zap.NewNop())
	appRecordService := appservice.NewRecordService(dao, appProfileService, zap.NewNop())
	appUserService := appservice.NewUserService(dao, tokens, settings.Storage, zap.NewNop())
	adminApartmentService := adminservice.NewApartmentService(dao, zap.NewNop())
	adminAppVersionService := adminservice.NewAppVersionService(dao, zap.NewNop())
	adminAuthService := adminservice.NewAuthService(dao, tokens, zap.NewNop())
	adminClassService := adminservice.NewClassService(dao, zap.NewNop())
	adminFeedbackService := adminservice.NewFeedbackService(dao, settings.Storage, zap.NewNop())
	adminLeaveTypeService := adminservice.NewLeaveTypeService(dao, zap.NewNop())
	adminOverviewService := adminservice.NewOverviewService(dao, zap.NewNop())
	adminRecordService := adminservice.NewRecordService(dao, zap.NewNop())
	adminStudentService := adminservice.NewStudentService(dao, settings.Storage, zap.NewNop())
	appVersionHandler := appcontroller.NewAppVersionHandler(appVersionService, settings.App.PackageDir, zap.NewNop())
	appAvatarHandler := appcontroller.NewAvatarHandler(appAvatarService, zap.NewNop())
	appFeedbackHandler := appcontroller.NewFeedbackHandler(appFeedbackService, zap.NewNop())
	appOptionsHandler := appcontroller.NewOptionsHandler(appOptionsService, commonOptionsService, zap.NewNop())
	appProfileHandler := appcontroller.NewProfileHandler(appProfileService, zap.NewNop())
	appRecordHandler := appcontroller.NewRecordHandler(appRecordService, zap.NewNop())
	appUserHandler := appcontroller.NewUserHandler(appUserService, zap.NewNop(), limiter)
	adminApartmentHandler := admincontroller.NewApartmentHandler(adminApartmentService, zap.NewNop())
	adminAppVersionHandler := admincontroller.NewAppVersionHandler(adminAppVersionService, zap.NewNop())
	adminAuthHandler := admincontroller.NewAuthHandler(adminAuthService, zap.NewNop(), limiter)
	adminClassHandler := admincontroller.NewClassHandler(adminClassService, zap.NewNop())
	adminFeedbackHandler := admincontroller.NewFeedbackHandler(adminFeedbackService, zap.NewNop())
	adminLeaveTypeHandler := admincontroller.NewLeaveTypeHandler(adminLeaveTypeService, zap.NewNop())
	adminOptionsHandler := admincontroller.NewOptionsHandler(commonOptionsService, zap.NewNop())
	adminOverviewHandler := admincontroller.NewOverviewHandler(adminOverviewService, zap.NewNop())
	adminRecordHandler := admincontroller.NewRecordHandler(adminRecordService, zap.NewNop())
	adminStudentHandler := admincontroller.NewStudentHandler(adminStudentService, zap.NewNop())
	return Dependencies{
		RateLimiter:          limiter,
		Logger:               zap.NewNop(),
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
		AuthMiddleware:       middleware.JWTAuthMiddleware(tokens, dao, zap.NewNop()),
		AppVersionMiddleware: middleware.AppVersionMiddleware(appVersionService, zap.NewNop()),
	}
}
