package mysql

import (
	"backend/internal/config"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type DAO struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewDAO(db *gorm.DB, logger *zap.Logger) *DAO {
	return &DAO{
		logger: logger,
		db:     db,
	}
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
	), &gorm.Config{
		TranslateError: true,
		// 业务边界已记录错误，关闭包含完整 SQL 和参数的默认输出。
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("mysql open failed: %w", err)
	}
	return db, nil
}
