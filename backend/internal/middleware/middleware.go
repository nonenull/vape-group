package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
				c.JSON(404, gin.H{"error": "Tenant not found"})
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

	return models.Tenant{}, "", gorm.ErrRecordNotFound
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
	return path == "/" ||
		path == "/health" ||
		path == "/favicon.ico" ||
		strings.HasPrefix(path, "/uploads/") ||
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

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现JWT验证逻辑
		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现管理员权限检查
		c.Next()
	}
}
