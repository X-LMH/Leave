package logger

import (
	"backend/internal/config"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	encodingConsole = "console"
	encodingJSON    = "json"
)

// NewLogger 创建终端日志及可选的按日 JSON 文件日志。
func NewLogger(cfg config.LogConfig) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("无效的 log.level: %w", err)
	}
	if cfg.Encoding != encodingConsole && cfg.Encoding != encodingJSON {
		return nil, fmt.Errorf("log.encoding 必须是 %s 或 %s，当前为 %q", encodingConsole, encodingJSON, cfg.Encoding)
	}

	settings := zap.NewProductionConfig()
	settings.Level = zap.NewAtomicLevelAt(level)
	settings.Encoding = cfg.Encoding
	settings.OutputPaths = []string{"stderr"}
	settings.ErrorOutputPaths = []string{"stderr"}
	settings.EncoderConfig.TimeKey = "time"
	settings.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	settings.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	settings.EncoderConfig.EncodeCaller = relativeCallerEncoder
	if cfg.Encoding == encodingConsole {
		// 颜色仅用于控制台，避免 JSON 字段混入 ANSI 转义字符。
		settings.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	// 暂不采样，避免以后排查业务问题时丢失重复事件。
	settings.Sampling = nil
	if strings.TrimSpace(cfg.Dir) == "" {
		return settings.Build()
	}
	writer, err := newDailyWriter(strings.TrimSpace(cfg.Dir))
	if err != nil {
		return nil, err
	}
	fileEncoder := settings.EncoderConfig
	fileEncoder.EncodeLevel = zapcore.LowercaseLevelEncoder
	return settings.Build(zap.WrapCore(func(console zapcore.Core) zapcore.Core {
		return zapcore.NewTee(
			console,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(fileEncoder),
				zapcore.AddSync(writer),
				settings.Level,
			),
		)
	}))
}

// relativeCallerEncoder 保留 Leave 根目录下的相对路径，方便定位源码。
func relativeCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	path := strings.ReplaceAll(caller.File, "\\", "/")
	if index := strings.LastIndex(path, "/backend/"); index >= 0 {
		enc.AppendString(path[index+1:] + ":" + strconv.Itoa(caller.Line))
		return
	}
	// -trimpath 构建使用模块路径 backend/...。
	if strings.HasPrefix(path, "backend/") {
		enc.AppendString(path + ":" + strconv.Itoa(caller.Line))
		return
	}
	zapcore.ShortCallerEncoder(caller, enc)
}
