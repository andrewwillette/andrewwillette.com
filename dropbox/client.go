// Package dropbox is a minimal client for the pieces of the Dropbox API this
// project needs: resolving a shared link to a stable file ID, checking whether
// a file still exists, listing a folder, and creating/fetching a shared link.
package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
)

const (
	apiBaseURL    = "https://api.dropboxapi.com/2"
	oauthTokenURL = "https://api.dropboxapi.com/oauth2/token"
)

// Client is a Dropbox API client that keeps itself authenticated using a
// long-lived refresh token, exchanging it for short-lived access tokens as needed.
type Client struct {
	appKey       string
	appSecret    string
	refreshToken string
	httpClient   *http.Client

	mu                sync.Mutex
	accessToken       string
	accessTokenExpiry time.Time
}

// NewClientFromConfig builds a Client from the app's configured Dropbox credentials.
// Returns nil if no refresh token is configured yet.
func NewClientFromConfig() *Client {
	if webCfg.C.DropboxRefreshToken == "" {
		return nil
	}
	return &Client{
		appKey:       webCfg.C.DropboxAppKey,
		appSecret:    webCfg.C.DropboxAppSecret,
		refreshToken: webCfg.C.DropboxRefreshToken,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// apiError mirrors the shape of Dropbox's structured error responses.
type apiError struct {
	ErrorSummary string          `json:"error_summary"`
	Error        json.RawMessage `json:"error"`
}

// Err is returned for non-2xx Dropbox API responses, carrying the raw error
// tag/body so callers can distinguish e.g. "not found" from other failures.
type Err struct {
	StatusCode int
	Summary    string
	Body       json.RawMessage
}

func (e *Err) Error() string {
	// Dropbox returns a JSON error_summary for most failures (409s especially),
	// but 400s are frequently a plain-text request-validation message instead —
	// fall back to the raw body so that text isn't silently dropped.
	if e.Summary != "" {
		return fmt.Sprintf("dropbox api error (status %d): %s", e.StatusCode, e.Summary)
	}
	return fmt.Sprintf("dropbox api error (status %d): %s", e.StatusCode, strings.TrimSpace(string(e.Body)))
}

// HasTag reports whether the structured error body contains ".tag": "<tag>"
// at the top level or within an "error" sub-object, the common shapes Dropbox uses.
func (e *Err) HasTag(tag string) bool {
	return strings.Contains(string(e.Body), `".tag":"`+tag+`"`)
}

func (c *Client) accessTokenValue(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.accessTokenExpiry) {
		return c.accessToken, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", c.refreshToken)
	form.Set("client_id", c.appKey)
	form.Set("client_secret", c.appSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dropbox oauth token refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	c.accessToken = tokenResp.AccessToken
	// Refresh a little early to avoid racing the expiry.
	c.accessTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - 60*time.Second)
	return c.accessToken, nil
}

// rpc calls a Dropbox RPC-style endpoint (POST, JSON in, JSON out).
// On a non-2xx response it returns *Err with the raw response body.
func (c *Client) rpc(ctx context.Context, endpoint string, reqBody, respBody interface{}) error {
	token, err := c.accessTokenValue(ctx)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return err
		}
	} else {
		buf.WriteString("null")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+endpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr apiError
		_ = json.Unmarshal(body, &apiErr)
		return &Err{StatusCode: resp.StatusCode, Summary: apiErr.ErrorSummary, Body: body}
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return err
		}
	}
	return nil
}
