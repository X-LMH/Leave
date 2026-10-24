package logger

import (
	"backend/internal/config"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 创建应用实例使用的控制台日志，由入口负责注入和同步。
func New(cfg config.LogConfig) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("无效的 log.level: %w", err)
	}
	if cfg.Encoding != "console" && cfg.Encoding != "json" {
		return nil, fmt.Errorf("log.encoding 必须是 console 或 json，当前为 %q", cfg.Encoding)
	}

	settings := zap.NewProductionConfig()
	settings.Level = zap.NewAtomicLevelAt(level)
	settings.Encoding = cfg.Encoding
	settings.OutputPaths = []string{"stderr"}
	settings.ErrorOutputPaths = []string{"stderr"}
	settings.EncoderConfig.TimeKey = "time"
	settings.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 暂不采样，避免以后排查业务问题时丢失重复事件。
	settings.Sampling = nil
	return settings.Build()
}
