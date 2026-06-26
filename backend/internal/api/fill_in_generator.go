package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"unicode"

	"codecopybook/internal/storage"
)

const fillInTemplateLimit = storage.FillInTemplateLimit

type fillInCandidateBlank struct {
	start     int
	end       int
	kind      string
	value     float64
	hint      string
	rationale string
}

func sourceContentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func generateFallbackFillInTemplate(content, language string, existing []*storage.FillInBlank) ([]*storage.FillInBlank, json.RawMessage, json.RawMessage) {
	runes := []rune(content)
	if len(runes) == 0 {
		return nil, json.RawMessage(`{}`), json.RawMessage(`{"reason":"empty source"}`)
	}

	usedStarts := map[int]bool{}
	for _, b := range existing {
		usedStarts[b.StartOffset] = true
	}

	candidates := make([]fillInCandidateBlank, 0)
	candidates = append(candidates, conditionCandidates(runes, language)...)
	candidates = append(candidates, callArgumentCandidates(runes, language)...)
	candidates = append(candidates, literalCandidates(runes)...)
	candidates = append(candidates, identifierCandidates(runes, language)...)

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].value == candidates[j].value {
			return candidates[i].start < candidates[j].start
		}
		return candidates[i].value > candidates[j].value
	})

	const maxBlanks = 5
	selected := make([]fillInCandidateBlank, 0, maxBlanks)
	for _, cand := range candidates {
		if len(selected) >= maxBlanks {
			break
		}
		if cand.start < 0 || cand.end > len(runes) || cand.start >= cand.end {
			continue
		}
		if usedStarts[cand.start] {
			continue
		}
		answer := strings.TrimSpace(string(runes[cand.start:cand.end]))
		if answer == "" || isLowValueAnswer(answer, language) {
			continue
		}
		if overlapsSelected(cand, selected) {
			continue
		}
		selected = append(selected, cand)
	}

	if len(selected) == 0 {
		return nil, json.RawMessage(`{"valueScore":0,"diversityScore":0}`), json.RawMessage(`{"reason":"no acceptable fallback blanks"}`)
	}

	sort.Slice(selected, func(i, j int) bool { return selected[i].start < selected[j].start })
	blanks := make([]*storage.FillInBlank, 0, len(selected))
	for i, cand := range selected {
		lineStart, lineEnd := lineRangeForRuneSpan(runes, cand.start, cand.end)
		blanks = append(blanks, &storage.FillInBlank{
			Position:               i,
			StartOffset:            cand.start,
			EndOffset:              cand.end,
			Answer:                 string(runes[cand.start:cand.end]),
			LineStart:              lineStart,
			LineEnd:                lineEnd,
			Kind:                   cand.kind,
			ValueScore:             cand.value,
			DifficultyContribution: 0.5,
			Hint:                   cand.hint,
			Rationale:              cand.rationale,
		})
	}

	scores := json.RawMessage(`{"valueScore":0.72,"diversityScore":0.55,"difficultyScore":0.45,"source":"fallback"}`)
	audit := json.RawMessage(`{"generationMethod":"fallback","promptVersion":"","accepted":true,"reason":"local heuristic template"}`)
	return blanks, scores, audit
}

func conditionCandidates(runes []rune, language string) []fillInCandidateBlank {
	var out []fillInCandidateBlank
	lineStart := 0
	for i := 0; i <= len(runes); i++ {
		if i < len(runes) && runes[i] != '\n' {
			continue
		}
		line := string(runes[lineStart:i])
		trimmed := strings.TrimSpace(line)
		leading := len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
		absoluteLineStart := lineStart + leading
		for _, prefix := range []string{"if ", "else if ", "for ", "while "} {
			if !strings.HasPrefix(trimmed, prefix) {
				continue
			}
			exprStart := absoluteLineStart + len([]rune(prefix))
			if strings.HasPrefix(trimmed, "else if ") {
				exprStart = absoluteLineStart + len([]rune("else if "))
			}
			exprEnd := i
			for exprEnd > exprStart && unicode.IsSpace(runes[exprEnd-1]) {
				exprEnd--
			}
			for exprEnd > exprStart {
				last := runes[exprEnd-1]
				if last == '{' || last == ':' {
					exprEnd--
					continue
				}
				break
			}
			if exprEnd-exprStart >= 5 && exprEnd-exprStart <= 100 {
				out = append(out, fillInCandidateBlank{
					start:     exprStart,
					end:       exprEnd,
					kind:      "condition",
					value:     0.92,
					hint:      "补全条件表达式",
					rationale: "条件表达式比单独关键字更能训练控制流理解",
				})
			}
		}
		lineStart = i + 1
	}
	return out
}

func callArgumentCandidates(runes []rune, language string) []fillInCandidateBlank {
	var out []fillInCandidateBlank
	for i := 0; i < len(runes); i++ {
		if runes[i] != '(' || i == 0 {
			continue
		}
		nameEnd := i
		nameStart := nameEnd - 1
		for nameStart >= 0 && (unicode.IsLetter(runes[nameStart]) || unicode.IsDigit(runes[nameStart]) || runes[nameStart] == '_' || runes[nameStart] == '.') {
			nameStart--
		}
		nameStart++
		if nameStart >= nameEnd {
			continue
		}
		name := string(runes[nameStart:nameEnd])
		lastPart := name
		if idx := strings.LastIndex(lastPart, "."); idx >= 0 {
			lastPart = lastPart[idx+1:]
		}
		if isKeyword(lastPart, language) {
			continue
		}
		close := findClosingParen(runes, i)
		if close <= i+1 || close-i > 120 {
			continue
		}
		argStart := i + 1
		argEnd := close
		for argStart < argEnd && unicode.IsSpace(runes[argStart]) {
			argStart++
		}
		for argEnd > argStart && unicode.IsSpace(runes[argEnd-1]) {
			argEnd--
		}
		if argEnd-argStart >= 4 {
			out = append(out, fillInCandidateBlank{
				start:     argStart,
				end:       argEnd,
				kind:      "call_argument",
				value:     0.86,
				hint:      "补全调用参数",
				rationale: "调用参数能训练 API 使用和数据流理解",
			})
		}
	}
	return out
}

func literalCandidates(runes []rune) []fillInCandidateBlank {
	var out []fillInCandidateBlank
	for i := 0; i < len(runes); i++ {
		if runes[i] == '"' || runes[i] == '\'' || runes[i] == '`' {
			quote := runes[i]
			j := i + 1
			for j < len(runes) {
				if runes[j] == quote && (quote == '`' || runes[j-1] != '\\') {
					break
				}
				j++
			}
			if j < len(runes) {
				if j+1-i >= 4 && j+1-i <= 80 {
					out = append(out, fillInCandidateBlank{
						start:     i,
						end:       j + 1,
						kind:      "literal",
						value:     0.78,
						hint:      "补全关键字面量",
						rationale: "字面量常承载配置、状态或协议信息",
					})
				}
				i = j
			}
			continue
		}
		if unicode.IsDigit(runes[i]) {
			j := i + 1
			for j < len(runes) && (unicode.IsDigit(runes[j]) || runes[j] == '_' || runes[j] == '.') {
				j++
			}
			if j-i >= 2 {
				out = append(out, fillInCandidateBlank{
					start:     i,
					end:       j,
					kind:      "literal",
					value:     0.68,
					hint:      "补全数字常量",
					rationale: "数字常量通常体现边界、状态或配置",
				})
			}
			i = j - 1
		}
	}
	return out
}

func identifierCandidates(runes []rune, language string) []fillInCandidateBlank {
	var out []fillInCandidateBlank
	for i := 0; i < len(runes); i++ {
		if !isIdentifierStart(runes[i]) {
			continue
		}
		j := i + 1
		for j < len(runes) && isIdentifierPart(runes[j]) {
			j++
		}
		ident := string(runes[i:j])
		if len([]rune(ident)) >= 5 && !isKeyword(ident, language) && !isCommonLowValueIdentifier(ident) {
			out = append(out, fillInCandidateBlank{
				start:     i,
				end:       j,
				kind:      "identifier",
				value:     0.64,
				hint:      "补全有意义的命名",
				rationale: "较长命名能训练代码意图和领域词记忆",
			})
		}
		i = j - 1
	}
	return out
}

func findClosingParen(runes []rune, open int) int {
	depth := 0
	for i := open; i < len(runes); i++ {
		switch runes[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		case '\n':
			if depth == 1 {
				return -1
			}
		}
	}
	return -1
}

func overlapsSelected(cand fillInCandidateBlank, selected []fillInCandidateBlank) bool {
	for _, existing := range selected {
		if cand.start < existing.end && existing.start < cand.end {
			return true
		}
	}
	return false
}

func lineRangeForRuneSpan(runes []rune, start, end int) (int, int) {
	line := 1
	lineStart := 1
	lineEnd := 1
	for i, r := range runes {
		if i == start {
			lineStart = line
		}
		if i == end-1 {
			lineEnd = line
			break
		}
		if r == '\n' {
			line++
		}
	}
	return lineStart, lineEnd
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isLowValueAnswer(answer, language string) bool {
	trimmed := strings.TrimSpace(answer)
	if trimmed == "" || isKeyword(trimmed, language) {
		return true
	}
	allPunctuation := true
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			allPunctuation = false
			break
		}
	}
	return allPunctuation
}

func isCommonLowValueIdentifier(ident string) bool {
	switch strings.ToLower(ident) {
	case "string", "number", "boolean", "object", "array", "error", "true", "false", "none", "null", "undefined", "self", "this":
		return true
	default:
		return false
	}
}

func isKeyword(word, language string) bool {
	word = strings.TrimSpace(word)
	if word == "" {
		return false
	}
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "while": true, "return": true, "func": true, "function": true,
		"int": true, "string": true, "bool": true, "boolean": true, "var": true, "let": true, "const": true,
		"type": true, "struct": true, "interface": true, "class": true, "def": true, "fn": true, "pub": true,
		"impl": true, "trait": true, "match": true, "case": true, "switch": true, "import": true, "from": true,
		"package": true, "use": true, "mod": true, "nil": true, "null": true, "none": true, "true": true, "false": true,
	}
	return keywords[strings.ToLower(word)]
}
