package models

import (
	"time"
)

// Recipient represents a message recipient
type Recipient struct {
	ID        int64     `json:"id"`
	OpenID    string    `json:"openId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	Title        string  `json:"title"`
	Content      string  `json:"content"`
	RecipientIDs []int64 `json:"recipientIds"`
}

// WeChatTemplateMessage represents a WeChat template message
type WeChatTemplateMessage struct {
	ToUser     string                 `json:"touser"`
	TemplateID string                 `json:"template_id"`
	Data       map[string]interface{} `json:"data"`
}

// WeChatAPIResponse represents a response from WeChat API
type WeChatAPIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgID   int64  `json:"msgid,omitempty"`
}

// ApiResponse represents a generic API response
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// WeChatConfig represents WeChat test account configuration
type WeChatConfig struct {
	AppID      string `json:"appId"`
	AppSecret  string `json:"appSecret"`
	TemplateID string `json:"templateId"`
}
