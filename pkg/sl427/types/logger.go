// pkg/sl427/types/logger.go
package types

// Logger 定义最小日志接口
type Logger interface {
	Printf(format string, v ...interface{})
}

// 默认的空日志实现
type noopLogger struct{}

func (l noopLogger) Printf(format string, v ...interface{}) {}

// DefaultLogger 默认使用空日志实现
var DefaultLogger Logger = noopLogger{}

// SetLogger 允许用户设置自定义日志实现
func SetLogger(l Logger) {
	if l != nil {
		DefaultLogger = l
	}
}
