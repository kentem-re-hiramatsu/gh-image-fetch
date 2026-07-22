package attachment

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// DefaultHost is the GitHub host used when the input does not specify one.
// GitHub Enterprise support can be added later by making this configurable.
const DefaultHost = "github.com"

const assetPathPrefix = "/user-attachments/assets/"

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Asset identifies a single user-attachments asset.
type Asset struct {
	Host string
	ID   string
}

// URL returns the download URL for the asset.
func (a Asset) URL() string {
	return fmt.Sprintf("https://%s%s%s", a.Host, assetPathPrefix, a.ID)
}

// Parse accepts either a full user-attachments URL
// (https://github.com/user-attachments/assets/<uuid>) or a bare UUID.
func Parse(input string) (Asset, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Asset{}, fmt.Errorf("asset URL or ID is empty")
	}

	if uuidRe.MatchString(input) {
		return Asset{Host: DefaultHost, ID: input}, nil
	}

	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return Asset{}, fmt.Errorf("invalid asset reference %q: expected a UUID or a URL like https://%s%s<uuid>", input, DefaultHost, assetPathPrefix)
	}
	if u.Scheme != "https" {
		return Asset{}, fmt.Errorf("unsupported URL scheme %q: only https is allowed", u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	if host != DefaultHost {
		return Asset{}, fmt.Errorf("unsupported host %q: only %s is supported for now", u.Hostname(), DefaultHost)
	}
	id, ok := strings.CutPrefix(u.EscapedPath(), assetPathPrefix)
	if !ok {
		return Asset{}, fmt.Errorf("unsupported URL path %q: expected %s<uuid>", u.EscapedPath(), assetPathPrefix)
	}
	id = strings.TrimSuffix(id, "/")
	if !uuidRe.MatchString(id) {
		return Asset{}, fmt.Errorf("invalid asset ID %q: expected a UUID", id)
	}
	return Asset{Host: host, ID: id}, nil
}
