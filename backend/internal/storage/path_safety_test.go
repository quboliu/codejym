package storage

import "testing"

func TestCleanStoragePath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"plain", "uploads/u/a/file.go", "uploads/u/a/file.go", true},
		{"clean dot", "uploads/u/a/./file.go", "uploads/u/a/file.go", true},
		{"windows separators", `uploads\u\a\file.go`, "uploads/u/a/file.go", true},
		{"empty", "", "", false},
		{"dot", ".", "", false},
		{"absolute", "/tmp/file.go", "", false},
		{"parent segment", "uploads/u/a/../file.go", "", false},
		{"leading parent", "../uploads/u/a/file.go", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := cleanStoragePath(c.in)
			if c.ok && err != nil {
				t.Fatalf("cleanStoragePath(%q) unexpected error: %v", c.in, err)
			}
			if !c.ok && err == nil {
				t.Fatalf("cleanStoragePath(%q) expected error, got %q", c.in, got)
			}
			if got != c.want {
				t.Fatalf("cleanStoragePath(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestCleanStoragePrefixKeepsTrailingSlash(t *testing.T) {
	got, err := cleanStoragePrefix("uploads/u/a/")
	if err != nil {
		t.Fatalf("cleanStoragePrefix returned error: %v", err)
	}
	if got != "uploads/u/a/" {
		t.Fatalf("cleanStoragePrefix = %q, want trailing slash preserved", got)
	}
}
