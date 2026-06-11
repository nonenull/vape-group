package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCanonicalTenantDomainAddsWWWForApexDomain(t *testing.T) {
	if got := canonicalTenantDomain("elfbar-sp2s.shop"); got != "www.elfbar-sp2s.shop" {
		t.Fatalf("expected www apex domain, got %q", got)
	}
}

func TestCanonicalTenantDomainKeepsExistingWWW(t *testing.T) {
	if got := canonicalTenantDomain("www.elfbar-sp2s.shop"); got != "www.elfbar-sp2s.shop" {
		t.Fatalf("expected existing www domain to be preserved, got %q", got)
	}
}

func TestBuildPrimaryDomainRedirectUsesCanonicalPrimaryDomain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("GET", "http://elfbar-sp2s.shop/fuck/tenants?tab=all", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	c.Request = req

	redirectURL := buildPrimaryDomainRedirect(c, canonicalTenantDomain("elfbar-sp2s.shop"))

	if redirectURL != "https://www.elfbar-sp2s.shop/fuck/tenants?tab=all" {
		t.Fatalf("unexpected redirect URL: %q", redirectURL)
	}
}
