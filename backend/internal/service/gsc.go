package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vape-group/backend/config"
)

type GSCService struct {
	tokenFile                string
	resourceSyncDelaySeconds time.Duration
	rateLimitRetries         int
	rateLimitBackoff         time.Duration
	client                   *http.Client
}

type GSCResult struct {
	SiteURL string `json:"site_url,omitempty"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type gscOAuthToken struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	TokenURI     string    `json:"token_uri"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	Expiry       time.Time `json:"expiry"`
}

type gscTokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func NewGSCService(cfg *config.Config) (*GSCService, error) {
	tokenFile := strings.TrimSpace(cfg.GSCOAuthTokenFile)
	if tokenFile == "" {
		return nil, errors.New("GSC_OAUTH_TOKEN_FILE is not configured")
	}

	delaySeconds, err := strconv.ParseFloat(strings.TrimSpace(cfg.GSCResourceSyncDelaySeconds), 64)
	if err != nil {
		delaySeconds = 6
	}

	retries, err := strconv.Atoi(strings.TrimSpace(cfg.GSCAPIRateLimitRetries))
	if err != nil || retries < 0 {
		retries = 3
	}

	backoffSeconds, err := strconv.ParseFloat(strings.TrimSpace(cfg.GSCAPIRateLimitBackoffSeconds), 64)
	if err != nil {
		backoffSeconds = 30
	}

	return &GSCService{
		tokenFile:                tokenFile,
		resourceSyncDelaySeconds: time.Duration(delaySeconds * float64(time.Second)),
		rateLimitRetries:         retries,
		rateLimitBackoff:         time.Duration(backoffSeconds * float64(time.Second)),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func normalizeGSCDomain(domain string) string {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	normalized = strings.TrimPrefix(normalized, "http://")
	normalized = strings.TrimPrefix(normalized, "https://")
	normalized = strings.TrimPrefix(normalized, "www.")
	normalized = strings.Trim(normalized, "/")
	return normalized
}

func gscSiteURLForDomain(domain string) string {
	normalized := normalizeGSCDomain(domain)
	if normalized == "" {
		return ""
	}
	return "https://www." + normalized + "/"
}

func (s *GSCService) EnsureSites(domains []string) []GSCResult {
	results := make([]GSCResult, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))

	for _, domain := range domains {
		siteURL := gscSiteURLForDomain(domain)
		if siteURL == "" {
			continue
		}
		if _, exists := seen[siteURL]; exists {
			continue
		}
		seen[siteURL] = struct{}{}

		result := s.ensureSite(siteURL)
		results = append(results, result)

		if result.Status == "rate_limited" {
			break
		}
		if s.resourceSyncDelaySeconds > 0 {
			time.Sleep(s.resourceSyncDelaySeconds)
		}
	}

	return results
}

func (s *GSCService) ensureSite(siteURL string) GSCResult {
	result := GSCResult{SiteURL: siteURL}
	token, err := s.loadToken()
	if err != nil {
		result.Status = "error"
		result.Message = err.Error()
		return result
	}

	endpoint := "https://www.googleapis.com/webmasters/v3/sites/" + url.PathEscape(siteURL)
	req, err := http.NewRequest(http.MethodPut, endpoint, nil)
	if err != nil {
		result.Status = "error"
		result.Message = err.Error()
		return result
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := s.doWithRateLimitRetry(req)
	if err != nil {
		result.Status = "error"
		result.Message = err.Error()
		return result
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyText := strings.TrimSpace(string(bodyBytes))

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		result.Status = "added"
		result.Message = "GSC resource added"
	case resp.StatusCode == http.StatusForbidden:
		result.Status = "exists_or_no_permission"
		if bodyText == "" {
			bodyText = "GSC site already exists or account has no permission"
		}
		result.Message = bodyText
	case resp.StatusCode == http.StatusTooManyRequests:
		result.Status = "rate_limited"
		if bodyText == "" {
			bodyText = "Search Console API rate limited"
		}
		result.Message = bodyText
	default:
		result.Status = "error"
		if bodyText == "" {
			bodyText = fmt.Sprintf("GSC request failed (%d)", resp.StatusCode)
		}
		result.Message = bodyText
	}

	return result
}

func (s *GSCService) doWithRateLimitRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	attempts := s.rateLimitRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 0; attempt < attempts; attempt++ {
		cloned := req.Clone(req.Context())
		resp, err := s.client.Do(cloned)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		} else {
			lastErr = fmt.Errorf("Search Console API rate limited")
			resp.Body.Close()
		}

		if attempt < attempts-1 && s.rateLimitBackoff > 0 {
			time.Sleep(s.rateLimitBackoff)
		}
	}

	return nil, lastErr
}

func (s *GSCService) loadToken() (*gscOAuthToken, error) {
	raw, err := os.ReadFile(s.tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read GSC token file: %w", err)
	}

	var token gscOAuthToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return nil, fmt.Errorf("failed to parse GSC token file: %w", err)
	}

	if strings.TrimSpace(token.Token) == "" {
		return nil, errors.New("GSC OAuth token is empty")
	}

	if tokenNeedsRefresh(token) {
		refreshed, err := s.refreshToken(token)
		if err != nil {
			return nil, err
		}
		token = *refreshed
	}

	return &token, nil
}

func tokenNeedsRefresh(token gscOAuthToken) bool {
	if token.Expiry.IsZero() {
		return false
	}
	return time.Now().Add(2 * time.Minute).After(token.Expiry)
}

func (s *GSCService) refreshToken(token gscOAuthToken) (*gscOAuthToken, error) {
	if strings.TrimSpace(token.RefreshToken) == "" {
		return nil, errors.New("GSC OAuth token expired and refresh_token is missing")
	}
	if strings.TrimSpace(token.TokenURI) == "" || strings.TrimSpace(token.ClientID) == "" || strings.TrimSpace(token.ClientSecret) == "" {
		return nil, errors.New("GSC OAuth token refresh configuration is incomplete")
	}

	form := url.Values{}
	form.Set("client_id", token.ClientID)
	form.Set("client_secret", token.ClientSecret)
	form.Set("refresh_token", token.RefreshToken)
	form.Set("grant_type", "refresh_token")

	req, err := http.NewRequest(http.MethodPost, token.TokenURI, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh GSC OAuth token: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to refresh GSC OAuth token (%d): %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var refreshed gscTokenRefreshResponse
	if err := json.Unmarshal(bodyBytes, &refreshed); err != nil {
		return nil, fmt.Errorf("failed to parse refreshed GSC OAuth token: %w", err)
	}
	if strings.TrimSpace(refreshed.AccessToken) == "" {
		return nil, errors.New("refreshed GSC OAuth token is empty")
	}

	token.Token = refreshed.AccessToken
	if strings.TrimSpace(refreshed.RefreshToken) != "" {
		token.RefreshToken = refreshed.RefreshToken
	}
	if refreshed.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second).UTC()
	}

	updatedRaw, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(s.tokenFile, updatedRaw, 0o600); err != nil {
		return nil, fmt.Errorf("failed to persist refreshed GSC OAuth token: %w", err)
	}

	return &token, nil
}
