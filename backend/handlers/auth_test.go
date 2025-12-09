package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wechat-notification/config"
	"wechat-notification/middleware"
	"wechat-notification/services"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Helper function to create a test config
func createTestConfig() *config.Config {
	return &config.Config{
		ServerAddress: ":8080",
		DatabasePath:  ":memory:",
		SessionSecret: "test-secret",
		OIDC: config.OIDCConfig{
			ProviderURL:  "https://example.com",
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			RedirectURL:  "http://localhost:8080/auth/callback",
		},
	}
}

// Helper function to setup router with auth middleware
func setupAuthRouter(sessionManager *services.SessionManager) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Protected endpoint for testing
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(sessionManager))
	{
		api.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	}

	return router
}

// Generator for random session IDs (non-empty strings)
func genSessionID() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 64
	})
}

// Generator for random user IDs
func genUserID() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 32
	})
}

// Generator for random email addresses
func genEmail() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			s = "user"
		}
		return s + "@example.com"
	})
}

// **Feature: wechat-notification, Property 1: 未认证访问重定向**
// *对于任意* 未携带有效会话的请求，访问受保护的 API 端点应返回 401 状态码或重定向到登录页面
// **验证: 需求 1.1**
func TestProperty1_UnauthenticatedAccessRedirect(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Requests without valid session should return 401 or redirect", prop.ForAll(
		func(invalidSessionID string) bool {
			sessionManager := services.NewSessionManager(24 * time.Hour)
			router := setupAuthRouter(sessionManager)

			// Test 1: Request with no session cookie
			req1, _ := http.NewRequest("GET", "/api/test", nil)
			req1.Header.Set("Accept", "application/json")
			w1 := httptest.NewRecorder()
			router.ServeHTTP(w1, req1)

			if w1.Code != http.StatusUnauthorized {
				return false
			}

			// Test 2: Request with invalid session cookie
			req2, _ := http.NewRequest("GET", "/api/test", nil)
			req2.Header.Set("Accept", "application/json")
			req2.AddCookie(&http.Cookie{
				Name:  middleware.SessionCookieName,
				Value: invalidSessionID,
			})
			w2 := httptest.NewRecorder()
			router.ServeHTTP(w2, req2)

			if w2.Code != http.StatusUnauthorized {
				return false
			}

			// Test 3: Request without JSON accept header should redirect
			req3, _ := http.NewRequest("GET", "/api/test", nil)
			w3 := httptest.NewRecorder()
			router.ServeHTTP(w3, req3)

			// Should redirect to login
			if w3.Code != http.StatusFound {
				return false
			}

			location := w3.Header().Get("Location")
			if location != "/auth/login" {
				return false
			}

			return true
		},
		genSessionID(),
	))

	properties.TestingRun(t)
}

// **Feature: wechat-notification, Property 2: 登出终止会话**
// *对于任意* 有效会话，执行登出操作后，使用相同会话令牌的后续请求应被拒绝
// **验证: 需求 1.4**
func TestProperty2_LogoutTerminatesSession(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("After logout, the same session token should be rejected", prop.ForAll(
		func(userID string, email string) bool {
			cfg := createTestConfig()
			sessionManager := services.NewSessionManager(24 * time.Hour)

			// Create a valid session
			session, err := sessionManager.CreateSession(userID, email)
			if err != nil {
				return false
			}

			// Verify session is valid before logout
			if !sessionManager.ValidateSession(session.ID) {
				return false
			}

			// Setup router with auth handler
			gin.SetMode(gin.TestMode)
			router := gin.New()

			authHandler := NewAuthHandlerWithDeps(cfg, nil, sessionManager)
			router.POST("/auth/logout", authHandler.Logout)

			api := router.Group("/api")
			api.Use(middleware.AuthMiddleware(sessionManager))
			{
				api.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})
			}

			// Verify we can access protected endpoint before logout
			req1, _ := http.NewRequest("GET", "/api/test", nil)
			req1.Header.Set("Accept", "application/json")
			req1.AddCookie(&http.Cookie{
				Name:  middleware.SessionCookieName,
				Value: session.ID,
			})
			w1 := httptest.NewRecorder()
			router.ServeHTTP(w1, req1)

			if w1.Code != http.StatusOK {
				return false
			}

			// Perform logout
			reqLogout, _ := http.NewRequest("POST", "/auth/logout", nil)
			reqLogout.AddCookie(&http.Cookie{
				Name:  middleware.SessionCookieName,
				Value: session.ID,
			})
			wLogout := httptest.NewRecorder()
			router.ServeHTTP(wLogout, reqLogout)

			if wLogout.Code != http.StatusOK {
				return false
			}

			// Verify session is no longer valid
			if sessionManager.ValidateSession(session.ID) {
				return false
			}

			// Verify we cannot access protected endpoint after logout
			req2, _ := http.NewRequest("GET", "/api/test", nil)
			req2.Header.Set("Accept", "application/json")
			req2.AddCookie(&http.Cookie{
				Name:  middleware.SessionCookieName,
				Value: session.ID,
			})
			w2 := httptest.NewRecorder()
			router.ServeHTTP(w2, req2)

			if w2.Code != http.StatusUnauthorized {
				return false
			}

			return true
		},
		genUserID(),
		genEmail(),
	))

	properties.TestingRun(t)
}

// Test that valid sessions allow access
func TestValidSessionAllowsAccess(t *testing.T) {
	sessionManager := services.NewSessionManager(24 * time.Hour)
	router := setupAuthRouter(sessionManager)

	// Create a valid session
	session, err := sessionManager.CreateSession("user123", "user@example.com")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Request with valid session
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  middleware.SessionCookieName,
		Value: session.ID,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// Test expired session is rejected
func TestExpiredSessionIsRejected(t *testing.T) {
	// Create session manager with very short TTL
	sessionManager := services.NewSessionManager(1 * time.Millisecond)
	router := setupAuthRouter(sessionManager)

	// Create a session
	session, err := sessionManager.CreateSession("user123", "user@example.com")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Wait for session to expire
	time.Sleep(10 * time.Millisecond)

	// Request with expired session
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  middleware.SessionCookieName,
		Value: session.ID,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
