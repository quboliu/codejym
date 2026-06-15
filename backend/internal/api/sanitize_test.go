package api

import (
	"path"
	"strings"
	"testing"
)

// TestSanitizeZipPath 验证 zip 条目名清洗：正常路径保留，越界路径被折叠/拒绝。
func TestSanitizeZipPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain file", "main.go", "main.go"},
		{"nested", "pkg/util/helper.go", "pkg/util/helper.go"},
		{"leading dot-slash", "./main.go", "main.go"},
		{"single parent escape collapses", "../main.go", "main.go"},
		{"deep parent escape collapses", "../../../../etc/passwd", "etc/passwd"},
		{"cross-tenant attempt stays in tree", "../../../otherUser/otherAsset/secret.txt", "otherUser/otherAsset/secret.txt"},
		{"absolute path stripped", "/etc/passwd", "etc/passwd"},
		{"windows backslash escape", `..\..\evil.txt`, "evil.txt"},
		{"embedded parent collapses", "a/../../b.txt", "b.txt"},
		{"empty", "", ""},
		{"dot", ".", ""},
		{"dotdot", "..", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sanitizeZipPath(c.in); got != c.want {
				t.Errorf("sanitizeZipPath(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestSanitizeZipPathContainment 是安全核心断言：无论输入多恶意，
// path.Join(basePath, sanitized) 都必须留在 basePath 之内，不能逃到别的用户目录。
func TestSanitizeZipPathContainment(t *testing.T) {
	base := "uploads/userA/assetA"
	malicious := []string{
		"../../../userB/assetB/owned.go",
		"../../../../../../../../etc/passwd",
		"../" + strings.Repeat("../", 20) + "root.txt",
		`..\..\..\userB\assetB\owned.go`,
		"normal/../../../escape.txt",
	}
	for _, m := range malicious {
		rel := sanitizeZipPath(m)
		if rel == "" {
			continue // 被拒绝，安全
		}
		// 清洗后不得再含 ".."，也不得为绝对路径
		if rel == ".." || strings.HasPrefix(rel, "..") || strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
			t.Fatalf("sanitizeZipPath(%q) = %q still contains traversal", m, rel)
		}
		joined := path.Clean(path.Join(base, rel))
		if joined != base && !strings.HasPrefix(joined, base+"/") {
			t.Fatalf("zip entry %q escaped base: joined=%q (base=%q)", m, joined, base)
		}
	}
}
