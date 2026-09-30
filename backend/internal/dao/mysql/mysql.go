package mysql

import (
	"backend/internal/config"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() (err error) {
	// 参数插值避免每条参数化 SQL 都先预处理，减少远程数据库的网络往返。
	db, err = gorm.Open(mysql.Open(
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true",
			config.Cfg.Mysql.User,
			config.Cfg.Mysql.Password,
			config.Cfg.Mysql.Host,
			config.Cfg.Mysql.Port,
			config.Cfg.Mysql.Database,
		),
	), &gorm.Config{TranslateError: true})
	fmt.Printf("user is: %s\n", config.Cfg.Mysql.User)
	if err != nil {
		return fmt.Errorf("mysql open failed, err:%v", err)
	}
	return nil
}
