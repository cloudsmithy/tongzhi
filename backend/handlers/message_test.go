package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"wechat-notification/models"
	"wechat-notification/repository"
	"wechat-notification/services"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
	mu           sync.Mutex
	sentMessages []string // OpenIDs that received messages
}

func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	// Parse the request body to extract the OpenID
	bodyBytes, _ := io.ReadAll(body)
	var msg models.WeChatTemplateMessage
	json.Unmarshal(bodyBytes, &msg)

	m.mu.Lock()
	m.sentMessages = append(m.sentMessages, msg.ToUser)
	m.mu.Unlock()

	// Return a successful response
	respBody := `{"errcode": 0, "errmsg": "ok", "msgid": 12345}`
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(respBody)),
	}, nil
}

func (m *MockHTTPClient) GetSentMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.sentMessages))
	copy(result, m.sentMessages)
	return result
}

func (m *MockHTTPClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = nil
}

// MockTokenHTTPClient is a mock HTTP client for token manager
type MockTokenHTTPClient struct{}

func (m *MockTokenHTTPClient) Get(url string) (*http.Response, error) {
	respBody := `{"access_token": "test_token", "expires_in": 7200}`
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(respBody)),
	}, nil
}

// Helper function to setup gin router with message handler
func setupMessageRouter(repo *repository.SQLiteRepository, wechatService *services.WeChatService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewMessageHandler(repo, wechatService)

	api := router.Group("/api")
	api.POST("/messages/send", handler.Send)

	return router
}

// Generator for valid message title (non-empty, non-whitespace)
func genValidTitle() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 100
	})
}

// Generator for valid message content (non-empty, non-whitespace)
func genValidContent() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 1000
	})
}

// Generator for recipient count (1 to 10, at least 1 for valid message)
func genValidRecipientCount() gopter.Gen {
	return gen.IntRange(1, 10)
}

// **Feature: wechat-notification, Property 3: 消息发送到所有选定接收者**
// *对于任意* 有效的消息（非空标题和内容）和非空接收者列表，发送操作应向列表中的每个接收者发送消息
// **验证: 需求 2.1, 4.2**
func TestProperty3_MessageSentToAllSelectedRecipients(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Send should deliver message to all selected recipients", prop.ForAll(
		func(title, content string, recipientCount int) bool {
			repo, cleanup := setupTestRepo(t)
			defer cleanup()

			// Create mock HTTP clients
			mockMessageClient := &MockHTTPClient{}
			mockTokenClient := &MockTokenHTTPClient{}

			// Create token manager with mock client
			tokenManager := services.NewTokenManagerWithClient("test_app_id", "test_app_secret", mockTokenClient)

			// Create WeChat service with mock client
			wechatService := services.NewWeChatServiceWithClient(tokenManager, "test_template_id", mockMessageClient)

			router := setupMessageRouter(repo, wechatService)

			// Create recipients in the database
			recipientIDs := make([]int64, 0, recipientCount)
			expectedOpenIDs := make(map[string]bool)

			for i := 0; i < recipientCount; i++ {
				openID := generateUniqueOpenID(i)
				recipient := &models.Recipient{
					OpenID: openID,
					Name:   generateUniqueName(i),
				}
				if err := repo.Create(recipient); err != nil {
					t.Logf("Failed to create recipient: %v", err)
					return false
				}
				recipientIDs = append(recipientIDs, recipient.ID)
				expectedOpenIDs[openID] = false // false means not yet sent
			}

			// Send message request
			reqBody := models.SendMessageRequest{
				Title:        title,
				Content:      content,
				RecipientIDs: recipientIDs,
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", "/api/messages/send", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Logf("Unexpected status code: %d, body: %s", w.Code, w.Body.String())
				return false
			}

			// Verify all recipients received the message
			sentMessages := mockMessageClient.GetSentMessages()

			// Check that we sent exactly the right number of messages
			if len(sentMessages) != recipientCount {
				t.Logf("Expected %d messages, got %d", recipientCount, len(sentMessages))
				return false
			}

			// Check that each expected OpenID received a message
			for _, openID := range sentMessages {
				if _, exists := expectedOpenIDs[openID]; !exists {
					t.Logf("Unexpected OpenID received message: %s", openID)
					return false
				}
				expectedOpenIDs[openID] = true
			}

			// Verify all expected OpenIDs were sent to
			for openID, sent := range expectedOpenIDs {
				if !sent {
					t.Logf("OpenID did not receive message: %s", openID)
					return false
				}
			}

			return true
		},
		genValidTitle(),
		genValidContent(),
		genValidRecipientCount(),
	))

	properties.TestingRun(t)
}

func generateUniqueOpenID(index int) string {
	return "openid_" + string(rune('a'+index%26)) + "_" + string(rune('0'+index/26))
}

func generateUniqueName(index int) string {
	return "name_" + string(rune('A'+index%26)) + "_" + string(rune('0'+index/26))
}
