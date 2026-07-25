package dropbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AuthorizeURL builds the URL to visit in a browser to authorize this app
// and obtain a one-time code, using the offline-access token flow so the
// resulting code can be exchanged for a long-lived refresh token.
//
// force_reapprove is set because Dropbox only issues a refresh_token on a
// genuinely fresh consent — re-running this flow for an app/user pair that's
// already authorized would otherwise silently return an access_token only.
func AuthorizeURL(appKey string) string {
	v := url.Values{}
	v.Set("client_id", appKey)
	v.Set("response_type", "code")
	v.Set("token_access_type", "offline")
	v.Set("force_reapprove", "true")
	return "https://www.dropbox.com/oauth2/authorize?" + v.Encode()
}

// ExchangeCodeForRefreshToken exchanges a one-time authorization code (obtained
// by visiting the AuthorizeURL) for a long-lived refresh token.
func ExchangeCodeForRefreshToken(ctx context.Context, appKey, appSecret, code string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", appKey)
	form.Set("client_secret", appSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dropbox authorization code exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.RefreshToken == "" {
		return "", fmt.Errorf("dropbox response did not include a refresh_token; ensure the app requests offline access")
	}
	return tokenResp.RefreshToken, nil
}
