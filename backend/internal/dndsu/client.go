// Package dndsu crawls dnd.su (2014 edition) and next.dnd.su (2024 edition)
// for monster statblocks. It is deliberately polite (rate-limited, custom
// User-Agent) and defensive: a parse failure on one field never aborts the
// whole card, and the raw HTML is always preserved alongside whatever
// parsed successfully.
package dndsu

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	userAgent  string
	minDelay   time.Duration
	lastReq    time.Time
}

func NewClient(userAgent string, minDelay time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 20 * time.Second},
		userAgent:  userAgent,
		minDelay:   minDelay,
	}
}

// Get fetches a URL, honoring the configured rate limit, and returns the
// raw response body.
func (c *Client) Get(ctx context.Context, url string) (string, error) {
	if wait := c.minDelay - time.Since(c.lastReq); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", "ru,en;q=0.8")

	resp, err := c.httpClient.Do(req)
	c.lastReq = time.Now()
	if err != nil {
		return "", fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get %s: status %d", url, resp.StatusCode)
	}
	return string(body), nil
}
