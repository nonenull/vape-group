package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vape-group/backend/config"
	"github.com/vape-group/backend/internal/models"
	"gorm.io/gorm"
)

// TenantMiddleware 租户中间件，从请求中识别租户
func TenantMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipTenantLookup(c) {
			c.Next()
			return
		}

		domain := normalizeRequestDomain(c)
		tenant, matchedDomain, err := resolveTenantByDomain(db, domain)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status := http.StatusNotFound
				if c.Request.URL.Path == "/__tenant_host_check" {
					status = http.StatusForbidden
				}
				c.JSON(status, gin.H{"error": "Tenant not found"})
				c.Abort()
				return
			}
			c.JSON(500, gin.H{"error": "Tenant lookup failed"})
			c.Abort()
			return
		}

		c.Set("tenant_matched_domain", matchedDomain)
		c.Set("tenant_primary_domain", tenant.Domain)

		if shouldRedirectToPrimaryDomain(c, matchedDomain, tenant.Domain) {
			redirectURL := buildPrimaryDomainRedirect(c, tenant.Domain)
			c.Redirect(http.StatusMovedPermanently, redirectURL)
			c.Abort()
			return
		}

		c.Set("tenant_id", tenant.ID)
		c.Set("tenant", tenant)
		c.Next()
	}
}

func normalizeRequestDomain(c *gin.Context) string {
	domain := c.GetHeader("X-Tenant-Domain")
	if domain == "" {
		domain = c.Request.Host
	}
	if strings.Contains(domain, ":") {
		domain = strings.Split(domain, ":")[0]
	}
	return strings.TrimSpace(strings.ToLower(domain))
}

func resolveTenantByDomain(db *gorm.DB, domain string) (models.Tenant, string, error) {
	var tenant models.Tenant
	err := db.Where("domain = ?", domain).Take(&tenant).Error
	if err == nil {
		return tenant, domain, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Tenant{}, "", err
	}

	var tenants []models.Tenant
	if err := db.Find(&tenants).Error; err != nil {
		return models.Tenant{}, "", err
	}

	for _, candidate := range tenants {
		for _, alias := range jsonArrayToStrings(candidate.BoundDomains) {
			if strings.TrimSpace(strings.ToLower(alias)) == domain {
				return candidate, domain, nil
			}
		}
	}

	alternateDomain := alternateTenantDomain(domain)
	if alternateDomain == "" || alternateDomain == domain {
		return models.Tenant{}, "", gorm.ErrRecordNotFound
	}

	for _, candidate := range tenants {
		if strings.TrimSpace(strings.ToLower(candidate.Domain)) == alternateDomain {
			return candidate, domain, nil
		}
		for _, alias := range jsonArrayToStrings(candidate.BoundDomains) {
			if strings.TrimSpace(strings.ToLower(alias)) == alternateDomain {
				return candidate, domain, nil
			}
		}
	}

	return models.Tenant{}, "", gorm.ErrRecordNotFound
}

func alternateTenantDomain(domain string) string {
	normalized := strings.TrimSpace(strings.ToLower(domain))
	if normalized == "" {
		return ""
	}
	if strings.HasPrefix(normalized, "www.") {
		return strings.TrimPrefix(normalized, "www.")
	}
	return "www." + normalized
}

func jsonArrayToStrings(values models.JSONArray) []string {
	result := make([]string, 0, len(values))
	for _, item := range values {
		if value, ok := item.(string); ok && value != "" {
			result = append(result, value)
		}
	}
	return result
}

func buildPrimaryDomainRedirect(c *gin.Context, primaryDomain string) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	target := scheme + "://" + primaryDomain + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		target += "?" + c.Request.URL.RawQuery
	}
	return target
}

func shouldSkipTenantLookup(c *gin.Context) bool {
	path := c.Request.URL.Path
	return path == "/health" ||
		path == "/favicon.ico" ||
		strings.HasPrefix(path, "/uploads/") ||
		strings.HasPrefix(path, "/wp-content/uploads/") ||
		strings.HasPrefix(path, "/api/admin") ||
		strings.HasPrefix(path, "/api/auth")
}

func shouldRedirectToPrimaryDomain(c *gin.Context, matchedDomain, primaryDomain string) bool {
	return matchedDomain != "" &&
		primaryDomain != "" &&
		matchedDomain != primaryDomain &&
		c.Request.URL.Path != "/__tenant_host_check"
}

// CORSMiddleware CORS跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-Domain")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

type adminAuthClaims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

func CreateAdminToken(adminID uint, jwtSecret string, now time.Time) (string, error) {
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(adminAuthClaims{
		Sub: strconv.FormatUint(uint64(adminID), 10),
		Iat: now.Unix(),
		Exp: now.Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := header + "." + payload
	signature := signJWT(signingInput, jwtSecret)
	return signingInput + "." + signature, nil
}

func ParseAdminToken(token, jwtSecret string) (uint, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, errors.New("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := signJWT(signingInput, jwtSecret)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return 0, errors.New("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, err
	}

	var claims adminAuthClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return 0, err
	}
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return 0, errors.New("token expired")
	}

	adminID, err := strconv.ParseUint(strings.TrimSpace(claims.Sub), 10, 64)
	if err != nil || adminID == 0 {
		return 0, errors.New("invalid token subject")
	}
	return uint(adminID), nil
}

func signJWT(input, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		adminID, err := ParseAdminToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		var admin models.AdminUser
		if err := db.Where("id = ? AND is_active = ?", adminID, true).Take(&admin).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin account not found"})
			c.Abort()
			return
		}

		c.Set("admin_user", admin)
		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("admin_user"); !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
