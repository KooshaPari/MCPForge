package protocol

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDocumentUri(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantURI     string
		wantPath    string
		expectError bool
	}{
		{
			name:     "empty input",
			input:    "",
			wantURI:  "",
			wantPath: "",
		},
		{
			name:     "normal file URI",
			input:    "file:///var/tmp/project/main.go",
			wantURI:  "file:///var/tmp/project/main.go",
			wantPath: "/var/tmp/project/main.go",
		},
		{
			name:     "two-slash URI normalized to three-slash",
			input:    "file://c:/repo/main.go",
			wantURI:  "file:///C:/repo/main.go",
			wantPath: "C:/repo/main.go",
		},
		{
			name:     "encoded drive colon normalized",
			input:    "file:///c%3A/repo/main.go",
			wantURI:  "file:///C:/repo/main.go",
			wantPath: "C:/repo/main.go",
		},
		{
			name:     "invalid scheme",
			input:    "http://example.com/file.txt",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotURI, err := ParseDocumentUri(tc.input)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(gotURI) != tc.wantURI {
				t.Fatalf("URI mismatch: got %q, want %q", string(gotURI), tc.wantURI)
			}
			if gotURI.Path() != tc.wantPath {
				t.Fatalf("path mismatch: got %q, want %q", gotURI.Path(), tc.wantPath)
			}
		})
	}
}

func TestURIFromPath(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "project", "main.go")

	tests := []struct {
		name     string
		path     string
		wantPath string
		wantURI  string
	}{
		{
			name:     "absolute path",
			path:     p,
			wantPath: p,
		},
		{
			name:     "windows-style drive path",
			path:     "c:/repo/main.go",
			wantPath: "C:/repo/main.go",
			wantURI:  "file:///C:/repo/main.go",
		},
		{
			name:     "empty path",
			path:     "",
			wantURI:  "",
			wantPath: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotURI := URIFromPath(tc.path)
			if tc.wantURI != "" && string(gotURI) != tc.wantURI {
				t.Fatalf("URI mismatch: got %q, want %q", string(gotURI), tc.wantURI)
			}
			if tc.wantPath == "" {
				if string(gotURI) != "" {
					t.Fatalf("expected empty URI/path, got %q", string(gotURI))
				}
				return
			}

			uriPath := gotURI.Path()
			if got, want := strings.ReplaceAll(uriPath, "\\", "/"), strings.ReplaceAll(tc.wantPath, "\\", "/"); got != want {
				t.Fatalf("path mismatch: got %q, want %q", got, want)
			}
		})
	}
}
