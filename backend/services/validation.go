package services

import (
	"errors"
	"strings"

	"wechat-notification/models"
)

// Validation errors
var (
	ErrEmptyRecipients   = errors.New("recipient list cannot be empty")
	ErrEmptyTemplateKey  = errors.New("template key cannot be empty")
	ErrEmptyKeywords     = errors.New("keywords cannot be empty")
)

// ValidationResult contains the result of message validation
type ValidationResult struct {
	Valid  bool
	Errors []error
}

// ValidateMessage validates a SendMessageRequest
// Returns a ValidationResult indicating whether the message is valid
// and any validation errors encountered
func ValidateMessage(req *models.SendMessageRequest) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []error{}}

	// Validate recipients list is not empty
	if len(req.RecipientIDs) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ErrEmptyRecipients)
	}

	// Validate template key is not empty
	if strings.TrimSpace(req.TemplateKey) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ErrEmptyTemplateKey)
	}

	// Validate keywords is not empty
	if len(req.Keywords) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ErrEmptyKeywords)
	}

	return result
}

// IsWhitespaceOnly checks if a string contains only whitespace characters
func IsWhitespaceOnly(s string) bool {
	return strings.TrimSpace(s) == ""
}
