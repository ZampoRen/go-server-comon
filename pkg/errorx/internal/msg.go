package internal

import (
	"fmt"
)

// withMessage 带消息的错误包装
type withMessage struct {
	cause error
	msg   string
}

func (w *withMessage) Unwrap() error {
	return w.cause
}

func (w *withMessage) Error() string {
	return fmt.Sprintf("%s\ncause=%s", w.msg, w.cause.Error())
}

// wrapf 包装错误并添加格式化消息
func wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	err = &withMessage{
		cause: err,
		msg:   fmt.Sprintf(format, args...),
	}

	return err
}

// Wrapf 包装错误并添加格式化消息，如果错误没有堆栈跟踪则添加
func Wrapf(err error, format string, args ...interface{}) error {
	return withStackTraceIfNotExists(wrapf(err, format, args...))
}
