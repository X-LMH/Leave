package mysql

import "errors"

var (
	ErrorUserExist        = errors.New("用户已存在")
	ErrorRegister         = errors.New("用户注册失败")
	ErrorInvalidPassword  = errors.New("学号或密码错误")
	ErrorUserNotExist     = errors.New("用户不存在")
	ErrorFinishProfile    = errors.New("完善用户信息失败")
	ErrorNotRightPassword = errors.New("密码错误")
	ErrorSubmitRecord     = errors.New("提交失败")
	ErrorRecordNotExist   = errors.New("请假记录为空")
)
