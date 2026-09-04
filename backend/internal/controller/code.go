package controller

type ResCode int

func (c ResCode) Message() string {
	message, ok := codeMessages[c]
	if !ok {
		message = codeMessages[CodeServerBusy]
	}
	return message
}

const (
	CodeSuccess ResCode = 0

	// 用户与认证类错误：1000-1999
	CodeUserExist        ResCode = 1001
	CodeUserNotExist     ResCode = 1002
	CodeInvalidPassword  ResCode = 1003
	CodeNeedLogin        ResCode = 1004
	CodeInvalidToken     ResCode = 1005
	CodeFinishData       ResCode = 1006
	CodeNotRightPassword ResCode = 1007

	// 通用请求类错误：2000-2999
	CodeInvalidParam ResCode = 2001

	// 请假记录类错误：3000-3999
	CodeRecordNotExist ResCode = 3001

	// 文件类错误：4000-4999
	CodeFileNotFound ResCode = 4001

	// 系统维护类错误：5000-5999
	CodeServiceFix ResCode = 5001

	// 服务异常类错误：9000-9999
	CodeServerBusy ResCode = 9001
)

var codeMessages = map[ResCode]string{
	CodeSuccess:          "success",
	CodeInvalidParam:     "请求参数错误",
	CodeServerBusy:       "服务器繁忙",
	CodeUserExist:        "用户已存在",
	CodeUserNotExist:     "用户名不存在，请先注册",
	CodeInvalidPassword:  "用户名或密码错误",
	CodeNeedLogin:        "请先登录",
	CodeInvalidToken:     "无效的Token",
	CodeFinishData:       "完善信息失败",
	CodeNotRightPassword: "密码错误",
	CodeRecordNotExist:   "记录不存在",
	CodeServiceFix:       "服务正在维护中...",
	CodeFileNotFound:     "文件未找到",
}
