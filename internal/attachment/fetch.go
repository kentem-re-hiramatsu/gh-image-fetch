package attachment

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

// maxRedirects bounds how many unauthenticated redirect hops we follow
// after the initial authenticated request.
const maxRedirects = 5

// Result is a successfully opened download stream.
// Callers must Close the Body.
type Result struct {
	Body        io.ReadCloser
	ContentType string
}

// Fetch requests the asset using gh's stored credentials.
//
// The Authorization header is sent only to the GitHub host itself. If GitHub
// redirects to external blob storage, the redirect is followed with a plain
// unauthenticated client so the token never leaks to third-party hosts
// (signed redirect URLs carry their own short-lived credentials).
func Fetch(asset Asset) (*Result, error) {
	client, err := api.NewHTTPClient(api.ClientOptions{Host: asset.Host})
	if err != nil {
		return nil, fmt.Errorf("gh authentication is not configured: run `gh auth login` first (%w)", err)
	}
	// Handle redirects ourselves so the authenticated client never follows
	// a redirect off the GitHub host.
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Get(asset.URL())
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", asset.Host, err)
	}

	for hops := 0; isRedirect(resp.StatusCode); hops++ {
		location := resp.Header.Get("Location")
		resp.Body.Close()
		if hops >= maxRedirects {
			return nil, fmt.Errorf("too many redirects while downloading the asset")
		}
		if location == "" {
			return nil, fmt.Errorf("got HTTP %d without a Location header", resp.StatusCode)
		}
		resp, err = followRedirect(location)
		if err != nil {
			return nil, err
		}
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		return &Result{Body: resp.Body, ContentType: resp.Header.Get("Content-Type")}, nil
	case resp.StatusCode == http.StatusUnauthorized:
		resp.Body.Close()
		return nil, fmt.Errorf("authentication failed (HTTP 401): your gh credentials may be expired or revoked; run `gh auth status` and `gh auth login` to re-authenticate")
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests:
		defer resp.Body.Close()
		if isRateLimited(resp) {
			return nil, fmt.Errorf("rate limited by GitHub (HTTP %d): %s", resp.StatusCode, rateLimitHint(resp))
		}
		return nil, fmt.Errorf("access denied (HTTP 403): your account does not have permission to access this asset")
	case resp.StatusCode == http.StatusNotFound:
		resp.Body.Close()
		return nil, fmt.Errorf("asset not found (HTTP 404): the asset does not exist, or it belongs to a private repository whose attachments cannot be fetched with a token (try downloading it in a browser)")
	default:
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected response from GitHub: HTTP %d", resp.StatusCode)
	}
}

// followRedirect fetches a redirect target with a client that sends no
// GitHub credentials and follows any further redirects normally.
func followRedirect(location string) (*http.Response, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(location)
	if err != nil {
		return nil, fmt.Errorf("failed to download from storage: %w", err)
	}
	return resp, nil
}

func isRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

func isRateLimited(resp *http.Response) bool {
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	return resp.Header.Get("X-RateLimit-Remaining") == "0"
}

func rateLimitHint(resp *http.Response) string {
	if retry := resp.Header.Get("Retry-After"); retry != "" {
		return fmt.Sprintf("retry after %s seconds", retry)
	}
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		return "wait until the rate limit window resets, then retry"
	}
	return "wait a while and retry"
}
