package mysql

import (
	"backend/config"
	"backend/models"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() (err error) {
	// 连接数据
	//dsn := "root:xmS10711!@tcp(localhost:3306)/leave_list?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Cfg.Mysql.User,
			config.Cfg.Mysql.Password,
			config.Cfg.Mysql.Host,
			config.Cfg.Mysql.Port,
			config.Cfg.Mysql.Database,
		),
	), &gorm.Config{})
	fmt.Printf("user is: %s\n", config.Cfg.Mysql.User)
	if err != nil {
		return fmt.Errorf("mysql open failed, err:%v", err)
	}
	// 建表
	if err = db.AutoMigrate(&models.User{}, &models.Profile{}, &models.Record{}); err != nil {
		return fmt.Errorf("mysql migrate failed, err:%v", err)
	}
	return nil
}
