package controller

import (
	"backend/config"
	"backend/dao/mysql"
	"backend/models"
	"backend/service"
	"errors"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// RegisterHandler 用户注册
func RegisterHandler(c *gin.Context) {
	// 参数绑定
	//ResponseError(c, CodeServiceFix)
	//return

	p := new(models.ParamRegister)
	if err := c.ShouldBindJSON(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if len(p.StudentID) == 0 || len(p.Password) == 0 || len(p.RePassword) == 0 {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 业务逻辑
	if err := service.Register(p); err != nil {
		if errors.Is(err, mysql.ErrorUserExist) {
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回响应
	ResponseSuccess(c, nil)
}

// LoginHandler 用户登录
func LoginHandler(c *gin.Context) {
	// 参数绑定
	p := new(models.ParamLogin)
	if err := c.ShouldBindJSON(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if len(p.StudentID) == 0 || len(p.Password) == 0 {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 业务逻辑
	token, err := service.Login(p)
	if err != nil {
		// 用户名是否存在
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		}
		// 密码是否正确
		if errors.Is(err, mysql.ErrorInvalidPassword) {
			ResponseError(c, CodeInvalidPassword)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, token)
}

// ChangePasswordHandler 修改密码
func ChangePasswordHandler(c *gin.Context) {
	// 参数校验
	p := new(models.ParamPassword)
	if err := c.ShouldBindJSON(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if len(p.Password) == 0 || len(p.RePassword) == 0 || p.NewPassword != p.RePassword {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 业务处理
	studentID, err := GetCurrentStuID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	if err := service.ChangePassword(p, studentID); err != nil {
		if errors.Is(err, mysql.ErrorNotRightPassword) {
			ResponseError(c, CodeNotRightPassword)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func VersionHandler(c *gin.Context) {
	version := config.GetVersionInfo()
	ResponseSuccess(c, version)
}

// UpdateHandler 处理APK文件下载请求（从程序同级目录读取）
func UpdateHandler(c *gin.Context) {
	// 获取当前工作目录（程序运行的同级目录）
	workDir, err := os.Getwd()
	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	// 拼接APK文件路径（同级目录下的Leave.apk）
	apkPath := filepath.Join(workDir, "Leave.apk")

	// 检查文件是否存在
	if _, err := os.Stat(apkPath); os.IsNotExist(err) {
		ResponseError(c, CodeFileNotFound)
		return
	}

	// 设置响应头，触发文件下载
	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Content-Disposition", "attachment; filename=Leave.apk")
	c.File(apkPath)
}
