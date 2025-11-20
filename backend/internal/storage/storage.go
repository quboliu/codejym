package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage persists metadata to PostgreSQL while keeping uploaded blobs on disk or S3.
type Storage struct {
	uploadsDir  string      // 保留用于兼容性（本地存储时使用）
	db          *pgxpool.Pool
	fileStorage FileStorage // 抽象文件存储接口
}

// New 创建 Storage 实例
// db: 数据库连接池
// root: 数据根目录（仅本地存储时使用）
// fileStorage: 文件存储实现（LocalStorage 或 S3Storage），如果为 nil 则使用默认本地存储
func New(db *pgxpool.Pool, root string, fileStorage FileStorage) (*Storage, error) {
	if db == nil {
		return nil, errors.New("storage: db pool is nil")
	}
	if root == "" {
		root = "data"
	}
	uploadsDir := filepath.Join(root, "uploads")

	// 如果未提供文件存储，使用默认本地存储
	if fileStorage == nil {
		if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
			return nil, err
		}
		var err error
		fileStorage, err = NewLocalStorage(uploadsDir)
		if err != nil {
			return nil, err
		}
	}

	return &Storage{
		uploadsDir:  uploadsDir,
		db:          db,
		fileStorage: fileStorage,
	}, nil
}

// Migrate ensures required tables exist.
func (s *Storage) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS assets (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			root_path TEXT NOT NULL,
			size_bytes BIGINT NOT NULL,
			file_count INT NOT NULL,
			source_name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_assets_user ON assets(user_id);`,
		`CREATE TABLE IF NOT EXISTS typing_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
			rel_path TEXT NOT NULL,
			cursor INT NOT NULL,
			errors INT NOT NULL,
			duration_seconds INT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_asset ON typing_sessions(user_id, asset_id);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Asset struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Name       string    `json:"name"`
	RootPath   string    `json:"rootPath"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	SizeBytes  int64     `json:"sizeBytes"`
	FileCount  int       `json:"fileCount"`
	SourceName string    `json:"sourceName"`
}

type Session struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	AssetID         string    `json:"assetId"`
	RelPath         string    `json:"relPath"`
	Cursor          int       `json:"cursor"`
	Errors          int       `json:"errors"`
	DurationSeconds int       `json:"durationSeconds"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// FileStorage 返回文件存储接口
func (s *Storage) FileStorage() FileStorage {
	return s.fileStorage
}

func (s *Storage) AssetDir(userID, assetID string) string {
	return filepath.Join(s.uploadsDir, userID, assetID)
}

func (s *Storage) UploadsDir() string {
	return s.uploadsDir
}

func (s *Storage) CreateUser(ctx context.Context, user *User) error {
	return s.db.QueryRow(
		ctx,
		`INSERT INTO users (id, email, name, password_hash) VALUES ($1, $2, $3, $4)
		 RETURNING created_at, updated_at`,
		user.ID, user.Email, user.Name, user.PasswordHash,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := s.db.QueryRow(ctx, `SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = $1`, email)
	return scanUser(row)
}

func (s *Storage) GetUserByID(ctx context.Context, id string) (*User, error) {
	row := s.db.QueryRow(ctx, `SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*User, error) {
	u := &User{}
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *Storage) RegisterAsset(ctx context.Context, asset *Asset) error {
	return s.db.QueryRow(
		ctx,
		`INSERT INTO assets (id, user_id, name, root_path, size_bytes, file_count, source_name)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at, updated_at`,
		asset.ID, asset.UserID, asset.Name, asset.RootPath, asset.SizeBytes, asset.FileCount, asset.SourceName,
	).Scan(&asset.CreatedAt, &asset.UpdatedAt)
}

func (s *Storage) ListAssets(ctx context.Context, userID string) ([]*Asset, error) {
	rows, err := s.db.Query(ctx, `SELECT id, user_id, name, root_path, size_bytes, file_count, source_name, created_at, updated_at FROM assets WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assets []*Asset
	for rows.Next() {
		a := &Asset{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.RootPath, &a.SizeBytes, &a.FileCount, &a.SourceName, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (s *Storage) GetAsset(ctx context.Context, userID, assetID string) (*Asset, error) {
	a := &Asset{}
	err := s.db.QueryRow(
		ctx,
		`SELECT id, user_id, name, root_path, size_bytes, file_count, source_name, created_at, updated_at
		 FROM assets WHERE id = $1 AND user_id = $2`,
		assetID, userID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.RootPath, &a.SizeBytes, &a.FileCount, &a.SourceName, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (s *Storage) DeleteAsset(ctx context.Context, userID, assetID string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM assets WHERE id = $1 AND user_id = $2`, assetID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Storage) UpdateAsset(ctx context.Context, asset *Asset) error {
	err := s.db.QueryRow(
		ctx,
		`UPDATE assets
		 SET name = $1, updated_at = now()
		 WHERE id = $2 AND user_id = $3
		 RETURNING updated_at`,
		asset.Name, asset.ID, asset.UserID,
	).Scan(&asset.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Storage) CreateSession(ctx context.Context, session *Session) error {
	return s.db.QueryRow(
		ctx,
		`INSERT INTO typing_sessions (id, user_id, asset_id, rel_path, cursor, errors, duration_seconds)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at, updated_at`,
		session.ID, session.UserID, session.AssetID, session.RelPath, session.Cursor, session.Errors, session.DurationSeconds,
	).Scan(&session.CreatedAt, &session.UpdatedAt)
}

func (s *Storage) GetSession(ctx context.Context, userID, sessionID string) (*Session, error) {
	sess := &Session{}
	err := s.db.QueryRow(
		ctx,
		`SELECT id, user_id, asset_id, rel_path, cursor, errors, duration_seconds, created_at, updated_at
		 FROM typing_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	).Scan(&sess.ID, &sess.UserID, &sess.AssetID, &sess.RelPath, &sess.Cursor, &sess.Errors, &sess.DurationSeconds, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sess, nil
}

// GetSessionByAssetAndPath retrieves an existing session by user, asset, and file path.
func (s *Storage) GetSessionByAssetAndPath(ctx context.Context, userID, assetID, relPath string) (*Session, error) {
	sess := &Session{}
	err := s.db.QueryRow(
		ctx,
		`SELECT id, user_id, asset_id, rel_path, cursor, errors, duration_seconds, created_at, updated_at
		 FROM typing_sessions WHERE user_id = $1 AND asset_id = $2 AND rel_path = $3
		 ORDER BY updated_at DESC LIMIT 1`,
		userID, assetID, relPath,
	).Scan(&sess.ID, &sess.UserID, &sess.AssetID, &sess.RelPath, &sess.Cursor, &sess.Errors, &sess.DurationSeconds, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sess, nil
}

func (s *Storage) UpdateSession(ctx context.Context, session *Session) error {
	err := s.db.QueryRow(
		ctx,
		`UPDATE typing_sessions
		 SET cursor = $1, errors = $2, duration_seconds = $3, updated_at = now()
		 WHERE id = $4 AND user_id = $5
		 RETURNING updated_at`,
		session.Cursor, session.Errors, session.DurationSeconds, session.ID, session.UserID,
	).Scan(&session.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func RandomID() (string, error) {
	var b [10]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

var ErrNotFound = errors.New("storage: not found")

func IsDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
