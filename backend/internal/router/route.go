package router

import (
	"backend/internal/controller"
	admincontroller "backend/internal/controller/admin"
	appcontroller "backend/internal/controller/app"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Dependencies 显式列出路由所需的已组装组件。
type Dependencies struct {
	RateLimiter          *middleware.RateLimiter
	Logger               *zap.Logger
	AppVersions          *appcontroller.AppVersionHandler
	AppAvatar            *appcontroller.AvatarHandler
	AppFeedback          *appcontroller.FeedbackHandler
	AppOptions           *appcontroller.OptionsHandler
	AppProfile           *appcontroller.ProfileHandler
	AppRecord            *appcontroller.RecordHandler
	AppUser              *appcontroller.UserHandler
	AdminApartment       *admincontroller.ApartmentHandler
	AdminAppVersion      *admincontroller.AppVersionHandler
	AdminAuth            *admincontroller.AuthHandler
	AdminClass           *admincontroller.ClassHandler
	AdminFeedback        *admincontroller.FeedbackHandler
	AdminLeaveType       *admincontroller.LeaveTypeHandler
	AdminOptions         *admincontroller.OptionsHandler
	AdminOverview        *admincontroller.OverviewHandler
	AdminRecord          *admincontroller.RecordHandler
	AdminStudent         *admincontroller.StudentHandler
	AuthMiddleware       gin.HandlerFunc
	AppVersionMiddleware gin.HandlerFunc
}

func SetupRouter(deps Dependencies) *gin.Engine {
	r := gin.New()
	// 仅信任同机 Nginx，由它覆盖 X-Real-IP；公网直连不能通过转发头伪造 IP。
	r.RemoteIPHeaders = []string{"X-Real-IP"}
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}

	// 全局中间件
	r.Use(middleware.RequestID(), middleware.AccessLogger(deps.Logger), middleware.Recovery(deps.Logger), middleware.CORSMiddleware())

	// API v1
	api := r.Group("/api/v1")

	// 健康检查接口
	api.GET("/health", controller.HealthHandler)

	// 版本接口必须不受版本校验保护，旧客户端才能获取更新信息。
	api.GET("/app-version", deps.AppVersions.GetCurrentAppVersionHandler)
	api.GET("/app-download", deps.AppVersions.DownloadCurrentAppHandler)
	api.HEAD("/app-download", deps.AppVersions.DownloadCurrentAppHandler)

	// App 接口需要校验客户端版本；管理平台不受 App 版本限制。
	app := api.Group("")
	// Public version and download routes are registered above and remain exempt.
	app.Use(appIPRateLimit(deps.RateLimiter), deps.AppVersionMiddleware)

	// App 公共接口
	{
		auth := app.Group("/auth")
		auth.POST("/register", deps.AppUser.RegisterHandler)
		auth.POST("/login", deps.AppUser.LoginHandler)
	}

	user := app.Group("")
	user.Use(deps.AuthMiddleware, deps.RateLimiter.Account(middleware.SideApp, middleware.OperationGeneral))
	{
		user.POST("/logout", deps.AppUser.LogoutHandler)
		user.POST("/client-info", deps.AppUser.UpdateAppInfoHandler)
		user.POST("/profile", deps.AppProfile.ProfileHandler)
		user.POST("/profile/avatar", deps.RateLimiter.Account(middleware.SideApp, middleware.OperationAvatarUpload), deps.AppAvatar.UploadAvatarHandler)
		user.POST("/password", deps.RateLimiter.Account(middleware.SideApp, middleware.OperationPasswordChange), deps.AppUser.ChangePasswordHandler)
		user.GET("/profile", deps.AppProfile.GetProfileHandler)
		user.GET("/home", deps.AppRecord.GetHomeHandler)

		user.GET("/class-options", deps.AppOptions.GetClassOptionsHandler)
		user.GET("/apartments", deps.AppOptions.GetApartmentOptionsHandler)
		user.GET("/leave-types", deps.AppOptions.GetLeaveTypeOptionsHandler)

		user.POST("/record", deps.RateLimiter.Account(middleware.SideApp, middleware.OperationRecordCreate), deps.AppRecord.CreateRecordHandler)
		user.GET("/record/:id", deps.AppRecord.GetRecordHandler)
		user.GET("/records", deps.AppRecord.GetRecordsLIstHandler)
		user.DELETE("/record/:id", deps.AppRecord.DeleteRecordHandler)

		user.POST("/feedback", deps.RateLimiter.Account(middleware.SideApp, middleware.OperationFeedbackCreate), deps.AppFeedback.CreateFeedbackHandler)
	}

	admin := api.Group("/admin")
	admin.Use(func(c *gin.Context) {
		operation := middleware.OperationGeneral
		if c.Request.Method == "POST" && c.FullPath() == "/api/v1/admin/auth/login" {
			operation = middleware.OperationLogin
		}
		if deps.RateLimiter.Check(c, middleware.SideAdmin, operation, middleware.DimensionIP, c.ClientIP()) {
			c.Next()
		}
	})
	adminAuth := admin.Group("/auth")
	{
		adminAuth.POST("/login", deps.AdminAuth.LoginHandler)

		adminAuth.Use(deps.AuthMiddleware, middleware.AdminAuthMiddleware(), deps.RateLimiter.Account(middleware.SideAdmin, middleware.OperationGeneral))
		adminAuth.POST("/password", deps.RateLimiter.Account(middleware.SideAdmin, middleware.OperationPasswordChange), deps.AdminAuth.ChangePasswordHandler)
	}

	admin.Use(deps.AuthMiddleware, middleware.AdminAuthMiddleware(), deps.RateLimiter.Account(middleware.SideAdmin, middleware.OperationGeneral))
	{
		admin.GET("/overview", deps.AdminOverview.GetOverviewHandler)
		admin.GET("/class-options", deps.AdminOptions.GetClassOptionsHandler)
		admin.GET("/apartment-options", deps.AdminOptions.GetApartmentOptionsHandler)
		admin.GET("/leave-type-options", deps.AdminOptions.GetLeaveTypeOptionsHandler)

		admin.GET("/app-versions", deps.AdminAppVersion.GetAppVersionsHandler)
		admin.GET("/app-versions/current", deps.AdminAppVersion.GetCurrentAppVersionHandler)
		admin.GET("/app-versions/:id", deps.AdminAppVersion.GetAppVersionHandler)

		admin.GET("/students", deps.AdminStudent.GetStudentsHandler)
		admin.GET("/records", deps.AdminRecord.GetRecordsHandler)
		admin.GET("/records/:id", deps.AdminRecord.GetRecordHandler)
		admin.PUT("/records/:id", deps.AdminRecord.UpdateRecordHandler)

		admin.GET("/students/:id", deps.AdminStudent.GetStudentHandler)
		admin.PUT("/students/:id", deps.AdminStudent.UpdateStudentHandler)
		admin.PUT("/students/:id/status", deps.AdminStudent.UpdateStudentStatusHandler)
		admin.POST("/students/:id/reset-password", deps.AdminStudent.ResetStudentPasswordHandler)

		admin.GET("/feedbacks", deps.AdminFeedback.GetFeedbacksHandler)
		admin.GET("/feedbacks/:id", deps.AdminFeedback.GetFeedbackHandler)
		admin.PATCH("/feedbacks/:id/handling", deps.AdminFeedback.UpdateFeedbackHandlingHandler)
		admin.GET("/feedbacks/:id/attachments/:index", deps.AdminFeedback.GetFeedbackAttachmentHandler)
		admin.DELETE("/feedbacks/:id", deps.AdminFeedback.DeleteFeedbackHandler)

		admin.GET("/apartments", deps.AdminApartment.GetApartmentsHandler)
		admin.POST("/apartments", deps.AdminApartment.CreateApartmentHandler)
		admin.PUT("/apartments/:id", deps.AdminApartment.UpdateApartmentHandler)
		admin.DELETE("/apartments/:id", deps.AdminApartment.DeleteApartmentHandler)

		admin.GET("/classes", deps.AdminClass.GetClassesHandler)
		admin.POST("/classes", deps.AdminClass.CreateClassHandler)
		admin.PUT("/classes/:id", deps.AdminClass.UpdateClassHandler)
		admin.DELETE("/classes/:id", deps.AdminClass.DeleteClassHandler)

		admin.GET("/leave-types", deps.AdminLeaveType.GetLeaveTypesHandler)
		admin.POST("/leave-types", deps.AdminLeaveType.CreateLeaveTypeHandler)
		admin.PUT("/leave-types/:id", deps.AdminLeaveType.UpdateLeaveTypeHandler)
		admin.DELETE("/leave-types/:id", deps.AdminLeaveType.DeleteLeaveTypeHandler)
	}

	return r
}

func appIPRateLimit(limiter *middleware.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		operation := middleware.OperationGeneral
		if c.Request.Method == "POST" {
			switch c.FullPath() {
			case "/api/v1/auth/login":
				operation = middleware.OperationLogin
			case "/api/v1/auth/register":
				operation = middleware.OperationRegister
			}
		}
		if limiter.Check(c, middleware.SideApp, operation, middleware.DimensionIP, c.ClientIP()) {
			c.Next()
		}
	}
}
