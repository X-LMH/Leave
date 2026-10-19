package mysql

import (
	"backend/internal/config"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func Open(cfg config.MySQLConfig) (*gorm.DB, error) {
	// 参数插值避免每条参数化 SQL 都先预处理，减少远程数据库的网络往返。
	db, err := gorm.Open(mysql.Open(
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		),
	), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("mysql open failed: %w", err)
	}
	return db, nil
}
