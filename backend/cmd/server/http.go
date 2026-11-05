package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// serveHTTP 在请求排空后才返回，调用方随后关闭 Redis、MySQL 并同步日志。
func serveHTTP(ctx context.Context, server *http.Server, listener net.Listener, timeout time.Duration, log *zap.Logger) error {
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(listener)
	}()

	var serveErr error
	select {
	case serveErr = <-serveDone:
	case <-ctx.Done():
		log.Info("收到退出信号，停止接收新请求", zap.Duration("shutdown_timeout", timeout))
	}

	// 使用独立 Context，让已收到退出信号的服务仍有时间完成在途请求。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownCtx)
	var closeErr error
	if shutdownErr != nil {
		log.Error("HTTP 优雅关闭失败，关闭剩余连接", zap.Error(shutdownErr))
		closeErr = server.Close()
	}
	if serveErr == nil {
		serveErr = <-serveDone
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	if err := errors.Join(serveErr, shutdownErr, closeErr); err != nil {
		return fmt.Errorf("HTTP 服务退出失败: %w", err)
	}
	log.Info("HTTP 在途请求已处理完毕")
	return nil
}
