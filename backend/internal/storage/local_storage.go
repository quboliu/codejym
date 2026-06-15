package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage 本地文件系统存储实现
type LocalStorage struct {
	rootDir string // 根目录，如 "/data/uploads"
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(rootDir string) (*LocalStorage, error) {
	if rootDir == "" {
		return nil, fmt.Errorf("local storage: root directory cannot be empty")
	}

	// 创建根目录
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return nil, fmt.Errorf("local storage: failed to create root directory: %w", err)
	}

	return &LocalStorage{
		rootDir: rootDir,
	}, nil
}

// SaveFile 保存文件到本地文件系统
func (s *LocalStorage) SaveFile(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	// 清理路径，防止路径遍历攻击
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "..") {
		return "", fmt.Errorf("local storage: invalid path (contains ..): %s", path)
	}

	fullPath := filepath.Join(s.rootDir, path)

	// 创建父目录
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("local storage: failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("local storage: failed to create file: %w", err)
	}
	defer file.Close()

	// 写入内容
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("local storage: failed to write file: %w", err)
	}

	return path, nil
}

// GetFile 从本地文件系统获取文件
func (s *LocalStorage) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// 清理路径
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "..") {
		return nil, fmt.Errorf("local storage: invalid path (contains ..): %s", path)
	}

	fullPath := filepath.Join(s.rootDir, path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local storage: file not found: %s", path)
		}
		return nil, fmt.Errorf("local storage: failed to open file: %w", err)
	}

	return file, nil
}

// DeleteFile 删除单个文件
func (s *LocalStorage) DeleteFile(ctx context.Context, path string) error {
	// 清理路径
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "..") {
		return fmt.Errorf("local storage: invalid path (contains ..): %s", path)
	}

	fullPath := filepath.Join(s.rootDir, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在视为成功
		}
		return fmt.Errorf("local storage: failed to delete file: %w", err)
	}

	return nil
}

// DeleteDir 递归删除目录
func (s *LocalStorage) DeleteDir(ctx context.Context, path string) error {
	// 清理路径
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "..") {
		return fmt.Errorf("local storage: invalid path (contains ..): %s", path)
	}

	fullPath := filepath.Join(s.rootDir, path)

	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("local storage: failed to delete directory: %w", err)
	}

	return nil
}

// Move 移动/重命名文件或目录
func (s *LocalStorage) Move(ctx context.Context, from, to string) error {
	from = filepath.Clean(from)
	to = filepath.Clean(to)
	if strings.HasPrefix(from, "..") || strings.HasPrefix(to, "..") {
		return fmt.Errorf("local storage: invalid path (contains ..)")
	}
	src := filepath.Join(s.rootDir, from)
	dst := filepath.Join(s.rootDir, to)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("local storage: failed to create destination directory: %w", err)
	}
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("local storage: failed to move: %w", err)
	}
	return nil
}

// GetURL 获取文件访问 URL（本地存储返回相对路径）
func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	// 本地存储返回相对路径，由应用层处理
	return path, nil
}

// ListFiles 列出目录下的所有文件
func (s *LocalStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	// 清理路径
	prefix = filepath.Clean(prefix)
	if strings.HasPrefix(prefix, "..") {
		return nil, fmt.Errorf("local storage: invalid path (contains ..): %s", prefix)
	}

	fullPrefix := filepath.Join(s.rootDir, prefix)
	var files []string

	err := filepath.Walk(fullPrefix, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 如果目录不存在，返回空列表而不是错误
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		// 跳过目录，只返回文件
		if !info.IsDir() {
			relPath, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}
			// 统一使用 Unix 风格路径
			relPath = filepath.ToSlash(relPath)
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("local storage: failed to list files: %w", err)
	}

	return files, nil
}
