package router

import (
	"backend/controller"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// 公共接口
	r.POST("/register", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)
	r.GET("/version", controller.VersionHandler)
	r.GET("/update", controller.UpdateHandler) // 下载接口

	// 需要认证的接口组
	authGroup := r.Group("/")
	authGroup.Use(middleware.JWTAuthMiddleware())
	{
		authGroup.POST("/profile", controller.ProfileHandler)
		authGroup.POST("/password", controller.ChangePasswordHandler)
		authGroup.GET("/profile", controller.GetProfileHandler)
		authGroup.POST("/record", controller.RecordHandler)
		authGroup.GET("/record/:id", controller.GetRecordHandler)
		authGroup.GET("/records", controller.GetRecordsLIstHandler)
		authGroup.DELETE("/record/:id", controller.DeleteRecordHandler)
	}

	return r
}
