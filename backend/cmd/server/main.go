package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"codecopybook/internal/api"
	"codecopybook/internal/storage"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "server listen address")
	dataDir := flag.String("data", envOr("DATA_DIR", "data"), "data directory for uploads and metadata")
	dbURL := flag.String("db", envOr("DATABASE_URL", ""), "PostgreSQL connection string")
	authSecret := flag.String("auth-secret", envOr("AUTH_SECRET", ""), "HMAC secret for auth tokens")
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("DATABASE_URL / -db is required")
	}
	if *authSecret == "" {
		log.Fatal("AUTH_SECRET / -auth-secret is required")
	}

	ctx := context.Background()
	poolCfg, err := pgxpool.ParseConfig(*dbURL)
	if err != nil {
		log.Fatalf("invalid DATABASE_URL: %v", err)
	}
	// 连接池容量。原先用 pgxpool 默认值 max(4, CPU 数)≈8，是进度保存并发的主要瓶颈。
	poolCfg.MaxConns = int32(envInt("DB_MAX_CONNS", 50))
	poolCfg.MinConns = int32(envInt("DB_MIN_CONNS", 5))
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()
	log.Printf("postgres pool configured: max_conns=%d min_conns=%d", poolCfg.MaxConns, poolCfg.MinConns)

	// 初始化文件存储（本地或 S3）
	var fileStorage storage.FileStorage
	storageType := envOr("STORAGE_TYPE", "local")

	switch storageType {
	case "s3":
		// S3 存储配置
		s3Endpoint := envOr("S3_ENDPOINT", "")
		s3AccessKey := envOr("S3_ACCESS_KEY", "")
		s3SecretKey := envOr("S3_SECRET_KEY", "")
		s3Bucket := envOr("S3_BUCKET", "")
		s3Region := envOr("S3_REGION", "us-east-1")
		s3URLPrefix := envOr("S3_URL_PREFIX", "")

		if s3Endpoint == "" || s3AccessKey == "" || s3SecretKey == "" || s3Bucket == "" {
			log.Fatal("S3 storage requires: S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY, S3_BUCKET")
		}

		fileStorage, err = storage.NewS3Storage(s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket, s3Region, s3URLPrefix)
		if err != nil {
			log.Fatalf("failed to initialize S3 storage: %v", err)
		}
		log.Printf("using S3 storage: endpoint=%s, bucket=%s", s3Endpoint, s3Bucket)

	case "local":
		// 本地存储配置
		uploadsDir := filepath.Join(*dataDir, "uploads")
		fileStorage, err = storage.NewLocalStorage(uploadsDir)
		if err != nil {
			log.Fatalf("failed to initialize local storage: %v", err)
		}
		log.Printf("using local storage: %s", uploadsDir)

	default:
		log.Fatalf("invalid STORAGE_TYPE: %s (must be 'local' or 's3')", storageType)
	}

	store, err := storage.New(pool, *dataDir, fileStorage)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	server := api.NewServer(store, log.New(os.Stdout, "[codecopybook] ", log.LstdFlags), []byte(*authSecret))
	handler := server.Handler()

	if frontendDir := envOr("FRONTEND_DIR", ""); frontendDir != "" {
		if stat, err := os.Stat(frontendDir); err == nil && stat.IsDir() {
			handler = mountFrontend(handler, frontendDir)
			log.Printf("serving frontend from %s", frontendDir)
		} else if err != nil && !os.IsNotExist(err) {
			log.Fatalf("failed to stat frontend dir: %v", err)
		}
	}

	log.Printf("starting server on %s (data dir: %s)", *addr, *dataDir)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func mountFrontend(apiHandler http.Handler, dir string) http.Handler {
	indexPath := filepath.Join(dir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
			apiHandler.ServeHTTP(w, r)
			return
		}
		clean := filepath.Clean(r.URL.Path)
		clean = strings.TrimPrefix(clean, "/")
		target := filepath.Join(dir, clean)
		rel, err := filepath.Rel(dir, target)
		if err != nil || strings.HasPrefix(rel, "..") {
			http.NotFound(w, r)
			return
		}
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			serveFile(w, r, target, info.IsDir())
			return
		}
		serveIndex(w, r, indexPath)
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, path string, isDir bool) {
	if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") {
		noCache(w)
	}
	http.ServeFile(w, r, path)
}

func serveIndex(w http.ResponseWriter, r *http.Request, indexPath string) {
	noCache(w)
	http.ServeFile(w, r, indexPath)
}

func noCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
