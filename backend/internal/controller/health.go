package controller

import (
	"backend/internal/response"

	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}
