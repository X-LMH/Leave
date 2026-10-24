package main

import (
	"errors"
	"runtime"
	"syscall"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type syncErrorWriter struct {
	err error
}

func (s syncErrorWriter) Write(p []byte) (int, error) { return len(p), nil }
func (s syncErrorWriter) Sync() error                 { return s.err }

func TestSyncLogger(t *testing.T) {
	failure := errors.New("flush failed")
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "success"},
		{name: "unsupported console", err: syscall.EINVAL},
		{name: "unsupported operation", err: syscall.ENOTSUP},
		{name: "real failure", err: failure, want: failure},
	}
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name string
			err  error
			want error
		}{name: "windows console", err: syscall.Errno(6)})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), syncErrorWriter{err: tt.err}, zap.InfoLevel)
			if err := syncLogger(zap.New(core)); !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}
