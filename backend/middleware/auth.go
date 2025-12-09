package middleware

import (
	"net/http"

	"wechat-notification/services"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "session_id"
	ContextKeySession = "session"
)

// AuthMiddleware validates user authentication using session manager
func AuthMiddleware(sessionManager *services.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session ID from cookie
		sessionID, err := c.Cookie(SessionCookieName)
		if err != nil || sessionID == "" {
			UnauthorizedResponse(c)
			return
		}

		// Validate session
		session := sessionManager.GetSession(sessionID)
		if session == nil {
			UnauthorizedResponse(c)
			return
		}

		// Store session in context for handlers to use
		c.Set(ContextKeySession, session)

		c.Next()
	}
}

// UnauthorizedResponse returns a 401 response
func UnauthorizedResponse(c *gin.Context) {
	// Check if request accepts JSON
	acceptHeader := c.GetHeader("Accept")
	if acceptHeader == "application/json" || c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
	} else {
		// For browser requests, redirect to login
		c.Redirect(http.StatusFound, "/auth/login")
	}
	c.Abort()
}

// GetSessionFromContext retrieves the session from the gin context
func GetSessionFromContext(c *gin.Context) *services.Session {
	session, exists := c.Get(ContextKeySession)
	if !exists {
		return nil
	}
	return session.(*services.Session)
}

// OptionalAuthMiddleware allows both authenticated and unauthenticated requests
// but sets session in context if available
func OptionalAuthMiddleware(sessionManager *services.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(SessionCookieName)
		if err == nil && sessionID != "" {
			session := sessionManager.GetSession(sessionID)
			if session != nil {
				c.Set(ContextKeySession, session)
			}
		}
		c.Next()
	}
}
