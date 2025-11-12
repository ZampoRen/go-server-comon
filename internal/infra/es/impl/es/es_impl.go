package es

import (
	"fmt"
	"os"

	"github.com/ZampoRen/go-server-comon/internal/infra/es"
)

// 类型别名
type (
	Client          = es.Client
	Types           = es.Types
	BulkIndexer     = es.BulkIndexer
	BulkIndexerItem = es.BulkIndexerItem
	BoolQuery       = es.BoolQuery
	Query           = es.Query
	Response        = es.Response
	Request         = es.Request
)

// New 创建 Elasticsearch 客户端
// 根据环境变量 ES_VERSION 决定创建 ES7 或 ES8 客户端
// 支持的值: v7, v8
func New() (Client, error) {
	v := os.Getenv("ES_VERSION")
	if v == "v8" {
		return newES8()
	} else if v == "v7" {
		return newES7()
	}

	return nil, fmt.Errorf("unsupported es version %s", v)
}
