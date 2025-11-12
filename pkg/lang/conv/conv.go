package conv

import (
	"encoding/json"
	"fmt"
)

// DebugJsonToStr 将对象转换为 JSON 字符串，用于调试
func DebugJsonToStr(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<json marshal error: %v>", err)
	}
	return string(data)
}
