package api

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"codecopybook/internal/storage"
)

// Server wires HTTP handlers with storage.
type Server struct {
	store  *storage.Storage
	logger *log.Logger
}

func NewServer(store *storage.Storage, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New(os.Stdout, "[api] ", log.LstdFlags)
	}
	return &Server{store: store, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/assets/upload", s.handleAssetUpload)
	mux.HandleFunc("/api/assets/paste", s.handleAssetPaste)
	mux.HandleFunc("/api/assets", s.handleAssets)
	mux.HandleFunc("/api/assets/", s.handleAssetByID)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/sessions/", s.handleSessionByID)
	return withCORS(logRequests(mux, s.logger))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	assets := s.store.ListAssets()
	resp := make([]assetDTO, 0, len(assets))
	for _, a := range assets {
		resp = append(resp, toAssetDTO(a))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAssetUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	tmp, err := os.CreateTemp("", "codecopybook-upload-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store upload")
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	size, err := io.Copy(tmp, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read upload")
		return
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to seek upload")
		return
	}

	assetID, err := storage.RandomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to allocate id")
		return
	}

	assetDir := s.store.AssetDir(assetID)
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare storage")
		return
	}

	var fileCount int
	var bytesTotal int64
	isZip, err := detectZip(tmp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to inspect upload")
		return
	}
	if isZip {
		if err := extractZip(tmp, assetDir, &fileCount, &bytesTotal); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid zip: %v", err))
			return
		}
		if fileCount == 0 {
			writeError(w, http.StatusBadRequest, "zip contains no files")
			return
		}
	} else {
		dstName := sanitizeFilename(header.Filename)
		if dstName == "" {
			dstName = fmt.Sprintf("asset-%s", assetID)
		}
		dstPath := filepath.Join(assetDir, dstName)
		if _, err := tmp.Seek(0, io.SeekStart); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reset upload")
			return
		}
		if err := copyFile(tmp, dstPath); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store file")
			return
		}
		fileCount = 1
		bytesTotal = size
	}

	asset := &storage.Asset{
		ID:         assetID,
		Name:       deriveAssetName(header),
		RootPath:   assetDir,
		SizeBytes:  bytesTotal,
		FileCount:  fileCount,
		SourceName: header.Filename,
	}
	if err := s.store.RegisterAsset(asset); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to persist asset")
		return
	}
	writeJSON(w, http.StatusCreated, toAssetDTO(asset))
}

func (s *Server) handleAssetPaste(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var payload struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	content := payload.Content
	if strings.TrimSpace(content) == "" {
		writeError(w, http.StatusBadRequest, "content cannot be empty")
		return
	}
	filename := sanitizeFilename(payload.Filename)
	if filename == "" {
		filename = "pasted-snippet.txt"
	}
	data := []byte(content)

	assetID, err := storage.RandomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to allocate id")
		return
	}
	assetDir := s.store.AssetDir(assetID)
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare storage")
		return
	}
	dstPath := filepath.Join(assetDir, filename)
	if err := os.WriteFile(dstPath, data, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store file")
		return
	}
	asset := &storage.Asset{
		ID:         assetID,
		Name:       deriveAssetNameFromFilename(filename),
		RootPath:   assetDir,
		SizeBytes:  int64(len(data)),
		FileCount:  1,
		SourceName: filename,
	}
	if err := s.store.RegisterAsset(asset); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to persist asset")
		return
	}
	writeJSON(w, http.StatusCreated, toAssetDTO(asset))
}

func (s *Server) handleAssetByID(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/assets/")
	if trimmed == "" {
		http.NotFound(w, r)
		return
	}
	segments := strings.Split(trimmed, "/")
	id := segments[0]
	switch len(segments) {
	case 1:
		s.handleAssetRoot(id, w, r)
	case 2:
		switch segments[1] {
		case "tree":
			s.handleAssetTree(id, w, r)
		case "file":
			s.handleAssetFile(id, w, r)
		default:
			http.NotFound(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleAssetRoot(id string, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		s.deleteAsset(id, w, r)
	case http.MethodGet:
		if asset, ok := s.store.GetAsset(id); ok {
			writeJSON(w, http.StatusOK, toAssetDTO(asset))
			return
		}
		writeError(w, http.StatusNotFound, "asset not found")
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodDelete)
	}
}

func (s *Server) deleteAsset(id string, w http.ResponseWriter, r *http.Request) {
	asset, ok := s.store.GetAsset(id)
	if !ok {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	if err := os.RemoveAll(asset.RootPath); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete files")
		return
	}
	if err := s.store.DeleteAsset(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete metadata")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAssetTree(id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	asset, ok := s.store.GetAsset(id)
	if !ok {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	nodes, err := buildTree(asset.RootPath, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (s *Server) handleAssetFile(id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	asset, ok := s.store.GetAsset(id)
	if !ok {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	rel := r.URL.Query().Get("path")
	if rel == "" {
		writeError(w, http.StatusBadRequest, "path query is required")
		return
	}
	data, err := readAssetFile(asset.RootPath, rel)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "file not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createSession(w, r)
	default:
		methodNotAllowed(w, http.MethodPost)
	}
}

func (s *Server) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if trimmed == "" {
		http.NotFound(w, r)
		return
	}
	id := strings.Split(trimmed, "/")[0]
	switch r.Method {
	case http.MethodGet:
		s.getSession(id, w, r)
	case http.MethodPatch:
		s.updateSession(id, w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPatch)
	}
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		AssetID string `json:"assetId"`
		Path    string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if payload.AssetID == "" || payload.Path == "" {
		writeError(w, http.StatusBadRequest, "assetId and path are required")
		return
	}
	asset, ok := s.store.GetAsset(payload.AssetID)
	if !ok {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	if _, err := readAssetFile(asset.RootPath, payload.Path); err != nil {
		writeError(w, http.StatusBadRequest, "file path invalid")
		return
	}
	sessID, err := storage.RandomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	session := &storage.Session{
		ID:      sessID,
		AssetID: payload.AssetID,
		RelPath: payload.Path,
	}
	if err := s.store.CreateSession(session); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) getSession(id string, w http.ResponseWriter, r *http.Request) {
	session, ok := s.store.GetSession(id)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) updateSession(id string, w http.ResponseWriter, r *http.Request) {
	session, ok := s.store.GetSession(id)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	var payload struct {
		Cursor          *int `json:"cursor"`
		Errors          *int `json:"errors"`
		DurationSeconds *int `json:"durationSeconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if payload.Cursor != nil {
		session.Cursor = *payload.Cursor
	}
	if payload.Errors != nil {
		session.Errors = *payload.Errors
	}
	if payload.DurationSeconds != nil {
		session.DurationSeconds = *payload.DurationSeconds
	}
	if err := s.store.UpdateSession(session); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update session")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

type assetDTO struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	SizeBytes  int64     `json:"sizeBytes"`
	FileCount  int       `json:"fileCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	SourceName string    `json:"sourceName"`
}

type fileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []fileNode `json:"children,omitempty"`
}

type fileContent struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

func toAssetDTO(a *storage.Asset) assetDTO {
	return assetDTO{
		ID:         a.ID,
		Name:       a.Name,
		SizeBytes:  a.SizeBytes,
		FileCount:  a.FileCount,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
		SourceName: a.SourceName,
	}
}

func methodNotAllowed(w http.ResponseWriter, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ","))
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func logRequests(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(lrw, r)
		logger.Printf("%s %s %d %v", r.Method, r.URL.Path, lrw.status, time.Since(start))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.status = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func detectZip(f *os.File) (bool, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	var sig [4]byte
	n, err := io.ReadFull(f, sig[:])
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return false, nil
		}
		return false, err
	}
	if n < len(sig) {
		return false, nil
	}
	return sig == [4]byte{'P', 'K', 0x03, 0x04}, nil
}

func extractZip(f *os.File, dest string, fileCount *int, bytesTotal *int64) error {
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
		targetPath := filepath.Join(dest, rel)
		if zipFile.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		rc, err := zipFile.Open()
		if err != nil {
			return err
		}
		if err := writeReaderToFile(rc, targetPath); err != nil {
			rc.Close()
			return err
		}
		rc.Close()
		*fileCount++
		*bytesTotal += int64(zipFile.UncompressedSize64)
	}
	return nil
}

func sanitizeZipPath(name string) string {
	clean := path.Clean(name)
	clean = strings.TrimPrefix(clean, "../")
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || clean == "" {
		return ""
	}
	return filepath.FromSlash(clean)
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "..", "")
	if name == "" || name == "." || name == string(os.PathSeparator) {
		return ""
	}
	return name
}

func copyFile(src *os.File, dstPath string) error {
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := io.Copy(out, src); err != nil {
		return err
	}
	return nil
}

func writeReaderToFile(r io.Reader, dstPath string) error {
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, r)
	return err
}

func deriveAssetNameFromFilename(name string) string {
	if idx := strings.LastIndex(name, "."); idx > 0 {
		name = name[:idx]
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "code-copy-asset"
	}
	return name
}

func deriveAssetName(header *multipart.FileHeader) string {
	return deriveAssetNameFromFilename(header.Filename)
}

func buildTree(root string, rel string) ([]fileNode, error) {
	dir := filepath.Join(root, rel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() == entries[j].IsDir() {
			return entries[i].Name() < entries[j].Name()
		}
		return entries[i].IsDir()
	})
	nodes := make([]fileNode, 0, len(entries))
	for _, entry := range entries {
		entryRel := filepath.Join(rel, entry.Name())
		node := fileNode{
			Name:  entry.Name(),
			Path:  filepath.ToSlash(entryRel),
			IsDir: entry.IsDir(),
		}
		if entry.IsDir() {
			children, err := buildTree(root, entryRel)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func readAssetFile(root, rel string) (*fileContent, error) {
	cleanRel := filepath.Clean(rel)
	if strings.HasPrefix(cleanRel, "..") {
		return nil, errors.New("invalid path")
	}
	if filepath.IsAbs(cleanRel) {
		return nil, errors.New("invalid path")
	}
	fullPath := filepath.Join(root, cleanRel)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return &fileContent{
		Name:     filepath.Base(cleanRel),
		Path:     filepath.ToSlash(cleanRel),
		Language: detectLanguage(cleanRel),
		Content:  string(data),
	}, nil
}

func detectLanguage(rel string) string {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "jsx"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".c":
		return "c"
	case ".cpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".sh":
		return "shell"
	case ".bash":
		return "shell"
	case ".yaml":
		return "yaml"
	case ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".txt":
		return "text"
	case ".toml":
		return "toml"
	case ".cfg", ".conf":
		return "config"
	default:
		return "text"
	}
}
