package services

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"wechat-notification/models"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Generator for valid OpenID strings (WeChat OpenIDs are typically 28 characters)
func genOpenID() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		return "o" + s // OpenIDs typically start with 'o'
	})
}

// Generator for valid template IDs
func genTemplateID() gopter.Gen {
	return gen.Identifier()
}

// Generator for message title
func genTitle() gopter.Gen {
	return gen.Identifier()
}

// Generator for message content
func genContent() gopter.Gen {
	return gen.Identifier()
}

// Generator for WeChatTemplateMessage
func genWeChatTemplateMessage() gopter.Gen {
	return gopter.CombineGens(
		genOpenID(),
		genTemplateID(),
		genTitle(),
		genContent(),
	).Map(func(values []interface{}) *models.WeChatTemplateMessage {
		openID := values[0].(string)
		templateID := values[1].(string)
		title := values[2].(string)
		content := values[3].(string)

		return &models.WeChatTemplateMessage{
			ToUser:     openID,
			TemplateID: templateID,
			Data: map[string]interface{}{
				"title": map[string]string{
					"value": title,
				},
				"content": map[string]string{
					"value": content,
				},
			},
		}
	})
}

// Generator for WeChatAPIResponse
func genWeChatAPIResponse() gopter.Gen {
	return gopter.CombineGens(
		gen.IntRange(0, 100),
		gen.AlphaString(),
		gen.Int64Range(1, 999999999),
	).Map(func(values []interface{}) *models.WeChatAPIResponse {
		return &models.WeChatAPIResponse{
			ErrCode: values[0].(int),
			ErrMsg:  values[1].(string),
			MsgID:   values[2].(int64),
		}
	})
}

// **Feature: wechat-notification, Property 13: 微信消息格式化**
// *对于任意* 消息标题和内容，格式化后的微信模板消息应包含所有必需字段且符合 API 规范
// **验证: 需求 6.1**
func TestProperty13_WeChatMessageFormatting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Formatted message should contain all required fields", prop.ForAll(
		func(openID, templateID, title, content string) bool {
			tokenManager := NewTokenManager("test_app_id", "test_app_secret")
			service := NewWeChatService(tokenManager, templateID)

			msg := service.FormatTemplateMessage(openID, title, content)

			// Check required fields
			if msg.ToUser != openID {
				return false
			}
			if msg.TemplateID != templateID {
				return false
			}
			if msg.Data == nil {
				return false
			}

			// Check title field
			titleData, ok := msg.Data["title"]
			if !ok {
				return false
			}
			titleMap, ok := titleData.(map[string]string)
			if !ok {
				return false
			}
			if titleMap["value"] != title {
				return false
			}

			// Check content field
			contentData, ok := msg.Data["content"]
			if !ok {
				return false
			}
			contentMap, ok := contentData.(map[string]string)
			if !ok {
				return false
			}
			if contentMap["value"] != content {
				return false
			}

			return true
		},
		genOpenID(),
		genTemplateID(),
		genTitle(),
		genContent(),
	))

	properties.TestingRun(t)
}


// **Feature: wechat-notification, Property 14: JSON 序列化往返**
// *对于任意* 有效的消息数据结构，序列化为 JSON 后再反序列化应得到等价的数据结构
// **验证: 需求 6.3, 6.4**
func TestProperty14_JSONSerializationRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Test WeChatTemplateMessage round-trip
	properties.Property("WeChatTemplateMessage JSON round-trip should preserve data", prop.ForAll(
		func(msg *models.WeChatTemplateMessage) bool {
			// Serialize
			data, err := SerializeMessage(msg)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := DeserializeMessage(data)
			if err != nil {
				return false
			}

			// Compare fields
			if restored.ToUser != msg.ToUser {
				return false
			}
			if restored.TemplateID != msg.TemplateID {
				return false
			}

			// Compare Data map - need to handle type conversion from JSON
			if len(restored.Data) != len(msg.Data) {
				return false
			}

			// Check title
			origTitle, ok1 := msg.Data["title"].(map[string]string)
			restoredTitle, ok2 := restored.Data["title"].(map[string]interface{})
			if ok1 && ok2 {
				if origTitle["value"] != restoredTitle["value"].(string) {
					return false
				}
			}

			// Check content
			origContent, ok1 := msg.Data["content"].(map[string]string)
			restoredContent, ok2 := restored.Data["content"].(map[string]interface{})
			if ok1 && ok2 {
				if origContent["value"] != restoredContent["value"].(string) {
					return false
				}
			}

			return true
		},
		genWeChatTemplateMessage(),
	))

	// Test WeChatAPIResponse round-trip
	properties.Property("WeChatAPIResponse JSON round-trip should preserve data", prop.ForAll(
		func(resp *models.WeChatAPIResponse) bool {
			// Serialize
			data, err := SerializeResponse(resp)
			if err != nil {
				return false
			}

			// Deserialize
			restored, err := DeserializeResponse(data)
			if err != nil {
				return false
			}

			// Compare fields
			return restored.ErrCode == resp.ErrCode &&
				restored.ErrMsg == resp.ErrMsg &&
				restored.MsgID == resp.MsgID
		},
		genWeChatAPIResponse(),
	))

	properties.TestingRun(t)
}

// Helper function to compare two WeChatTemplateMessage structs
func messagesEqual(a, b *models.WeChatTemplateMessage) bool {
	if a.ToUser != b.ToUser || a.TemplateID != b.TemplateID {
		return false
	}
	return reflect.DeepEqual(a.Data, b.Data)
}


// MockHTTPClient for testing token refresh
type MockHTTPClient struct {
	GetFunc  func(url string) (*http.Response, error)
	PostFunc func(url, contentType string, body io.Reader) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	if m.GetFunc != nil {
		return m.GetFunc(url)
	}
	return nil, nil
}

func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	if m.PostFunc != nil {
		return m.PostFunc(url, contentType, body)
	}
	return nil, nil
}


// **Feature: wechat-notification, Property 15: 令牌自动刷新**
// *对于任意* 过期的访问令牌，发送消息前系统应自动刷新令牌，确保使用有效令牌
// **验证: 需求 6.2**
func TestProperty15_TokenAutoRefresh(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Expired token should trigger refresh before use", prop.ForAll(
		func(initialToken, newToken string, expiresIn int) bool {
			// Skip if tokens are empty
			if len(initialToken) == 0 || len(newToken) == 0 {
				return true
			}

			refreshCalled := false

			// Create mock HTTP client that returns a new token
			mockClient := &MockHTTPClient{
				GetFunc: func(url string) (*http.Response, error) {
					refreshCalled = true
					responseBody := fmt.Sprintf(`{"access_token":"%s","expires_in":%d}`, newToken, expiresIn)
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(responseBody)),
					}, nil
				},
			}

			tokenManager := NewTokenManagerWithClient("test_app_id", "test_app_secret", mockClient)

			// Set an expired token
			tokenManager.SetToken(initialToken, -1*time.Hour) // Already expired

			// Verify token is expired
			if !tokenManager.IsExpired() {
				return false
			}

			// Get token - should trigger refresh
			token, err := tokenManager.GetAccessToken()
			if err != nil {
				return false
			}

			// Verify refresh was called
			if !refreshCalled {
				return false
			}

			// Verify we got the new token
			if token != newToken {
				return false
			}

			// Verify token is no longer expired
			if tokenManager.IsExpired() {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(3600, 7200), // expires_in between 1-2 hours
	))

	// Test that valid (non-expired) token does not trigger refresh
	properties.Property("Valid token should not trigger refresh", prop.ForAll(
		func(validToken string) bool {
			if len(validToken) == 0 {
				return true
			}

			refreshCalled := false

			mockClient := &MockHTTPClient{
				GetFunc: func(url string) (*http.Response, error) {
					refreshCalled = true
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"access_token":"new_token","expires_in":7200}`)),
					}, nil
				},
			}

			tokenManager := NewTokenManagerWithClient("test_app_id", "test_app_secret", mockClient)

			// Set a valid (non-expired) token
			tokenManager.SetToken(validToken, 2*time.Hour)

			// Verify token is not expired
			if tokenManager.IsExpired() {
				return false
			}

			// Get token - should NOT trigger refresh
			token, err := tokenManager.GetAccessToken()
			if err != nil {
				return false
			}

			// Verify refresh was NOT called
			if refreshCalled {
				return false
			}

			// Verify we got the original token
			return token == validToken
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}
