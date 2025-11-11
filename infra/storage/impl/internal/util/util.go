package util

import (
	"net/url"
	"strings"
)

// MapToQuery 将 map 转换为 query string
func MapToQuery(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	var parts []string
	for k, v := range m {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	return strings.Join(parts, "&")
}
