package handlers

import (
	"wechat-notification/models"
	"wechat-notification/services"
)

// SendResult represents the result of sending a message to a single recipient
type SendResult struct {
	RecipientID   int64  `json:"recipientId"`
	RecipientName string `json:"recipientName"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	MsgID         int64  `json:"msgId,omitempty"`
}

// SendResponse represents the response for message sending
type SendResponse struct {
	TotalCount  int          `json:"totalCount"`
	TotalSent   int          `json:"totalSent"`
	TotalFailed int          `json:"totalFailed"`
	Results     []SendResult `json:"results"`
}

// SendMessages sends messages to recipients and returns the response
func SendMessages(wechatSvc *services.WeChatService, recipients []models.Recipient, title, content string) SendResponse {
	var openIDs []string
	for _, r := range recipients {
		openIDs = append(openIDs, r.OpenID)
	}

	results, _ := wechatSvc.SendMessageToMultiple(openIDs, title, content)

	var sendResults []SendResult
	successCount, failureCount := 0, 0

	for _, r := range recipients {
		result := results[r.OpenID]
		success := result != nil && result.ErrCode == 0

		sendResult := SendResult{
			RecipientID:   r.ID,
			RecipientName: r.Name,
			Success:       success,
		}

		if success {
			successCount++
			if result != nil {
				sendResult.MsgID = result.MsgID
			}
		} else {
			failureCount++
			if result != nil {
				sendResult.Error = result.ErrMsg
			}
		}

		sendResults = append(sendResults, sendResult)
	}

	return SendResponse{
		TotalCount:  len(recipients),
		TotalSent:   successCount,
		TotalFailed: failureCount,
		Results:     sendResults,
	}
}
