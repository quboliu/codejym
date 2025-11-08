package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"codecopybook/internal/api"
	"codecopybook/internal/storage"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "server listen address")
	dataDir := flag.String("data", envOr("DATA_DIR", "data"), "data directory for uploads and metadata")
	flag.Parse()

	store, err := storage.New(*dataDir)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	server := api.NewServer(store, log.New(os.Stdout, "[codecopybook] ", log.LstdFlags))
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

func mountFrontend(apiHandler http.Handler, dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
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
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}
