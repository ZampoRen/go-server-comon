package internal

import (
	"errors"
	"fmt"
	"strings"
)

// StatusError 状态错误接口
type StatusError interface {
	error
	Code() int32
}

// statusError 状态错误实现
type statusError struct {
	statusCode int32
	message    string

	ext Extension
}

// withStatus 带状态码的错误包装
type withStatus struct {
	status *statusError

	stack string
	cause error
}

// Extension 扩展信息
type Extension struct {
	IsAffectStability bool              // 是否影响稳定性
	Extra             map[string]string // 额外信息
}

func (w *statusError) Code() int32 {
	return w.statusCode
}

func (w *statusError) IsAffectStability() bool {
	return w.ext.IsAffectStability
}

func (w *statusError) Msg() string {
	return w.message
}

func (w *statusError) Error() string {
	return fmt.Sprintf("code=%d message=%s", w.statusCode, w.message)
}

func (w *statusError) Extra() map[string]string {
	return w.ext.Extra
}

// Unwrap 支持 go errors.Unwrap()
func (w *withStatus) Unwrap() error {
	return w.cause
}

// Is 支持 go errors.Is()
func (w *withStatus) Is(target error) bool {
	var ws StatusError
	if errors.As(target, &ws) && w.status.Code() == ws.Code() {
		return true
	}
	return false
}

// As 支持 go errors.As()
func (w *withStatus) As(target interface{}) bool {
	if errors.As(w.status, target) {
		return true
	}

	return false
}

func (w *withStatus) StackTrace() string {
	return w.stack
}

func (w *withStatus) Error() string {
	b := strings.Builder{}
	b.WriteString(w.status.Error())

	if w.cause != nil {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("cause=%s", w.cause))
	}

	if w.stack != "" {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("stack=%s", w.stack))
	}

	return b.String()
}

// Option 选项函数
type Option func(ws *withStatus)

// Param 创建参数选项，用于替换错误消息中的占位符
func Param(k, v string) Option {
	return func(ws *withStatus) {
		if ws == nil || ws.status == nil {
			return
		}
		ws.status.message = strings.Replace(ws.status.message, fmt.Sprintf("{%s}", k), v, -1)
	}
}

// Extra 创建额外信息选项
func Extra(k, v string) Option {
	return func(ws *withStatus) {
		if ws == nil || ws.status == nil {
			return
		}
		if ws.status.ext.Extra == nil {
			ws.status.ext.Extra = make(map[string]string)
		}
		ws.status.ext.Extra[k] = v
	}
}

// NewByCode 通过错误码创建新错误
func NewByCode(code int32, options ...Option) error {
	ws := &withStatus{
		status: getStatusByCode(code),
		cause:  nil,
		stack:  stack(),
	}

	for _, opt := range options {
		opt(ws)
	}

	return ws
}

// WrapByCode 用状态码包装错误
func WrapByCode(err error, code int32, options ...Option) error {
	if err == nil {
		return nil
	}

	ws := &withStatus{
		status: getStatusByCode(code),
		cause:  err,
	}

	for _, opt := range options {
		opt(ws)
	}

	// 如果堆栈已存在则跳过
	var stackTracer StackTracer
	if errors.As(err, &stackTracer) {
		return ws
	}

	ws.stack = stack()

	return ws
}

// getStatusByCode 通过错误码获取状态错误
func getStatusByCode(code int32) *statusError {
	codeDefinition, ok := CodeDefinitions[code]
	if ok {
		// 预定义的错误码
		return &statusError{
			statusCode: code,
			message:    codeDefinition.Message,
			ext: Extension{
				IsAffectStability: codeDefinition.IsAffectStability,
			},
		}
	}

	return &statusError{
		statusCode: code,
		message:    DefaultErrorMsg,
		ext: Extension{
			IsAffectStability: DefaultIsAffectStability,
		},
	}
}
