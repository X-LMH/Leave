package router

import (
	"backend/internal/controller"
	admincontroller "backend/internal/controller/admin"
	appcontroller "backend/internal/controller/app"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Dependencies 显式列出路由所需的已组装组件。
type Dependencies struct {
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
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())

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
	app.Use(deps.AppVersionMiddleware)

	// App 公共接口
	{
		auth := app.Group("/auth")
		auth.POST("/register", deps.AppUser.RegisterHandler)
		auth.POST("/login", deps.AppUser.LoginHandler)
	}

	user := app.Group("")
	user.Use(deps.AuthMiddleware)
	{
		user.POST("/logout", deps.AppUser.LogoutHandler)
		user.POST("/client-info", deps.AppUser.UpdateAppInfoHandler)
		user.POST("/profile", deps.AppProfile.ProfileHandler)
		user.POST("/profile/avatar", deps.AppAvatar.UploadAvatarHandler)
		user.POST("/password", deps.AppUser.ChangePasswordHandler)
		user.GET("/profile", deps.AppProfile.GetProfileHandler)

		user.GET("/class-options", deps.AppOptions.GetClassOptionsHandler)
		user.GET("/apartments", deps.AppOptions.GetApartmentOptionsHandler)
		user.GET("/leave-types", deps.AppOptions.GetLeaveTypeOptionsHandler)

		user.POST("/record", deps.AppRecord.CreateRecordHandler)
		user.GET("/record/:id", deps.AppRecord.GetRecordHandler)
		user.GET("/records", deps.AppRecord.GetRecordsLIstHandler)
		user.DELETE("/record/:id", deps.AppRecord.DeleteRecordHandler)

		user.POST("/feedback", deps.AppFeedback.CreateFeedbackHandler)
	}

	admin := api.Group("/admin")
	adminAuth := admin.Group("/auth")
	{
		adminAuth.POST("/login", deps.AdminAuth.LoginHandler)

		adminAuth.Use(deps.AuthMiddleware, middleware.AdminAuthMiddleware())
		adminAuth.POST("/password", deps.AdminAuth.ChangePasswordHandler)
	}

	admin.Use(deps.AuthMiddleware, middleware.AdminAuthMiddleware())
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
