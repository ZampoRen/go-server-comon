package storage

import (
	"time"
)

// GetOptFn 获取选项函数
type GetOptFn func(option *GetOption)

// GetOption 获取选项
type GetOption struct {
	Expire      int64 // 过期时间（秒）
	WithURL     bool  // 是否包含 URL
	WithTagging bool  // 是否包含标签
}

// WithExpire 设置过期时间
func WithExpire(expire int64) GetOptFn {
	return func(o *GetOption) {
		o.Expire = expire
	}
}

// WithURL 设置是否包含 URL
func WithURL(withURL bool) GetOptFn {
	return func(o *GetOption) {
		o.WithURL = withURL
	}
}

// WithGetTagging 设置是否包含标签
func WithGetTagging(withTagging bool) GetOptFn {
	return func(o *GetOption) {
		o.WithTagging = withTagging
	}
}

// PutOption 上传选项
type PutOption struct {
	ContentType        *string           // 内容类型
	ContentEncoding    *string           // 内容编码
	ContentDisposition *string           // 内容处置
	ContentLanguage    *string           // 内容语言
	Expires            *time.Time        // 过期时间
	Tagging            map[string]string // 标签
	ObjectSize         int64             // 对象大小
}

// PutOptFn 上传选项函数
type PutOptFn func(option *PutOption)

// WithTagging 设置标签
func WithTagging(tag map[string]string) PutOptFn {
	return func(o *PutOption) {
		if len(tag) > 0 {
			o.Tagging = make(map[string]string, len(tag))
			for k, v := range tag {
				o.Tagging[k] = v
			}
		}
	}
}

// WithContentType 设置内容类型
func WithContentType(v string) PutOptFn {
	return func(o *PutOption) {
		o.ContentType = &v
	}
}

// WithObjectSize 设置对象大小
func WithObjectSize(v int64) PutOptFn {
	return func(o *PutOption) {
		o.ObjectSize = v
	}
}

// WithContentEncoding 设置内容编码
func WithContentEncoding(v string) PutOptFn {
	return func(o *PutOption) {
		o.ContentEncoding = &v
	}
}

// WithContentDisposition 设置内容处置
func WithContentDisposition(v string) PutOptFn {
	return func(o *PutOption) {
		o.ContentDisposition = &v
	}
}

// WithContentLanguage 设置内容语言
func WithContentLanguage(v string) PutOptFn {
	return func(o *PutOption) {
		o.ContentLanguage = &v
	}
}

// WithExpires 设置过期时间
func WithExpires(v time.Time) PutOptFn {
	return func(o *PutOption) {
		o.Expires = &v
	}
}
