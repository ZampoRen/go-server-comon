package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	// ErrObjectNotFound 对象未找到错误
	ErrObjectNotFound = errors.New("object not found")
)

// Storage 存储接口
type Storage interface {
	// PutObject 上传对象到指定的键
	PutObject(ctx context.Context, objectKey string, content []byte, opts ...PutOptFn) error
	// PutObjectWithReader 使用 Reader 上传对象到指定的键
	PutObjectWithReader(ctx context.Context, objectKey string, content io.Reader, opts ...PutOptFn) error
	// GetObject 获取指定键的对象
	GetObject(ctx context.Context, objectKey string) ([]byte, error)
	// DeleteObject 删除指定键的对象
	DeleteObject(ctx context.Context, objectKey string) error
	// GetObjectUrl 返回对象的预签名 URL
	// URL 在指定的有效期内有效
	GetObjectUrl(ctx context.Context, objectKey string, opts ...GetOptFn) (string, error)
	// HeadObject 返回指定键的对象元数据
	HeadObject(ctx context.Context, objectKey string, opts ...GetOptFn) (*FileInfo, error)
	// ListAllObjects 返回指定前缀的所有对象
	// 可能返回大量对象，建议使用 ListObjectsPaginated 以获得更好的性能
	ListAllObjects(ctx context.Context, prefix string, opts ...GetOptFn) ([]*FileInfo, error)
	// ListObjectsPaginated 返回支持分页的对象列表
	// 处理大量对象时使用此方法
	ListObjectsPaginated(ctx context.Context, input *ListObjectsPaginatedInput, opts ...GetOptFn) (*ListObjectsPaginatedOutput, error)
}

// SecurityToken 安全令牌
type SecurityToken struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
	ExpiredTime     string `json:"expired_time"`
	CurrentTime     string `json:"current_time"`
}

// ListObjectsPaginatedInput 分页列出对象的输入参数
type ListObjectsPaginatedInput struct {
	Prefix   string // 前缀
	PageSize int    // 每页大小
	Cursor   string // 游标
}

// ListObjectsPaginatedOutput 分页列出对象的输出结果
type ListObjectsPaginatedOutput struct {
	Files       []*FileInfo // 文件列表
	Cursor      string      // 游标
	IsTruncated bool        // false: 所有结果已返回，true: 还有更多结果
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string            `json:"key"`           // 对象键
	LastModified time.Time         `json:"last_modified"` // 最后修改时间
	ETag         string            `json:"etag"`          // ETag
	Size         int64             `json:"size"`          // 大小
	URL          string            `json:"url"`           // URL
	Tagging      map[string]string `json:"tagging"`       // 标签
}
