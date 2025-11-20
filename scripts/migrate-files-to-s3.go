package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"codecopybook/internal/storage"
)

func main() {
	ctx := context.Background()

	// 连接数据库
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbpool.Close()

	// 初始化 S3 存储
	s3Storage, err := storage.NewS3Storage(
		getEnvOrExit("S3_ENDPOINT"),
		getEnvOrExit("S3_ACCESS_KEY"),
		getEnvOrExit("S3_SECRET_KEY"),
		getEnvOrExit("S3_BUCKET"),
		getEnvOrDefault("S3_REGION", "us-east-1"),
		os.Getenv("S3_URL_PREFIX"), // 可选
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}

	// 扫描本地文件
	uploadsDir := getEnvOrDefault("LOCAL_UPLOADS_DIR", "./data/uploads")
	log.Printf("Scanning local uploads directory: %s", uploadsDir)

	var fileCount, successCount, failCount int
	var totalBytes int64

	err = filepath.Walk(uploadsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("ERROR: Failed to access %s: %v", path, err)
			return nil // 继续处理其他文件
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		fileCount++
		totalBytes += info.Size()

		// 计算相对路径
		relPath, err := filepath.Rel(uploadsDir, path)
		if err != nil {
			log.Printf("ERROR: Failed to get relative path for %s: %v", path, err)
			failCount++
			return nil
		}

		// 统一使用 Unix 风格路径
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// 读取文件
		file, err := os.Open(path)
		if err != nil {
			log.Printf("ERROR: Failed to open %s: %v", path, err)
			failCount++
			return nil
		}
		defer file.Close()

		// 上传到 S3
		s3Path := "uploads/" + relPath
		_, err = s3Storage.SaveFile(ctx, s3Path, file, detectContentType(path))
		if err != nil {
			log.Printf("ERROR: Failed to upload %s: %v", relPath, err)
			failCount++
			return nil
		}

		successCount++
		if successCount%10 == 0 {
			log.Printf("Progress: %d/%d files uploaded (%.1f%%), %d failed",
				successCount, fileCount, float64(successCount)/float64(fileCount)*100, failCount)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk directory: %v", err)
	}

	// 打印迁移摘要
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Migration Summary")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total files scanned:  %d\n", fileCount)
	fmt.Printf("Successfully uploaded: %d\n", successCount)
	fmt.Printf("Failed:               %d\n", failCount)
	fmt.Printf("Total size:           %.2f MB\n", float64(totalBytes)/(1024*1024))
	fmt.Println(strings.Repeat("=", 50))

	if failCount > 0 {
		log.Printf("WARNING: %d files failed to upload. Please check the errors above.", failCount)
	}

	// 更新数据库中的文件路径（将本地路径改为 S3 路径）
	if successCount > 0 {
		fmt.Println("\nUpdating database asset paths...")

		bucket := getEnvOrExit("S3_BUCKET")
		updateQuery := `
			UPDATE assets
			SET root_path = 's3://' || $1 || '/' || root_path
			WHERE root_path NOT LIKE 's3://%'
			AND root_path NOT LIKE 'http://%'
			AND root_path NOT LIKE 'https://%'
		`

		result, err := dbpool.Exec(ctx, updateQuery, bucket)
		if err != nil {
			log.Fatalf("Failed to update database: %v", err)
		}

		fmt.Printf("Updated %d asset records in database\n", result.RowsAffected())
	}

	fmt.Println("\n✅ Migration completed!")
	fmt.Println("Next steps:")
	fmt.Println("1. Verify files are accessible in your application")
	fmt.Println("2. Test upload and download functionality")
	fmt.Println("3. Once confirmed, you can safely delete local uploads:")
	fmt.Printf("   rm -rf %s\n", uploadsDir)
}

// getEnvOrExit 获取必需的环境变量，如果不存在则退出
func getEnvOrExit(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// detectContentType 根据文件扩展名检测 Content-Type
func detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "text/x-go"
	case ".js":
		return "application/javascript"
	case ".ts":
		return "application/typescript"
	case ".py":
		return "text/x-python"
	case ".java":
		return "text/x-java"
	case ".c", ".h":
		return "text/x-c"
	case ".cpp", ".cc", ".cxx":
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
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
