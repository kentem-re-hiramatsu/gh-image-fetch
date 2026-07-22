package attachment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveDestRejectsTraversal(t *testing.T) {
	for _, dest := range []string{
		"../evil.png",
		"..\\evil.png",
		"images/../../evil.png",
		"images\\..\\..\\evil.png",
		"..",
	} {
		if _, err := ResolveDest(dest, "id", "image/png"); err == nil {
			t.Errorf("ResolveDest(%q) succeeded, want traversal error", dest)
		}
	}
}

func TestResolveDestDirectory(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveDest(dir, "abc-123", "image/png")
	if err != nil {
		t.Fatalf("ResolveDest returned error: %v", err)
	}
	want := filepath.Join(dir, "abc-123.png")
	if got != want {
		t.Fatalf("ResolveDest = %q, want %q", got, want)
	}
}

func TestResolveDestFilePath(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "shot.png")
	got, err := ResolveDest(dest, "abc-123", "image/png")
	if err != nil {
		t.Fatalf("ResolveDest returned error: %v", err)
	}
	if got != dest {
		t.Fatalf("ResolveDest = %q, want %q", got, dest)
	}
}

func TestResolveDestMissingParent(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "no-such-dir", "shot.png")
	if _, err := ResolveDest(dest, "abc-123", "image/png"); err == nil {
		t.Fatal("ResolveDest succeeded, want missing-parent error")
	}
}

func TestExtensionFor(t *testing.T) {
	tests := map[string]string{
		"image/png":                ".png",
		"image/jpeg":               ".jpg",
		"image/gif; charset=binary": ".gif",
		"application/pdf":          ".pdf",
		"":                         ".bin",
		"garbage":                  ".bin",
		// Never trust extensions outside the allowlist: these must not
		// become .html / .exe / .js.
		"text/html":                ".bin",
		"application/x-msdownload": ".bin",
		"text/javascript":          ".bin",
	}
	for ct, want := range tests {
		if got := extensionFor(ct); got != want {
			t.Errorf("extensionFor(%q) = %q, want %q", ct, got, want)
		}
	}
}

func TestDefaultFileName(t *testing.T) {
	now := time.Date(2026, 7, 22, 9, 30, 15, 0, time.Local)
	got := DefaultFileName(now, "27ecac64-b73f-4ad7-ac47-a4071db12c76", "image/png")
	want := "20260722-093015-27ecac64.png"
	if got != want {
		t.Fatalf("DefaultFileName = %q, want %q", got, want)
	}
	if got := DefaultFileName(now, "short", ""); got != "20260722-093015-short.bin" {
		t.Fatalf("DefaultFileName short id = %q", got)
	}
}

func TestSaveRejectsOversizedDownload(t *testing.T) {
	old := maxDownloadBytes
	maxDownloadBytes = 10
	defer func() { maxDownloadBytes = old }()

	dir := t.TempDir()
	path := filepath.Join(dir, "big.bin")
	if _, err := Save(path, strings.NewReader("0123456789ABC")); err == nil {
		t.Fatal("Save succeeded, want size-limit error")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("oversized file was left on disk: %v", err)
	}

	if n, err := Save(path, strings.NewReader("0123456789")); err != nil || n != 10 {
		t.Fatalf("Save at exactly the limit = (%d, %v), want (10, nil)", n, err)
	}
}

func TestSaveOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(path, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := Save(path, strings.NewReader("new"))
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if n != 3 {
		t.Fatalf("Save wrote %d bytes, want 3", n)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("file content = %q, want %q", got, "new")
	}
}
