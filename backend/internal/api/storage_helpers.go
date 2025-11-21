package api

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
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

	log.Printf("[DEBUG] buildTreeFromStorage: basePath=%s, rel=%s, prefix=%s", basePath, rel, prefix)

	files, err := fileStorage.ListFiles(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	log.Printf("[DEBUG] ListFiles returned %d files", len(files))
	if len(files) > 0 {
		log.Printf("[DEBUG] First file: %s", files[0])
	}

	// 构建目录树结构
	tree := buildFileTree(files, prefix)
	log.Printf("[DEBUG] buildFileTree returned %d nodes", len(tree))
	return tree, nil
}

// buildFileTree 从文件列表构建树形结构
func buildFileTree(files []string, basePath string) []fileNode {
	// 规范化 basePath：去掉前导 /，确保与 S3 返回的路径格式一致
	basePath = strings.TrimPrefix(basePath, "/")

	log.Printf("[DEBUG] buildFileTree: basePath=%s, files count=%d", basePath, len(files))

	// 移除 basePath 前缀
	relFiles := make([]string, 0, len(files))
	for _, f := range files {
		log.Printf("[DEBUG] Processing file: %s, hasPrefix(%s): %v", f, basePath, strings.HasPrefix(f, basePath))
		if strings.HasPrefix(f, basePath) {
			relPath := strings.TrimPrefix(f, basePath)
			log.Printf("[DEBUG] Added relPath: %s", relPath)
			relFiles = append(relFiles, relPath)
		}
	}

	log.Printf("[DEBUG] relFiles count: %d", len(relFiles))

	// 构建树形结构
	root := make(map[string]*fileNode)

	for _, file := range relFiles {
		parts := strings.Split(file, "/")
		if len(parts) == 0 {
			continue
		}

		// 处理目录
		for i := 0; i < len(parts)-1; i++ {
			dirPath := strings.Join(parts[:i+1], "/")
			if _, exists := root[dirPath]; !exists {
				root[dirPath] = &fileNode{
					Name:  parts[i],
					Path:  dirPath,
					IsDir: true,
				}
			}
		}

		// 处理文件
		filePath := file
		root[filePath] = &fileNode{
			Name:  parts[len(parts)-1],
			Path:  filePath,
			IsDir: false,
		}
	}

	// 构建父子关系
	nodes := make([]fileNode, 0)
	for nodePath, node := range root {
		parentPath := path.Dir(nodePath)
		if parentPath == "." || parentPath == "" || parentPath == "/" {
			// 根级节点
			nodes = append(nodes, *node)
		} else if parent, exists := root[parentPath]; exists {
			// 添加到父节点
			if parent.Children == nil {
				parent.Children = make([]fileNode, 0)
			}
			parent.Children = append(parent.Children, *node)
		}
	}

	// 排序（目录在前，文件在后）
	sortFileNodes(nodes)
	return nodes
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
	cleanRel := filepath.Clean(rel)
	if strings.HasPrefix(cleanRel, "..") {
		return nil, fmt.Errorf("invalid path")
	}
	if filepath.IsAbs(cleanRel) {
		return nil, fmt.Errorf("invalid path")
	}

	// 构建存储路径
	storagePath := path.Join(basePath, filepath.ToSlash(cleanRel))

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

	// 查找注释范围（不删除注释，只标记位置）
	skipRanges := FindCommentRanges(content, language)

	return &fileContent{
		Name:       filepath.Base(cleanRel),
		Path:       filepath.ToSlash(cleanRel),
		Language:   language,
		Content:    content,
		SkipRanges: skipRanges,
	}, nil
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

