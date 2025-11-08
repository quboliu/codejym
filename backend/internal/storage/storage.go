package storage

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"crypto/rand"
)

// Asset represents an uploaded source bundle (file or extracted folder).
type Asset struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	RootPath   string    `json:"rootPath"`
	CreatedAt  time.Time `json:"createdAt"`
	SizeBytes  int64     `json:"sizeBytes"`
	FileCount  int       `json:"fileCount"`
	UpdatedAt  time.Time `json:"updatedAt"`
	SourceName string    `json:"sourceName"`
}

// Session tracks a typing practice run.
type Session struct {
	ID              string    `json:"id"`
	AssetID         string    `json:"assetId"`
	RelPath         string    `json:"relPath"`
	Cursor          int       `json:"cursor"`
	Errors          int       `json:"errors"`
	DurationSeconds int       `json:"durationSeconds"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Storage struct {
	rootDir      string
	uploadsDir   string
	assetsFile   string
	sessionsFile string

	mu       sync.RWMutex
	assets   map[string]*Asset
	sessions map[string]*Session
}

func New(root string) (*Storage, error) {
	if root == "" {
		root = "data"
	}
	uploadsDir := filepath.Join(root, "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return nil, err
	}
	s := &Storage{
		rootDir:      root,
		uploadsDir:   uploadsDir,
		assetsFile:   filepath.Join(root, "assets.json"),
		sessionsFile: filepath.Join(root, "sessions.json"),
		assets:       make(map[string]*Asset),
		sessions:     make(map[string]*Session),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) load() error {
	if err := s.loadAssets(); err != nil {
		return err
	}
	return s.loadSessions()
}

func (s *Storage) loadAssets() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	assets := make(map[string]*Asset)
	if data, err := os.ReadFile(s.assetsFile); err == nil {
		if err := json.Unmarshal(data, &assets); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.assets = assets
	return nil
}

func (s *Storage) loadSessions() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessions := make(map[string]*Session)
	if data, err := os.ReadFile(s.sessionsFile); err == nil {
		if err := json.Unmarshal(data, &sessions); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.sessions = sessions
	return nil
}

func (s *Storage) saveAssetsLocked() error {
	return writeJSONAtomic(s.assetsFile, s.assets)
}

func (s *Storage) saveSessionsLocked() error {
	return writeJSONAtomic(s.sessionsFile, s.sessions)
}

// AssetDir returns the filesystem directory for an asset id.
func (s *Storage) AssetDir(assetID string) string {
	return filepath.Join(s.uploadsDir, assetID)
}

func (s *Storage) ListAssets() []*Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Asset, 0, len(s.assets))
	for _, a := range s.assets {
		out = append(out, cloneAsset(a))
	}
	return out
}

func (s *Storage) GetAsset(id string) (*Asset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.assets[id]
	if !ok {
		return nil, false
	}
	return cloneAsset(a), true
}

func (s *Storage) RegisterAsset(asset *Asset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	asset.CreatedAt = now
	asset.UpdatedAt = now
	s.assets[asset.ID] = cloneAsset(asset)
	return s.saveAssetsLocked()
}

func (s *Storage) UpdateAssetStats(id string, fileCount int, sizeBytes int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.assets[id]
	if !ok {
		return ErrNotFound
	}
	a.FileCount = fileCount
	a.SizeBytes = sizeBytes
	a.UpdatedAt = time.Now().UTC()
	return s.saveAssetsLocked()
}

func (s *Storage) DeleteAsset(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assets[id]; !ok {
		return ErrNotFound
	}
	delete(s.assets, id)
	if err := s.saveAssetsLocked(); err != nil {
		return err
	}
	// Remove related sessions
	for sid, sess := range s.sessions {
		if sess.AssetID == id {
			delete(s.sessions, sid)
		}
	}
	return s.saveSessionsLocked()
}

func (s *Storage) CreateSession(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	session.CreatedAt = now
	session.UpdatedAt = now
	s.sessions[session.ID] = cloneSession(session)
	return s.saveSessionsLocked()
}

func (s *Storage) GetSession(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	return cloneSession(sess), true
}

func (s *Storage) UpdateSession(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[session.ID]; !ok {
		return ErrNotFound
	}
	session.UpdatedAt = time.Now().UTC()
	s.sessions[session.ID] = cloneSession(session)
	return s.saveSessionsLocked()
}

func (s *Storage) ListSessionsByAsset(assetID string) []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*Session{}
	for _, sess := range s.sessions {
		if sess.AssetID == assetID {
			out = append(out, cloneSession(sess))
		}
	}
	return out
}

func (s *Storage) UploadsDir() string {
	return s.uploadsDir
}

func cloneAsset(a *Asset) *Asset {
	if a == nil {
		return nil
	}
	cp := *a
	return &cp
}

func cloneSession(ses *Session) *Session {
	if ses == nil {
		return nil
	}
	cp := *ses
	return &cp
}

// RandomID creates a base32-ish identifier.
func RandomID() (string, error) {
	var b [10]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func writeJSONAtomic(path string, data any) error {
	tmp := path + ".tmp"
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, buf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

var ErrNotFound = errors.New("storage: not found")
