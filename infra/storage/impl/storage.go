package impl

import (
	"context"
	"fmt"

	"github.com/ZampoRen/go-server-comon/infra/storage"
	"github.com/ZampoRen/go-server-comon/infra/storage/impl/aliyun"
	"github.com/ZampoRen/go-server-comon/infra/storage/impl/tencent"
	"github.com/ZampoRen/go-server-comon/infra/storage/impl/volcengine"
	"github.com/ZampoRen/go-server-comon/pkg/envkey"
)

// Storage 存储接口类型别名
type Storage = storage.Storage

// New 根据环境变量创建存储客户端
// 支持的类型: tos, aliyun, tencent
// 环境变量:
//   - STORAGE_TYPE: 存储类型 (tos/aliyun/tencent)
//   - STORAGE_BUCKET: 存储桶名称
//   - TOS_ACCESS_KEY, TOS_SECRET_KEY, TOS_ENDPOINT, TOS_REGION: 火山引擎 TOS 配置
//   - ALIYUN_ACCESS_KEY, ALIYUN_SECRET_KEY, ALIYUN_ENDPOINT, ALIYUN_REGION: 阿里云 OSS 配置
//   - TENCENT_ACCESS_KEY, TENCENT_SECRET_KEY, TENCENT_ENDPOINT, TENCENT_REGION: 腾讯云 COS 配置
func New(ctx context.Context) (Storage, error) {
	storageType := envkey.GetStringD("STORAGE_TYPE", "")
	bucketName := envkey.GetStringD("STORAGE_BUCKET", "")

	switch storageType {
	case "tos":
		return volcengine.New(
			ctx,
			envkey.GetStringD("TOS_ACCESS_KEY", ""),
			envkey.GetStringD("TOS_SECRET_KEY", ""),
			bucketName,
			envkey.GetStringD("TOS_ENDPOINT", ""),
			envkey.GetStringD("TOS_REGION", ""),
		)
	case "aliyun":
		return aliyun.New(
			ctx,
			envkey.GetStringD("ALIYUN_ACCESS_KEY", ""),
			envkey.GetStringD("ALIYUN_SECRET_KEY", ""),
			bucketName,
			envkey.GetStringD("ALIYUN_ENDPOINT", ""),
			envkey.GetStringD("ALIYUN_REGION", ""),
		)
	case "tencent":
		return tencent.New(
			ctx,
			envkey.GetStringD("TENCENT_ACCESS_KEY", ""),
			envkey.GetStringD("TENCENT_SECRET_KEY", ""),
			bucketName,
			envkey.GetStringD("TENCENT_ENDPOINT", ""),
			envkey.GetStringD("TENCENT_REGION", ""),
		)
	default:
		return nil, fmt.Errorf("unknown storage type: %s, supported types: tos, aliyun, tencent", storageType)
	}
}

// NewWithType 根据指定类型创建存储客户端
func NewWithType(ctx context.Context, storageType string, ak, sk, bucketName, endpoint, region string) (Storage, error) {
	switch storageType {
	case "tos":
		return volcengine.New(ctx, ak, sk, bucketName, endpoint, region)
	case "aliyun":
		return aliyun.New(ctx, ak, sk, bucketName, endpoint, region)
	case "tencent":
		return tencent.New(ctx, ak, sk, bucketName, endpoint, region)
	default:
		return nil, fmt.Errorf("unknown storage type: %s, supported types: tos, aliyun, tencent", storageType)
	}
}
