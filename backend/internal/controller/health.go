package controller

import (
	"backend/internal/response"

	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	// TODO: version is test 字段
	response.Success(c, gin.H{
		"status":  "ok",
		"version": "6",
	})
}
