package errorx

import (
	"fmt"
	"strings"

	"github.com/ZampoRen/go-server-comon/pkg/errorx/internal"
)

// StatusError 是一个带有状态码的错误接口，你可以通过 New 或 WrapByCode
// 创建一个错误，并通过 FromStatusError 将其转换回 StatusError 以获取
// 错误状态码等信息。
type StatusError interface {
	error
	Code() int32
	Msg() string
	IsAffectStability() bool
	Extra() map[string]string
}

// Option 用于配置 StatusError
type Option = internal.Option

// KV 创建一个键值对选项，用于替换错误消息中的占位符
func KV(k, v string) Option {
	return internal.Param(k, v)
}

// KVf 创建一个格式化的键值对选项，用于替换错误消息中的占位符
func KVf(k, v string, a ...any) Option {
	formatValue := fmt.Sprintf(v, a...)
	return internal.Param(k, formatValue)
}

// Extra 创建一个额外信息选项，用于添加额外的错误信息
func Extra(k, v string) Option {
	return internal.Extra(k, v)
}

// New 通过状态码获取配置文件中预定义的错误，并在调用 New 的位置生成堆栈跟踪
func New(code int32, options ...Option) error {
	return internal.NewByCode(code, options...)
}

// WrapByCode 返回一个错误，在调用 WrapByCode 的位置用堆栈跟踪和状态码注释 err
func WrapByCode(err error, statusCode int32, options ...Option) error {
	if err == nil {
		return nil
	}

	return internal.WrapByCode(err, statusCode, options...)
}

// Wrapf 返回一个错误，在调用 Wrapf 的位置用堆栈跟踪和格式说明符注释 err
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return internal.Wrapf(err, format, args...)
}

// ErrorWithoutStack 返回不带堆栈信息的错误消息
func ErrorWithoutStack(err error) string {
	if err == nil {
		return ""
	}
	errMsg := err.Error()
	index := strings.Index(errMsg, "stack=")
	if index != -1 {
		errMsg = errMsg[:index]
	}
	return errMsg
}
