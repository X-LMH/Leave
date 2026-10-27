package logger

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// Error 隐藏 MySQL 错误中的字段值，保留可用于排障的错误码。
func Error(err error) zap.Field {
	var databaseError *mysql.MySQLError
	if errors.As(err, &databaseError) {
		return zap.String("error", fmt.Sprintf("MySQL error %d (SQLSTATE %s)", databaseError.Number, databaseError.SQLState))
	}
	return zap.Error(err)
}
