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

	// 公共接口
	{
		api.POST("/register", controller.RegisterHandler)
		api.POST("/login", controller.LoginHandler)
		//api.GET("/version", controller.VersionHandler)
		//api.GET("/update", controller.UpdateHandler)
	}

	//auth := api.Group("")
	//auth.Use(middleware.JWTAuthMiddleware())
	//{
	//	auth.POST("/profile", controller.ProfileHandler)
	//	auth.POST("/password", controller.ChangePasswordHandler)
	//	auth.GET("/profile", controller.GetProfileHandler)
	//
	//	auth.POST("/record", controller.RecordHandler)
	//	auth.GET("/record/:id", controller.GetRecordHandler)
	//	auth.GET("/records", controller.GetRecordsLIstHandler)
	//	auth.DELETE("/record/:id", controller.DeleteRecordHandler)
	//}

	return r
}
