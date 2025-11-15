package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"codecopybook/internal/storage"
)

type legacyAsset struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RootPath   string `json:"rootPath"`
	SizeBytes  int64  `json:"sizeBytes"`
	FileCount  int    `json:"fileCount"`
	SourceName string `json:"sourceName"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type legacySession struct {
	ID              string `json:"id"`
	AssetID         string `json:"assetId"`
	RelPath         string `json:"relPath"`
	Cursor          int    `json:"cursor"`
	Errors          int    `json:"errors"`
	DurationSeconds int    `json:"durationSeconds"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

func main() {
	var (
		dbURL        = flag.String("db", "", "PostgreSQL connection string (required)")
		dataDirFlag  = flag.String("data", "data", "data directory containing uploads and legacy JSON")
		assetsFile   = flag.String("assets", "", "path to legacy assets.json (defaults to <data>/assets.json)")
		sessionsFile = flag.String("sessions", "", "path to legacy sessions.json (defaults to <data>/sessions.json)")
		userEmail    = flag.String("user-email", "", "email to own migrated data (required)")
		userName     = flag.String("user-name", "", "display name when creating the user")
		userPassword = flag.String("user-password", "", "password when creating the user (ignored if user exists)")
	)
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("missing required -db flag")
	}
	if *userEmail == "" {
		log.Fatal("missing required -user-email flag")
	}

	dataDir, err := filepath.Abs(*dataDirFlag)
	if err != nil {
		log.Fatalf("failed to resolve data dir: %v", err)
	}
	if *assetsFile == "" {
		*assetsFile = filepath.Join(dataDir, "assets.json")
	}
	if *sessionsFile == "" {
		*sessionsFile = filepath.Join(dataDir, "sessions.json")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dbURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	store, err := storage.New(pool, dataDir)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	user, err := store.GetUserByEmail(ctx, normalizeEmail(*userEmail))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			if *userPassword == "" {
				log.Fatal("user does not exist; provide -user-password to create it")
			}
			if *userName == "" {
				*userName = *userEmail
			}
			user = createUser(ctx, store, *userEmail, *userName, *userPassword)
			log.Printf("created user %s (%s)", user.Name, user.Email)
		} else {
			log.Fatalf("failed to look up user: %v", err)
		}
	} else {
		log.Printf("using existing user %s (%s)", user.Name, user.Email)
	}

	assets, err := loadLegacyAssets(*assetsFile)
	if err != nil {
		log.Fatalf("failed to load legacy assets: %v", err)
	}
	sessions, err := loadLegacySessions(*sessionsFile)
	if err != nil {
		log.Fatalf("failed to load legacy sessions: %v", err)
	}

	log.Printf("found %d legacy assets, %d legacy sessions", len(assets), len(sessions))

	var importedAssets, importedSessions int
	for _, asset := range assets {
		if err := migrateAsset(ctx, pool, user.ID, dataDir, asset); err != nil {
			log.Fatalf("asset %s migration failed: %v", asset.ID, err)
		}
		importedAssets++
	}
	for _, sess := range sessions {
		if err := migrateSession(ctx, pool, user.ID, sess); err != nil {
			log.Fatalf("session %s migration failed: %v", sess.ID, err)
		}
		importedSessions++
	}

	log.Printf("migration finished: %d assets, %d sessions imported for %s", importedAssets, importedSessions, user.Email)
}

func createUser(ctx context.Context, store *storage.Storage, email, name, password string) *storage.User {
	email = normalizeEmail(email)
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}
	userID, err := storage.RandomID()
	if err != nil {
		log.Fatalf("failed to allocate user id: %v", err)
	}
	user := &storage.User{
		ID:           userID,
		Email:        email,
		Name:         name,
		PasswordHash: string(hashed),
	}
	if err := store.CreateUser(ctx, user); err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	return user
}

func loadLegacyAssets(path string) (map[string]legacyAsset, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]legacyAsset{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var payload map[string]legacyAsset
	if err := json.NewDecoder(file).Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func loadLegacySessions(path string) (map[string]legacySession, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]legacySession{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var payload map[string]legacySession
	if err := json.NewDecoder(file).Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func migrateAsset(ctx context.Context, pool *pgxpool.Pool, userID, dataDir string, legacy legacyAsset) error {
	createdAt := parseTimestamp(legacy.CreatedAt)
	updatedAt := parseTimestamp(legacy.UpdatedAt)

	oldPath := legacy.RootPath
	if oldPath == "" {
		oldPath = filepath.Join(dataDir, "uploads", legacy.ID)
	}
	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		// try relative path
		candidate := filepath.Join(dataDir, "uploads", legacy.ID)
		if _, err := os.Stat(candidate); err == nil {
			oldPath = candidate
		}
	}

	newPath := filepath.Join(dataDir, "uploads", userID, legacy.ID)
	if err := ensureLegacyFiles(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move files: %w", err)
	}

	_, err := pool.Exec(
		ctx,
		`INSERT INTO assets (id, user_id, name, root_path, size_bytes, file_count, source_name, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 ON CONFLICT (id) DO UPDATE SET
			user_id=EXCLUDED.user_id,
			name=EXCLUDED.name,
			root_path=EXCLUDED.root_path,
			size_bytes=EXCLUDED.size_bytes,
			file_count=EXCLUDED.file_count,
			source_name=EXCLUDED.source_name,
			updated_at=EXCLUDED.updated_at`,
		legacy.ID,
		userID,
		legacy.Name,
		newPath,
		legacy.SizeBytes,
		legacy.FileCount,
		legacy.SourceName,
		createdAt,
		updatedAt,
	)
	return err
}

func migrateSession(ctx context.Context, pool *pgxpool.Pool, userID string, legacy legacySession) error {
	createdAt := parseTimestamp(legacy.CreatedAt)
	updatedAt := parseTimestamp(legacy.UpdatedAt)
	_, err := pool.Exec(
		ctx,
		`INSERT INTO typing_sessions (id, user_id, asset_id, rel_path, cursor, errors, duration_seconds, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 ON CONFLICT (id) DO UPDATE SET
			user_id=EXCLUDED.user_id,
			asset_id=EXCLUDED.asset_id,
			rel_path=EXCLUDED.rel_path,
			cursor=EXCLUDED.cursor,
			errors=EXCLUDED.errors,
			duration_seconds=EXCLUDED.duration_seconds,
			updated_at=EXCLUDED.updated_at`,
		legacy.ID,
		userID,
		legacy.AssetID,
		legacy.RelPath,
		legacy.Cursor,
		legacy.Errors,
		legacy.DurationSeconds,
		createdAt,
		updatedAt,
	)
	return err
}

func ensureLegacyFiles(oldPath, newPath string) error {
	if oldPath == "" {
		return os.MkdirAll(newPath, 0o755)
	}
	if oldPath == newPath {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(newPath, 0o755)
	}
	if err := os.RemoveAll(newPath); err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		// fallback to copy
		if err := copyDir(oldPath, newPath); err != nil {
			return fmt.Errorf("rename failed: %w", err)
		}
		if err := os.RemoveAll(oldPath); err != nil {
			log.Printf("warning: failed to clean up old path %s: %v", oldPath, err)
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func parseTimestamp(value string) time.Time {
	if value == "" {
		return time.Now().UTC()
	}
	if ts, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return ts
	}
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts
	}
	return time.Now().UTC()
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
