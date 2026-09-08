package router

import (
	"backend/internal/controller"
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
	api.GET("/app-version", controller.GetCurrentAppVersionHandler)
	api.GET("/app-download", controller.DownloadCurrentAppHandler)
	api.HEAD("/app-download", controller.DownloadCurrentAppHandler)
	api.Use(middleware.AppVersionMiddleware())

	// 公共接口
	{
		api.POST("/register", controller.RegisterHandler)
		api.POST("/login", controller.LoginHandler)
	}

	auth := api.Group("")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		auth.POST("/logout", controller.LogoutHandler)
		auth.POST("/profile", controller.ProfileHandler)
		auth.POST("/password", controller.ChangePasswordHandler)
		auth.GET("/profile", controller.GetProfileHandler)

		auth.GET("/classes", controller.GetClassOptionsHandler)
		auth.GET("/apartments", controller.GetApartmentOptionsHandler)
		auth.GET("/leave_types", controller.GetLeaveTypeOptionsHandler)

		auth.POST("/record", controller.CreateRecordHandler)
		auth.GET("/record/:id", controller.GetRecordHandler)
		auth.GET("/records", controller.GetRecordsLIstHandler)
		auth.DELETE("/record/:id", controller.DeleteRecordHandler)
	}

	return r
}
