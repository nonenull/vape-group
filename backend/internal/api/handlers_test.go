package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/vape-group/backend/config"
)

func TestFetchAllGoDaddyDomainsFollowsPagination(t *testing.T) {
	t.Helper()

	pageSize := 1000
	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.Header.Get("Authorization"); got != "sso-key key:secret" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		if r.URL.Path != "/v1/domains" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if got := query.Get("limit"); got != strconv.Itoa(pageSize) {
			t.Fatalf("unexpected limit query: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")

		switch query.Get("marker") {
		case "":
			writeJSON(t, w, buildDomainPage(pageSize, "domain"))
		case "domain-1000.example.com":
			writeJSON(t, w, []goDaddyDomainSummary{
				{Domain: "domain-1001.example.com", Expires: "2027-01-02T03:04:05Z"},
				{Domain: "domain-1002.example.com", Expires: "2027-01-03T03:04:05Z"},
			})
		default:
			t.Fatalf("unexpected marker query: %q", query.Get("marker"))
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		GoDaddyAPIKey:     "key",
		GoDaddyAPISecret:  "secret",
		GoDaddyAPIBaseURL: server.URL,
	}

	domains, err := fetchAllGoDaddyDomains(cfg)
	if err != nil {
		t.Fatalf("fetchAllGoDaddyDomains returned error: %v", err)
	}

	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
	if len(domains) != pageSize+2 {
		t.Fatalf("expected %d domains, got %d", pageSize+2, len(domains))
	}
	if domains[0].Domain != "domain-0001.example.com" {
		t.Fatalf("unexpected first domain: %q", domains[0].Domain)
	}
	if domains[len(domains)-1].Domain != "domain-1002.example.com" {
		t.Fatalf("unexpected last domain: %q", domains[len(domains)-1].Domain)
	}
}

func buildDomainPage(size int, prefix string) []goDaddyDomainSummary {
	domains := make([]goDaddyDomainSummary, 0, size)
	for i := 1; i <= size; i++ {
		domains = append(domains, goDaddyDomainSummary{
			Domain:  prefix + "-" + leftPad4(i) + ".example.com",
			Expires: "2027-01-01T03:04:05Z",
		})
	}
	return domains
}

func leftPad4(value int) string {
	switch {
	case value >= 1000:
		return strconv.Itoa(value)
	case value >= 100:
		return "0" + strconv.Itoa(value)
	case value >= 10:
		return "00" + strconv.Itoa(value)
	default:
		return "000" + strconv.Itoa(value)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("failed to encode response: %v", err)
	}
}

func TestFetchAllGoDaddyDomainsStopsOnShortPage(t *testing.T) {
	t.Helper()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.URL.Query().Get("marker"); got != "" {
			t.Fatalf("did not expect marker on short-page test, got %q", got)
		}
		writeJSON(t, w, []goDaddyDomainSummary{
			{Domain: "only.example.com", Expires: "2027-01-01T03:04:05Z"},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		GoDaddyAPIKey:     "key",
		GoDaddyAPISecret:  "secret",
		GoDaddyAPIBaseURL: server.URL,
	}

	domains, err := fetchAllGoDaddyDomains(cfg)
	if err != nil {
		t.Fatalf("fetchAllGoDaddyDomains returned error: %v", err)
	}

	if requests != 1 {
		t.Fatalf("expected 1 request, got %d", requests)
	}
	if len(domains) != 1 || domains[0].Domain != "only.example.com" {
		t.Fatalf("unexpected domains: %+v", domains)
	}
}

func TestFetchAllGoDaddyDomainsEncodesMarker(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			t.Fatalf("failed to parse query: %v", err)
		}
		if got := query.Get("limit"); got != "1000" {
			t.Fatalf("unexpected limit: %q", got)
		}
		writeJSON(t, w, []goDaddyDomainSummary{})
	}))
	defer server.Close()

	cfg := &config.Config{
		GoDaddyAPIKey:     "key",
		GoDaddyAPISecret:  "secret",
		GoDaddyAPIBaseURL: server.URL,
	}

	if _, err := fetchAllGoDaddyDomains(cfg); err != nil {
		t.Fatalf("fetchAllGoDaddyDomains returned error: %v", err)
	}
}
