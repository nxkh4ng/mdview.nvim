package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestBaseDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	filePath := filepath.Join(dir, "test.md")
	if err := os.WriteFile(filePath, []byte("# hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	baseDirMu.Lock()
	currentBaseDir = dir
	baseDirMu.Unlock()

	t.Cleanup(func() {
		baseDirMu.Lock()
		currentBaseDir = ""
		baseDirMu.Unlock()
	})
	return filePath
}

func request(t *testing.T, url string) *httptest.ResponseRecorder {
	t.Helper()

	handler := handleLocalFiles()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, url, nil)
	handler(w, r)
	return w
}

func TestHandleLocalFiles_ValidFile(t *testing.T) {
	filePath := setupTestBaseDir(t)

	w := request(t, "/local/"+filePath)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "# hello" {
		t.Errorf("expected raw content, got:\n%s", w.Body.String())
	}
}

func TestHandleLocalFiles_PathTraversal(t *testing.T) {
	setupTestBaseDir(t)

	w := request(t, "/local/../../../etc/passwd")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for path traversal, got %d", w.Code)
	}
}

func TestHandleLocalFiles_NoBaseDir(t *testing.T) {
	w := request(t, "/local/some/file.md")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when no baseDir, got %d", w.Code)
	}
}

func TestHandleLocalFiles_MissingLocalPrefix(t *testing.T) {
	setupTestBaseDir(t)

	w := request(t, "/notlocal/file.md")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for wrong prefix, got %d", w.Code)
	}
}

func TestHandleLocalFiles_NonExistentFile(t *testing.T) {
	setupTestBaseDir(t)

	w := request(t, "/local/nonexistent.md")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent file, got %d", w.Code)
	}
}

func TestHandleLocalFiles_DotPath(t *testing.T) {
	setupTestBaseDir(t)

	w := request(t, "/local/./././etc/passwd")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for dot path traversal, got %d", w.Code)
	}
}

func TestHandleLocalFiles_AbsolutePathTraversal(t *testing.T) {
	filePath := setupTestBaseDir(t)
	baseDir := filepath.Dir(filePath)

	fakeBase := baseDir + "-extra"
	url := "/local/" + fakeBase + "/test.md"
	w := request(t, url)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for path traversal, got %d (url=%s)", w.Code, url)
	}
}

func TestHandleLocalFiles_PrefixCheckBypass(t *testing.T) {
	filePath := setupTestBaseDir(t)
	baseDir := filepath.Dir(filePath)

	fakeDir := baseDir + "-secret"
	os.MkdirAll(fakeDir, 0755)
	fakeFile := filepath.Join(fakeDir, "sneaky.md")
	os.WriteFile(fakeFile, []byte("leaked"), 0644)

	url := "/local" + fakeDir + "/sneaky.md"

	w := request(t, url)

	if w.Code == http.StatusOK {
		t.Errorf("file served from outside baseDir! body=%q", w.Body.String())
	}
}
