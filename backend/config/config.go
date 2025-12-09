package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	ServerAddress      string
	DatabasePath       string
	OIDC               OIDCConfig
	WeChat             WeChatConfig
	SessionSecret      string
	CORSAllowedOrigins []string
	DevMode            bool // Skip authentication when true
}

// OIDCConfig holds OIDC provider configuration
type OIDCConfig struct {
	ProviderURL  string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// WeChatConfig holds WeChat API configuration
type WeChatConfig struct {
	AppID      string
	AppSecret  string
	TemplateID string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	oidcProviderURL := getEnv("OIDC_PROVIDER_URL", "")
	devMode := getEnv("DEV_MODE", "") == "true" || oidcProviderURL == ""

	cfg := &Config{
		ServerAddress:      getEnv("SERVER_ADDRESS", ":8080"),
		DatabasePath:       getEnv("DATABASE_PATH", "./data/notification.db"),
		SessionSecret:      getEnv("SESSION_SECRET", "default-secret-change-in-production"),
		CORSAllowedOrigins: parseCSV(getEnv("CORS_ALLOWED_ORIGINS", "*")),
		DevMode:            devMode,
		OIDC: OIDCConfig{
			ProviderURL:  oidcProviderURL,
			ClientID:     getEnv("OIDC_CLIENT_ID", ""),
			ClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("OIDC_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		},
		WeChat: WeChatConfig{
			AppID:      getEnv("WECHAT_APP_ID", ""),
			AppSecret:  getEnv("WECHAT_APP_SECRET", ""),
			TemplateID: getEnv("WECHAT_TEMPLATE_ID", ""),
		},
	}
	return cfg, nil
}

// parseCSV parses a comma-separated string into a slice of strings
func parseCSV(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
