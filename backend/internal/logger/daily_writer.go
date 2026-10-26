package logger

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// dailyWriter 每次写入后同步并关闭文件，无需管理跨日句柄和退出清理。
// 适用于当前低频日志；锁保证同一进程内多条日志不会交错写入。
type dailyWriter struct {
	mu  sync.Mutex
	dir string
	now func() time.Time
}

func newDailyWriter(dir string) (*dailyWriter, error) {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, err
	}
	return &dailyWriter{dir: dir, now: time.Now}, nil
}

func (w *dailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	path := filepath.Join(w.dir, w.now().Format(time.DateOnly)+".jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return 0, err
	}
	n, err := file.Write(p)
	return n, errors.Join(err, file.Sync(), file.Close())
}
