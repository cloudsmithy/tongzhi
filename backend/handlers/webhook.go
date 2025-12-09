package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"wechat-notification/models"
	"wechat-notification/repository"
	"wechat-notification/services"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles webhook endpoints
type WebhookHandler struct {
	repo      *repository.SQLiteRepository
	wechatSvc *services.WeChatService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(repo *repository.SQLiteRepository, wechatSvc *services.WeChatService) *WebhookHandler {
	return &WebhookHandler{repo: repo, wechatSvc: wechatSvc}
}

// WebhookSendRequest represents the webhook send request
type WebhookSendRequest struct {
	Title        string  `json:"title" binding:"required"`
	Content      string  `json:"content" binding:"required"`
	RecipientIDs []int64 `json:"recipientIds"` // Optional, if empty sends to all recipients
}

// Send handles webhook message sending
// POST /webhook/send
func (h *WebhookHandler) Send(c *gin.Context) {
	// Validate token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.ApiResponse{
			Success: false, Error: "Missing authorization header", Code: "UNAUTHORIZED",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		c.JSON(http.StatusUnauthorized, models.ApiResponse{
			Success: false, Error: "Invalid authorization format, use: Bearer <token>", Code: "UNAUTHORIZED",
		})
		return
	}

	// Verify token
	savedToken, _ := h.repo.GetConfig("webhook_token")
	if savedToken == "" || token != savedToken {
		c.JSON(http.StatusUnauthorized, models.ApiResponse{
			Success: false, Error: "Invalid webhook token", Code: "UNAUTHORIZED",
		})
		return
	}

	// Check WeChat config
	wechatConfig, _ := h.repo.GetWeChatConfig()
	if wechatConfig == nil || wechatConfig.AppID == "" || wechatConfig.AppSecret == "" || wechatConfig.TemplateID == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false, Error: "WeChat configuration not set. Please configure AppID, AppSecret and TemplateID first.", Code: "CONFIG_NOT_SET",
		})
		return
	}

	// Parse request
	var req WebhookSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false, Error: "Invalid request: title and content are required", Code: "INVALID_REQUEST",
		})
		return
	}

	// Validate message content
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false, Error: "Title and content cannot be empty", Code: "VALIDATION_ERROR",
		})
		return
	}

	// Get recipients
	var recipients []models.Recipient
	var err error

	if len(req.RecipientIDs) > 0 {
		// Get specific recipients by IDs
		recipients, err = h.repo.GetByIDs(req.RecipientIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ApiResponse{
				Success: false, Error: "Failed to get recipients", Code: "DATABASE_ERROR",
			})
			return
		}
	} else {
		// Get all recipients
		recipients, err = h.repo.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ApiResponse{
				Success: false, Error: "Failed to get recipients", Code: "DATABASE_ERROR",
			})
			return
		}
	}

	if len(recipients) == 0 {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false, Error: "No recipients found", Code: "NO_RECIPIENTS",
		})
		return
	}

	// Send messages using shared logic
	response := SendMessages(h.wechatSvc, recipients, req.Title, req.Content)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    response,
	})
}

// GetToken returns the current webhook token (masked)
// GET /api/webhook/token
func (h *WebhookHandler) GetToken(c *gin.Context) {
	token, _ := h.repo.GetConfig("webhook_token")
	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"hasToken": token != "",
			"token":    token, // Show full token for copying
		},
	})
}

// GenerateToken generates a new webhook token
// POST /api/webhook/token
func (h *WebhookHandler) GenerateToken(c *gin.Context) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false, Error: "Failed to generate token", Code: "INTERNAL_ERROR",
		})
		return
	}
	token := hex.EncodeToString(bytes)

	// Save token
	if err := h.repo.SetConfig("webhook_token", token); err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false, Error: "Failed to save token", Code: "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    map[string]string{"token": token},
	})
}
