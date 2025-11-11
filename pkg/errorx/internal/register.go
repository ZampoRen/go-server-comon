package internal

const (
	// DefaultErrorMsg 默认错误消息
	DefaultErrorMsg = "Service Internal Error"
	// DefaultIsAffectStability 默认是否影响稳定性
	DefaultIsAffectStability = true
)

var (
	// ServiceInternalErrorCode 服务内部错误码
	ServiceInternalErrorCode int32 = 1
	// CodeDefinitions 错误码定义映射
	CodeDefinitions = make(map[int32]*CodeDefinition)
)

// CodeDefinition 错误码定义
type CodeDefinition struct {
	Code              int32  // 错误码
	Message           string // 错误消息
	IsAffectStability bool   // 是否影响稳定性
}

// RegisterOption 注册选项函数
type RegisterOption func(definition *CodeDefinition)

// WithAffectStability 设置是否影响稳定性
func WithAffectStability(affectStability bool) RegisterOption {
	return func(definition *CodeDefinition) {
		definition.IsAffectStability = affectStability
	}
}

// Register 注册错误码定义
func Register(code int32, msg string, opts ...RegisterOption) {
	definition := &CodeDefinition{
		Code:              code,
		Message:           msg,
		IsAffectStability: DefaultIsAffectStability,
	}

	for _, opt := range opts {
		opt(definition)
	}

	CodeDefinitions[code] = definition
}

// SetDefaultErrorCode 设置默认错误码
func SetDefaultErrorCode(code int32) {
	ServiceInternalErrorCode = code
}
