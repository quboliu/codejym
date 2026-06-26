package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	"codecopybook/internal/storage"
)

type fillInGenerationInfo struct {
	method     string
	provider   string
	model      string
	intent     string
	difficulty string
}

type modelCandidateTemplate struct {
	Intent     string                `json:"intent"`
	Difficulty string                `json:"difficulty"`
	Blanks     []modelCandidateBlank `json:"blanks"`
	SelfScores any                   `json:"selfScores,omitempty"`
}

type modelCandidateBlank struct {
	StartOffset int     `json:"startOffset"`
	EndOffset   int     `json:"endOffset"`
	Kind        string  `json:"kind"`
	Hint        string  `json:"hint"`
	Rationale   string  `json:"rationale"`
	ValueScore  float64 `json:"valueScore"`
}

func (s *Server) generateFillInTemplate(r *http.Request, userID, content, language string, existing []*storage.FillInBlank) ([]*storage.FillInBlank, json.RawMessage, json.RawMessage, fillInGenerationInfo) {
	for attempt := 1; attempt <= 3; attempt++ {
		blanks, scores, audit, info, err := s.generateModelFillInTemplate(r.Context(), userID, content, language, existing, attempt)
		if err == nil && len(blanks) > 0 {
			return blanks, scores, audit, info
		}
		if err != nil {
			s.logger.Printf("fill-in model candidate attempt %d skipped/failed: %v", attempt, err)
			if strings.Contains(err.Error(), "model key unavailable") || strings.Contains(err.Error(), "source access disabled") {
				break
			}
		}
	}

	blanks, scores, audit := generateFallbackFillInTemplate(content, language, existing)
	return blanks, scores, audit, fillInGenerationInfo{
		method:   storage.FillInGenerationFallback,
		provider: "local",
		model:    "heuristic-v1",
	}
}

func (s *Server) generateModelFillInTemplate(ctx context.Context, userID, content, language string, existing []*storage.FillInBlank, attempt int) ([]*storage.FillInBlank, json.RawMessage, json.RawMessage, fillInGenerationInfo, error) {
	cfg, apiKey, err := s.effectiveModelConfig(ctx, userID)
	if err != nil {
		return nil, nil, nil, fillInGenerationInfo{}, err
	}
	if apiKey == "" {
		return nil, nil, nil, fillInGenerationInfo{}, fmt.Errorf("model key unavailable")
	}
	if !cfg.SourceAccessEnabled {
		return nil, nil, nil, fillInGenerationInfo{}, fmt.Errorf("model source access disabled")
	}

	prompt := buildFillInPrompt(content, language, existing, attempt)
	raw, err := s.callConfiguredModel(ctx, cfg, apiKey, prompt)
	if err != nil {
		return nil, nil, nil, fillInGenerationInfo{}, err
	}
	candidate, err := parseModelCandidate(raw)
	if err != nil {
		return nil, nil, nil, fillInGenerationInfo{}, err
	}
	blanks, err := validateModelCandidate(candidate, content, language, existing)
	if err != nil {
		return nil, nil, nil, fillInGenerationInfo{}, err
	}
	scores := json.RawMessage(`{"valueScore":0.8,"diversityScore":0.65,"difficultyScore":0.6,"source":"model"}`)
	auditBytes, _ := json.Marshal(map[string]any{
		"generationMethod": "model",
		"provider":         cfg.Provider,
		"model":            cfg.Model,
		"promptVersion":    "fill-in-v1",
		"candidateAttempt": attempt,
		"accepted":         true,
		"raw":              candidate,
	})
	return blanks, scores, auditBytes, fillInGenerationInfo{
		method:     storage.FillInGenerationModel,
		provider:   cfg.Provider,
		model:      cfg.Model,
		intent:     strings.TrimSpace(candidate.Intent),
		difficulty: normalizeDifficulty(candidate.Difficulty),
	}, nil
}

func (s *Server) effectiveModelConfig(ctx context.Context, userID string) (*storage.UserModelConfig, string, error) {
	cfg, err := s.store.GetUserModelConfig(ctx, userID)
	if err != nil && !strings.Contains(err.Error(), "not found") && err != storage.ErrNotFound {
		return nil, "", err
	}
	if cfg == nil {
		cfg = &storage.UserModelConfig{
			UserID:              userID,
			Provider:            "deepseek",
			Model:               defaultModelForProvider("deepseek"),
			SourceAccessEnabled: true,
		}
	}
	apiKey := ""
	if cfg.EncryptedAPIKey != "" {
		var err error
		apiKey, err = s.decryptModelKey(cfg.EncryptedAPIKey)
		if err != nil {
			return nil, "", err
		}
	} else {
		apiKey = developmentModelAPIKey(cfg.Provider)
	}
	return cfg, apiKey, nil
}

func buildFillInPrompt(content, language string, existing []*storage.FillInBlank, attempt int) string {
	const maxChars = 14000
	source := content
	if len(source) > maxChars {
		source = source[:maxChars]
	}
	summary := make([]map[string]any, 0, len(existing))
	for _, blank := range existing {
		summary = append(summary, map[string]any{
			"startOffset": blank.StartOffset,
			"endOffset":   blank.EndOffset,
			"kind":        blank.Kind,
			"lineStart":   blank.LineStart,
			"lineEnd":     blank.LineEnd,
		})
	}
	summaryJSON, _ := json.Marshal(summary)
	return fmt.Sprintf(`You generate reusable fill-in coding practice templates.

Return strict JSON only. Do not use markdown fences.

JSON schema:
{
  "intent": "short learning goal",
  "difficulty": "easy|medium|hard",
  "blanks": [
    {
      "startOffset": 0,
      "endOffset": 0,
      "kind": "identifier|call_argument|condition|field_access|literal|return_value|error_handling|comment|other",
      "hint": "short hint that does not reveal the answer",
      "rationale": "why this blank has practice value",
      "valueScore": 0.0
    }
  ],
  "selfScores": {
    "practiceValue": 0.0,
    "diversity": 0.0,
    "difficultyFit": 0.0
  }
}

Use character offsets into the source text. Choose continuous semantic spans only.
High-value blanks: meaningful names, conditions, API arguments, field chains, literals, return values, error handling, domain logic.
Low-value blanks to avoid: standalone keywords such as if/for/return/int, punctuation, whitespace, braces, boilerplate.
Create 3 to 5 blanks. Keep blanks non-overlapping. Avoid previous blank regions.
This is candidate attempt %d of 3. If previous attempts were rejected, choose a meaningfully different region and learning goal.

Language: %s
Existing blank summary: %s

Source:
%s`, attempt, language, string(summaryJSON), source)
}

func (s *Server) callConfiguredModel(ctx context.Context, cfg *storage.UserModelConfig, apiKey, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	switch cfg.Provider {
	case "anthropic":
		return callAnthropicModel(ctx, cfg, apiKey, prompt)
	default:
		return callOpenAICompatibleModel(ctx, cfg, apiKey, prompt)
	}
}

func callOpenAICompatibleModel(ctx context.Context, cfg *storage.UserModelConfig, apiKey, prompt string) (string, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		if cfg.Provider == "deepseek" {
			baseURL = "https://api.deepseek.com"
		} else {
			baseURL = "https://api.openai.com/v1"
		}
	}
	endpoint := baseURL + "/chat/completions"
	body := map[string]any{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You return strict JSON only."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("model request failed: %s %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("model returned empty content")
	}
	return decoded.Choices[0].Message.Content, nil
}

func callAnthropicModel(ctx context.Context, cfg *storage.UserModelConfig, apiKey, prompt string) (string, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	body := map[string]any{
		"model":      cfg.Model,
		"max_tokens": 1800,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("model request failed: %s %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	var decoded struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	for _, item := range decoded.Content {
		if strings.TrimSpace(item.Text) != "" {
			return item.Text, nil
		}
	}
	return "", fmt.Errorf("model returned empty content")
}

func parseModelCandidate(raw string) (*modelCandidateTemplate, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	var candidate modelCandidateTemplate
	if err := json.Unmarshal([]byte(clean), &candidate); err != nil {
		return nil, err
	}
	if len(candidate.Blanks) == 0 {
		return nil, fmt.Errorf("candidate has no blanks")
	}
	return &candidate, nil
}

func validateModelCandidate(candidate *modelCandidateTemplate, content, language string, existing []*storage.FillInBlank) ([]*storage.FillInBlank, error) {
	runes := []rune(content)
	existingStarts := map[int]bool{}
	for _, blank := range existing {
		existingStarts[blank.StartOffset] = true
	}
	modelBlanks := candidate.Blanks
	sort.Slice(modelBlanks, func(i, j int) bool { return modelBlanks[i].StartOffset < modelBlanks[j].StartOffset })
	selected := make([]fillInCandidateBlank, 0, len(modelBlanks))
	out := make([]*storage.FillInBlank, 0, len(modelBlanks))
	for _, blank := range modelBlanks {
		if len(out) >= 5 {
			break
		}
		if blank.StartOffset < 0 || blank.EndOffset > len(runes) || blank.StartOffset >= blank.EndOffset {
			return nil, fmt.Errorf("candidate blank range invalid")
		}
		if existingStarts[blank.StartOffset] {
			continue
		}
		cand := fillInCandidateBlank{start: blank.StartOffset, end: blank.EndOffset}
		if overlapsSelected(cand, selected) {
			return nil, fmt.Errorf("candidate blanks overlap")
		}
		answer := string(runes[blank.StartOffset:blank.EndOffset])
		if isLowValueAnswer(answer, language) || mostlyWhitespaceOrPunctuation(answer) {
			continue
		}
		lineStart, lineEnd := lineRangeForRuneSpan(runes, blank.StartOffset, blank.EndOffset)
		valueScore := blank.ValueScore
		if valueScore <= 0 {
			valueScore = 0.65
		}
		out = append(out, &storage.FillInBlank{
			Position:               len(out),
			StartOffset:            blank.StartOffset,
			EndOffset:              blank.EndOffset,
			Answer:                 answer,
			LineStart:              lineStart,
			LineEnd:                lineEnd,
			Kind:                   normalizeBlankKind(blank.Kind),
			ValueScore:             valueScore,
			DifficultyContribution: 0.5,
			Hint:                   strings.TrimSpace(blank.Hint),
			Rationale:              strings.TrimSpace(blank.Rationale),
		})
		selected = append(selected, cand)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("candidate had no acceptable blanks")
	}
	return out, nil
}

func normalizeDifficulty(difficulty string) string {
	switch strings.ToLower(strings.TrimSpace(difficulty)) {
	case storage.FillInDifficultyEasy:
		return storage.FillInDifficultyEasy
	case storage.FillInDifficultyHard:
		return storage.FillInDifficultyHard
	default:
		return storage.FillInDifficultyMedium
	}
}

func normalizeBlankKind(kind string) string {
	switch strings.TrimSpace(kind) {
	case "identifier", "call_argument", "condition", "field_access", "literal", "return_value", "error_handling", "comment":
		return kind
	default:
		return "other"
	}
}

func mostlyWhitespaceOrPunctuation(value string) bool {
	meaningful := 0
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			meaningful++
		}
	}
	return meaningful == 0
}
