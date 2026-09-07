package mysql

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrorUserExist         = errors.New("用户已存在")
	ErrorInvalidPassword   = errors.New("学号或密码错误")
	ErrorUserDisabled      = errors.New("用户已停用")
	ErrorUserNotExist      = errors.New("用户不存在")
	ErrorNotRightPassword  = errors.New("密码错误")
	ErrorSubmitRecord      = errors.New("提交失败")
	ErrorRecordNotExist    = errors.New("请假记录为空")
	ErrorLeaveTypeNotExist = errors.New("请假类型不存在")
)

// IsDuplicateKeyError 判断写入操作是否违反了唯一约束。
// 开启 GORM TranslateError 后，该判断不依赖具体数据库的错误码。
func IsDuplicateKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
