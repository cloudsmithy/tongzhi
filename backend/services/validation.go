package services

import (
	"errors"
	"strings"

	"wechat-notification/models"
)

// Validation errors
var (
	ErrEmptyRecipients = errors.New("recipient list cannot be empty")
	ErrEmptyTitle      = errors.New("title cannot be empty or whitespace only")
	ErrEmptyContent    = errors.New("content cannot be empty or whitespace only")
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

	// Validate title is not empty or whitespace only
	if strings.TrimSpace(req.Title) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ErrEmptyTitle)
	}

	// Validate content is not empty or whitespace only
	if strings.TrimSpace(req.Content) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ErrEmptyContent)
	}

	return result
}

// IsWhitespaceOnly checks if a string contains only whitespace characters
func IsWhitespaceOnly(s string) bool {
	return strings.TrimSpace(s) == ""
}
