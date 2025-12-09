package handlers

import (
	"net/http"
	"time"

	"wechat-notification/config"
	"wechat-notification/services"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "session_id"
	StateCookieName   = "oauth_state"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	config         *config.Config
	oidcProvider   *services.OIDCProvider
	sessionManager *services.SessionManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config) *AuthHandler {
	oidcConfig := services.OIDCConfig{
		ProviderURL:  cfg.OIDC.ProviderURL,
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,
		RedirectURL:  cfg.OIDC.RedirectURL,
	}

	return &AuthHandler{
		config:         cfg,
		oidcProvider:   services.NewOIDCProvider(oidcConfig),
		sessionManager: services.NewSessionManager(24 * time.Hour),
	}
}

// NewAuthHandlerWithDeps creates an auth handler with injected dependencies (for testing)
func NewAuthHandlerWithDeps(cfg *config.Config, oidcProvider *services.OIDCProvider, sessionManager *services.SessionManager) *AuthHandler {
	return &AuthHandler{
		config:         cfg,
		oidcProvider:   oidcProvider,
		sessionManager: sessionManager,
	}
}

// Login redirects to OIDC provider
// GET /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	// Check if OIDC is configured
	if !h.oidcProvider.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OIDC provider not configured",
			"code":  "OIDC_NOT_CONFIGURED",
		})
		return
	}

	// Generate state for CSRF protection
	state, err := services.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate state",
			"code":  "STATE_GENERATION_FAILED",
		})
		return
	}

	// Get authorization URL
	authURL, err := h.oidcProvider.GetAuthorizationURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get authorization URL",
			"code":  "AUTH_URL_FAILED",
		})
		return
	}

	// Store state in cookie for validation
	c.SetCookie(StateCookieName, state, 600, "/", "", false, true)

	// Redirect to OIDC provider
	c.Redirect(http.StatusFound, authURL)
}

// Callback handles OIDC callback
// GET /auth/callback
func (h *AuthHandler) Callback(c *gin.Context) {
	// Check for error from OIDC provider
	if errParam := c.Query("error"); errParam != "" {
		errDesc := c.Query("error_description")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errDesc,
			"code":  errParam,
		})
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing authorization code",
			"code":  "MISSING_CODE",
		})
		return
	}

	// Validate state
	state := c.Query("state")
	storedState, err := c.Cookie(StateCookieName)
	if err != nil || state == "" || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid state parameter",
			"code":  "INVALID_STATE",
		})
		return
	}

	// Validate state with provider
	if !h.oidcProvider.ValidateState(state) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State validation failed",
			"code":  "STATE_VALIDATION_FAILED",
		})
		return
	}

	// Clear state cookie
	c.SetCookie(StateCookieName, "", -1, "/", "", false, true)

	// Exchange code for tokens
	tokenResp, err := h.oidcProvider.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to exchange authorization code",
			"code":  "TOKEN_EXCHANGE_FAILED",
		})
		return
	}

	// Get user info - try ID token first, then userinfo endpoint
	var userInfo *services.UserInfo
	if tokenResp.IDToken != "" {
		userInfo, err = h.oidcProvider.GetUserInfoFromIDToken(tokenResp.IDToken)
	}
	if userInfo == nil || err != nil {
		userInfo, err = h.oidcProvider.GetUserInfo(tokenResp.AccessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user information",
				"code":  "USERINFO_FAILED",
			})
			return
		}
	}

	// Create session
	session, err := h.sessionManager.CreateSession(userInfo.Sub, userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create session",
			"code":  "SESSION_CREATION_FAILED",
		})
		return
	}

	// Set session cookie
	c.SetCookie(SessionCookieName, session.ID, int(24*time.Hour.Seconds()), "/", "", false, true)

	// Redirect to home page
	c.Redirect(http.StatusFound, "/")
}

// Logout terminates the user session
// POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session ID from cookie
	sessionID, err := c.Cookie(SessionCookieName)
	if err == nil && sessionID != "" {
		// Delete session
		h.sessionManager.DeleteSession(sessionID)
	}

	// Clear session cookie
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetSessionManager returns the session manager (for middleware use)
func (h *AuthHandler) GetSessionManager() *services.SessionManager {
	return h.sessionManager
}

// ValidateSession validates a session ID
func (h *AuthHandler) ValidateSession(sessionID string) bool {
	return h.sessionManager.ValidateSession(sessionID)
}
