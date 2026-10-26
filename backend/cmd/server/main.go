package main

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/logger"
	"backend/internal/redis"
	"backend/internal/router"
	"errors"
	"fmt"
	"log"
	"runtime"
	"syscall"

	"go.uber.org/zap"
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
	appLogger, err := logger.NewLogger(cfg.Log)
	if err != nil {
		return fmt.Errorf("日志初始化失败: %w", err)
	}
	defer func() {
		if syncErr := syncLogger(appLogger); syncErr != nil {
			err = errors.Join(err, fmt.Errorf("同步日志失败: %w", syncErr))
		}
	}()
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
	dao := mysql.NewDAO(db, appLogger)
	r := router.SetupRouter(buildDependencies(cfg, dao, appLogger))
	return r.Run(fmt.Sprintf(":%d", cfg.App.Port))
}

func syncLogger(logger *zap.Logger) error {
	err := logger.Sync()
	// stderr 可能是终端或管道，不支持文件刷盘；其他错误仍由入口报告。
	if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTSUP) {
		return nil
	}
	// Windows 的控制台句柄调用 FlushFileBuffers 会返回 ERROR_INVALID_HANDLE。
	if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(6)) {
		return nil
	}
	return err
}
