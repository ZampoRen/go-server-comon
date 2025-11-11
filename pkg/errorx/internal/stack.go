package internal

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// StackTracer 堆栈跟踪接口
type StackTracer interface {
	StackTrace() string
}

// withStack 带堆栈的错误包装
type withStack struct {
	cause error
	stack string
}

func (w *withStack) Unwrap() error {
	return w.cause
}

func (w *withStack) StackTrace() string {
	return w.stack
}

func (w *withStack) Error() string {
	return fmt.Sprintf("%s\nstack=%s", w.cause.Error(), w.stack)
}

// stack 生成堆栈跟踪信息
func stack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:])

	b := strings.Builder{}
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i])

		file, line := fn.FileLine(pcs[i])
		name := trimPathPrefix(fn.Name())
		b.WriteString(fmt.Sprintf("%s:%d %s\n", file, line, name))
	}

	return b.String()
}

// trimPathPrefix 修剪路径前缀
func trimPathPrefix(s string) string {
	i := strings.LastIndex(s, "/")
	s = s[i+1:]
	i = strings.Index(s, ".")
	return s[i+1:]
}

// withStackTraceIfNotExists 如果错误没有堆栈跟踪则添加
func withStackTraceIfNotExists(err error) error {
	if err == nil {
		return nil
	}

	// 如果堆栈已存在则跳过
	var stackTracer StackTracer
	if errors.As(err, &stackTracer) {
		return err
	}

	return &withStack{
		err,
		stack(),
	}
}
