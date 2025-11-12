package sonic

import "github.com/bytedance/sonic"

var config = sonic.Config{
	UseInt64: true,
}.Froze()

// Marshal 返回 v 的 JSON 编码字节
func Marshal(val interface{}) ([]byte, error) {
	return config.Marshal(val)
}

// MarshalIndent 类似于 Marshal，但应用 Indent 来格式化输出
// 输出中的每个 JSON 元素将在新行上开始，以 prefix 开头
// 后跟一个或多个根据缩进嵌套的 indent 副本
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return config.MarshalIndent(v, prefix, indent)
}

// MarshalString 返回 v 的 JSON 编码字符串
func MarshalString(val interface{}) (string, error) {
	return config.MarshalToString(val)
}

// Unmarshal 解析 JSON 编码的数据并将结果存储在 v 指向的值中
// 注意：此 API 默认复制给定的缓冲区
// 如果您想更高效地传递 JSON，请使用 UnmarshalString
func Unmarshal(buf []byte, val interface{}) error {
	return config.Unmarshal(buf, val)
}

// UnmarshalString 类似于 Unmarshal，但 buf 是字符串
func UnmarshalString(buf string, val interface{}) error {
	return config.UnmarshalFromString(buf, val)
}
