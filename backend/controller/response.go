package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    ResCode     `json:"code"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ResponseError(c *gin.Context, code ResCode) {
	res := &Response{
		Code:    code,
		Message: code.Message(),
		Data:    nil,
	}
	c.JSON(http.StatusOK, res)
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	res := &Response{
		Code:    CodeSuccess,
		Message: CodeSuccess.Message(),
		Data:    data,
	}
	c.JSON(http.StatusOK, res)
}
