/**
 * TypeScript types for the WeChat notification system
 * Aligned with backend models in backend/models/models.go
 */

// Recipient represents a message recipient
export interface Recipient {
  id: number;
  openId: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

// Request to create a new recipient
export interface CreateRecipientRequest {
  openId: string;
  name: string;
}

// Request to update an existing recipient
export interface UpdateRecipientRequest {
  name: string;
}

// Request to send a message
export interface SendMessageRequest {
  title: string;
  content: string;
  recipientIds: number[];
}

// Generic API response wrapper
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  code?: string;
}

// WeChat API response (for message sending results)
export interface WeChatAPIResponse {
  errcode: number;
  errmsg: string;
  msgid?: number;
}

// Message send result for each recipient
export interface MessageSendResult {
  recipientId: number;
  recipientName: string;
  success: boolean;
  error?: string;
}

// Overall message send response
export interface SendMessageResponse {
  totalSent: number;
  totalFailed: number;
  results: MessageSendResult[];
}

// Auth status
export interface AuthStatus {
  authenticated: boolean;
  user?: {
    id: string;
    name: string;
    email?: string;
  };
}

// WeChat configuration
export interface WeChatConfig {
  appId: string;
  appSecret: string;
  templateId: string;
}

// Webhook token response
export interface WebhookTokenResponse {
  hasToken: boolean;
  token: string;
}
