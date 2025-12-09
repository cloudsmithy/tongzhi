package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wechat-notification/models"
)

const (
	// WeChatSendMessageURL is the URL to send template messages
	WeChatSendMessageURL = "https://api.weixin.qq.com/cgi-bin/message/template/send"
)

// MessageHTTPClient interface for making HTTP requests (allows mocking in tests)
type MessageHTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// WeChatService handles WeChat API interactions
type WeChatService struct {
	tokenManager *TokenManager
	templateID   string
	httpClient   MessageHTTPClient
}

// NewWeChatService creates a new WeChat service
func NewWeChatService(tokenManager *TokenManager, templateID string) *WeChatService {
	return &WeChatService{
		tokenManager: tokenManager,
		templateID:   templateID,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// NewWeChatServiceWithClient creates a new WeChat service with a custom HTTP client
func NewWeChatServiceWithClient(tokenManager *TokenManager, templateID string, client MessageHTTPClient) *WeChatService {
	return &WeChatService{
		tokenManager: tokenManager,
		templateID:   templateID,
		httpClient:   client,
	}
}

// SendMessage sends a template message to a recipient with dynamic keywords
func (s *WeChatService) SendMessage(openID, templateID string, keywords map[string]string) (*models.WeChatAPIResponse, error) {
	// Get access token (will auto-refresh if expired)
	token, err := s.tokenManager.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Format the message
	msg := s.FormatTemplateMessage(openID, templateID, keywords)

	// Serialize to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Build the request URL with access token
	url := fmt.Sprintf("%s?access_token=%s", WeChatSendMessageURL, token)

	// Send the request
	resp, err := s.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var apiResp models.WeChatAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if apiResp.ErrCode != 0 {
		return &apiResp, fmt.Errorf("WeChat API error: code=%d, msg=%s", apiResp.ErrCode, apiResp.ErrMsg)
	}

	return &apiResp, nil
}

// SendMessageToMultiple sends a template message to multiple recipients concurrently
func (s *WeChatService) SendMessageToMultiple(openIDs []string, templateID string, keywords map[string]string) (map[string]*models.WeChatAPIResponse, error) {
	results := make(map[string]*models.WeChatAPIResponse)
	resultChan := make(chan struct {
		openID string
		resp   *models.WeChatAPIResponse
	}, len(openIDs))

	// Send messages concurrently
	for _, openID := range openIDs {
		go func(id string) {
			resp, err := s.SendMessage(id, templateID, keywords)
			if err != nil {
				resultChan <- struct {
					openID string
					resp   *models.WeChatAPIResponse
				}{id, &models.WeChatAPIResponse{ErrCode: -1, ErrMsg: err.Error()}}
			} else {
				resultChan <- struct {
					openID string
					resp   *models.WeChatAPIResponse
				}{id, resp}
			}
		}(openID)
	}

	// Collect results
	for range openIDs {
		r := <-resultChan
		results[r.openID] = r.resp
	}

	return results, nil
}

// FormatTemplateMessage formats a message for WeChat template API with dynamic keywords
// keywords map: {"first": "头部", "keyword1": "值1", "keyword2": "值2", "remark": "备注"}
func (s *WeChatService) FormatTemplateMessage(openID, templateID string, keywords map[string]string) *models.WeChatTemplateMessage {
	data := make(map[string]interface{})
	for key, value := range keywords {
		data[key] = map[string]string{
			"value": value,
		}
	}

	return &models.WeChatTemplateMessage{
		ToUser:     openID,
		TemplateID: templateID,
		Data:       data,
	}
}

// SerializeMessage serializes a WeChatTemplateMessage to JSON bytes
func SerializeMessage(msg *models.WeChatTemplateMessage) ([]byte, error) {
	return json.Marshal(msg)
}

// DeserializeMessage deserializes JSON bytes to a WeChatTemplateMessage
func DeserializeMessage(data []byte) (*models.WeChatTemplateMessage, error) {
	var msg models.WeChatTemplateMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// SerializeResponse serializes a WeChatAPIResponse to JSON bytes
func SerializeResponse(resp *models.WeChatAPIResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeResponse deserializes JSON bytes to a WeChatAPIResponse
func DeserializeResponse(data []byte) (*models.WeChatAPIResponse, error) {
	var resp models.WeChatAPIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PrettyPrintMessage formats a WeChatTemplateMessage as a readable JSON string for debugging
func PrettyPrintMessage(msg *models.WeChatTemplateMessage) (string, error) {
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// PrettyPrintResponse formats a WeChatAPIResponse as a readable JSON string for debugging
func PrettyPrintResponse(resp *models.WeChatAPIResponse) (string, error) {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpdateTemplateID updates the template ID
func (s *WeChatService) UpdateTemplateID(templateID string) {
	s.templateID = templateID
}
