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
	AdminUsername    string
	AdminPassword    string
	AdminName        string
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
	NPMAPIURL        string
	NPMEmail         string
	NPMPassword      string
	ECPayLogisticsMerchantID string
	ECPayLogisticsHashKey    string
	ECPayLogisticsHashIV     string
	ECPayLogisticsSubType    string
	ECPayLogisticsStage      bool
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
		AdminUsername:    getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin123456"),
		AdminName:        getEnv("ADMIN_NAME", "Platform Admin"),
		UploadDir:        getEnv("UPLOAD_DIR", "uploads"),
		ProductListCacheSeconds: getEnv("PRODUCT_LIST_CACHE_SECONDS", "30"),
		DevTenantDomains: getEnvList("DEV_TENANT_DOMAINS", []string{}),
		DeepSeekAPIKey:   getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:  getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:    getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		GoDaddyAPIKey:    getEnv("GODADDY_API_KEY", ""),
		GoDaddyAPISecret: getEnv("GODADDY_API_SECRET", ""),
		GoDaddyAPIBaseURL: getEnv("GODADDY_API_BASE_URL", "https://api.godaddy.com"),
		DNSCheckServer:   getEnv("DNS_CHECK_SERVER", "168.95.1.1:53"),
		DNSBlockedIP:     getEnv("DNS_BLOCKED_IP", "182.173.0.181"),
		NPMAPIURL:        getEnv("NPM_API_URL", ""),
		NPMEmail:         getEnv("NPM_EMAIL", ""),
		NPMPassword:      getEnv("NPM_PASSWORD", ""),
		ECPayLogisticsMerchantID: getEnv("ECPAY_LOGISTICS_MERCHANT_ID", ""),
		ECPayLogisticsHashKey:    getEnv("ECPAY_LOGISTICS_HASH_KEY", ""),
		ECPayLogisticsHashIV:     getEnv("ECPAY_LOGISTICS_HASH_IV", ""),
		ECPayLogisticsSubType:    getEnv("ECPAY_LOGISTICS_SUB_TYPE", "UNIMART"),
		ECPayLogisticsStage:      getEnvBool("ECPAY_LOGISTICS_STAGE", true),
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

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}
