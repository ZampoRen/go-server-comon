package fileutil

import (
	"context"

	"github.com/ZampoRen/go-server-comon/internal/infra/storage"
)

// AssembleFileUrl 为文件列表组装 URL
func AssembleFileUrl(ctx context.Context, urlExpire *int64, files []*storage.FileInfo, s storage.Storage) ([]*storage.FileInfo, error) {
	if files == nil || s == nil {
		return files, nil
	}

	// 使用简单的并发方式获取 URL
	// 注意：这里简化了实现，实际可以使用 taskgroup 等并发库
	for _, f := range files {
		expire := int64(7 * 60 * 60 * 24) // 默认 7 天
		if urlExpire != nil && *urlExpire > 0 {
			expire = *urlExpire
		}

		url, err := s.GetObjectUrl(ctx, f.Key, storage.WithExpire(expire))
		if err != nil {
			return nil, err
		}

		f.URL = url
	}

	return files, nil
}
