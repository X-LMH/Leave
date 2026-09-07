package request

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	CtxStuID = "student_id"
	CtxRole  = "role"
)

var (
	ErrorUserNotLogin         = errors.New("用户未登录")
	ErrorUserPermissionDenied = errors.New("用户权限不足")
)

// GetCurrentStuID 获取当前登录学生ID。
func GetCurrentStuID(c *gin.Context) (studentID string, err error) {
	id, ok := c.Get(CtxStuID)
	if !ok {
		return "", ErrorUserNotLogin
	}
	studentID, ok = id.(string)
	if !ok {
		return "", ErrorUserNotLogin
	}
	return studentID, nil
}

// GetCurrentRole 获取当前登录用户角色。
func GetCurrentRole(c *gin.Context) (role string, err error) {
	id, ok := c.Get(CtxRole)
	if !ok {
		return "", ErrorUserPermissionDenied
	}
	role, ok = id.(string)
	if !ok {
		return "", ErrorUserPermissionDenied
	}
	return role, nil
}
