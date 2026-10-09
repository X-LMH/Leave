package router

import (
	"backend/internal/controller"
	admincontroller "backend/internal/controller/admin"
	appcontroller "backend/internal/controller/app"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())

	// API v1
	api := r.Group("/api/v1")

	// 健康检查接口
	api.GET("/health", controller.HealthHandler)

	// 版本接口必须不受版本校验保护，旧客户端才能获取更新信息。
	api.GET("/app-version", appcontroller.GetCurrentAppVersionHandler)
	api.GET("/app-download", appcontroller.DownloadCurrentAppHandler)
	api.HEAD("/app-download", appcontroller.DownloadCurrentAppHandler)

	// App 接口需要校验客户端版本；管理平台不受 App 版本限制。
	app := api.Group("")
	app.Use(middleware.AppVersionMiddleware())

	// App 公共接口
	{
		auth := app.Group("/auth")
		auth.POST("/register", appcontroller.RegisterHandler)
		auth.POST("/login", appcontroller.LoginHandler)
	}

	user := app.Group("")
	user.Use(middleware.JWTAuthMiddleware())
	{
		user.POST("/logout", appcontroller.LogoutHandler)
		user.POST("/client-info", appcontroller.UpdateAppInfoHandler)
		user.POST("/profile", appcontroller.ProfileHandler)
		user.POST("/profile/avatar", appcontroller.UploadAvatarHandler)
		user.POST("/password", appcontroller.ChangePasswordHandler)
		user.GET("/profile", appcontroller.GetProfileHandler)

		user.GET("/class-options", appcontroller.GetClassOptionsHandler)
		user.GET("/apartments", appcontroller.GetApartmentOptionsHandler)
		user.GET("/leave-types", appcontroller.GetLeaveTypeOptionsHandler)

		user.POST("/record", appcontroller.CreateRecordHandler)
		user.GET("/record/:id", appcontroller.GetRecordHandler)
		user.GET("/records", appcontroller.GetRecordsLIstHandler)
		user.DELETE("/record/:id", appcontroller.DeleteRecordHandler)

		user.POST("/feedback", appcontroller.CreateFeedbackHandler)
	}

	admin := api.Group("/admin")
	adminAuth := admin.Group("/auth")
	adminAuth.POST("/login", admincontroller.LoginHandler)

	admin.Use(middleware.JWTAuthMiddleware(), middleware.AdminAuthMiddleware())
	{
		admin.GET("/class-options", admincontroller.GetClassOptionsHandler)
		admin.GET("/apartment-options", admincontroller.GetApartmentOptionsHandler)
		admin.GET("/leave-type-options", admincontroller.GetLeaveTypeOptionsHandler)

		admin.GET("/app-versions", admincontroller.GetAppVersionsHandler)
		admin.GET("/app-versions/current", admincontroller.GetCurrentAppVersionHandler)
		admin.GET("/app-versions/:id", admincontroller.GetAppVersionHandler)

		admin.GET("/students", admincontroller.GetStudentsHandler)
		admin.GET("/students/:id", admincontroller.GetStudentHandler)
		admin.PUT("/students/:id", admincontroller.UpdateStudentHandler)
		admin.PUT("/students/:id/status", admincontroller.UpdateStudentStatusHandler)

		admin.GET("/classes", admincontroller.GetClassesHandler)
		admin.POST("/classes", admincontroller.CreateClassHandler)
		admin.PUT("/classes/:id", admincontroller.UpdateClassHandler)
		admin.DELETE("/classes/:id", admincontroller.DeleteClassHandler)

		admin.GET("/leave-types", admincontroller.GetLeaveTypesHandler)
		admin.POST("/leave-types", admincontroller.CreateLeaveTypeHandler)
		admin.PUT("/leave-types/:id", admincontroller.UpdateLeaveTypeHandler)
		admin.DELETE("/leave-types/:id", admincontroller.DeleteLeaveTypeHandler)
	}

	return r
}
