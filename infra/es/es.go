package es

import (
	"context"
)

// Client Elasticsearch 客户端接口
type Client interface {
	// Create 创建文档
	Create(ctx context.Context, index, id string, document any) error
	// Update 更新文档
	Update(ctx context.Context, index, id string, document any) error
	// Delete 删除文档
	Delete(ctx context.Context, index, id string) error
	// Search 搜索文档
	Search(ctx context.Context, index string, req *Request) (*Response, error)
	// Exists 检查索引是否存在
	Exists(ctx context.Context, index string) (bool, error)
	// CreateIndex 创建索引
	CreateIndex(ctx context.Context, index string, properties map[string]any) error
	// DeleteIndex 删除索引
	DeleteIndex(ctx context.Context, index string) error
	// Types 返回类型工具
	Types() Types
	// NewBulkIndexer 创建批量索引器
	NewBulkIndexer(index string) (BulkIndexer, error)
}

// Types 类型工具接口
type Types interface {
	// NewLongNumberProperty 创建长整型数字属性
	NewLongNumberProperty() any
	// NewTextProperty 创建文本属性
	NewTextProperty() any
	// NewUnsignedLongNumberProperty 创建无符号长整型数字属性
	NewUnsignedLongNumberProperty() any
}

// BulkIndexer 批量索引器接口
type BulkIndexer interface {
	// Add 添加索引项
	Add(ctx context.Context, item BulkIndexerItem) error
	// Close 关闭批量索引器
	Close(ctx context.Context) error
}
