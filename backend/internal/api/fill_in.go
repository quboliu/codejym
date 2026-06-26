package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"codecopybook/internal/storage"
)

type fillInEnterResponse struct {
	Template fillInTemplateDTO `json:"template"`
	Source   *fileContent      `json:"source"`
	Blanks   []fillInBlankDTO  `json:"blanks"`
	Session  fillInSessionDTO  `json:"session"`
}

type fillInTemplateDTO struct {
	ID               string `json:"id"`
	Difficulty       string `json:"difficulty"`
	Intent           string `json:"intent"`
	GenerationMethod string `json:"generationMethod"`
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	Status           string `json:"status"`
}

type fillInBlankDTO struct {
	ID           string `json:"id"`
	StartOffset  int    `json:"startOffset"`
	EndOffset    int    `json:"endOffset"`
	LineStart    int    `json:"lineStart"`
	LineEnd      int    `json:"lineEnd"`
	Kind         string `json:"kind"`
	Hint         string `json:"hint,omitempty"`
	Status       string `json:"status"`
	CurrentInput string `json:"currentInput"`
	ErrorCount   int    `json:"errorCount"`
	Revealed     bool   `json:"revealed"`
	Answer       string `json:"answer,omitempty"`
}

type fillInSessionDTO struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	CompletionOutcome string `json:"completionOutcome"`
	CompletedBlanks   int    `json:"completedBlanks"`
	TotalBlanks       int    `json:"totalBlanks"`
}

func (s *Server) handleFillInEnter(w http.ResponseWriter, r *http.Request) {
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
		AssetID string `json:"assetId"`
		Path    string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	resp, err := s.enterFillInPractice(r, user, payload.AssetID, payload.Path, "")
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleFillInSessionByID(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authorized")
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/fill-in/sessions/")
	segments := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		http.NotFound(w, r)
		return
	}
	sessionID := segments[0]
	if len(segments) == 1 {
		http.NotFound(w, r)
		return
	}
	switch {
	case len(segments) == 3 && segments[1] == "answers":
		s.handleFillInAnswer(user, sessionID, segments[2], w, r)
	case len(segments) == 4 && segments[1] == "blanks" && segments[3] == "reveal":
		s.handleFillInReveal(user, sessionID, segments[2], w, r)
	case len(segments) == 2 && segments[1] == "reset":
		s.handleFillInReset(user, sessionID, w, r)
	case len(segments) == 2 && segments[1] == "switch-template":
		s.handleFillInSwitchTemplate(user, sessionID, w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) enterFillInPractice(r *http.Request, user *storage.User, assetID, relPath, preferredTemplateID string) (*fillInEnterResponse, error) {
	if assetID == "" || relPath == "" {
		return nil, errBadRequest("assetId and path are required")
	}
	asset, err := s.store.GetAsset(r.Context(), user.ID, assetID)
	if err != nil {
		return nil, err
	}
	source, err := readAssetFileFromStorage(r.Context(), s.store.FileStorage(), asset.RootPath, relPath)
	if err != nil {
		return nil, err
	}
	contentHash := sourceContentHash(source.Content)
	version := &storage.SourceFileVersion{
		UserID:      user.ID,
		AssetID:     asset.ID,
		RelPath:     source.Path,
		ContentHash: contentHash,
		Language:    source.Language,
		SizeBytes:   int64(len([]byte(source.Content))),
	}
	if err := s.store.UpsertSourceFileVersion(r.Context(), version); err != nil {
		return nil, err
	}

	active, err := s.store.ListActiveFillInTemplates(r.Context(), user.ID, version.ID)
	if err != nil {
		return nil, err
	}
	if len(active) < fillInTemplateLimit {
		existingBlanks, err := s.collectTemplateBlanks(r, active)
		if err != nil {
			return nil, err
		}
		blanks, scores, audit, generation := s.generateFillInTemplate(r, user.ID, source.Content, source.Language, existingBlanks)
		if len(blanks) > 0 {
			template := &storage.FillInTemplate{
				UserID:           user.ID,
				SourceVersionID:  version.ID,
				AssetID:          asset.ID,
				RelPath:          source.Path,
				ContentHash:      contentHash,
				Language:         source.Language,
				Status:           storage.FillInTemplateStatusActive,
				Difficulty:       storage.FillInDifficultyMedium,
				Intent:           "练习有意义的命名、表达式和调用参数",
				GenerationMethod: generation.method,
				Provider:         generation.provider,
				Model:            generation.model,
				ScoresJSON:       scores,
				AuditJSON:        audit,
			}
			if generation.intent != "" {
				template.Intent = generation.intent
			}
			if generation.difficulty != "" {
				template.Difficulty = generation.difficulty
			}
			if err := s.store.CreateFillInTemplate(r.Context(), template, blanks); err != nil {
				if !errors.Is(err, storage.ErrTemplateLimitExceeded) {
					return nil, err
				}
				active, err = s.store.ListActiveFillInTemplates(r.Context(), user.ID, version.ID)
				if err != nil {
					return nil, err
				}
			} else {
				active = append(active, template)
			}
		}
	}
	if len(active) == 0 {
		return nil, errBadRequest("no fill-in template could be generated for this file")
	}

	template, err := s.selectFillInTemplate(r, user.ID, active, preferredTemplateID)
	if err != nil {
		return nil, err
	}
	template, err = s.store.GetFillInTemplate(r.Context(), user.ID, template.ID)
	if err != nil {
		return nil, err
	}
	session, err := s.store.GetOrCreateFillInSession(r.Context(), user.ID, template.ID)
	if err != nil {
		return nil, err
	}
	return s.buildFillInEnterResponse(r, source, template, session, false)
}

func (s *Server) handleFillInAnswer(user *storage.User, sessionID, blankID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var payload struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	session, template, blank, err := s.loadFillInSessionTemplateBlank(r, user.ID, sessionID, blankID)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	correct := payload.Input == blank.Answer
	status := storage.FillInBlankIncorrect
	if correct {
		status = storage.FillInBlankCorrect
	}
	answer, err := s.store.UpsertFillInAnswer(r.Context(), &storage.FillInBlankAnswer{
		SessionID:    session.ID,
		BlankID:      blank.ID,
		CurrentInput: payload.Input,
		Status:       status,
		Revealed:     false,
	}, !correct)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	updatedSession, err := s.store.UpdateFillInSessionCompletion(r.Context(), user.ID, session.ID, len(template.Blanks))
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"blankId":       blank.ID,
		"correct":       correct,
		"status":        answer.Status,
		"errorCount":    answer.ErrorCount,
		"sessionStatus": updatedSession.Status,
		"outcome":       updatedSession.CompletionOutcome,
	})
}

func (s *Server) handleFillInReveal(user *storage.User, sessionID, blankID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	session, template, blank, err := s.loadFillInSessionTemplateBlank(r, user.ID, sessionID, blankID)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	answer, err := s.store.UpsertFillInAnswer(r.Context(), &storage.FillInBlankAnswer{
		SessionID:    session.ID,
		BlankID:      blank.ID,
		CurrentInput: blank.Answer,
		Status:       storage.FillInBlankRevealed,
		Revealed:     true,
	}, false)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	updatedSession, err := s.store.UpdateFillInSessionCompletion(r.Context(), user.ID, session.ID, len(template.Blanks))
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"blankId":       blank.ID,
		"answer":        blank.Answer,
		"status":        answer.Status,
		"sessionStatus": updatedSession.Status,
		"outcome":       updatedSession.CompletionOutcome,
	})
}

func (s *Server) handleFillInReset(user *storage.User, sessionID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	session, err := s.store.ResetFillInSession(r.Context(), user.ID, sessionID)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	template, err := s.store.GetFillInTemplate(r.Context(), user.ID, session.TemplateID)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, fillInSessionDTO{
		ID:                session.ID,
		Status:            session.Status,
		CompletionOutcome: session.CompletionOutcome,
		CompletedBlanks:   0,
		TotalBlanks:       len(template.Blanks),
	})
}

func (s *Server) handleFillInSwitchTemplate(user *storage.User, sessionID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	session, err := s.store.GetFillInSession(r.Context(), user.ID, sessionID)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	template, err := s.store.GetFillInTemplate(r.Context(), user.ID, session.TemplateID)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	resp, err := s.enterFillInPractice(r, user, template.AssetID, template.RelPath, template.ID)
	if err != nil {
		s.writeFillInError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) collectTemplateBlanks(r *http.Request, templates []*storage.FillInTemplate) ([]*storage.FillInBlank, error) {
	var out []*storage.FillInBlank
	for _, template := range templates {
		blanks, err := s.store.ListFillInBlanks(r.Context(), template.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, blanks...)
	}
	return out, nil
}

func (s *Server) selectFillInTemplate(r *http.Request, userID string, templates []*storage.FillInTemplate, excludeTemplateID string) (*storage.FillInTemplate, error) {
	var fallback *storage.FillInTemplate
	for _, template := range templates {
		if template.ID == excludeTemplateID {
			continue
		}
		if fallback == nil {
			fallback = template
		}
		session, err := s.store.GetFillInSessionByTemplate(r.Context(), userID, template.ID)
		if errors.Is(err, storage.ErrNotFound) {
			return template, nil
		}
		if err != nil {
			return nil, err
		}
		if session.Status != storage.FillInSessionCompleted {
			return template, nil
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	return templates[0], nil
}

func (s *Server) buildFillInEnterResponse(r *http.Request, source *fileContent, template *storage.FillInTemplate, session *storage.FillInSession, revealAnswers bool) (*fillInEnterResponse, error) {
	answers, err := s.store.ListFillInAnswers(r.Context(), session.ID)
	if err != nil {
		return nil, err
	}
	blankDTOs := make([]fillInBlankDTO, 0, len(template.Blanks))
	completed := 0
	for _, blank := range template.Blanks {
		status := storage.FillInBlankEmpty
		currentInput := ""
		errorCount := 0
		revealed := false
		answer := ""
		if existing := answers[blank.ID]; existing != nil {
			status = existing.Status
			currentInput = existing.CurrentInput
			errorCount = existing.ErrorCount
			revealed = existing.Revealed
			if status == storage.FillInBlankCorrect || status == storage.FillInBlankRevealed {
				completed++
			}
			if revealAnswers || existing.Revealed {
				answer = blank.Answer
			}
		}
		blankDTOs = append(blankDTOs, fillInBlankDTO{
			ID:           blank.ID,
			StartOffset:  blank.StartOffset,
			EndOffset:    blank.EndOffset,
			LineStart:    blank.LineStart,
			LineEnd:      blank.LineEnd,
			Kind:         blank.Kind,
			Hint:         blank.Hint,
			Status:       status,
			CurrentInput: currentInput,
			ErrorCount:   errorCount,
			Revealed:     revealed,
			Answer:       answer,
		})
	}
	return &fillInEnterResponse{
		Template: fillInTemplateDTO{
			ID:               template.ID,
			Difficulty:       template.Difficulty,
			Intent:           template.Intent,
			GenerationMethod: template.GenerationMethod,
			Provider:         template.Provider,
			Model:            template.Model,
			Status:           template.Status,
		},
		Source: source,
		Blanks: blankDTOs,
		Session: fillInSessionDTO{
			ID:                session.ID,
			Status:            session.Status,
			CompletionOutcome: session.CompletionOutcome,
			CompletedBlanks:   completed,
			TotalBlanks:       len(template.Blanks),
		},
	}, nil
}

func (s *Server) loadFillInSessionTemplateBlank(r *http.Request, userID, sessionID, blankID string) (*storage.FillInSession, *storage.FillInTemplate, *storage.FillInBlank, error) {
	session, err := s.store.GetFillInSession(r.Context(), userID, sessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	template, err := s.store.GetFillInTemplate(r.Context(), userID, session.TemplateID)
	if err != nil {
		return nil, nil, nil, err
	}
	blank, err := s.store.GetFillInBlank(r.Context(), userID, template.ID, blankID)
	if err != nil {
		return nil, nil, nil, err
	}
	return session, template, blank, nil
}

type requestError struct {
	status int
	msg    string
}

func (e requestError) Error() string {
	return e.msg
}

func errBadRequest(msg string) error {
	return requestError{status: http.StatusBadRequest, msg: msg}
}

func (s *Server) writeFillInError(w http.ResponseWriter, err error) {
	var reqErr requestError
	if errors.As(err, &reqErr) {
		writeError(w, reqErr.status, reqErr.msg)
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "fill-in resource not found")
		return
	}
	if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "not found") {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}
