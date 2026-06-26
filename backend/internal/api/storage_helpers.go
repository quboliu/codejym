package api

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"codecopybook/internal/storage"
)

// extractZipToStorage 解压 ZIP 文件并上传到存储（FileStorage 接口）
func extractZipToStorage(ctx context.Context, f *os.File, fileStorage storage.FileStorage, basePath string, fileCount *int, bytesTotal *int64) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	reader, err := zip.NewReader(f, info.Size())
	if err != nil {
		return err
	}

	// 解压配额：防止 zip 炸弹 / 海量小文件耗尽磁盘
	const (
		maxZipEntries          = 10000
		maxZipUncompressedSize = 256 << 20 // 256MB
	)
	if len(reader.File) > maxZipEntries {
		return fmt.Errorf("zip 文件数超过上限 (%d)", maxZipEntries)
	}
	var declared uint64
	for _, zf := range reader.File {
		declared += zf.UncompressedSize64
	}
	if declared > maxZipUncompressedSize {
		return fmt.Errorf("zip 解压后体积超过上限 (%dMB)", maxZipUncompressedSize>>20)
	}

	for _, zipFile := range reader.File {
		rel := sanitizeZipPath(zipFile.Name)
		if rel == "" {
			continue
		}

		// 跳过目录（S3 不需要显式创建目录）
		if zipFile.FileInfo().IsDir() {
			continue
		}

		// 构建存储路径
		storagePath := path.Join(basePath, filepath.ToSlash(rel))

		// 打开 ZIP 中的文件
		rc, err := zipFile.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in zip: %w", rel, err)
		}

		// 检测 Content-Type
		contentType := detectContentTypeFromFilename(rel)

		// 上传到存储
		_, err = fileStorage.SaveFile(ctx, storagePath, rc, contentType)
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to save file %s: %w", rel, err)
		}

		*fileCount++
		*bytesTotal += int64(zipFile.UncompressedSize64)
	}

	return nil
}

// saveFileToStorage 保存单个文件到存储
func saveFileToStorage(ctx context.Context, src io.Reader, fileStorage storage.FileStorage, storagePath string, contentType string) error {
	_, err := fileStorage.SaveFile(ctx, storagePath, src, contentType)
	return err
}

// buildTreeFromStorage 从存储构建文件树（支持 S3 和本地存储）
func buildTreeFromStorage(ctx context.Context, fileStorage storage.FileStorage, basePath string, rel string) ([]fileNode, error) {
	// 列出所有文件
	prefix := path.Join(basePath, rel)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	files, err := fileStorage.ListFiles(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// 构建目录树结构
	return buildFileTree(files, prefix), nil
}

// buildFileTree 从扁平文件列表构建树形结构。
// 用指针树一次性建好父子关系（顺序无关），最后再递归转成值类型——
// 避免旧实现「在 map 随机遍历中按值拷贝节点」导致父目录先被拷贝、子节点丢失的 bug。
func buildFileTree(files []string, basePath string) []fileNode {
	basePath = strings.TrimPrefix(basePath, "/")

	type treeNode struct {
		name     string
		path     string
		isDir    bool
		children map[string]*treeNode
	}
	root := &treeNode{children: map[string]*treeNode{}}

	for _, f := range files {
		if !strings.HasPrefix(f, basePath) {
			continue
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(f, basePath), "/")
		if rel == "" {
			continue
		}
		parts := strings.Split(rel, "/")
		cur := root
		for i, p := range parts {
			isLast := i == len(parts)-1
			// 跳过目录占位用的 .keep（其父目录已建立，空文件夹仍可见）
			if isLast && p == ".keep" {
				break
			}
			child, ok := cur.children[p]
			if !ok {
				child = &treeNode{
					name:     p,
					path:     strings.Join(parts[:i+1], "/"),
					isDir:    !isLast,
					children: map[string]*treeNode{},
				}
				cur.children[p] = child
			}
			if !isLast {
				child.isDir = true
			}
			cur = child
		}
	}

	var convert func(n *treeNode) []fileNode
	convert = func(n *treeNode) []fileNode {
		out := make([]fileNode, 0, len(n.children))
		for _, c := range n.children {
			fn := fileNode{Name: c.name, Path: c.path, IsDir: c.isDir}
			if c.isDir {
				fn.Children = convert(c)
			}
			out = append(out, fn)
		}
		sortFileNodes(out)
		return out
	}
	return convert(root)
}

// sortFileNodes 递归排序文件节点
func sortFileNodes(nodes []fileNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].IsDir == nodes[j].IsDir {
			return nodes[i].Name < nodes[j].Name
		}
		return nodes[i].IsDir
	})

	for i := range nodes {
		if nodes[i].Children != nil {
			sortFileNodes(nodes[i].Children)
		}
	}
}

// readAssetFileFromStorage 从存储读取文件内容
func readAssetFileFromStorage(ctx context.Context, fileStorage storage.FileStorage, basePath, rel string) (*fileContent, error) {
	cleanRel, ok := cleanRelPath(rel)
	if !ok {
		return nil, fmt.Errorf("invalid path")
	}

	// 构建存储路径
	storagePath := path.Join(basePath, cleanRel)

	// 从存储读取文件
	reader, err := fileStorage.GetFile(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer reader.Close()

	// 读取内容
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	language := detectLanguage(cleanRel)
	content := string(data)

	return &fileContent{
		Name:     filepath.Base(cleanRel),
		Path:     filepath.ToSlash(cleanRel),
		Language: language,
		Content:  content,
	}, nil
}

func cleanRelPath(relPath string) (string, bool) {
	normalized := strings.TrimSpace(strings.ReplaceAll(relPath, "\\", "/"))
	if normalized == "" || path.IsAbs(normalized) {
		return "", false
	}
	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return "", false
		}
	}
	clean := path.Clean(normalized)
	if clean == "." || clean == "" {
		return "", false
	}
	return clean, true
}

func isValidBaseName(name string) bool {
	normalized := strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	return normalized != "" &&
		normalized != "." &&
		normalized != ".." &&
		!strings.Contains(normalized, "/") &&
		!strings.ContainsRune(normalized, 0)
}

// detectContentTypeFromFilename 根据文件名检测 Content-Type
func detectContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "text/x-go"
	case ".js":
		return "application/javascript"
	case ".ts":
		return "application/typescript"
	case ".tsx":
		return "text/tsx"
	case ".jsx":
		return "text/jsx"
	case ".py":
		return "text/x-python"
	case ".java":
		return "text/x-java"
	case ".c", ".h":
		return "text/x-c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "text/x-c++"
	case ".rs":
		return "text/x-rust"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".md", ".markdown":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".yaml", ".yml":
		return "application/x-yaml"
	case ".sh":
		return "application/x-sh"
	case ".sql":
		return "application/sql"
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}
