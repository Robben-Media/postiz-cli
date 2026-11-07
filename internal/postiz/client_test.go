package postiz

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUploadFileSelectedPath(t *testing.T) {
	for _, name := range []string{"regular", "symlink", "search-only", "parent-symlink-dotdot"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()

			target := filepath.Join(dir, "media.txt")
			if err := os.WriteFile(target, []byte("selected media"), 0o600); err != nil {
				t.Fatal(err)
			}

			selected := target
			if name == "symlink" {
				selected = filepath.Join(t.TempDir(), "alias.txt")
				if err := os.Symlink(target, selected); err != nil {
					t.Fatal(err)
				}
			}

			if name == "parent-symlink-dotdot" {
				realDir := filepath.Join(dir, "real")

				child := filepath.Join(realDir, "child")
				if err := os.MkdirAll(child, 0o700); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(target, []byte("wrong media"), 0o600); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(filepath.Join(realDir, "media.txt"), []byte("selected media"), 0o600); err != nil {
					t.Fatal(err)
				}

				if err := os.Symlink(child, filepath.Join(dir, "parent-alias")); err != nil {
					t.Fatal(err)
				}
				// Joining would clean away the filesystem-sensitive parent traversal.
				selected = dir + "/parent-alias/../media.txt"
			}

			if name == "search-only" {
				if runtime.GOOS == "windows" {
					t.Skip("Unix directory permissions")
				}

				if os.Geteuid() == 0 {
					t.Skip("root bypasses directory permissions")
				}

				t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

				if err := os.Chmod(dir, 0o100); err != nil {
					t.Fatal(err)
				}
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/uploads" {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}

				if r.Header.Get("Authorization") != "Bearer test-key" {
					t.Error("missing upload credentials")
				}

				f, header, err := r.FormFile("file")
				if err != nil {
					t.Error(err)
					http.Error(w, "invalid multipart", http.StatusBadRequest)

					return
				}
				defer f.Close()

				content, err := io.ReadAll(f)
				if err != nil {
					t.Error(err)
				}

				if string(content) != "selected media" || header.Filename != filepath.Base(selected) {
					t.Errorf("uploaded %q as %q", content, header.Filename)
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"id":"uploaded","url":"https://example.test/media"}`)
			}))
			defer server.Close()

			t.Setenv("POSTIZ_BASE_URL", server.URL)

			got, err := NewClient("test-key").UploadFile(context.Background(), selected)
			if err != nil {
				t.Fatal(err)
			}

			if got.ID != "uploaded" {
				t.Fatalf("upload = %#v", got)
			}
		})
	}
}

func TestUploadFileMissing(t *testing.T) {
	got, err := NewClient("test-key").UploadFile(context.Background(), filepath.Join(t.TempDir(), "missing"))
	if err == nil || got != nil {
		t.Fatalf("missing file = %v, %v", got, err)
	}
}
