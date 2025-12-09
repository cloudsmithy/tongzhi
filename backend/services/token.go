package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	// WeChatTokenURL is the URL to get access token from WeChat API
	WeChatTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
	// TokenBufferTime is the buffer time before token expiration to trigger refresh
	TokenBufferTime = 5 * time.Minute
)

// TokenResponse represents the response from WeChat token API
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// HTTPClient interface for making HTTP requests (allows mocking in tests)
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// TokenManager manages WeChat access tokens
type TokenManager struct {
	appID       string
	appSecret   string
	accessToken string
	expiresAt   time.Time
	mu          sync.RWMutex
	httpClient  HTTPClient
}

// NewTokenManager creates a new token manager
func NewTokenManager(appID, appSecret string) *TokenManager {
	return &TokenManager{
		appID:      appID,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewTokenManagerWithClient creates a new token manager with a custom HTTP client
func NewTokenManagerWithClient(appID, appSecret string, client HTTPClient) *TokenManager {
	return &TokenManager{
		appID:      appID,
		appSecret:  appSecret,
		httpClient: client,
	}
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetAccessToken() (string, error) {
	tm.mu.RLock()
	if tm.accessToken != "" && time.Now().Add(TokenBufferTime).Before(tm.expiresAt) {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	return tm.refreshToken()
}

// refreshToken fetches a new access token from WeChat API
func (tm *TokenManager) refreshToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock
	if tm.accessToken != "" && time.Now().Add(TokenBufferTime).Before(tm.expiresAt) {
		return tm.accessToken, nil
	}

	// Build the request URL
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s",
		WeChatTokenURL, tm.appID, tm.appSecret)

	resp, err := tm.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to request access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("WeChat API error: code=%d, msg=%s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	tm.accessToken = tokenResp.AccessToken
	// WeChat tokens typically expire in 7200 seconds (2 hours)
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return tm.accessToken, nil
}

// IsExpired checks if the current token is expired or will expire soon
func (tm *TokenManager) IsExpired() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.accessToken == "" || time.Now().Add(TokenBufferTime).After(tm.expiresAt)
}

// ForceRefresh forces a token refresh regardless of expiration status
func (tm *TokenManager) ForceRefresh() (string, error) {
	tm.mu.Lock()
	tm.accessToken = ""
	tm.expiresAt = time.Time{}
	tm.mu.Unlock()
	return tm.refreshToken()
}

// SetToken sets the token directly (useful for testing)
func (tm *TokenManager) SetToken(token string, expiresIn time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.accessToken = token
	tm.expiresAt = time.Now().Add(expiresIn)
}

// GetExpiresAt returns the expiration time of the current token
func (tm *TokenManager) GetExpiresAt() time.Time {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.expiresAt
}

// UpdateCredentials updates the app credentials and clears the cached token
func (tm *TokenManager) UpdateCredentials(appID, appSecret string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.appID = appID
	tm.appSecret = appSecret
	tm.accessToken = ""
	tm.expiresAt = time.Time{}
}
