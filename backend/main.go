package main

import (
	"backend/config"
	"backend/router"
	"fmt"
	"log"

	"backend/dao/mysql"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatal("初始化配置失败, err:", err)
		return
	}
	// 初始化数据库
	if err := mysql.Init(); err != nil {
		log.Fatal("数据库初始化配置失败, err:", err)
		return
	}

	// 启动路由
	r := router.SetupRouter()
	// 启动
	if err := r.Run(); err != nil {
		fmt.Printf("Run mysql failed, err:%v\n", err)
		return
	}

}
