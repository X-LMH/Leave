package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const CtxStuID = "student_id"
const CtxRole = "role"

var (
	ErrorUserNotLogin         = errors.New("用户未登录")
	ErrorUserPermissionDenied = errors.New("用户权限不足")
)

// GetCurrentStuID 获取当前登录学生ID
func GetCurrentStuID(c *gin.Context) (studentID string, err error) {
	id, ok := c.Get(CtxStuID)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	studentID, ok = id.(string)

	if !ok {
		err = ErrorUserNotLogin
		return
	}
	return
}

func GetCurrentRole(c *gin.Context) (role string, err error) {
	id, ok := c.Get(CtxRole)
	if !ok {
		err = ErrorUserPermissionDenied
		return
	}
	role, ok = id.(string)
	if !ok {
		err = ErrorUserPermissionDenied
		return
	}

	return
}
