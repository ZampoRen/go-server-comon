package es

import (
	"encoding/json"
	"io"
)

// BulkIndexerItem 批量索引器项
type BulkIndexerItem struct {
	Index           string        // 索引名称
	Action          string        // 操作类型
	DocumentID      string        // 文档 ID
	Routing         string        // 路由
	Version         *int64        // 版本
	VersionType     string        // 版本类型
	Body            io.ReadSeeker // 文档内容
	RetryOnConflict *int          // 冲突重试次数
}

// Request 搜索请求
type Request struct {
	Size        *int        // 返回结果数量
	Query       *Query      // 查询条件
	MinScore    *float64    // 最小分数
	Sort        []SortFiled // 排序字段
	SearchAfter []any       // 搜索后游标
	From        *int        // 起始位置
}

// SortFiled 排序字段
type SortFiled struct {
	Field string // 字段名
	Asc   bool   // 是否升序
}

// Response 搜索响应
type Response struct {
	Hits     HitsMetadata `json:"hits"`                // 命中结果
	MaxScore *float64     `json:"max_score,omitempty"` // 最大分数
}

// HitsMetadata 命中结果元数据
type HitsMetadata struct {
	Hits     []Hit    `json:"hits"`                // 命中列表
	MaxScore *float64 `json:"max_score,omitempty"` // 最大分数
	// Total 总命中数信息，仅在搜索请求中 `track_total_hits` 不为 `false` 时存在
	Total *TotalHits `json:"total,omitempty"`
}

// Hit 命中结果
type Hit struct {
	Id_     *string         `json:"_id,omitempty"`     // 文档 ID
	Score_  *float64        `json:"_score,omitempty"`  // 分数
	Source_ json.RawMessage `json:"_source,omitempty"` // 源文档
}

// TotalHits 总命中数
type TotalHits struct {
	Relation TotalHitsRelation `json:"relation"` // 关系类型
	Value    int64             `json:"value"`    // 值
}

// TotalHitsRelation 总命中数关系类型
type TotalHitsRelation struct {
	Name string // 名称
}
