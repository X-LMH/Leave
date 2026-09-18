package main

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/router"
	"fmt"
	"log"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatal("初始化配置失败, err:", err)
	}

	if err := mysql.Init(); err != nil {
		log.Fatal("数据库初始化配置失败, err:", err)
	}

	r := router.SetupRouter()
	if err := r.Run(":10000"); err != nil {
		fmt.Printf("Run server failed, err:%v\n", err)
	}
}
