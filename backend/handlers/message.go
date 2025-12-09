package handlers

import (
	"net/http"

	"wechat-notification/models"
	"wechat-notification/repository"
	"wechat-notification/services"

	"github.com/gin-gonic/gin"
)

// MessageHandler handles message endpoints
type MessageHandler struct {
	repo          *repository.SQLiteRepository
	wechatService *services.WeChatService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(repo *repository.SQLiteRepository, wechatService *services.WeChatService) *MessageHandler {
	return &MessageHandler{
		repo:          repo,
		wechatService: wechatService,
	}
}

// Send sends a message to selected recipients
// POST /api/messages/send
func (h *MessageHandler) Send(c *gin.Context) {
	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate the message request
	validationResult := services.ValidateMessage(&req)
	if !validationResult.Valid {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   validationResult.Errors[0].Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Fetch recipients from database
	var recipients []models.Recipient
	for _, id := range req.RecipientIDs {
		recipient, err := h.repo.GetByID(id)
		if err != nil {
			if err == repository.ErrNotFound {
				c.JSON(http.StatusBadRequest, models.ApiResponse{
					Success: false,
					Error:   "One or more recipients not found",
					Code:    "RECIPIENT_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ApiResponse{
				Success: false,
				Error:   "Failed to retrieve recipients",
				Code:    "DATABASE_ERROR",
			})
			return
		}
		recipients = append(recipients, *recipient)
	}

	// Send messages using shared logic
	response := SendMessages(h.wechatService, recipients, req.Title, req.Content)

	// Determine response status
	if response.TotalFailed == 0 {
		c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: response})
	} else if response.TotalSent > 0 {
		c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: response, Error: "Some messages failed to send", Code: "PARTIAL_SUCCESS"})
	} else {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Success: false, Data: response, Error: "Failed to send messages", Code: "SEND_FAILED"})
	}
}
