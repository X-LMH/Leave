package main

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/logger"
	"backend/internal/redis"
	"backend/internal/router"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	checkRedis := flag.Bool("check-redis", false, "检查配置及 Redis 连接后退出，不启动 HTTP 服务")
	flag.Parse()
	if *checkRedis {
		if err := checkRedisConnection(); err != nil {
			log.Print(err)
			os.Exit(1)
		}
		log.Print("Redis 连接检查通过")
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func checkRedisConnection() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}
	client, err := redis.Open(cfg.Redis)
	if err != nil {
		return fmt.Errorf("Redis 连接检查失败: %w", err)
	}
	return client.Close()
}

func run() (err error) {
	cfg, err := config.Load()
	if err != nil {
		// 配置和日志尚未就绪时，只能使用标准错误输出。
		log.Print("初始化配置失败: ", err)
		return fmt.Errorf("初始化配置失败: %w", err)
	}
	appLogger, err := logger.NewLogger(cfg.Log)
	if err != nil {
		log.Print("日志初始化失败: ", err)
		return fmt.Errorf("日志初始化失败: %w", err)
	}
	defer func() {
		if err != nil {
			appLogger.Error("服务退出", logger.Error(err))
		} else {
			appLogger.Info("服务退出")
		}
		if syncErr := syncLogger(appLogger); syncErr != nil {
			log.Print("同步日志失败: ", syncErr)
			err = errors.Join(err, fmt.Errorf("同步日志失败: %w", syncErr))
		}
	}()
	appLogger.Info("服务初始化开始")
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
	appLogger.Info("MySQL 初始化成功")
	redisClient, err := redis.Open(cfg.Redis)
	if err != nil {
		return fmt.Errorf("Redis 初始化失败: %w", err)
	}
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("关闭 Redis 连接失败: %w", closeErr))
		}
	}()
	appLogger.Info("Redis 初始化成功")
	dao := mysql.NewDAO(db, appLogger)
	r := router.SetupRouter(buildDependencies(cfg, dao, appLogger, redisClient))
	// Run 成功后会阻塞；此日志只表示即将尝试监听。
	appLogger.Info("准备监听 HTTP 请求", zap.Int("port", cfg.App.Port))
	if err := r.Run(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
		return fmt.Errorf("HTTP 服务运行失败: %w", err)
	}
	return nil
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
