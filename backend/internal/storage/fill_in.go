package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	FillInTemplateLimit = 8

	FillInTemplateStatusCandidate = "candidate"
	FillInTemplateStatusActive    = "active"
	FillInTemplateStatusRetired   = "retired"

	FillInDifficultyEasy   = "easy"
	FillInDifficultyMedium = "medium"
	FillInDifficultyHard   = "hard"

	FillInGenerationModel    = "model"
	FillInGenerationFallback = "fallback"

	FillInSessionInProgress = "in_progress"
	FillInSessionCompleted  = "completed"

	FillInOutcomeIndependent = "independent_completion"
	FillInOutcomeAssisted    = "assisted_completion"

	FillInBlankEmpty     = "empty"
	FillInBlankIncorrect = "incorrect"
	FillInBlankCorrect   = "correct"
	FillInBlankRevealed  = "revealed"
)

var ErrTemplateLimitExceeded = errors.New("storage: fill-in template limit exceeded")

type SourceFileVersion struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	AssetID     string    `json:"assetId"`
	RelPath     string    `json:"relPath"`
	ContentHash string    `json:"contentHash"`
	Language    string    `json:"language"`
	SizeBytes   int64     `json:"sizeBytes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type FillInTemplate struct {
	ID               string          `json:"id"`
	UserID           string          `json:"userId"`
	SourceVersionID  string          `json:"sourceVersionId"`
	AssetID          string          `json:"assetId"`
	RelPath          string          `json:"relPath"`
	ContentHash      string          `json:"contentHash"`
	Language         string          `json:"language"`
	Status           string          `json:"status"`
	Difficulty       string          `json:"difficulty"`
	Intent           string          `json:"intent"`
	GenerationMethod string          `json:"generationMethod"`
	Provider         string          `json:"provider"`
	Model            string          `json:"model"`
	ScoresJSON       json.RawMessage `json:"scores"`
	AuditJSON        json.RawMessage `json:"audit"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	Blanks           []*FillInBlank  `json:"blanks,omitempty"`
}

type FillInBlank struct {
	ID                     string  `json:"id"`
	TemplateID             string  `json:"templateId"`
	Position               int     `json:"position"`
	StartOffset            int     `json:"startOffset"`
	EndOffset              int     `json:"endOffset"`
	Answer                 string  `json:"answer,omitempty"`
	LineStart              int     `json:"lineStart"`
	LineEnd                int     `json:"lineEnd"`
	Kind                   string  `json:"kind"`
	ValueScore             float64 `json:"valueScore"`
	DifficultyContribution float64 `json:"difficultyContribution"`
	Hint                   string  `json:"hint,omitempty"`
	Rationale              string  `json:"rationale,omitempty"`
}

type FillInSession struct {
	ID                string    `json:"id"`
	UserID            string    `json:"userId"`
	TemplateID        string    `json:"templateId"`
	Status            string    `json:"status"`
	CompletionOutcome string    `json:"completionOutcome"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type FillInBlankAnswer struct {
	SessionID    string    `json:"sessionId"`
	BlankID      string    `json:"blankId"`
	CurrentInput string    `json:"currentInput"`
	Status       string    `json:"status"`
	ErrorCount   int       `json:"errorCount"`
	Revealed     bool      `json:"revealed"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type UserModelConfig struct {
	UserID              string    `json:"userId"`
	Provider            string    `json:"provider"`
	Model               string    `json:"model"`
	BaseURL             string    `json:"baseUrl"`
	EncryptedAPIKey     string    `json:"-"`
	KeyHint             string    `json:"keyHint"`
	SourceAccessEnabled bool      `json:"sourceAccessEnabled"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (s *Storage) UpsertSourceFileVersion(ctx context.Context, version *SourceFileVersion) error {
	if version.ID == "" {
		id, err := RandomID()
		if err != nil {
			return err
		}
		version.ID = id
	}
	return s.db.QueryRow(
		ctx,
		`INSERT INTO source_file_versions (id, user_id, asset_id, rel_path, content_hash, language, size_bytes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id, asset_id, rel_path, content_hash)
		 DO UPDATE SET language = EXCLUDED.language, size_bytes = EXCLUDED.size_bytes, updated_at = now()
		 RETURNING id, created_at, updated_at`,
		version.ID, version.UserID, version.AssetID, version.RelPath, version.ContentHash, version.Language, version.SizeBytes,
	).Scan(&version.ID, &version.CreatedAt, &version.UpdatedAt)
}

func (s *Storage) ListActiveFillInTemplates(ctx context.Context, userID, sourceVersionID string) ([]*FillInTemplate, error) {
	rows, err := s.db.Query(
		ctx,
		`SELECT id, user_id, source_version_id, asset_id, rel_path, content_hash, language, status, difficulty, intent,
		        generation_method, provider, model, scores_json, audit_json, created_at, updated_at
		 FROM fill_in_templates
		 WHERE user_id = $1 AND source_version_id = $2 AND status = $3
		 ORDER BY created_at ASC`,
		userID, sourceVersionID, FillInTemplateStatusActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var templates []*FillInTemplate
	for rows.Next() {
		t, err := scanFillInTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func (s *Storage) CreateFillInTemplate(ctx context.Context, template *FillInTemplate, blanks []*FillInBlank) error {
	if template.ID == "" {
		id, err := RandomID()
		if err != nil {
			return err
		}
		template.ID = id
	}
	if len(template.ScoresJSON) == 0 {
		template.ScoresJSON = json.RawMessage(`{}`)
	}
	if len(template.AuditJSON) == 0 {
		template.AuditJSON = json.RawMessage(`{}`)
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, template.SourceVersionID); err != nil {
		return err
	}
	var activeCount int
	if err := tx.QueryRow(
		ctx,
		`SELECT count(*) FROM fill_in_templates WHERE user_id = $1 AND source_version_id = $2 AND status = $3`,
		template.UserID, template.SourceVersionID, FillInTemplateStatusActive,
	).Scan(&activeCount); err != nil {
		return err
	}
	if template.Status == FillInTemplateStatusActive && activeCount >= FillInTemplateLimit {
		return ErrTemplateLimitExceeded
	}

	err = tx.QueryRow(
		ctx,
		`INSERT INTO fill_in_templates
		 (id, user_id, source_version_id, asset_id, rel_path, content_hash, language, status, difficulty, intent,
		  generation_method, provider, model, scores_json, audit_json)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		 RETURNING created_at, updated_at`,
		template.ID, template.UserID, template.SourceVersionID, template.AssetID, template.RelPath, template.ContentHash,
		template.Language, template.Status, template.Difficulty, template.Intent, template.GenerationMethod, template.Provider,
		template.Model, template.ScoresJSON, template.AuditJSON,
	).Scan(&template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		return err
	}
	for i, blank := range blanks {
		if blank.ID == "" {
			id, err := RandomID()
			if err != nil {
				return err
			}
			blank.ID = id
		}
		blank.TemplateID = template.ID
		blank.Position = i
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO fill_in_blanks
			 (id, template_id, position, start_offset, end_offset, answer, line_start, line_end, kind,
			  value_score, difficulty_contribution, hint, rationale)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			blank.ID, blank.TemplateID, blank.Position, blank.StartOffset, blank.EndOffset, blank.Answer, blank.LineStart,
			blank.LineEnd, blank.Kind, blank.ValueScore, blank.DifficultyContribution, blank.Hint, blank.Rationale,
		); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	template.Blanks = blanks
	return nil
}

func (s *Storage) GetFillInTemplate(ctx context.Context, userID, templateID string) (*FillInTemplate, error) {
	row := s.db.QueryRow(
		ctx,
		`SELECT id, user_id, source_version_id, asset_id, rel_path, content_hash, language, status, difficulty, intent,
		        generation_method, provider, model, scores_json, audit_json, created_at, updated_at
		 FROM fill_in_templates WHERE id = $1 AND user_id = $2`,
		templateID, userID,
	)
	t, err := scanFillInTemplate(row)
	if err != nil {
		return nil, err
	}
	blanks, err := s.ListFillInBlanks(ctx, templateID)
	if err != nil {
		return nil, err
	}
	t.Blanks = blanks
	return t, nil
}

func (s *Storage) ListFillInBlanks(ctx context.Context, templateID string) ([]*FillInBlank, error) {
	rows, err := s.db.Query(
		ctx,
		`SELECT id, template_id, position, start_offset, end_offset, answer, line_start, line_end, kind,
		        value_score, difficulty_contribution, hint, rationale
		 FROM fill_in_blanks WHERE template_id = $1 ORDER BY position ASC`,
		templateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blanks []*FillInBlank
	for rows.Next() {
		b := &FillInBlank{}
		if err := rows.Scan(&b.ID, &b.TemplateID, &b.Position, &b.StartOffset, &b.EndOffset, &b.Answer, &b.LineStart, &b.LineEnd, &b.Kind, &b.ValueScore, &b.DifficultyContribution, &b.Hint, &b.Rationale); err != nil {
			return nil, err
		}
		blanks = append(blanks, b)
	}
	return blanks, rows.Err()
}

func (s *Storage) GetFillInBlank(ctx context.Context, userID, templateID, blankID string) (*FillInBlank, error) {
	b := &FillInBlank{}
	err := s.db.QueryRow(
		ctx,
		`SELECT b.id, b.template_id, b.position, b.start_offset, b.end_offset, b.answer, b.line_start, b.line_end,
		        b.kind, b.value_score, b.difficulty_contribution, b.hint, b.rationale
		 FROM fill_in_blanks b
		 JOIN fill_in_templates t ON t.id = b.template_id
		 WHERE b.id = $1 AND b.template_id = $2 AND t.user_id = $3`,
		blankID, templateID, userID,
	).Scan(&b.ID, &b.TemplateID, &b.Position, &b.StartOffset, &b.EndOffset, &b.Answer, &b.LineStart, &b.LineEnd, &b.Kind, &b.ValueScore, &b.DifficultyContribution, &b.Hint, &b.Rationale)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *Storage) GetOrCreateFillInSession(ctx context.Context, userID, templateID string) (*FillInSession, error) {
	existing, err := s.GetFillInSessionByTemplate(ctx, userID, templateID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	sessionID, err := RandomID()
	if err != nil {
		return nil, err
	}
	sess := &FillInSession{ID: sessionID, UserID: userID, TemplateID: templateID, Status: FillInSessionInProgress}
	err = s.db.QueryRow(
		ctx,
		`INSERT INTO fill_in_sessions (id, user_id, template_id, status)
		 VALUES ($1,$2,$3,$4)
		 ON CONFLICT (user_id, template_id) DO UPDATE SET updated_at = fill_in_sessions.updated_at
		 RETURNING id, user_id, template_id, status, completion_outcome, created_at, updated_at`,
		sess.ID, sess.UserID, sess.TemplateID, sess.Status,
	).Scan(&sess.ID, &sess.UserID, &sess.TemplateID, &sess.Status, &sess.CompletionOutcome, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Storage) GetFillInSession(ctx context.Context, userID, sessionID string) (*FillInSession, error) {
	sess := &FillInSession{}
	err := s.db.QueryRow(
		ctx,
		`SELECT id, user_id, template_id, status, completion_outcome, created_at, updated_at
		 FROM fill_in_sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	).Scan(&sess.ID, &sess.UserID, &sess.TemplateID, &sess.Status, &sess.CompletionOutcome, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sess, nil
}

func (s *Storage) GetFillInSessionByTemplate(ctx context.Context, userID, templateID string) (*FillInSession, error) {
	sess := &FillInSession{}
	err := s.db.QueryRow(
		ctx,
		`SELECT id, user_id, template_id, status, completion_outcome, created_at, updated_at
		 FROM fill_in_sessions WHERE user_id = $1 AND template_id = $2`,
		userID, templateID,
	).Scan(&sess.ID, &sess.UserID, &sess.TemplateID, &sess.Status, &sess.CompletionOutcome, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sess, nil
}

func (s *Storage) ListFillInAnswers(ctx context.Context, sessionID string) (map[string]*FillInBlankAnswer, error) {
	rows, err := s.db.Query(
		ctx,
		`SELECT session_id, blank_id, current_input, status, error_count, revealed, updated_at
		 FROM fill_in_blank_answers WHERE session_id = $1`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	answers := map[string]*FillInBlankAnswer{}
	for rows.Next() {
		a := &FillInBlankAnswer{}
		if err := rows.Scan(&a.SessionID, &a.BlankID, &a.CurrentInput, &a.Status, &a.ErrorCount, &a.Revealed, &a.UpdatedAt); err != nil {
			return nil, err
		}
		answers[a.BlankID] = a
	}
	return answers, rows.Err()
}

func (s *Storage) UpsertFillInAnswer(ctx context.Context, answer *FillInBlankAnswer, incrementError bool) (*FillInBlankAnswer, error) {
	errIncrement := 0
	if incrementError {
		errIncrement = 1
	}
	out := &FillInBlankAnswer{}
	err := s.db.QueryRow(
		ctx,
		`INSERT INTO fill_in_blank_answers (session_id, blank_id, current_input, status, error_count, revealed)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (session_id, blank_id) DO UPDATE SET
		   current_input = EXCLUDED.current_input,
		   status = EXCLUDED.status,
		   error_count = fill_in_blank_answers.error_count + $7,
		   revealed = fill_in_blank_answers.revealed OR EXCLUDED.revealed,
		   updated_at = now()
		 RETURNING session_id, blank_id, current_input, status, error_count, revealed, updated_at`,
		answer.SessionID, answer.BlankID, answer.CurrentInput, answer.Status, errIncrement, answer.Revealed, errIncrement,
	).Scan(&out.SessionID, &out.BlankID, &out.CurrentInput, &out.Status, &out.ErrorCount, &out.Revealed, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Storage) ResetFillInSession(ctx context.Context, userID, sessionID string) (*FillInSession, error) {
	sess, err := s.GetFillInSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM fill_in_blank_answers WHERE session_id = $1`, sessionID); err != nil {
		return nil, err
	}
	err = tx.QueryRow(
		ctx,
		`UPDATE fill_in_sessions
		 SET status = $1, completion_outcome = '', updated_at = now()
		 WHERE id = $2 AND user_id = $3
		 RETURNING status, completion_outcome, updated_at`,
		FillInSessionInProgress, sessionID, userID,
	).Scan(&sess.Status, &sess.CompletionOutcome, &sess.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Storage) UpdateFillInSessionCompletion(ctx context.Context, userID, sessionID string, totalBlanks int) (*FillInSession, error) {
	answers, err := s.ListFillInAnswers(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	completedCount := 0
	revealed := false
	for _, answer := range answers {
		if answer.Status == FillInBlankCorrect || answer.Status == FillInBlankRevealed {
			completedCount++
		}
		if answer.Revealed || answer.Status == FillInBlankRevealed {
			revealed = true
		}
	}
	status := FillInSessionInProgress
	outcome := ""
	if totalBlanks > 0 && completedCount >= totalBlanks {
		status = FillInSessionCompleted
		if revealed {
			outcome = FillInOutcomeAssisted
		} else {
			outcome = FillInOutcomeIndependent
		}
	}
	sess := &FillInSession{}
	err = s.db.QueryRow(
		ctx,
		`UPDATE fill_in_sessions
		 SET status = $1, completion_outcome = $2, updated_at = now()
		 WHERE id = $3 AND user_id = $4
		 RETURNING id, user_id, template_id, status, completion_outcome, created_at, updated_at`,
		status, outcome, sessionID, userID,
	).Scan(&sess.ID, &sess.UserID, &sess.TemplateID, &sess.Status, &sess.CompletionOutcome, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sess, nil
}

func (s *Storage) UpsertUserModelConfig(ctx context.Context, cfg *UserModelConfig) error {
	return s.db.QueryRow(
		ctx,
		`INSERT INTO user_model_configs
		 (user_id, provider, model, base_url, encrypted_api_key, key_hint, source_access_enabled)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (user_id) DO UPDATE SET
		   provider = EXCLUDED.provider,
		   model = EXCLUDED.model,
		   base_url = EXCLUDED.base_url,
		   encrypted_api_key = EXCLUDED.encrypted_api_key,
		   key_hint = EXCLUDED.key_hint,
		   source_access_enabled = EXCLUDED.source_access_enabled,
		   updated_at = now()
		 RETURNING created_at, updated_at`,
		cfg.UserID, cfg.Provider, cfg.Model, cfg.BaseURL, cfg.EncryptedAPIKey, cfg.KeyHint, cfg.SourceAccessEnabled,
	).Scan(&cfg.CreatedAt, &cfg.UpdatedAt)
}

func (s *Storage) GetUserModelConfig(ctx context.Context, userID string) (*UserModelConfig, error) {
	cfg := &UserModelConfig{}
	err := s.db.QueryRow(
		ctx,
		`SELECT user_id, provider, model, base_url, encrypted_api_key, key_hint, source_access_enabled, created_at, updated_at
		 FROM user_model_configs WHERE user_id = $1`,
		userID,
	).Scan(&cfg.UserID, &cfg.Provider, &cfg.Model, &cfg.BaseURL, &cfg.EncryptedAPIKey, &cfg.KeyHint, &cfg.SourceAccessEnabled, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return cfg, nil
}

func (s *Storage) DeleteUserModelConfig(ctx context.Context, userID string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM user_model_configs WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanFillInTemplate(row pgx.Row) (*FillInTemplate, error) {
	t := &FillInTemplate{}
	err := row.Scan(&t.ID, &t.UserID, &t.SourceVersionID, &t.AssetID, &t.RelPath, &t.ContentHash, &t.Language, &t.Status, &t.Difficulty, &t.Intent, &t.GenerationMethod, &t.Provider, &t.Model, &t.ScoresJSON, &t.AuditJSON, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}
