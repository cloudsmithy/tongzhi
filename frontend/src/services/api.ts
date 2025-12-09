import axios, { AxiosInstance, AxiosError } from 'axios';
import {
  Recipient,
  CreateRecipientRequest,
  UpdateRecipientRequest,
  SendMessageRequest,
  ApiResponse,
  SendMessageResponse,
  AuthStatus,
  WeChatConfig,
  WebhookTokenResponse,
  MessageTemplate,
} from '../types';

// API base URL - can be configured via environment variable
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';
const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL || '/auth';

// Create axios instance with default config
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Include cookies for session authentication
});

// Auth client for authentication endpoints
const authClient: AxiosInstance = axios.create({
  baseURL: AUTH_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

// Response interceptor to handle authentication errors
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Redirect to login on authentication failure
      window.location.href = `${AUTH_BASE_URL}/login`;
    }
    return Promise.reject(error);
  }
);


// ============ Recipient API ============

/**
 * Get all recipients
 * GET /api/recipients
 */
export async function getRecipients(): Promise<Recipient[]> {
  const response = await apiClient.get<ApiResponse<Recipient[]>>('/recipients');
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to fetch recipients');
  }
  return response.data.data || [];
}

/**
 * Create a new recipient
 * POST /api/recipients
 */
export async function createRecipient(data: CreateRecipientRequest): Promise<Recipient> {
  const response = await apiClient.post<ApiResponse<Recipient>>('/recipients', data);
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to create recipient');
  }
  return response.data.data!;
}

/**
 * Update an existing recipient
 * PUT /api/recipients/:id
 */
export async function updateRecipient(id: number, data: UpdateRecipientRequest): Promise<Recipient> {
  const response = await apiClient.put<ApiResponse<Recipient>>(`/recipients/${id}`, data);
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to update recipient');
  }
  return response.data.data!;
}

/**
 * Delete a recipient
 * DELETE /api/recipients/:id
 */
export async function deleteRecipient(id: number): Promise<void> {
  const response = await apiClient.delete<ApiResponse<void>>(`/recipients/${id}`);
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to delete recipient');
  }
}

// ============ Message API ============

/**
 * Send a message to selected recipients
 * POST /api/messages/send
 */
export async function sendMessage(data: SendMessageRequest): Promise<SendMessageResponse> {
  const response = await apiClient.post<ApiResponse<SendMessageResponse>>('/messages/send', data);
  if (!response.data.success && !response.data.data) {
    throw new Error(response.data.error || 'Failed to send message');
  }
  return response.data.data!;
}

// ============ Auth API ============

/**
 * Get login URL - redirects to OIDC provider
 */
export function getLoginUrl(): string {
  return `${AUTH_BASE_URL}/login`;
}

/**
 * Logout the current user
 * POST /auth/logout
 */
export async function logout(): Promise<void> {
  await authClient.post('/logout');
}

/**
 * Check authentication status
 * This is a client-side check based on whether API calls succeed
 */
export async function checkAuthStatus(): Promise<AuthStatus> {
  try {
    // Try to fetch recipients as a way to check if authenticated
    await apiClient.get('/recipients');
    return { authenticated: true };
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return { authenticated: false };
    }
    // For other errors, assume authenticated but with network issues
    throw error;
  }
}

// ============ Config API ============

/**
 * Get WeChat configuration
 * GET /api/config/wechat
 */
export async function getWeChatConfig(): Promise<WeChatConfig> {
  const response = await apiClient.get<ApiResponse<WeChatConfig>>('/config/wechat');
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to fetch config');
  }
  return response.data.data || { appId: '', appSecret: '', templateId: '' };
}

/**
 * Save WeChat configuration
 * POST /api/config/wechat
 */
export async function saveWeChatConfig(config: WeChatConfig): Promise<void> {
  const response = await apiClient.post<ApiResponse<void>>('/config/wechat', config);
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to save config');
  }
}

// ============ Webhook API ============

/**
 * Get webhook token
 * GET /api/webhook/token
 */
export async function getWebhookToken(): Promise<WebhookTokenResponse> {
  const response = await apiClient.get<ApiResponse<WebhookTokenResponse>>('/webhook/token');
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to fetch webhook token');
  }
  return response.data.data || { hasToken: false, token: '' };
}

/**
 * Generate new webhook token
 * POST /api/webhook/token
 */
export async function generateWebhookToken(): Promise<string> {
  const response = await apiClient.post<ApiResponse<{ token: string }>>('/webhook/token');
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to generate webhook token');
  }
  return response.data.data?.token || '';
}

// ============ Template API ============

/**
 * Get all templates
 * GET /api/templates
 */
export async function getTemplates(): Promise<MessageTemplate[]> {
  const response = await apiClient.get<ApiResponse<MessageTemplate[]>>('/templates');
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to fetch templates');
  }
  return response.data.data || [];
}

/**
 * Create a new template
 * POST /api/templates
 */
export async function createTemplate(data: { key: string; templateId: string; name: string }): Promise<MessageTemplate> {
  const response = await apiClient.post<ApiResponse<MessageTemplate>>('/templates', data);
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to create template');
  }
  return response.data.data!;
}

/**
 * Delete a template
 * DELETE /api/templates/:id
 */
export async function deleteTemplate(id: number): Promise<void> {
  const response = await apiClient.delete<ApiResponse<void>>(`/templates/${id}`);
  if (!response.data.success) {
    throw new Error(response.data.error || 'Failed to delete template');
  }
}

// Export the axios instances for advanced usage
export { apiClient, authClient };
