package config

import (
	"os"
	"strings"
)

type Config struct {
	Port             string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPass           string
	DBName           string
	JWTSecret        string
	UploadDir        string
	ProductListCacheSeconds string
	DevTenantDomains []string
	DeepSeekAPIKey   string
	DeepSeekBaseURL  string
	DeepSeekModel    string
	GoDaddyAPIKey    string
	GoDaddyAPISecret string
	GoDaddyAPIBaseURL string
	DNSCheckServer   string
	DNSBlockedIP     string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "3306"),
		DBUser:           getEnv("DB_USER", "root"),
		DBPass:           getEnv("DB_PASS", "password"),
		DBName:           getEnv("DB_NAME", "vape_group"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		UploadDir:        getEnv("UPLOAD_DIR", "uploads"),
		ProductListCacheSeconds: getEnv("PRODUCT_LIST_CACHE_SECONDS", "30"),
		DevTenantDomains: getEnvList("DEV_TENANT_DOMAINS", []string{"localhost", "127.0.0.1"}),
		DeepSeekAPIKey:   getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:  getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:    getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		GoDaddyAPIKey:    getEnv("GODADDY_API_KEY", ""),
		GoDaddyAPISecret: getEnv("GODADDY_API_SECRET", ""),
		GoDaddyAPIBaseURL: getEnv("GODADDY_API_BASE_URL", "https://api.godaddy.com"),
		DNSCheckServer:   getEnv("DNS_CHECK_SERVER", "168.95.1.1:53"),
		DNSBlockedIP:     getEnv("DNS_BLOCKED_IP", "182.173.0.181"),
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvList(key string, defaultValue []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}

	var result []string
	for _, part := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}

	return result
}
