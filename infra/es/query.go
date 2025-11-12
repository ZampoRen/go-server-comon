package es

const (
	// QueryTypeEqual 等值查询
	QueryTypeEqual = "equal"
	// QueryTypeMatch 匹配查询
	QueryTypeMatch = "match"
	// QueryTypeMultiMatch 多字段匹配查询
	QueryTypeMultiMatch = "multi_match"
	// QueryTypeNotExists 不存在查询
	QueryTypeNotExists = "not_exists"
	// QueryTypeContains 包含查询
	QueryTypeContains = "contains"
	// QueryTypeIn 包含在查询
	QueryTypeIn = "in"
)

// KV 键值对
type KV struct {
	Key   string // 键
	Value any    // 值
}

// QueryType 查询类型
type QueryType string

// Query 查询
type Query struct {
	KV              KV              // 键值对
	Type            QueryType       // 查询类型
	MultiMatchQuery MultiMatchQuery // 多字段匹配查询
	Bool            *BoolQuery      // 布尔查询
}

// BoolQuery 布尔查询
type BoolQuery struct {
	Filter             []Query // 过滤条件
	Must               []Query // 必须匹配
	MustNot            []Query // 必须不匹配
	Should             []Query // 应该匹配
	MinimumShouldMatch *int    // 最小应该匹配数
}

// MultiMatchQuery 多字段匹配查询
type MultiMatchQuery struct {
	Fields   []string // 字段列表
	Type     string   // 类型，如 best_fields
	Query    string   // 查询内容
	Operator string   // 操作符
}

const (
	// Or 或操作
	Or = "or"
	// And 与操作
	And = "and"
)

// NewEqualQuery 创建等值查询
func NewEqualQuery(k string, v any) Query {
	return Query{
		KV:   KV{Key: k, Value: v},
		Type: QueryTypeEqual,
	}
}

// NewMatchQuery 创建匹配查询
func NewMatchQuery(k string, v any) Query {
	return Query{
		KV:   KV{Key: k, Value: v},
		Type: QueryTypeMatch,
	}
}

// NewMultiMatchQuery 创建多字段匹配查询
func NewMultiMatchQuery(fields []string, query, typeStr, operator string) Query {
	return Query{
		Type: QueryTypeMultiMatch,
		MultiMatchQuery: MultiMatchQuery{
			Fields:   fields,
			Query:    query,
			Operator: operator,
			Type:     typeStr,
		},
	}
}

// NewNotExistsQuery 创建不存在查询
func NewNotExistsQuery(k string) Query {
	return Query{
		KV:   KV{Key: k},
		Type: QueryTypeNotExists,
	}
}

// NewContainsQuery 创建包含查询
func NewContainsQuery(k string, v any) Query {
	return Query{
		KV:   KV{Key: k, Value: v},
		Type: QueryTypeContains,
	}
}

// NewInQuery 创建包含在查询
func NewInQuery[T any](k string, v []T) Query {
	arr := make([]any, 0, len(v))
	for _, item := range v {
		arr = append(arr, item)
	}
	return Query{
		KV:   KV{Key: k, Value: arr},
		Type: QueryTypeIn,
	}
}
