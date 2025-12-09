package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"wechat-notification/models"
	"wechat-notification/repository"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Helper function to create a test repository with a temporary database
func setupTestRepo(t *testing.T) (*repository.SQLiteRepository, func()) {
	tmpFile, err := os.CreateTemp("", "test_handler_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	repo, err := repository.NewSQLiteRepository(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create repository: %v", err)
	}

	cleanup := func() {
		repo.Close()
		os.Remove(tmpFile.Name())
	}

	return repo, cleanup
}

// Helper function to setup gin router with recipient handler
func setupRouter(repo *repository.SQLiteRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewRecipientHandler(repo)

	api := router.Group("/api")
	api.GET("/recipients", handler.GetAll)
	api.POST("/recipients", handler.Create)
	api.PUT("/recipients/:id", handler.Update)
	api.DELETE("/recipients/:id", handler.Delete)

	return router
}

// Generator for valid OpenID strings (non-empty alphanumeric)
func genValidOpenID() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 64
	})
}

// Generator for valid name strings (non-empty)
func genValidName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 100
	})
}

// Generator for recipient count (0 to 10)
func genRecipientCount() gopter.Gen {
	return gen.IntRange(0, 10)
}

// **Feature: wechat-notification, Property 6: 接收者列表完整性**
// *对于任意* 数据库中的接收者集合，获取接收者列表应返回所有接收者，不多不少
// **验证: 需求 3.1**
func TestProperty6_RecipientListCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("GetAll should return exactly all recipients in the database", prop.ForAll(
		func(count int) bool {
			repo, cleanup := setupTestRepo(t)
			defer cleanup()

			router := setupRouter(repo)

			// Generate and add unique recipients to the database
			type recipientData struct {
				OpenID string
				Name   string
			}
			addedRecipients := make(map[int64]recipientData)

			for i := 0; i < count; i++ {
				openID := fmt.Sprintf("openid_%d", i)
				name := fmt.Sprintf("name_%d", i)
				recipient := &models.Recipient{
					OpenID: openID,
					Name:   name,
				}
				if err := repo.Create(recipient); err != nil {
					return false
				}
				addedRecipients[recipient.ID] = recipientData{OpenID: openID, Name: name}
			}

			// Call GetAll via HTTP
			req, _ := http.NewRequest("GET", "/api/recipients", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				return false
			}

			var response models.ApiResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				return false
			}

			if !response.Success {
				return false
			}

			// Parse the data as recipients
			dataBytes, err := json.Marshal(response.Data)
			if err != nil {
				return false
			}

			var recipients []models.Recipient
			if err := json.Unmarshal(dataBytes, &recipients); err != nil {
				return false
			}

			// Verify count matches
			if len(recipients) != count {
				return false
			}

			// Verify all added recipients are present
			for _, r := range recipients {
				data, exists := addedRecipients[r.ID]
				if !exists {
					return false
				}
				if r.OpenID != data.OpenID || r.Name != data.Name {
					return false
				}
			}

			return true
		},
		genRecipientCount(),
	))

	properties.TestingRun(t)
}
