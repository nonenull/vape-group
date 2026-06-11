package service

import "testing"

func TestNormalizeNPMDomainsAddsWWWVariant(t *testing.T) {
	domains := normalizeNPMDomains("example.com", []string{"shop.example.com"})

	expected := []string{
		"example.com",
		"www.example.com",
		"shop.example.com",
		"www.shop.example.com",
	}
	if len(domains) != len(expected) {
		t.Fatalf("expected %d domains, got %d: %#v", len(expected), len(domains), domains)
	}
	for index, item := range expected {
		if domains[index] != item {
			t.Fatalf("expected domains[%d] to be %q, got %q", index, item, domains[index])
		}
	}
}

func TestNormalizeNPMDomainsSkipsWWWForLocalhostAndIPs(t *testing.T) {
	domains := normalizeNPMDomains("tenant1.localhost", []string{"127.0.0.1", "www.example.com"})

	expected := []string{
		"tenant1.localhost",
		"127.0.0.1",
		"www.example.com",
	}
	if len(domains) != len(expected) {
		t.Fatalf("expected %d domains, got %d: %#v", len(expected), len(domains), domains)
	}
	for index, item := range expected {
		if domains[index] != item {
			t.Fatalf("expected domains[%d] to be %q, got %q", index, item, domains[index])
		}
	}
}

func TestBuildBaseUpdatePayloadForcesSSLWhenReissuingCertificate(t *testing.T) {
	service := &NPMService{}
	host := &npmProxyHost{
		ForwardHost: "frontend",
		ForwardPort: 3000,
	}

	payload := service.buildBaseUpdatePayload(host, []string{"example.com"}, NPMUpdateModeReissueSSL, nil)

	if payload["certificate_id"] != "new" {
		t.Fatalf("expected certificate_id to be %q, got %#v", "new", payload["certificate_id"])
	}
	if forced, ok := payload["ssl_forced"].(bool); !ok || !forced {
		t.Fatalf("expected ssl_forced=true when reissuing certificate, got %#v", payload["ssl_forced"])
	}
}

func TestBuildBaseUpdatePayloadKeepsSSLRedirectOffWithoutOverride(t *testing.T) {
	service := &NPMService{}
	host := &npmProxyHost{
		ForwardHost: "frontend",
		ForwardPort: 3000,
		SSLForced:   false,
	}

	payload := service.buildBaseUpdatePayload(host, []string{"example.com"}, NPMUpdateModeKeepSSL, nil)

	if forced, ok := payload["ssl_forced"].(bool); !ok || forced {
		t.Fatalf("expected ssl_forced=false when keeping SSL without override, got %#v", payload["ssl_forced"])
	}
}

func TestBuildBaseUpdatePayloadForcesSSLWithCertificateOverride(t *testing.T) {
	service := &NPMService{}
	host := &npmProxyHost{
		ForwardHost: "frontend",
		ForwardPort: 3000,
	}

	payload := service.buildBaseUpdatePayload(host, []string{"example.com"}, NPMUpdateModeKeepSSL, 42)

	if payload["certificate_id"] != 42 {
		t.Fatalf("expected certificate_id override to be preserved, got %#v", payload["certificate_id"])
	}
	if forced, ok := payload["ssl_forced"].(bool); !ok || !forced {
		t.Fatalf("expected ssl_forced=true when certificate override is present, got %#v", payload["ssl_forced"])
	}
}
