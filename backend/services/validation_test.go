package services

import (
	"reflect"
	"testing"
	"unicode"

	"wechat-notification/models"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Generator for non-empty strings (valid titles/content)
func genNonEmptyString() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})
}

// Generator for whitespace-only strings
func genWhitespaceString() gopter.Gen {
	return gen.IntRange(1, 20).Map(func(length int) string {
		whitespaceChars := []rune{' ', '\t', '\n', '\r'}
		result := make([]rune, length)
		for i := 0; i < length; i++ {
			result[i] = whitespaceChars[i%len(whitespaceChars)]
		}
		return string(result)
	})
}

// Generator for non-empty recipient ID lists
func genNonEmptyRecipientIDs() gopter.Gen {
	return gen.IntRange(1, 10).FlatMap(func(v interface{}) gopter.Gen {
		count := v.(int)
		ids := make([]int64, count)
		for i := 0; i < count; i++ {
			ids[i] = int64(i + 1)
		}
		return gen.Const(ids)
	}, reflect.TypeOf([]int64{}))
}

// **Feature: wechat-notification, Property 4: 空接收者列表验证**
// *对于任意* 消息发送请求，如果接收者列表为空，系统应拒绝该请求并返回验证错误
// **验证: 需求 2.2**
func TestProperty4_EmptyRecipientsValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Message with empty recipients should be rejected", prop.ForAll(
		func(title, content string) bool {
			req := &models.SendMessageRequest{
				Title:        title,
				Content:      content,
				RecipientIDs: []int64{}, // Empty recipients list
			}

			result := ValidateMessage(req)

			// Should be invalid
			if result.Valid {
				return false
			}

			// Should contain ErrEmptyRecipients
			hasEmptyRecipientsError := false
			for _, err := range result.Errors {
				if err == ErrEmptyRecipients {
					hasEmptyRecipientsError = true
					break
				}
			}

			return hasEmptyRecipientsError
		},
		genNonEmptyString(),
		genNonEmptyString(),
	))

	properties.TestingRun(t)
}


// **Feature: wechat-notification, Property 5: 空白消息验证**
// *对于任意* 仅包含空白字符的标题或内容，系统应拒绝该消息并返回验证错误
// **验证: 需求 2.3**
func TestProperty5_WhitespaceMessageValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Test whitespace-only title
	properties.Property("Message with whitespace-only title should be rejected", prop.ForAll(
		func(whitespaceTitle, validContent string, recipientIDs []int64) bool {
			req := &models.SendMessageRequest{
				Title:        whitespaceTitle,
				Content:      validContent,
				RecipientIDs: recipientIDs,
			}

			result := ValidateMessage(req)

			// Should be invalid
			if result.Valid {
				return false
			}

			// Should contain ErrEmptyTitle
			hasEmptyTitleError := false
			for _, err := range result.Errors {
				if err == ErrEmptyTitle {
					hasEmptyTitleError = true
					break
				}
			}

			return hasEmptyTitleError
		},
		genWhitespaceString(),
		genNonEmptyString(),
		genNonEmptyRecipientIDs(),
	))

	// Test whitespace-only content
	properties.Property("Message with whitespace-only content should be rejected", prop.ForAll(
		func(validTitle, whitespaceContent string, recipientIDs []int64) bool {
			req := &models.SendMessageRequest{
				Title:        validTitle,
				Content:      whitespaceContent,
				RecipientIDs: recipientIDs,
			}

			result := ValidateMessage(req)

			// Should be invalid
			if result.Valid {
				return false
			}

			// Should contain ErrEmptyContent
			hasEmptyContentError := false
			for _, err := range result.Errors {
				if err == ErrEmptyContent {
					hasEmptyContentError = true
					break
				}
			}

			return hasEmptyContentError
		},
		genNonEmptyString(),
		genWhitespaceString(),
		genNonEmptyRecipientIDs(),
	))

	// Test empty string title
	properties.Property("Message with empty title should be rejected", prop.ForAll(
		func(validContent string, recipientIDs []int64) bool {
			req := &models.SendMessageRequest{
				Title:        "",
				Content:      validContent,
				RecipientIDs: recipientIDs,
			}

			result := ValidateMessage(req)
			return !result.Valid && containsError(result.Errors, ErrEmptyTitle)
		},
		genNonEmptyString(),
		genNonEmptyRecipientIDs(),
	))

	// Test empty string content
	properties.Property("Message with empty content should be rejected", prop.ForAll(
		func(validTitle string, recipientIDs []int64) bool {
			req := &models.SendMessageRequest{
				Title:        validTitle,
				Content:      "",
				RecipientIDs: recipientIDs,
			}

			result := ValidateMessage(req)
			return !result.Valid && containsError(result.Errors, ErrEmptyContent)
		},
		genNonEmptyString(),
		genNonEmptyRecipientIDs(),
	))

	properties.TestingRun(t)
}

// Helper function to check if an error slice contains a specific error
func containsError(errors []error, target error) bool {
	for _, err := range errors {
		if err == target {
			return true
		}
	}
	return false
}

// Test that valid messages pass validation
func TestValidMessagePasses(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Valid message should pass validation", prop.ForAll(
		func(title, content string, recipientIDs []int64) bool {
			req := &models.SendMessageRequest{
				Title:        title,
				Content:      content,
				RecipientIDs: recipientIDs,
			}

			result := ValidateMessage(req)
			return result.Valid && len(result.Errors) == 0
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyRecipientIDs(),
	))

	properties.TestingRun(t)
}

// Test IsWhitespaceOnly helper function
func TestIsWhitespaceOnly(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Whitespace-only strings should return true", prop.ForAll(
		func(s string) bool {
			return IsWhitespaceOnly(s)
		},
		genWhitespaceString(),
	))

	properties.Property("Non-whitespace strings should return false", prop.ForAll(
		func(s string) bool {
			// Check if string has at least one non-whitespace character
			hasNonWhitespace := false
			for _, r := range s {
				if !unicode.IsSpace(r) {
					hasNonWhitespace = true
					break
				}
			}
			if hasNonWhitespace {
				return !IsWhitespaceOnly(s)
			}
			return true // Skip strings that are actually whitespace-only
		},
		genNonEmptyString(),
	))

	properties.TestingRun(t)
}
