package main

import (
	"log"
	"time"

	"wechat-notification/config"
	"wechat-notification/handlers"
	"wechat-notification/middleware"
	"wechat-notification/repository"
	"wechat-notification/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	repo, err := repository.NewSQLiteRepository(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close()

	// Initialize services
	tokenManager := services.NewTokenManager(cfg.WeChat.AppID, cfg.WeChat.AppSecret)
	wechatService := services.NewWeChatService(tokenManager, cfg.WeChat.TemplateID)

	// Load WeChat config from database if available
	dbConfig, _ := repo.GetWeChatConfig()
	if dbConfig != nil && dbConfig.AppID != "" {
		tokenManager.UpdateCredentials(dbConfig.AppID, dbConfig.AppSecret)
		wechatService.UpdateTemplateID(dbConfig.TemplateID)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	recipientHandler := handlers.NewRecipientHandler(repo)
	messageHandler := handlers.NewMessageHandler(repo, wechatService)
	configHandler := handlers.NewConfigHandler(repo, tokenManager, wechatService)
	webhookHandler := handlers.NewWebhookHandler(repo, wechatService)
	templateHandler := handlers.NewTemplateHandler(repo)

	// Setup router
	r := gin.Default()

	// Configure CORS
	r.Use(middleware.CORSMiddleware(middleware.CORSConfig{
		AllowedOrigins: cfg.CORSAllowedOrigins,
	}))

	// Auth routes (public)
	r.GET("/auth/login", authHandler.Login)
	r.GET("/auth/callback", authHandler.Callback)
	r.POST("/auth/logout", authHandler.Logout)

	// Redirect root to frontend (for development)
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "http://localhost:5173")
	})

	// Protected API routes
	api := r.Group("/api")
	if !cfg.DevMode {
		api.Use(middleware.AuthMiddleware(authHandler.GetSessionManager()))
	} else {
		log.Println("WARNING: Running in dev mode - authentication is disabled")
	}
	{
		api.GET("/recipients", recipientHandler.GetAll)
		api.POST("/recipients", recipientHandler.Create)
		api.PUT("/recipients/:id", recipientHandler.Update)
		api.DELETE("/recipients/:id", recipientHandler.Delete)
		api.POST("/messages/send", messageHandler.Send)
		api.GET("/config/wechat", configHandler.GetWeChatConfig)
		api.POST("/config/wechat", configHandler.SaveWeChatConfig)
		api.GET("/webhook/token", webhookHandler.GetToken)
		api.POST("/webhook/token", webhookHandler.GenerateToken)
		api.GET("/templates", templateHandler.List)
		api.POST("/templates", templateHandler.Create)
		api.DELETE("/templates/:id", templateHandler.Delete)
	}

	// Public webhook endpoint (uses its own token auth + rate limiting)
	webhookLimiter := middleware.NewRateLimiter(10, time.Second, 20) // 10 req/s, burst 20
	r.POST("/api/webhook/send", middleware.RateLimitMiddleware(webhookLimiter), webhookHandler.Send)

	log.Printf("Server starting on %s (dev mode: %v)", cfg.ServerAddress, cfg.DevMode)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
