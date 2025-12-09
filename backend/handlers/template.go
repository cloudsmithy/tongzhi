package handlers

import (
	"net/http"
	"strconv"

	"wechat-notification/models"
	"wechat-notification/repository"

	"github.com/gin-gonic/gin"
)

// TemplateHandler handles template endpoints
type TemplateHandler struct {
	repo *repository.SQLiteRepository
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(repo *repository.SQLiteRepository) *TemplateHandler {
	return &TemplateHandler{repo: repo}
}

// CreateTemplateRequest represents a request to create a template
type CreateTemplateRequest struct {
	Key        string `json:"key" binding:"required"`
	TemplateID string `json:"templateId" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

// List returns all templates
// GET /api/templates
func (h *TemplateHandler) List(c *gin.Context) {
	templates, err := h.repo.GetAllTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false, Error: "Failed to get templates", Code: "DATABASE_ERROR",
		})
		return
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: templates})
}

// Create creates a new template
// POST /api/templates
func (h *TemplateHandler) Create(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false, Error: "Invalid request", Code: "INVALID_REQUEST",
		})
		return
	}

	template := &models.MessageTemplate{
		Key:        req.Key,
		TemplateID: req.TemplateID,
		Name:       req.Name,
	}

	if err := h.repo.CreateTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false, Error: "Failed to create template", Code: "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, models.ApiResponse{Success: true, Data: template})
}

// Delete deletes a template
// DELETE /api/templates/:id
func (h *TemplateHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false, Error: "Invalid ID", Code: "INVALID_ID",
		})
		return
	}

	if err := h.repo.DeleteTemplate(id); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ApiResponse{
				Success: false, Error: "Template not found", Code: "NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ApiResponse{
			Success: false, Error: "Failed to delete template", Code: "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{Success: true})
}
