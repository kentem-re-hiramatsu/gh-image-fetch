package attachment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	}
	for ct, want := range tests {
		if got := extensionFor(ct); got != want {
			t.Errorf("extensionFor(%q) = %q, want %q", ct, got, want)
		}
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
