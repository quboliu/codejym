package api

import (
	"archive/zip"
	"context"
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

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"codecopybook/internal/storage"
)

// Server wires HTTP handlers with storage.
type Server struct {
	store      *storage.Storage
	logger     *log.Logger
	authSecret []byte
	authTTL    time.Duration
}

type userContextKey struct{}

// 默认 Token 超时时间：24 小时（比原来的 30 天更安全）
const defaultAuthTokenTTL = 24 * time.Hour

// 获取 Token 超时时间（支持环境变量 AUTH_TOKEN_TTL 配置）
func getAuthTokenTTL() time.Duration {
	if ttlStr := os.Getenv("AUTH_TOKEN_TTL"); ttlStr != "" {
		// 支持格式：30m, 24h, 7d
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			return ttl
		}
		log.Printf("warning: invalid AUTH_TOKEN_TTL format, using default %v", defaultAuthTokenTTL)
	}
	return defaultAuthTokenTTL
}

func NewServer(store *storage.Storage, logger *log.Logger, authSecret []byte) *Server {
	if logger == nil {
		logger = log.New(os.Stdout, "[api] ", log.LstdFlags)
	}
	return &Server{
		store:      store,
		logger:     logger,
		authSecret: authSecret,
		authTTL:    getAuthTokenTTL(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/auth/signup", s.handleSignup)
	mux.HandleFunc("/api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/auth/me", s.withAuth(s.handleAuthMe))
	mux.HandleFunc("/api/assets/upload", s.withAuth(s.handleAssetUpload))
	mux.HandleFunc("/api/assets/paste", s.withAuth(s.handleAssetPaste))
	mux.HandleFunc("/api/assets", s.withAuth(s.handleAssets))
	mux.HandleFunc("/api/assets/", s.withAuth(s.handleAssetByID))
	mux.HandleFunc("/api/sessions", s.withAuth(s.handleSessions))
	mux.HandleFunc("/api/sessions/", s.withAuth(s.handleSessionByID))

	// 应用安全中间件
	handler := withSecurityHeaders(withCORS(logRequests(mux, s.logger)))
	return handler
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	email := strings.TrimSpace(strings.ToLower(payload.Email))
	password := strings.TrimSpace(payload.Password)
	name := strings.TrimSpace(payload.Name)
	if email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(password) < 6 {
		writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}
	if name == "" {
		name = email
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	userID, err := storage.RandomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to allocate id")
		return
	}
	user := &storage.User{
		ID:           userID,
		Email:        email,
		Name:         name,
		PasswordHash: string(hashed),
	}
	if err := s.store.CreateUser(r.Context(), user); err != nil {
		if storage.IsDuplicate(err) {
			writeError(w, http.StatusConflict, "email already registered")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	// 创建默认训练组
	if err := s.createDefaultAsset(r.Context(), user.ID); err != nil {
		// 记录错误但不阻止注册流程
		s.logger.Printf("warning: failed to create default asset for user %s: %v", user.ID, err)
	}

	token, err := s.issueToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	user.PasswordHash = ""
	writeJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  toUserDTO(user),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	email := strings.TrimSpace(strings.ToLower(payload.Email))
	password := payload.Password
	if email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	user, err := s.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to query user")
		}
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := s.issueToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	user.PasswordHash = ""
	writeJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  toUserDTO(user),
	})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(user))
}

func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.listAssets(user, w, r)
	case http.MethodPost:
		s.createAsset(user, w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) listAssets(user *storage.User, w http.ResponseWriter, r *http.Request) {
	assets, err := s.store.ListAssets(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load assets")
		return
	}
	resp := make([]assetDTO, 0, len(assets))
	for _, a := range assets {
		resp = append(resp, toAssetDTO(a))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) createAsset(user *storage.User, w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	assetID, err := storage.RandomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to allocate id")
		return
	}

	// 统一使用相对存储路径（与 paste/upload 一致），所有读写都走 FileStorage。
	// 空训练组无需预建目录：首次写文件时 FileStorage 会自动创建。
	basePath := fmt.Sprintf("uploads/%s/%s", user.ID, assetID)

	asset := &storage.Asset{
		ID:         assetID,
		UserID:     user.ID,
		Name:       name,
		RootPath:   basePath,
		SizeBytes:  0,
		FileCount:  0,
		SourceName: "",
	}
	if err := s.store.RegisterAsset(r.Context(), asset); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to persist asset")
		return
	}
	writeJSON(w, http.StatusCreated, toAssetDTO(asset))
}

// maxUploadBytes 上传请求体的硬上限（防止超大上传打满磁盘/内存）。
const maxUploadBytes = 64 << 20 // 64MB

func (s *Server) handleAssetUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
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

	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}

	// 存储路径：uploads/{userID}/{assetID}/
	basePath := fmt.Sprintf("uploads/%s/%s", user.ID, assetID)

	var fileCount int
	var bytesTotal int64
	isZip, err := detectZip(tmp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to inspect upload")
		return
	}

	fileStorage := s.store.FileStorage()
	ctx := r.Context()

	if isZip {
		// 解压并上传到存储
		if err := extractZipToStorage(ctx, tmp, fileStorage, basePath, &fileCount, &bytesTotal); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid zip: %v", err))
			return
		}
		if fileCount == 0 {
			writeError(w, http.StatusBadRequest, "zip contains no files")
			return
		}
	} else {
		// 单文件上传
		dstName := sanitizeFilename(header.Filename)
		if dstName == "" {
			dstName = fmt.Sprintf("asset-%s", assetID)
		}
		storagePath := fmt.Sprintf("%s/%s", basePath, dstName)
		contentType := detectContentTypeFromFilename(dstName)

		if _, err := tmp.Seek(0, io.SeekStart); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reset upload")
			return
		}
		if err := saveFileToStorage(ctx, tmp, fileStorage, storagePath, contentType); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store file")
			return
		}
		fileCount = 1
		bytesTotal = size
	}

	asset := &storage.Asset{
		ID:         assetID,
		UserID:     user.ID,
		Name:       deriveAssetName(header),
		RootPath:   basePath, // 存储路径：uploads/{userID}/{assetID}
		SizeBytes:  bytesTotal,
		FileCount:  fileCount,
		SourceName: header.Filename,
	}
	if err := s.store.RegisterAsset(r.Context(), asset); err != nil {
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

	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}

	assetID, err := storage.RandomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to allocate id")
		return
	}

	// 存储路径：uploads/{userID}/{assetID}/{filename}
	basePath := fmt.Sprintf("uploads/%s/%s", user.ID, assetID)
	storagePath := fmt.Sprintf("%s/%s", basePath, filename)
	contentType := detectContentTypeFromFilename(filename)

	// 上传到存储
	fileStorage := s.store.FileStorage()
	if err := saveFileToStorage(r.Context(), strings.NewReader(content), fileStorage, storagePath, contentType); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store file")
		return
	}

	asset := &storage.Asset{
		ID:         assetID,
		UserID:     user.ID,
		Name:       deriveAssetNameFromFilename(filename),
		RootPath:   basePath, // 存储路径：uploads/{userID}/{assetID}
		SizeBytes:  int64(len(data)),
		FileCount:  1,
		SourceName: filename,
	}
	if err := s.store.RegisterAsset(r.Context(), asset); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to persist asset")
		return
	}
	writeJSON(w, http.StatusCreated, toAssetDTO(asset))
}

func (s *Server) handleAssetByID(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/assets/")
	if trimmed == "" {
		http.NotFound(w, r)
		return
	}
	segments := strings.Split(trimmed, "/")
	id := segments[0]
	switch len(segments) {
	case 1:
		s.handleAssetRoot(user, id, w, r)
	case 2:
		switch segments[1] {
		case "tree":
			s.handleAssetTree(user, id, w, r)
		case "file":
			s.handleAssetFile(user, id, w, r)
		case "rename":
			s.handleAssetRename(user, id, w, r)
		case "mkdir":
			s.handleAssetMkdir(user, id, w, r)
		case "move-file":
			s.handleAssetMoveFile(user, id, w, r)
		case "rename-file":
			s.handleAssetRenameFile(user, id, w, r)
		case "delete-file":
			s.handleAssetDeleteFile(user, id, w, r)
		case "upload":
			s.handleAssetUploadToExisting(user, id, w, r)
		case "paste":
			s.handleAssetPasteToExisting(user, id, w, r)
		default:
			http.NotFound(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleAssetRoot(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		s.deleteAsset(user, id, w, r)
	case http.MethodGet:
		asset, err := s.store.GetAsset(r.Context(), user.ID, id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				writeError(w, http.StatusNotFound, "asset not found")
			} else {
				writeError(w, http.StatusInternalServerError, "failed to load asset")
			}
			return
		}
		writeJSON(w, http.StatusOK, toAssetDTO(asset))
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodDelete)
	}
}

func (s *Server) deleteAsset(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}
	// 走 FileStorage（相对路径），兼容本地与 S3；原来的 os.RemoveAll(相对路径)
	// 会解析到进程 CWD 而非存储根，导致文件残留、永不释放。
	if err := s.store.FileStorage().DeleteDir(r.Context(), asset.RootPath); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete files")
		return
	}
	if err := s.store.DeleteAsset(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete metadata")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAssetTree(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}

	// 使用新的存储抽象层获取文件树
	fileStorage := s.store.FileStorage()
	nodes, err := buildTreeFromStorage(r.Context(), fileStorage, asset.RootPath, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (s *Server) handleAssetFile(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}
	rel := r.URL.Query().Get("path")
	if rel == "" {
		writeError(w, http.StatusBadRequest, "path query is required")
		return
	}

	// 使用新的存储抽象层读取文件
	fileStorage := s.store.FileStorage()
	data, err := readAssetFileFromStorage(r.Context(), fileStorage, asset.RootPath, rel)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "file not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.querySession(user, w, r)
	case http.MethodPost:
		s.createSession(user, w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if trimmed == "" {
		http.NotFound(w, r)
		return
	}
	id := strings.Split(trimmed, "/")[0]
	switch r.Method {
	case http.MethodGet:
		s.getSession(user, id, w, r)
	case http.MethodPatch:
		s.updateSession(user, id, w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPatch)
	}
}

func (s *Server) querySession(user *storage.User, w http.ResponseWriter, r *http.Request) {
	assetID := r.URL.Query().Get("assetId")
	path := r.URL.Query().Get("path")
	if assetID == "" || path == "" {
		writeError(w, http.StatusBadRequest, "assetId and path query parameters are required")
		return
	}
	session, err := s.store.GetSessionByAssetAndPath(r.Context(), user.ID, assetID, path)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to query session")
		}
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) createSession(user *storage.User, w http.ResponseWriter, r *http.Request) {
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
	asset, err := s.store.GetAsset(r.Context(), user.ID, payload.AssetID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}

	// 验证文件是否存在（使用新的存储抽象层）
	fileStorage := s.store.FileStorage()
	if _, err := readAssetFileFromStorage(r.Context(), fileStorage, asset.RootPath, payload.Path); err != nil {
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
		UserID:  user.ID,
		AssetID: payload.AssetID,
		RelPath: payload.Path,
	}
	if err := s.store.CreateSession(r.Context(), session); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) getSession(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	session, err := s.store.GetSession(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load session")
		}
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) updateSession(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Cursor          *int `json:"cursor"`
		Errors          *int `json:"errors"`
		DurationSeconds *int `json:"durationSeconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	// 单次往返完成部分更新（替代 GetSession + UpdateSession 两次查询）
	session, err := s.store.UpdateSessionFields(r.Context(), user.ID, id, payload.Cursor, payload.Errors, payload.DurationSeconds)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to update session")
		}
		return
	}
	writeJSON(w, http.StatusOK, session)
}

type authResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

type userDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type assetDTO struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
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
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Language   string      `json:"language"`
	Content    string      `json:"content"`
	SkipRanges []skipRange `json:"skipRanges"`
}

type skipRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func toAssetDTO(a *storage.Asset) assetDTO {
	return assetDTO{
		ID:         a.ID,
		UserID:     a.UserID,
		Name:       a.Name,
		SizeBytes:  a.SizeBytes,
		FileCount:  a.FileCount,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
		SourceName: a.SourceName,
	}
}

func toUserDTO(u *storage.User) userDTO {
	return userDTO{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
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

// 安全头中间件 - 禁用 Cookies 并添加安全头
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 明确拒绝 Cookies - 系统不使用 Cookies 认证
		w.Header().Set("Set-Cookie", "Path=/; HttpOnly; Max-Age=0")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// 安全相关头
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// 明确说明认证方式：使用 Authorization Header Bearer Token，不使用 Cookies
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"CodeJYM API\"")
		w.Header().Set("X-Auth-Method", "JWT Bearer Token (no cookies)")

		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		authHeader := r.Header.Get("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization")
			return
		}
		user, err := s.parseToken(strings.TrimSpace(parts[1]))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		// 身份信息来自已签名的 JWT，无需查库（无状态鉴权）
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next(w, r.WithContext(ctx))
	}
}

// authClaims 在 JWT 中携带身份信息（email/name），使鉴权变为无状态：
// withAuth 不再每个请求都查 user 表（进度保存热路径少一次 DB 往返）。
// 代价：用户改名后旧 token 内的 name 会陈旧到过期为止（可接受）。
type authClaims struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	jwt.RegisteredClaims
}

func (s *Server) issueToken(u *storage.User) (string, error) {
	if len(s.authSecret) == 0 {
		return "", errors.New("auth secret not configured")
	}
	now := time.Now()
	claims := authClaims{
		Email: u.Email,
		Name:  u.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.authTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.authSecret)
}

func (s *Server) parseToken(tokenStr string) (*storage.User, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return s.authSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*authClaims)
	if !ok || !token.Valid || claims.Subject == "" {
		return nil, errors.New("invalid token")
	}
	return &storage.User{ID: claims.Subject, Email: claims.Email, Name: claims.Name}, nil
}

func currentUser(r *http.Request) *storage.User {
	if r == nil {
		return nil
	}
	user, _ := r.Context().Value(userContextKey{}).(*storage.User)
	return user
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
	// 以 "/" 为根做 Clean：任何试图越界的 "../" 都会在此被折叠掉，
	// 无法逃出 basePath（修复 zip-slip 越权写）。原实现只 TrimPrefix 一次 "../"，
	// 残留的 ".." 会在 path.Join 时把写入位置带到其他用户目录。
	// 显式把反斜杠也当作分隔符（ZIP 规范用 "/"，但恶意条目可能用 "\" 在 Linux 上绕过）
	clean := path.Clean("/" + strings.ReplaceAll(name, "\\", "/"))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || clean == "" {
		return ""
	}
	// 兜底：Clean 之后不应再残留任何 ".." 段
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
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

// handleAssetRename 重命名训练组
func (s *Server) handleAssetRename(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		methodNotAllowed(w, http.MethodPost, http.MethodPatch)
		return
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	newName := strings.TrimSpace(payload.Name)
	if newName == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}
	asset.Name = newName
	if err := s.store.UpdateAsset(r.Context(), asset); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update asset")
		return
	}
	writeJSON(w, http.StatusOK, toAssetDTO(asset))
}

// handleAssetMkdir 在训练组内创建目录
func (s *Server) handleAssetMkdir(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var payload struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	relPath := strings.TrimSpace(payload.Path)
	if relPath == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	// 验证路径安全性
	if !isValidRelPath(relPath) {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}
	fullPath := filepath.Join(asset.RootPath, filepath.FromSlash(relPath))
	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create directory")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "directory created"})
}

// handleAssetMoveFile 移动文件或文件夹
func (s *Server) handleAssetMoveFile(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var payload struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	fromPath := strings.TrimSpace(payload.From)
	toPath := strings.TrimSpace(payload.To)
	if fromPath == "" || toPath == "" {
		writeError(w, http.StatusBadRequest, "from and to paths are required")
		return
	}
	// 验证路径安全性
	if !isValidRelPath(fromPath) || !isValidRelPath(toPath) {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}
	fullFromPath := filepath.Join(asset.RootPath, filepath.FromSlash(fromPath))
	fullToPath := filepath.Join(asset.RootPath, filepath.FromSlash(toPath))

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(fullToPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create destination directory")
		return
	}

	if err := os.Rename(fullFromPath, fullToPath); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to move file: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "file moved successfully"})
}

// handleAssetRenameFile 重命名文件或文件夹
func (s *Server) handleAssetRenameFile(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		methodNotAllowed(w, http.MethodPost, http.MethodPatch)
		return
	}
	var payload struct {
		Path    string `json:"path"`
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	relPath := strings.TrimSpace(payload.Path)
	newName := strings.TrimSpace(payload.NewName)
	if relPath == "" || newName == "" {
		writeError(w, http.StatusBadRequest, "path and newName are required")
		return
	}
	// 验证路径安全性
	if !isValidRelPath(relPath) {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	// 确保新名称不包含路径分隔符
	if strings.ContainsAny(newName, "/\\") {
		writeError(w, http.StatusBadRequest, "newName cannot contain path separators")
		return
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}
	fullPath := filepath.Join(asset.RootPath, filepath.FromSlash(relPath))
	dir := filepath.Dir(fullPath)
	newFullPath := filepath.Join(dir, newName)

	if err := os.Rename(fullPath, newFullPath); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to rename: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "renamed successfully"})
}

// handleAssetDeleteFile 删除文件或文件夹
func (s *Server) handleAssetDeleteFile(user *storage.User, id string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodDelete, http.MethodPost)
		return
	}
	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		writeError(w, http.StatusBadRequest, "path query parameter is required")
		return
	}
	// 验证路径安全性
	if !isValidRelPath(relPath) {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}
	fullPath := filepath.Join(asset.RootPath, filepath.FromSlash(relPath))

	if err := os.RemoveAll(fullPath); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete: %v", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleAssetUploadToExisting 上传文件到现有训练组
func (s *Server) handleAssetUploadToExisting(user *storage.User, assetID string, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	// 获取现有的 asset
	asset, err := s.store.GetAsset(r.Context(), user.ID, assetID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
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

	var fileCount int
	var bytesTotal int64
	isZip, err := detectZip(tmp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to inspect upload")
		return
	}

	fileStorage := s.store.FileStorage()
	ctx := r.Context()
	basePath := asset.RootPath // 复用现有的存储路径

	if isZip {
		// 解压并上传到存储
		if err := extractZipToStorage(ctx, tmp, fileStorage, basePath, &fileCount, &bytesTotal); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid zip: %v", err))
			return
		}
		if fileCount == 0 {
			writeError(w, http.StatusBadRequest, "zip contains no files")
			return
		}
	} else {
		// 单文件上传
		dstName := sanitizeFilename(header.Filename)
		if dstName == "" {
			dstName = fmt.Sprintf("file-%d", time.Now().Unix())
		}
		storagePath := fmt.Sprintf("%s/%s", basePath, dstName)
		contentType := detectContentTypeFromFilename(dstName)

		if _, err := tmp.Seek(0, io.SeekStart); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reset upload")
			return
		}
		if err := saveFileToStorage(ctx, tmp, fileStorage, storagePath, contentType); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store file")
			return
		}
		fileCount = 1
		bytesTotal = size
	}

	// 更新 asset 统计信息
	asset.FileCount += fileCount
	asset.SizeBytes += bytesTotal
	if err := s.store.UpdateAsset(r.Context(), asset); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update asset")
		return
	}

	writeJSON(w, http.StatusOK, toAssetDTO(asset))
}

// handleAssetPasteToExisting 粘贴代码到现有训练组
func (s *Server) handleAssetPasteToExisting(user *storage.User, assetID string, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	// 获取现有的 asset
	asset, err := s.store.GetAsset(r.Context(), user.ID, assetID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to load asset")
		}
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

	// 存储路径
	basePath := asset.RootPath // 复用现有的存储路径
	storagePath := fmt.Sprintf("%s/%s", basePath, filename)
	contentType := detectContentTypeFromFilename(filename)

	// 上传到存储
	fileStorage := s.store.FileStorage()
	if err := saveFileToStorage(r.Context(), strings.NewReader(content), fileStorage, storagePath, contentType); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store file")
		return
	}

	// 更新 asset 统计信息
	asset.FileCount += 1
	asset.SizeBytes += int64(len(data))
	if err := s.store.UpdateAsset(r.Context(), asset); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update asset")
		return
	}

	writeJSON(w, http.StatusOK, toAssetDTO(asset))
}

// isValidRelPath 验证相对路径的安全性
func isValidRelPath(relPath string) bool {
	cleanPath := filepath.Clean(relPath)
	// 拒绝绝对路径
	if filepath.IsAbs(cleanPath) {
		return false
	}
	// 拒绝包含 .. 的路径（防止路径遍历攻击）
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, string(filepath.Separator)+"..") {
		return false
	}
	return true
}

// createDefaultAsset 为新用户创建默认训练组
func (s *Server) createDefaultAsset(ctx context.Context, userID string) error {
	assetID, err := storage.RandomID()
	if err != nil {
		return fmt.Errorf("failed to generate asset id: %w", err)
	}

	// 统一相对路径 + 走 FileStorage：保证写入路径与读取路径一致，
	// 否则会出现双重前缀（/data/uploads/data/uploads/...）导致 welcome.md 读不到。
	basePath := fmt.Sprintf("uploads/%s/%s", userID, assetID)

	// 创建一个示例文件
	welcomeContent := `# 欢迎使用 CodeJYM！

这是你的默认训练组。你可以：

1. 上传代码文件进行练习
2. 上传整个项目的 ZIP 包
3. 创建新的训练组来组织不同的练习内容

开始你的代码临摹之旅吧！
`
	if _, err := s.store.FileStorage().SaveFile(ctx, basePath+"/welcome.md", strings.NewReader(welcomeContent), "text/markdown; charset=utf-8"); err != nil {
		return fmt.Errorf("failed to create welcome file: %w", err)
	}

	asset := &storage.Asset{
		ID:         assetID,
		UserID:     userID,
		Name:       "默认训练组",
		RootPath:   basePath,
		SizeBytes:  int64(len(welcomeContent)),
		FileCount:  1,
		SourceName: "welcome.md",
	}

	if err := s.store.RegisterAsset(ctx, asset); err != nil {
		return fmt.Errorf("failed to register asset: %w", err)
	}

	return nil
}
