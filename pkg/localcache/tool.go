package localcache

// AnyValue 将 any 类型转换为泛型类型 V
// 如果 err 不为 nil，返回零值和错误
// 如果类型断言失败，会 panic
func AnyValue[V any](v any, err error) (V, error) {
	if err != nil {
		var zero V
		return zero, err
	}
	val, ok := v.(V)
	if !ok {
		var zero V
		return zero, &TypeAssertionError{Value: v, Type: "V"}
	}
	return val, nil
}

// TypeAssertionError 类型断言错误
type TypeAssertionError struct {
	Value any
	Type  string
}

func (e *TypeAssertionError) Error() string {
	return "type assertion failed"
}
