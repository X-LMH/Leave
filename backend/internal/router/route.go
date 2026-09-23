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
	api.Use(middleware.AppVersionMiddleware())

	// 公共接口
	{
		api.POST("/register", appcontroller.RegisterHandler)
		api.POST("/login", appcontroller.LoginHandler)
	}

	auth := api.Group("")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		auth.POST("/logout", appcontroller.LogoutHandler)
		auth.POST("/client-info", appcontroller.UpdateAppInfoHandler)
		auth.POST("/profile", appcontroller.ProfileHandler)
		auth.POST("/password", appcontroller.ChangePasswordHandler)
		auth.GET("/profile", appcontroller.GetProfileHandler)

		auth.GET("/class-options", appcontroller.GetClassOptionsHandler)
		auth.GET("/apartments", appcontroller.GetApartmentOptionsHandler)
		auth.GET("/leave-types", appcontroller.GetLeaveTypeOptionsHandler)

		auth.POST("/record", appcontroller.CreateRecordHandler)
		auth.GET("/record/:id", appcontroller.GetRecordHandler)
		auth.GET("/records", appcontroller.GetRecordsLIstHandler)
		auth.DELETE("/record/:id", appcontroller.DeleteRecordHandler)

		auth.POST("/feedback", appcontroller.CreateFeedbackHandler)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware(), middleware.AdminAuthMiddleware())
	{
		admin.GET("/classes", admincontroller.GetClassesHandler)
	}

	return r
}
