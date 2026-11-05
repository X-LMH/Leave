package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dao/redis"
	"backend/internal/logger"
	"backend/internal/router"

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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: r,
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("HTTP 服务监听失败: %w", err)
	}

	appLogger.Info("开始监听 HTTP 请求", zap.Int("port", cfg.App.Port))

	// 部署脚本最多等待 20 秒，预留 5 秒用于资源清理和进程退出。
	return serveHTTP(ctx, server, listener, 15*time.Second, appLogger)
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
