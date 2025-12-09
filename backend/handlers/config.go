package handlers

import (
	"net/http"

	"wechat-notification/models"
	"wechat-notification/repository"
	"wechat-notification/services"

	"github.com/gin-gonic/gin"
)

// ConfigHandler handles configuration endpoints
type ConfigHandler struct {
	repo         *repository.SQLiteRepository
	tokenManager *services.TokenManager
	wechatSvc    *services.WeChatService
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(repo *repository.SQLiteRepository, tokenManager *services.TokenManager, wechatSvc *services.WeChatService) *ConfigHandler {
	return &ConfigHandler{repo: repo, tokenManager: tokenManager, wechatSvc: wechatSvc}
}

// GetWeChatConfig returns the current WeChat configuration
// GET /api/config/wechat
func (h *ConfigHandler) GetWeChatConfig(c *gin.Context) {
	config, err := h.repo.GetWeChatConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Error:   "Failed to retrieve configuration",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	// Mask the app secret for security
	maskedConfig := &models.WeChatConfig{
		AppID:      config.AppID,
		AppSecret:  maskSecret(config.AppSecret),
		TemplateID: config.TemplateID,
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    maskedConfig,
	})
}

// SaveWeChatConfig saves the WeChat configuration
// POST /api/config/wechat
func (h *ConfigHandler) SaveWeChatConfig(c *gin.Context) {
	var config models.WeChatConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// If secret is masked, keep the old one
	if config.AppSecret == "" || config.AppSecret == "******" {
		oldConfig, _ := h.repo.GetWeChatConfig()
		if oldConfig != nil {
			config.AppSecret = oldConfig.AppSecret
		}
	}

	if err := h.repo.SaveWeChatConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Error:   "Failed to save configuration",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	// Update token manager and wechat service with new config
	h.tokenManager.UpdateCredentials(config.AppID, config.AppSecret)
	h.wechatSvc.UpdateTemplateID(config.TemplateID)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    gin.H{"message": "Configuration saved successfully"},
	})
}

func maskSecret(secret string) string {
	if len(secret) == 0 {
		return ""
	}
	return "******"
}
