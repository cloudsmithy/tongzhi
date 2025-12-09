package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OIDCConfig holds OIDC configuration
type OIDCConfig struct {
	ProviderURL  string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// OIDCProvider represents an OIDC provider
type OIDCProvider struct {
	config           OIDCConfig
	discoveryDoc     *OIDCDiscoveryDocument
	discoveryMu      sync.RWMutex
	stateStore       map[string]time.Time
	stateMu          sync.RWMutex
}

// OIDCDiscoveryDocument represents the OIDC discovery document
type OIDCDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

// OIDCTokenResponse represents the OIDC token endpoint response
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfo represents user information from OIDC provider
type UserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// NewOIDCProvider creates a new OIDC provider
func NewOIDCProvider(config OIDCConfig) *OIDCProvider {
	return &OIDCProvider{
		config:     config,
		stateStore: make(map[string]time.Time),
	}
}

// GetAuthorizationURL returns the authorization URL for OIDC login
func (p *OIDCProvider) GetAuthorizationURL(state string) (string, error) {
	doc, err := p.getDiscoveryDocument()
	if err != nil {
		return "", fmt.Errorf("failed to get discovery document: %w", err)
	}

	// Store state for validation
	p.stateMu.Lock()
	p.stateStore[state] = time.Now().Add(10 * time.Minute)
	p.stateMu.Unlock()

	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", doc.AuthorizationEndpoint, params.Encode()), nil
}

// ValidateState validates and consumes a state parameter
func (p *OIDCProvider) ValidateState(state string) bool {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	expiry, exists := p.stateStore[state]
	if !exists {
		return false
	}

	delete(p.stateStore, state)

	return time.Now().Before(expiry)
}

// ExchangeCode exchanges an authorization code for tokens
func (p *OIDCProvider) ExchangeCode(code string) (*OIDCTokenResponse, error) {
	doc, err := p.getDiscoveryDocument()
	if err != nil {
		return nil, fmt.Errorf("failed to get discovery document: %w", err)
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	req, err := http.NewRequest("POST", doc.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp OIDCTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfoFromIDToken extracts user info from ID token (JWT)
func (p *OIDCProvider) GetUserInfoFromIDToken(idToken string) (*UserInfo, error) {
	// Split JWT: header.payload.signature
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID token format")
	}

	// Decode payload (base64url)
	payload := parts[1]
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Try standard encoding
		decoded, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ID token payload: %w", err)
		}
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	return &UserInfo{
		Sub:   claims.Sub,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}

// GetUserInfo retrieves user information using an access token
func (p *OIDCProvider) GetUserInfo(accessToken string) (*UserInfo, error) {
	doc, err := p.getDiscoveryDocument()
	if err != nil {
		return nil, fmt.Errorf("failed to get discovery document: %w", err)
	}

	req, err := http.NewRequest("GET", doc.UserinfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed: %s", string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// getDiscoveryDocument fetches and caches the OIDC discovery document
func (p *OIDCProvider) getDiscoveryDocument() (*OIDCDiscoveryDocument, error) {
	p.discoveryMu.RLock()
	if p.discoveryDoc != nil {
		doc := p.discoveryDoc
		p.discoveryMu.RUnlock()
		return doc, nil
	}
	p.discoveryMu.RUnlock()

	p.discoveryMu.Lock()
	defer p.discoveryMu.Unlock()

	// Double-check after acquiring write lock
	if p.discoveryDoc != nil {
		return p.discoveryDoc, nil
	}

	discoveryURL := strings.TrimSuffix(p.config.ProviderURL, "/") + "/.well-known/openid-configuration"
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery request failed: %s", string(body))
	}

	var doc OIDCDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode discovery document: %w", err)
	}

	p.discoveryDoc = &doc
	return &doc, nil
}

// IsConfigured returns true if OIDC is properly configured
func (p *OIDCProvider) IsConfigured() bool {
	return p.config.ProviderURL != "" && 
		p.config.ClientID != "" && 
		p.config.ClientSecret != ""
}
