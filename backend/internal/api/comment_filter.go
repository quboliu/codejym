package api

import (
	"regexp"
	"strings"
)

// FindCommentRanges 根据语言类型找出注释的位置范围（不删除注释）
func FindCommentRanges(content string, language string) []skipRange {
	switch strings.ToLower(language) {
	case "go", "golang":
		return findGoCommentRanges(content)
	case "javascript", "js", "typescript", "ts", "jsx", "tsx":
		return findJavaScriptCommentRanges(content)
	case "python", "py":
		return findPythonCommentRanges(content)
	case "java":
		return findJavaCommentRanges(content)
	case "c", "cpp", "c++", "cc", "cxx":
		return findCCommentRanges(content)
	case "rust", "rs":
		return findRustCommentRanges(content)
	case "ruby", "rb":
		return findRubyCommentRanges(content)
	case "shell", "bash", "sh":
		return findShellCommentRanges(content)
	case "sql":
		return findSQLCommentRanges(content)
	case "php":
		return findPHPCommentRanges(content)
	case "swift":
		return findSwiftCommentRanges(content)
	case "kotlin", "kt":
		return findKotlinCommentRanges(content)
	default:
		// 对于不支持的语言，返回空数组（无注释跳过）
		return []skipRange{}
	}
}

// CommentFilter 根据语言类型过滤注释
func CommentFilter(content string, language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		return filterGoComments(content)
	case "javascript", "js", "typescript", "ts", "jsx", "tsx":
		return filterJavaScriptComments(content)
	case "python", "py":
		return filterPythonComments(content)
	case "java":
		return filterJavaComments(content)
	case "c", "cpp", "c++", "cc", "cxx":
		return filterCComments(content)
	case "rust", "rs":
		return filterRustComments(content)
	case "ruby", "rb":
		return filterRubyComments(content)
	case "shell", "bash", "sh":
		return filterShellComments(content)
	case "sql":
		return filterSQLComments(content)
	case "php":
		return filterPHPComments(content)
	case "swift":
		return filterSwiftComments(content)
	case "kotlin", "kt":
		return filterKotlinComments(content)
	default:
		// 对于不支持的语言，返回原内容
		return content
	}
}

// filterGoComments 过滤 Go 语言注释
func filterGoComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 //
	content = removeSingleLineComments(content, `//`)
	return content
}

// filterJavaScriptComments 过滤 JavaScript/TypeScript 注释
func filterJavaScriptComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 //
	content = removeSingleLineComments(content, `//`)
	return content
}

// filterPythonComments 过滤 Python 注释
func filterPythonComments(content string) string {
	// 处理多行注释 """ """ 或 ''' '''
	content = removeMultiLineComments(content, `"""`, `"""`)
	content = removeMultiLineComments(content, `'''`, `'''`)
	// 处理单行注释 #
	content = removeSingleLineComments(content, `#`)
	return content
}

// filterJavaComments 过滤 Java 注释
func filterJavaComments(content string) string {
	// 处理多行注释 /* */ 和 JavaDoc /** */
	content = removeMultiLineComments(content, `\/\*\*?`, `\*\/`)
	// 处理单行注释 //
	content = removeSingleLineComments(content, `//`)
	return content
}

// filterCComments 过滤 C/C++ 注释
func filterCComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 //
	content = removeSingleLineComments(content, `//`)
	return content
}

// filterRustComments 过滤 Rust 注释
func filterRustComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 //
	content = removeSingleLineComments(content, `//`)
	return content
}

// filterRubyComments 过滤 Ruby 注释
func filterRubyComments(content string) string {
	// 处理多行注释 =begin =end
	content = removeMultiLineComments(content, `^=begin`, `^=end`)
	// 处理单行注释 #
	content = removeSingleLineComments(content, `#`)
	return content
}

// filterShellComments 过滤 Shell 脚本注释
func filterShellComments(content string) string {
	// 处理单行注释 #（保留 shebang）
	lines := strings.Split(content, "\n")
	var result []string
	for i, line := range lines {
		// 保留第一行的 shebang
		if i == 0 && strings.HasPrefix(strings.TrimSpace(line), "#!") {
			result = append(result, line)
			continue
		}
		// 移除注释
		if idx := strings.Index(line, "#"); idx >= 0 {
			trimmed := strings.TrimSpace(line[:idx])
			if trimmed != "" {
				result = append(result, line[:idx])
			}
		} else if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// filterSQLComments 过滤 SQL 注释
func filterSQLComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 --
	content = removeSingleLineComments(content, `--`)
	return content
}

// filterPHPComments 过滤 PHP 注释
func filterPHPComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 // 和 #
	content = removeSingleLineComments(content, `//`)
	content = removeSingleLineComments(content, `#`)
	return content
}

// filterSwiftComments 过滤 Swift 注释
func filterSwiftComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 //
	content = removeSingleLineComments(content, `//`)
	return content
}

// filterKotlinComments 过滤 Kotlin 注释
func filterKotlinComments(content string) string {
	// 处理多行注释 /* */
	content = removeMultiLineComments(content, `\/\*`, `\*\/`)
	// 处理单行注释 //
	content = removeSingleLineComments(content, `//`)
	return content
}

// removeMultiLineComments 移除多行注释
func removeMultiLineComments(content string, startPattern string, endPattern string) string {
	// 使用正则表达式匹配多行注释（包括换行符）
	pattern := startPattern + `[\s\S]*?` + endPattern
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, "")
}

// removeSingleLineComments 移除单行注释
func removeSingleLineComments(content string, commentPrefix string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		// 查找注释开始位置
		commentIdx := strings.Index(line, commentPrefix)
		if commentIdx >= 0 {
			// 保留注释前的代码
			beforeComment := strings.TrimSpace(line[:commentIdx])
			if beforeComment != "" {
				result = append(result, line[:commentIdx])
			}
		} else {
			// 没有注释，保留整行（如果不是空行）
			if strings.TrimSpace(line) != "" {
				result = append(result, line)
			}
		}
	}

	return strings.Join(result, "\n")
}

// ==================== 查找注释范围的函数 ====================

// findMultiLineCommentRanges 查找多行注释的范围
func findMultiLineCommentRanges(content string, startPattern string, endPattern string) []skipRange {
	pattern := startPattern + `[\s\S]*?` + endPattern
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringIndex(content, -1)
	
	ranges := make([]skipRange, 0, len(matches))
	for _, match := range matches {
		ranges = append(ranges, skipRange{
			Start: match[0],
			End:   match[1],
		})
	}
	return ranges
}

// findSingleLineCommentRanges 查找单行注释的范围
func findSingleLineCommentRanges(content string, commentPrefix string) []skipRange {
	ranges := []skipRange{}
	lines := strings.Split(content, "\n")
	pos := 0
	
	for i, line := range lines {
		commentIdx := strings.Index(line, commentPrefix)
		if commentIdx >= 0 {
			// 找到注释，从注释开始到行尾都是注释
			start := pos + commentIdx
			end := pos + len(line)
			ranges = append(ranges, skipRange{
				Start: start,
				End:   end,
			})
		}
		// 移动到下一行（+1 是换行符）
		pos += len(line)
		if i < len(lines)-1 {
			pos += 1 // 换行符
		}
	}
	return ranges
}

// mergeRanges 合并重叠的范围并排序
func mergeRanges(ranges []skipRange) []skipRange {
	if len(ranges) == 0 {
		return ranges
	}
	
	// 按start排序
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[i].Start > ranges[j].Start {
				ranges[i], ranges[j] = ranges[j], ranges[i]
			}
		}
	}
	
	// 合并重叠范围
	merged := []skipRange{ranges[0]}
	for i := 1; i < len(ranges); i++ {
		last := &merged[len(merged)-1]
		current := ranges[i]
		
		if current.Start <= last.End {
			// 重叠，合并
			if current.End > last.End {
				last.End = current.End
			}
		} else {
			// 不重叠，添加新范围
			merged = append(merged, current)
		}
	}
	
	return merged
}

// findGoCommentRanges 查找 Go 语言注释
func findGoCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "//")
	return mergeRanges(append(multiLine, singleLine...))
}

// findJavaScriptCommentRanges 查找 JavaScript/TypeScript 注释
func findJavaScriptCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "//")
	return mergeRanges(append(multiLine, singleLine...))
}

// findPythonCommentRanges 查找 Python 注释
func findPythonCommentRanges(content string) []skipRange {
	multiLine1 := findMultiLineCommentRanges(content, `"""`, `"""`)
	multiLine2 := findMultiLineCommentRanges(content, `'''`, `'''`)
	singleLine := findSingleLineCommentRanges(content, "#")
	return mergeRanges(append(append(multiLine1, multiLine2...), singleLine...))
}

// findJavaCommentRanges 查找 Java 注释
func findJavaCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*\*?`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "//")
	return mergeRanges(append(multiLine, singleLine...))
}

// findCCommentRanges 查找 C/C++ 注释
func findCCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "//")
	return mergeRanges(append(multiLine, singleLine...))
}

// findRustCommentRanges 查找 Rust 注释
func findRustCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "//")
	return mergeRanges(append(multiLine, singleLine...))
}

// findRubyCommentRanges 查找 Ruby 注释
func findRubyCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `^=begin`, `^=end`)
	singleLine := findSingleLineCommentRanges(content, "#")
	return mergeRanges(append(multiLine, singleLine...))
}

// findShellCommentRanges 查找 Shell 脚本注释
func findShellCommentRanges(content string) []skipRange {
	ranges := []skipRange{}
	lines := strings.Split(content, "\n")
	pos := 0
	
	for i, line := range lines {
		// 保留第一行的 shebang
		if i == 0 && strings.HasPrefix(strings.TrimSpace(line), "#!") {
			pos += len(line) + 1
			continue
		}
		
		commentIdx := strings.Index(line, "#")
		if commentIdx >= 0 {
			start := pos + commentIdx
			end := pos + len(line)
			ranges = append(ranges, skipRange{
				Start: start,
				End:   end,
			})
		}
		
		pos += len(line)
		if i < len(lines)-1 {
			pos += 1
		}
	}
	return ranges
}

// findSQLCommentRanges 查找 SQL 注释
func findSQLCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "--")
	return mergeRanges(append(multiLine, singleLine...))
}

// findPHPCommentRanges 查找 PHP 注释
func findPHPCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine1 := findSingleLineCommentRanges(content, "//")
	singleLine2 := findSingleLineCommentRanges(content, "#")
	return mergeRanges(append(append(multiLine, singleLine1...), singleLine2...))
}

// findSwiftCommentRanges 查找 Swift 注释
func findSwiftCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "//")
	return mergeRanges(append(multiLine, singleLine...))
}

// findKotlinCommentRanges 查找 Kotlin 注释
func findKotlinCommentRanges(content string) []skipRange {
	multiLine := findMultiLineCommentRanges(content, `\/\*`, `\*\/`)
	singleLine := findSingleLineCommentRanges(content, "//")
	return mergeRanges(append(multiLine, singleLine...))
}
