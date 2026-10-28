package middleware

const (
	SideApp   = "app"
	SideAdmin = "admin"
)

const (
	OperationGeneral        = "general"
	OperationLogin          = "login"
	OperationRegister       = "register"
	OperationPasswordChange = "password-change"
	OperationRecordCreate   = "record-create"
	OperationFeedbackCreate = "feedback-create"
	OperationAvatarUpload   = "avatar-upload"
)

const (
	DimensionIP      = "ip"
	DimensionAccount = "account"
)

// rateLimitRule 定义首次请求开始计时的固定窗口。
type rateLimitRule struct {
	Limit         int64
	WindowSeconds int64
}

// defaultRateLimits 按操作和维度维护规则，不从配置文件读取。
// App 与管理端共用阈值、分别计数；专项限制叠加普通额度。
func defaultRateLimits() map[string]map[string]rateLimitRule {
	return map[string]map[string]rateLimitRule{
		OperationGeneral: {
			// 同一出口或反向代理可能承载多个用户，IP 额度保留余量。
			DimensionIP:      {Limit: 600, WindowSeconds: 60},
			DimensionAccount: {Limit: 60, WindowSeconds: 60},
		},
		// 登录成功、失败均消耗额度。
		OperationLogin: {
			DimensionIP:      {Limit: 60, WindowSeconds: 60},
			DimensionAccount: {Limit: 10, WindowSeconds: 60},
		},
		OperationRegister: {
			DimensionIP:      {Limit: 20, WindowSeconds: 60},
			DimensionAccount: {Limit: 3, WindowSeconds: 600},
		},
		OperationPasswordChange: {
			DimensionAccount: {Limit: 5, WindowSeconds: 600},
		},
		OperationRecordCreate: {
			DimensionAccount: {Limit: 5, WindowSeconds: 60},
		},
		OperationFeedbackCreate: {
			DimensionAccount: {Limit: 3, WindowSeconds: 600},
		},
		OperationAvatarUpload: {
			DimensionAccount: {Limit: 3, WindowSeconds: 60},
		},
	}
}
