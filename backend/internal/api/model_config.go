package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"codecopybook/internal/storage"
)

type modelConfigResponse struct {
	Provider            string `json:"provider"`
	Model               string `json:"model"`
	BaseURL             string `json:"baseUrl"`
	KeyHint             string `json:"keyHint"`
	HasKey              bool   `json:"hasKey"`
	SourceAccessEnabled bool   `json:"sourceAccessEnabled"`
	UsingDevelopmentKey bool   `json:"usingDevelopmentKey"`
}

func (s *Server) handleModelConfig(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.getModelConfig(user, w, r)
	case http.MethodPost:
		s.upsertModelConfig(user, w, r)
	case http.MethodDelete:
		s.deleteModelConfig(user, w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost, http.MethodDelete)
	}
}

func (s *Server) handleModelConfigTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}
	var payload struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		BaseURL  string `json:"baseUrl"`
		APIKey   string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	provider := normalizeProvider(payload.Provider)
	if provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}
	model := strings.TrimSpace(payload.Model)
	if model == "" {
		model = defaultModelForProvider(provider)
	}
	apiKey := strings.TrimSpace(payload.APIKey)
	if apiKey == "" {
		if existing, err := s.store.GetUserModelConfig(r.Context(), user.ID); err == nil && existing.EncryptedAPIKey != "" {
			var decryptErr error
			apiKey, decryptErr = s.decryptModelKey(existing.EncryptedAPIKey)
			if decryptErr != nil {
				writeError(w, http.StatusInternalServerError, "failed to decrypt saved api key")
				return
			}
		}
	}
	if apiKey == "" {
		apiKey = developmentModelAPIKey(provider)
	}
	if apiKey == "" {
		writeError(w, http.StatusBadRequest, "api key is required for connection test")
		return
	}
	cfg := &storage.UserModelConfig{
		UserID:              user.ID,
		Provider:            provider,
		Model:               model,
		BaseURL:             strings.TrimSpace(payload.BaseURL),
		SourceAccessEnabled: true,
	}
	raw, err := s.callConfiguredModel(r.Context(), cfg, apiKey, `Return strict JSON only: {"ok":true}`)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"provider": provider,
		"model":    model,
		"sample":   strings.TrimSpace(raw),
	})
}

func (s *Server) getModelConfig(user *storage.User, w http.ResponseWriter, r *http.Request) {
	cfg, err := s.store.GetUserModelConfig(r.Context(), user.ID)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "failed to load model config")
		return
	}
	if cfg == nil {
		writeJSON(w, http.StatusOK, s.defaultModelConfigResponse())
		return
	}
	writeJSON(w, http.StatusOK, modelConfigResponse{
		Provider:            cfg.Provider,
		Model:               cfg.Model,
		BaseURL:             cfg.BaseURL,
		KeyHint:             cfg.KeyHint,
		HasKey:              cfg.EncryptedAPIKey != "",
		SourceAccessEnabled: cfg.SourceAccessEnabled,
		UsingDevelopmentKey: cfg.EncryptedAPIKey == "" && developmentModelAPIKey(cfg.Provider) != "",
	})
}

func (s *Server) upsertModelConfig(user *storage.User, w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Provider            string `json:"provider"`
		Model               string `json:"model"`
		BaseURL             string `json:"baseUrl"`
		APIKey              string `json:"apiKey"`
		SourceAccessEnabled bool   `json:"sourceAccessEnabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	provider := normalizeProvider(payload.Provider)
	if provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}
	model := strings.TrimSpace(payload.Model)
	if model == "" {
		model = defaultModelForProvider(provider)
	}
	baseURL := strings.TrimSpace(payload.BaseURL)

	var encryptedKey, keyHint string
	if strings.TrimSpace(payload.APIKey) != "" {
		var err error
		encryptedKey, err = s.encryptModelKey(strings.TrimSpace(payload.APIKey))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to encrypt api key")
			return
		}
		keyHint = maskModelKey(payload.APIKey)
	} else if existing, err := s.store.GetUserModelConfig(r.Context(), user.ID); err == nil {
		encryptedKey = existing.EncryptedAPIKey
		keyHint = existing.KeyHint
	}

	cfg := &storage.UserModelConfig{
		UserID:              user.ID,
		Provider:            provider,
		Model:               model,
		BaseURL:             baseURL,
		EncryptedAPIKey:     encryptedKey,
		KeyHint:             keyHint,
		SourceAccessEnabled: payload.SourceAccessEnabled,
	}
	if err := s.store.UpsertUserModelConfig(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save model config")
		return
	}
	writeJSON(w, http.StatusOK, modelConfigResponse{
		Provider:            cfg.Provider,
		Model:               cfg.Model,
		BaseURL:             cfg.BaseURL,
		KeyHint:             cfg.KeyHint,
		HasKey:              cfg.EncryptedAPIKey != "",
		SourceAccessEnabled: cfg.SourceAccessEnabled,
		UsingDevelopmentKey: cfg.EncryptedAPIKey == "" && developmentModelAPIKey(cfg.Provider) != "",
	})
}

func (s *Server) deleteModelConfig(user *storage.User, w http.ResponseWriter, r *http.Request) {
	err := s.store.DeleteUserModelConfig(r.Context(), user.ID)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "failed to delete model config")
		return
	}
	writeJSON(w, http.StatusOK, s.defaultModelConfigResponse())
}

func (s *Server) defaultModelConfigResponse() modelConfigResponse {
	provider := "deepseek"
	return modelConfigResponse{
		Provider:            provider,
		Model:               defaultModelForProvider(provider),
		BaseURL:             "",
		HasKey:              false,
		SourceAccessEnabled: true,
		UsingDevelopmentKey: developmentModelAPIKey(provider) != "",
	}
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "deepseek":
		return "deepseek"
	case "openai", "openai-compatible", "gpt":
		return "openai-compatible"
	case "anthropic", "claude":
		return "anthropic"
	default:
		return ""
	}
}

func defaultModelForProvider(provider string) string {
	switch provider {
	case "deepseek":
		return "deepseek-chat"
	case "anthropic":
		return "claude-3-5-sonnet-latest"
	default:
		return "gpt-4o-mini"
	}
}

func developmentModelAPIKey(provider string) string {
	switch provider {
	case "deepseek":
		return strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	case "anthropic":
		return strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	default:
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}
}

func (s *Server) encryptModelKey(plain string) (string, error) {
	secret := strings.TrimSpace(os.Getenv("MODEL_CONFIG_SECRET"))
	if secret == "" {
		secret = string(s.authSecret)
	}
	if secret == "" {
		return "", errors.New("model config secret not configured")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *Server) decryptModelKey(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}
	secret := strings.TrimSpace(os.Getenv("MODEL_CONFIG_SECRET"))
	if secret == "" {
		secret = string(s.authSecret)
	}
	if secret == "" {
		return "", errors.New("model config secret not configured")
	}
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func maskModelKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 8 {
		return "****"
	}
	return trimmed[:3] + "..." + trimmed[len(trimmed)-4:]
}
