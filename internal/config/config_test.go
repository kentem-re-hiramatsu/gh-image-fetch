package config

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateConfig points os.UserConfigDir at a temp directory for the test.
func isolateConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)          // Windows
	t.Setenv("XDG_CONFIG_HOME", dir)  // Linux
	t.Setenv("HOME", dir)             // macOS fallback
	t.Setenv(EnvDir, "")
	return dir
}

func TestDefaultDirUnset(t *testing.T) {
	isolateConfig(t)
	dir, err := DefaultDir()
	if err != nil {
		t.Fatalf("DefaultDir returned error: %v", err)
	}
	if dir != "" {
		t.Fatalf("DefaultDir = %q, want empty", dir)
	}
}

func TestSetDirAndLoad(t *testing.T) {
	base := isolateConfig(t)
	target := filepath.Join(base, "downloads")

	abs, err := SetDir(target)
	if err != nil {
		t.Fatalf("SetDir returned error: %v", err)
	}
	if abs != target {
		t.Fatalf("SetDir = %q, want %q", abs, target)
	}

	got, err := DefaultDir()
	if err != nil {
		t.Fatalf("DefaultDir returned error: %v", err)
	}
	if got != target {
		t.Fatalf("DefaultDir = %q, want %q", got, target)
	}

	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file was not written: %v", err)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	base := isolateConfig(t)
	if _, err := SetDir(filepath.Join(base, "from-file")); err != nil {
		t.Fatal(err)
	}
	envDir := filepath.Join(base, "from-env")
	t.Setenv(EnvDir, envDir)

	got, err := DefaultDir()
	if err != nil {
		t.Fatalf("DefaultDir returned error: %v", err)
	}
	if got != envDir {
		t.Fatalf("DefaultDir = %q, want env value %q", got, envDir)
	}
}

func TestSetDirRejectsTraversal(t *testing.T) {
	isolateConfig(t)
	for _, dir := range []string{"../evil", "..\\evil", "a/../../b", ""} {
		if _, err := SetDir(dir); err == nil {
			t.Errorf("SetDir(%q) succeeded, want error", dir)
		}
	}
}

func TestDefaultDirRejectsTraversalInEnv(t *testing.T) {
	isolateConfig(t)
	t.Setenv(EnvDir, "../evil")
	if _, err := DefaultDir(); err == nil {
		t.Fatal("DefaultDir succeeded with traversal in env, want error")
	}
}
