package mysql

import "gorm.io/gorm"

// Paginate 接收已在请求入口校验的分页参数，仅设置查询分页条件。
func Paginate(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((page - 1) * pageSize).
			Limit(pageSize)
	}
}
