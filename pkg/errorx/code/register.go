package code

import (
	"github.com/ZampoRen/go-server-comon/pkg/errorx/internal"
)

// RegisterOptionFn 注册选项函数类型别名
type RegisterOptionFn = internal.RegisterOption

// WithAffectStability 设置稳定性标志，true: 会影响系统稳定性并在接口错误率中体现，false: 不会影响稳定性
func WithAffectStability(affectStability bool) RegisterOptionFn {
	return internal.WithAffectStability(affectStability)
}

// Register 注册用户预定义的错误码信息，在初始化时调用对应 PSM 服务的 code_gen 子模块
func Register(code int32, msg string, opts ...RegisterOptionFn) {
	internal.Register(code, msg, opts...)
}

// SetDefaultErrorCode 设置默认错误码，用于替换带有 PSM 信息染色的默认代码
func SetDefaultErrorCode(code int32) {
	internal.SetDefaultErrorCode(code)
}
