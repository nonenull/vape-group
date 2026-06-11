package service

import "testing"

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
