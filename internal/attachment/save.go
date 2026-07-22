package attachment

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// maxDownloadBytes caps a single download. GitHub user-attachments are at
// most ~100 MB, so anything larger is unexpected. A var so tests can
// override it.
var maxDownloadBytes int64 = 256 << 20

// preferredExts is the allowlist of content types whose extension we trust.
// Anything else is saved as .bin so a hostile Content-Type can never yield
// an executable or browser-runnable extension (.exe, .html, .js, ...).
var preferredExts = map[string]string{
	"image/png":     ".png",
	"image/jpeg":    ".jpg",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
	"video/mp4":     ".mp4",
	"video/quicktime": ".mov",
	"application/pdf": ".pdf",
	"text/plain":      ".txt",
}

// ResolveDest turns the user-supplied destination into a concrete file path.
// If dest is an existing directory, the file is named <assetID><ext> inside
// it, with ext guessed from contentType. Paths containing ".." are rejected.
func ResolveDest(dest, assetID, contentType string) (string, error) {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return "", fmt.Errorf("destination path is empty")
	}
	for _, part := range strings.FieldsFunc(dest, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return "", fmt.Errorf("destination %q must not contain \"..\" path elements", dest)
		}
	}

	info, err := os.Stat(dest)
	if err == nil && info.IsDir() {
		return filepath.Join(dest, assetID+extensionFor(contentType)), nil
	}
	if strings.HasSuffix(dest, "/") || strings.HasSuffix(dest, `\`) {
		return "", fmt.Errorf("destination directory %q does not exist", dest)
	}
	parent := filepath.Dir(dest)
	if _, err := os.Stat(parent); err != nil {
		return "", fmt.Errorf("parent directory %q does not exist", parent)
	}
	return dest, nil
}

// DefaultFileName names a file saved into the configured default directory:
// <yyyymmdd-hhmmss>-<first 8 chars of the asset ID><ext>.
func DefaultFileName(now time.Time, assetID, contentType string) string {
	short := assetID
	if len(short) > 8 {
		short = short[:8]
	}
	return now.Format("20060102-150405") + "-" + short + extensionFor(contentType)
}

// Save writes the stream to path, overwriting any existing file.
// Downloads larger than maxDownloadBytes are aborted and removed.
func Save(path string, body io.Reader) (int64, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("cannot create %q: %w", path, err)
	}
	n, err := io.Copy(f, io.LimitReader(body, maxDownloadBytes+1))
	if err == nil && n > maxDownloadBytes {
		err = fmt.Errorf("asset is larger than the %d MB limit", maxDownloadBytes>>20)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(path)
		return 0, fmt.Errorf("failed to write %q: %w", path, err)
	}
	return n, nil
}

func extensionFor(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		return ".bin"
	}
	if ext, ok := preferredExts[mediaType]; ok {
		return ext
	}
	return ".bin"
}
