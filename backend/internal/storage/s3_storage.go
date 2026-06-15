package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage S3 对象存储实现（兼容缤纷云 S4）
type S3Storage struct {
	client    *s3.Client
	bucket    string
	urlPrefix string // CDN URL 前缀，如 "https://cdn.example.com"
}

// NewS3Storage 创建 S3 存储实例
func NewS3Storage(endpoint, accessKey, secretKey, bucket, region, urlPrefix string) (*S3Storage, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("s3 storage: endpoint cannot be empty")
	}
	if accessKey == "" {
		return nil, fmt.Errorf("s3 storage: access key cannot be empty")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("s3 storage: secret key cannot be empty")
	}
	if bucket == "" {
		return nil, fmt.Errorf("s3 storage: bucket cannot be empty")
	}
	if region == "" {
		region = "us-east-1" // 默认区域
	}

	// 加载 AWS 配置
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("s3 storage: failed to load config: %w", err)
	}

	// 创建 S3 客户端
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // 使用路径风格访问（缤纷云兼容）
	})

	return &S3Storage{
		client:    client,
		bucket:    bucket,
		urlPrefix: strings.TrimSuffix(urlPrefix, "/"),
	}, nil
}

// SaveFile 上传文件到 S3
func (s *S3Storage) SaveFile(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	// 清理路径，统一使用 Unix 风格斜杠
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "/")
	// 防止路径遍历（与本地存储一致）
	if path == ".." || strings.HasPrefix(path, "../") {
		return "", fmt.Errorf("s3 storage: invalid path (contains ..): %s", path)
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("s3 storage: failed to upload file: %w", err)
	}

	return path, nil
}

// GetFile 从 S3 下载文件
func (s *S3Storage) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// 清理路径
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "/")

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("s3 storage: failed to get file: %w", err)
	}

	return result.Body, nil
}

// DeleteFile 删除 S3 上的单个文件
func (s *S3Storage) DeleteFile(ctx context.Context, path string) error {
	// 清理路径
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "/")

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("s3 storage: failed to delete file: %w", err)
	}

	return nil
}

// DeleteDir 删除 S3 上的目录（递归删除所有匹配前缀的对象）
func (s *S3Storage) DeleteDir(ctx context.Context, path string) error {
	// 清理路径
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "/")

	// 确保路径以斜杠结尾（匹配目录）
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// 列出所有文件
	files, err := s.ListFiles(ctx, path)
	if err != nil {
		return err
	}

	// 批量删除（S3 支持一次删除最多 1000 个对象）
	const batchSize = 1000
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		objects := make([]types.ObjectIdentifier, len(batch))
		for j, file := range batch {
			objects[j] = types.ObjectIdentifier{
				Key: aws.String(file),
			}
		}

		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		}

		_, err := s.client.DeleteObjects(ctx, input)
		if err != nil {
			return fmt.Errorf("s3 storage: failed to delete directory: %w", err)
		}
	}

	return nil
}

// Move 移动/重命名对象（S3 无原生 move：对前缀下每个对象 copy 到新键再删除原键）
func (s *S3Storage) Move(ctx context.Context, from, to string) error {
	from = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(from)), "/")
	to = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(to)), "/")
	if from == ".." || strings.HasPrefix(from, "../") || to == ".." || strings.HasPrefix(to, "../") {
		return fmt.Errorf("s3 storage: invalid path (contains ..)")
	}
	// 收集要移动的对象：精确对象 + 该前缀（目录）下的所有对象
	keys := map[string]struct{}{from: {}}
	if objs, err := s.ListFiles(ctx, from+"/"); err == nil {
		for _, k := range objs {
			keys[k] = struct{}{}
		}
	}
	for srcKey := range keys {
		var dstKey string
		switch {
		case srcKey == from:
			dstKey = to
		case strings.HasPrefix(srcKey, from+"/"):
			dstKey = to + strings.TrimPrefix(srcKey, from)
		default:
			continue
		}
		if _, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
			Bucket:     aws.String(s.bucket),
			CopySource: aws.String(s.bucket + "/" + srcKey),
			Key:        aws.String(dstKey),
		}); err != nil {
			// 精确对象可能不存在（纯目录移动）——跳过即可
			if srcKey == from {
				continue
			}
			return fmt.Errorf("s3 storage: failed to copy %s: %w", srcKey, err)
		}
		if err := s.DeleteFile(ctx, srcKey); err != nil {
			return fmt.Errorf("s3 storage: failed to delete original %s: %w", srcKey, err)
		}
	}
	return nil
}

// GetURL 获取文件访问 URL
func (s *S3Storage) GetURL(ctx context.Context, path string) (string, error) {
	// 清理路径
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "/")

	// 如果配置了 CDN URL，直接返回
	if s.urlPrefix != "" {
		return fmt.Sprintf("%s/%s", s.urlPrefix, path), nil
	}

	// 否则返回 S3 直接访问 URL
	// 注意：这需要 Bucket 配置为公开访问或使用预签名 URL
	// 这里简化处理，实际生产环境应该生成预签名 URL
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, path), nil
}

// ListFiles 列出目录下的所有文件
func (s *S3Storage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	// 清理路径
	prefix = filepath.ToSlash(filepath.Clean(prefix))
	prefix = strings.TrimPrefix(prefix, "/")

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("s3 storage: failed to list files: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key != nil {
				files = append(files, *obj.Key)
			}
		}
	}

	return files, nil
}
