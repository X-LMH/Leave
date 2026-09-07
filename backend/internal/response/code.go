package response

import "net/http"

// Code 是客户端可用于细分错误原因的稳定业务码。
type Code int

const (
	CodeSuccess Code = 0

	// 用户与认证类错误：1000-1999
	CodeUserExist        Code = 1001
	CodeUserNotExist     Code = 1002
	CodeInvalidPassword  Code = 1003
	CodeNeedLogin        Code = 1004
	CodeInvalidToken     Code = 1005
	CodeNotRightPassword Code = 1007

	// 通用请求类错误：2000-2999
	CodeInvalidParam     Code = 2001
	CodeRouteNotFound    Code = 2002
	CodeMethodNotAllowed Code = 2003

	// 请假记录类错误：3000-3999
	CodeRecordNotExist    Code = 3001
	CodeProfileIncomplete Code = 3002

	// 文件类错误：4000-4999
	CodeFileNotFound Code = 4001

	// 系统维护类错误：5000-5999
	CodeServiceFix Code = 5001

	// 服务异常类错误：9000-9999
	CodeServerBusy Code = 9001
)

type codeMeta struct {
	message    string
	httpStatus int
}

var metaByCode = map[Code]codeMeta{
	CodeSuccess:           {message: "success", httpStatus: http.StatusOK},
	CodeInvalidParam:      {message: "请求参数错误", httpStatus: http.StatusBadRequest},
	CodeRouteNotFound:     {message: "路由不存在", httpStatus: http.StatusNotFound},
	CodeMethodNotAllowed:  {message: "请求方法不被允许", httpStatus: http.StatusMethodNotAllowed},
	CodeServerBusy:        {message: "服务器繁忙", httpStatus: http.StatusInternalServerError},
	CodeUserExist:         {message: "用户已存在", httpStatus: http.StatusConflict},
	CodeUserNotExist:      {message: "用户名不存在，请先注册", httpStatus: http.StatusNotFound},
	CodeInvalidPassword:   {message: "用户名或密码错误", httpStatus: http.StatusUnauthorized},
	CodeNeedLogin:         {message: "用户未登录", httpStatus: http.StatusUnauthorized},
	CodeInvalidToken:      {message: "无效的Token", httpStatus: http.StatusUnauthorized},
	CodeNotRightPassword:  {message: "密码错误", httpStatus: http.StatusBadRequest},
	CodeRecordNotExist:    {message: "记录不存在", httpStatus: http.StatusNotFound},
	CodeProfileIncomplete: {message: "请先完善个人信息", httpStatus: http.StatusBadRequest},
	CodeServiceFix:        {message: "服务正在维护中...", httpStatus: http.StatusServiceUnavailable},
	CodeFileNotFound:      {message: "文件未找到", httpStatus: http.StatusNotFound},
}

// Message 返回业务码对应的客户端消息；未知业务码安全回退为服务器繁忙。
func (c Code) Message() string {
	return metaFor(c).message
}

// HTTPStatus 返回业务码对应的 HTTP 状态；未知业务码安全回退为 500。
func (c Code) HTTPStatus() int {
	return metaFor(c).httpStatus
}

func metaFor(code Code) codeMeta {
	meta, ok := metaByCode[code]
	if !ok {
		return metaByCode[CodeServerBusy]
	}
	return meta
}
