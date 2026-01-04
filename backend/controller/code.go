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
	CodeSuccess ResCode = iota
	CodeInvalidParam
	CodeServerBusy
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeNeedLogin
	CodeInvalidToken
	CodeFinishData
	CodeNotRightPassword
	CodeRecordNotExist
	CodeServiceFix
	CodeFileNotFound
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
