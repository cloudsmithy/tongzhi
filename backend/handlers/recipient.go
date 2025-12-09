package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"wechat-notification/models"
	"wechat-notification/repository"

	"github.com/gin-gonic/gin"
)

// RecipientHandler handles recipient endpoints
type RecipientHandler struct {
	repo *repository.SQLiteRepository
}

// NewRecipientHandler creates a new recipient handler
func NewRecipientHandler(repo *repository.SQLiteRepository) *RecipientHandler {
	return &RecipientHandler{repo: repo}
}

// CreateRecipientRequest represents the request body for creating a recipient
type CreateRecipientRequest struct {
	OpenID string `json:"openId" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

// UpdateRecipientRequest represents the request body for updating a recipient
type UpdateRecipientRequest struct {
	OpenID string `json:"openId"`
	Name   string `json:"name"`
}

// GetAll returns all recipients
// GET /api/recipients
func (h *RecipientHandler) GetAll(c *gin.Context) {
	recipients, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Error:   "Failed to retrieve recipients",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    recipients,
	})
}

// Create adds a new recipient
// POST /api/recipients
func (h *RecipientHandler) Create(c *gin.Context) {
	var req CreateRecipientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid request format: openId and name are required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate OpenID is not empty or whitespace
	if strings.TrimSpace(req.OpenID) == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "OpenID cannot be empty or whitespace only",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate Name is not empty or whitespace
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Name cannot be empty or whitespace only",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	recipient := &models.Recipient{
		OpenID: strings.TrimSpace(req.OpenID),
		Name:   strings.TrimSpace(req.Name),
	}

	if err := h.repo.Create(recipient); err != nil {
		if errors.Is(err, repository.ErrDuplicateOpenID) {
			c.JSON(http.StatusConflict, models.ApiResponse{
				Success: false,
				Error:   "A recipient with this OpenID already exists",
				Code:    "DUPLICATE_OPENID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Error:   "Failed to create recipient",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, models.ApiResponse{
		Success: true,
		Data:    recipient,
	})
}

// Update modifies an existing recipient
// PUT /api/recipients/:id
func (h *RecipientHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid recipient ID",
			Code:    "INVALID_ID",
		})
		return
	}

	// Get existing recipient
	existing, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, models.ApiResponse{
				Success: false,
				Error:   "Recipient not found",
				Code:    "NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Error:   "Failed to retrieve recipient",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	var req UpdateRecipientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Update fields if provided
	if req.OpenID != "" {
		trimmedOpenID := strings.TrimSpace(req.OpenID)
		if trimmedOpenID == "" {
			c.JSON(http.StatusBadRequest, models.ApiResponse{
				Success: false,
				Error:   "OpenID cannot be empty or whitespace only",
				Code:    "VALIDATION_ERROR",
			})
			return
		}
		existing.OpenID = trimmedOpenID
	}

	if req.Name != "" {
		trimmedName := strings.TrimSpace(req.Name)
		if trimmedName == "" {
			c.JSON(http.StatusBadRequest, models.ApiResponse{
				Success: false,
				Error:   "Name cannot be empty or whitespace only",
				Code:    "VALIDATION_ERROR",
			})
			return
		}
		existing.Name = trimmedName
	}

	if err := h.repo.Update(existing); err != nil {
		if errors.Is(err, repository.ErrDuplicateOpenID) {
			c.JSON(http.StatusConflict, models.ApiResponse{
				Success: false,
				Error:   "A recipient with this OpenID already exists",
				Code:    "DUPLICATE_OPENID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Error:   "Failed to update recipient",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    existing,
	})
}

// Delete removes a recipient
// DELETE /api/recipients/:id
func (h *RecipientHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "Invalid recipient ID",
			Code:    "INVALID_ID",
		})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, models.ApiResponse{
				Success: false,
				Error:   "Recipient not found",
				Code:    "NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false,
			Error:   "Failed to delete recipient",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    gin.H{"message": "Recipient deleted successfully"},
	})
}
