package api

import (
	"regexp"
	"strings"
)

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
