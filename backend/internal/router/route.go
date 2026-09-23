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
		auth := api.Group("/auth")
		auth.POST("/register", appcontroller.RegisterHandler)
		auth.POST("/login", appcontroller.LoginHandler)
	}

	user := api.Group("")
	user.Use(middleware.JWTAuthMiddleware())
	{
		user.POST("/logout", appcontroller.LogoutHandler)
		user.POST("/client-info", appcontroller.UpdateAppInfoHandler)
		user.POST("/profile", appcontroller.ProfileHandler)
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
	admin.Use(middleware.JWTAuthMiddleware(), middleware.AdminAuthMiddleware())
	{
		admin.GET("/classes", admincontroller.GetClassesHandler)
	}

	return r
}
