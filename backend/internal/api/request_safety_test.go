package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCleanRelPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"plain", "src/main.go", "src/main.go", true},
		{"clean dot", "./src/main.go", "src/main.go", true},
		{"windows separators", `src\main.go`, "src/main.go", true},
		{"empty", "", "", false},
		{"dot", ".", "", false},
		{"absolute", "/etc/passwd", "", false},
		{"parent segment", "src/../main.go", "", false},
		{"leading parent", "../main.go", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := cleanRelPath(c.in)
			if ok != c.ok {
				t.Fatalf("cleanRelPath(%q) ok = %v, want %v (got %q)", c.in, ok, c.ok, got)
			}
			if got != c.want {
				t.Fatalf("cleanRelPath(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestIsValidBaseName(t *testing.T) {
	valid := []string{"main.go", "README.md", "..hidden"}
	for _, name := range valid {
		if !isValidBaseName(name) {
			t.Fatalf("isValidBaseName(%q) = false, want true", name)
		}
	}
	invalid := []string{"", ".", "..", "dir/file.go", `dir\file.go`}
	for _, name := range invalid {
		if isValidBaseName(name) {
			t.Fatalf("isValidBaseName(%q) = true, want false", name)
		}
	}
}

func TestCORSAllowsAuthorizationHeader(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodOptions, "/api/assets", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
		t.Fatalf("Access-Control-Allow-Headers = %q", got)
	}
}
