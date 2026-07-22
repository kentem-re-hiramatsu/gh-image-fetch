// Package config stores tool settings such as the default download
// directory. It never stores credentials; authentication is always
// delegated to the gh CLI via go-gh.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvDir overrides the configured default download directory when set.
const EnvDir = "GH_IMAGE_FETCH_DIR"

type file struct {
	Dir string `json:"dir,omitempty"`
}

// Path returns the location of the config file
// (e.g. %AppData%\gh-image-fetch\config.json on Windows).
func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine user config directory: %w", err)
	}
	return filepath.Join(base, "gh-image-fetch", "config.json"), nil
}

// DefaultDir returns the effective default download directory:
// the GH_IMAGE_FETCH_DIR environment variable if set, otherwise the value
// from the config file. Returns "" when neither is configured.
func DefaultDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv(EnvDir)); dir != "" {
		if err := validateDir(dir); err != nil {
			return "", fmt.Errorf("invalid %s: %w", EnvDir, err)
		}
		return dir, nil
	}
	cfg, err := load()
	if err != nil {
		return "", err
	}
	return cfg.Dir, nil
}

// SetDir validates dir, converts it to an absolute path, and persists it.
func SetDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if err := validateDir(dir); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve %q: %w", dir, err)
	}
	cfg, err := load()
	if err != nil {
		return "", err
	}
	cfg.Dir = abs

	path, err := Path()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("cannot write config file: %w", err)
	}
	return abs, nil
}

func load() (file, error) {
	var cfg file
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("cannot read config file %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config file %q is corrupted: %w", path, err)
	}
	return cfg, nil
}

func validateDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory path is empty")
	}
	for _, part := range strings.FieldsFunc(dir, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return fmt.Errorf("directory %q must not contain \"..\" path elements", dir)
		}
	}
	return nil
}
