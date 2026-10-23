package main

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/redis"
	"backend/internal/router"
	"errors"
	"fmt"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}
	db, err := mysql.Open(cfg.Mysql)
	if err != nil {
		return fmt.Errorf("数据库初始化失败: %w", err)
	}
	connection, err := db.DB()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := connection.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("关闭数据库连接失败: %w", closeErr))
		}
	}()
	redisClient, err := redis.Open(cfg.Redis)
	if err != nil {
		return fmt.Errorf("Redis 初始化失败: %w", err)
	}
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("关闭 Redis 连接失败: %w", closeErr))
		}
	}()
	dao := mysql.NewDAO(db)
	r := router.SetupRouter(buildDependencies(cfg, dao))
	return r.Run(fmt.Sprintf(":%d", cfg.App.Port))
}
