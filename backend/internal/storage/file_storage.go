package storage

import (
	"context"
	"io"
)

// FileStorage 抽象文件存储接口
// 实现：LocalStorage（本地文件系统）、S3Storage（对象存储）
type FileStorage interface {
	// SaveFile 保存文件到存储
	// path: 文件存储路径（相对路径，如 "uploads/userID/assetID/file.go"）
	// reader: 文件内容读取器
	// contentType: MIME 类型（如 "text/plain", "application/zip"）
	// 返回：存储路径（用于后续访问）和错误
	SaveFile(ctx context.Context, path string, reader io.Reader, contentType string) (string, error)

	// GetFile 获取文件内容
	// path: 文件存储路径
	// 返回：可读取的文件流（调用者负责关闭）和错误
	GetFile(ctx context.Context, path string) (io.ReadCloser, error)

	// DeleteFile 删除单个文件
	DeleteFile(ctx context.Context, path string) error

	// DeleteDir 递归删除目录及其所有内容
	// 注意：S3 没有真正的目录概念，会删除所有匹配前缀的对象
	DeleteDir(ctx context.Context, path string) error

	// GetURL 获取文件访问 URL
	// 本地存储：返回相对路径
	// S3 存储：返回 CDN URL 或预签名 URL
	GetURL(ctx context.Context, path string) (string, error)

	// ListFiles 列出目录下的所有文件
	// prefix: 目录前缀（如 "uploads/userID/"）
	// 返回：文件路径列表和错误
	ListFiles(ctx context.Context, prefix string) ([]string, error)
}
