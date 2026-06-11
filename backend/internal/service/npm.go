package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/vape-group/backend/config"
)

type NPMService struct {
	apiURL   string
	email    string
	password string
	client   *http.Client
	token    string
}

type NPMResult struct {
	Status         string        `json:"status"`
	Message        string        `json:"message,omitempty"`
	NPMUpdated     bool          `json:"npm_updated"`
	ProxyHostID    any           `json:"proxy_host_id,omitempty"`
	UpdatedDomains []string      `json:"updated_domains,omitempty"`
	SSLResult      *NPMSSLResult `json:"ssl_result,omitempty"`
}

type NPMSSLResult struct {
	Updated       bool   `json:"updated"`
	CertificateID any    `json:"certificate_id,omitempty"`
	Message       string `json:"message,omitempty"`
}

type npmTokenResponse struct {
	Token string `json:"token"`
}

type npmProxyHost struct {
	ID                    any              `json:"id"`
	DomainNames           []string         `json:"domain_names"`
	ForwardScheme         string           `json:"forward_scheme"`
	ForwardHost           string           `json:"forward_host"`
	ForwardPort           any              `json:"forward_port"`
	CertificateID         any              `json:"certificate_id"`
	SSLForced             bool             `json:"ssl_forced"`
	HTTP2Support          bool             `json:"http2_support"`
	HSTSEnabled           bool             `json:"hsts_enabled"`
	HSTSSubdomains        bool             `json:"hsts_subdomains"`
	BlockExploits         bool             `json:"block_exploits"`
	CachingEnabled        bool             `json:"caching_enabled"`
	AllowWebsocketUpgrade bool             `json:"allow_websocket_upgrade"`
	TrustForwardedProto   bool             `json:"trust_forwarded_proto"`
	AccessListID          any              `json:"access_list_id"`
	Meta                  map[string]any   `json:"meta"`
	Locations             []map[string]any `json:"locations"`
	AdvancedConfig        string           `json:"advanced_config"`
}

type npmPayloadVariant struct {
	name string
	data map[string]any
}

type npmCertificateResponse struct {
	ID any `json:"id"`
}

type NPMUpdateMode string

const (
	NPMUpdateModeReissueSSL NPMUpdateMode = "reissue_ssl"
	NPMUpdateModeKeepSSL    NPMUpdateMode = "keep_ssl"
)

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func NewNPMService(cfg *config.Config) (*NPMService, error) {
	if strings.TrimSpace(cfg.NPMAPIURL) == "" {
		return nil, errors.New("NPM_API_URL is not configured")
	}
	if strings.TrimSpace(cfg.NPMEmail) == "" || strings.TrimSpace(cfg.NPMPassword) == "" {
		return nil, errors.New("NPM credentials are not configured")
	}

	return &NPMService{
		apiURL:   strings.TrimRight(strings.TrimSpace(cfg.NPMAPIURL), "/"),
		email:    strings.TrimSpace(cfg.NPMEmail),
		password: strings.TrimSpace(cfg.NPMPassword),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *NPMService) authenticate() error {
	if s.token != "" {
		return nil
	}

	payload := map[string]string{
		"identity": s.email,
		"secret":   s.password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.apiURL+"/tokens", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate with NPM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		message, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to authenticate with NPM: %s", strings.TrimSpace(string(message)))
	}

	var tokenResp npmTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}
	if strings.TrimSpace(tokenResp.Token) == "" {
		return errors.New("NPM auth succeeded but token is empty")
	}

	s.token = tokenResp.Token
	return nil
}

func (s *NPMService) doJSON(method, path string, payload any, out any) error {
	if err := s.authenticate(); err != nil {
		return err
	}

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, s.apiURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		message, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("NPM request failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func normalizeNPMDomains(primaryDomain string, boundDomains []string) []string {
	seen := make(map[string]struct{}, (len(boundDomains)+1)*2)
	result := make([]string, 0, (len(boundDomains)+1)*2)
	for _, candidate := range append([]string{primaryDomain}, boundDomains...) {
		domain := strings.ToLower(strings.TrimSpace(candidate))
		if domain == "" {
			continue
		}
		for _, expanded := range expandNPMDomainAliases(domain) {
			if _, exists := seen[expanded]; exists {
				continue
			}
			seen[expanded] = struct{}{}
			result = append(result, expanded)
		}
	}
	return result
}

func expandNPMDomainAliases(domain string) []string {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	if normalized == "" {
		return nil
	}

	result := []string{normalized}
	if shouldAddWWWVariant(normalized) {
		result = append(result, "www."+normalized)
	}
	return result
}

func shouldAddWWWVariant(domain string) bool {
	if domain == "" {
		return false
	}
	if strings.HasPrefix(domain, "www.") {
		return false
	}
	if net.ParseIP(domain) != nil {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	if strings.HasSuffix(domain, ".localhost") || domain == "localhost" {
		return false
	}
	return true
}

func appendUniqueDomain(values []string, domain string) []string {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	if normalized == "" {
		return normalizeNPMDomains("", values)
	}

	result := normalizeNPMDomains("", values)
	for _, item := range result {
		if strings.EqualFold(item, normalized) {
			return result
		}
	}
	return append(result, normalized)
}

func (s *NPMService) getAllProxyHosts() ([]npmProxyHost, error) {
	var hosts []npmProxyHost
	if err := s.doJSON(http.MethodGet, "/nginx/proxy-hosts", nil, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (s *NPMService) getProxyHostByDomain(domain string) (*npmProxyHost, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return nil, nil
	}

	hosts, err := s.getAllProxyHosts()
	if err != nil {
		return nil, err
	}
	for _, host := range hosts {
		for _, item := range host.DomainNames {
			if strings.EqualFold(strings.TrimSpace(item), domain) {
				matched := host
				return &matched, nil
			}
		}
	}
	return nil, nil
}

func (s *NPMService) GetProxyHostByID(proxyHostID uint) (*npmProxyHost, error) {
	var host npmProxyHost
	if err := s.doJSON(http.MethodGet, fmt.Sprintf("/nginx/proxy-hosts/%d", proxyHostID), nil, &host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (s *NPMService) GetProxyHostForDomains(primaryDomain string, boundDomains []string) (*npmProxyHost, []string, error) {
	allDomains := normalizeNPMDomains(primaryDomain, boundDomains)
	for _, domain := range allDomains {
		host, err := s.getProxyHostByDomain(domain)
		if err != nil {
			return nil, allDomains, err
		}
		if host != nil {
			return host, allDomains, nil
		}
	}
	return nil, allDomains, nil
}

func ProxyHostContainsDomain(host *npmProxyHost, domain string) bool {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	if host == nil || normalized == "" {
		return false
	}
	for _, item := range host.DomainNames {
		if strings.EqualFold(strings.TrimSpace(item), normalized) {
			return true
		}
	}
	return false
}

func (s *NPMService) buildBaseUpdatePayload(host *npmProxyHost, domains []string, mode NPMUpdateMode, certificateOverride any) map[string]any {
	meta := host.Meta
	if meta == nil {
		meta = map[string]any{}
	}

	certificateID := host.CertificateID
	if certificateOverride != nil {
		certificateID = certificateOverride
	}
	if mode == NPMUpdateModeReissueSSL {
		certificateID = "new"
	}

	return map[string]any{
		"domain_names":            domains,
		"forward_scheme":          firstNonEmptyString(host.ForwardScheme, "http"),
		"forward_host":            host.ForwardHost,
		"forward_port":            host.ForwardPort,
		"certificate_id":          certificateID,
		"ssl_forced":              shouldForceSSL(host, mode, certificateOverride),
		"http2_support":           host.HTTP2Support,
		"hsts_enabled":            host.HSTSEnabled,
		"hsts_subdomains":         host.HSTSSubdomains,
		"block_exploits":          host.BlockExploits,
		"caching_enabled":         host.CachingEnabled,
		"allow_websocket_upgrade": host.AllowWebsocketUpgrade,
		"trust_forwarded_proto":   host.TrustForwardedProto,
		"access_list_id":          host.AccessListID,
		"meta":                    meta,
		"locations":               host.Locations,
		"advanced_config":         host.AdvancedConfig,
	}
}

func shouldForceSSL(host *npmProxyHost, mode NPMUpdateMode, certificateOverride any) bool {
	if host != nil && host.SSLForced {
		return true
	}
	if mode == NPMUpdateModeReissueSSL {
		return true
	}
	return certificateOverride != nil
}

func copyPayload(source map[string]any, keys ...string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, ok := source[key]; ok {
			result[key] = value
		}
	}
	return result
}

func (s *NPMService) buildUpdatePayloadVariants(host *npmProxyHost, domains []string, mode NPMUpdateMode, certificateOverride any) []npmPayloadVariant {
	base := s.buildBaseUpdatePayload(host, domains, mode, certificateOverride)
	return []npmPayloadVariant{
		{
			name: "full",
			data: base,
		},
		{
			name: "without-meta",
			data: copyPayload(
				base,
				"domain_names",
				"forward_scheme",
				"forward_host",
				"forward_port",
				"certificate_id",
				"ssl_forced",
				"http2_support",
				"hsts_enabled",
				"hsts_subdomains",
				"block_exploits",
				"caching_enabled",
				"allow_websocket_upgrade",
				"trust_forwarded_proto",
				"access_list_id",
				"advanced_config",
			),
		},
		{
			name: "compact",
			data: copyPayload(
				base,
				"domain_names",
				"forward_scheme",
				"forward_host",
				"forward_port",
				"certificate_id",
				"ssl_forced",
				"http2_support",
				"hsts_enabled",
				"hsts_subdomains",
				"block_exploits",
				"caching_enabled",
				"allow_websocket_upgrade",
			),
		},
		{
			name: "minimal",
			data: copyPayload(
				base,
				"domain_names",
				"forward_scheme",
				"forward_host",
				"forward_port",
				"certificate_id",
			),
		},
	}
}

func (s *NPMService) requestLetsEncryptCertificate(domains []string) (any, error) {
	payload := map[string]any{
		"provider":     "letsencrypt",
		"domain_names": domains,
		"meta": map[string]any{
			"letsencrypt_email": s.email,
			"letsencrypt_agree": true,
			"dns_challenge":     false,
		},
	}

	var response npmCertificateResponse
	if err := s.doJSON(http.MethodPost, "/nginx/certificates", payload, &response); err != nil {
		return nil, err
	}
	if response.ID == nil {
		return nil, errors.New("NPM certificate request succeeded but no certificate id was returned")
	}
	return response.ID, nil
}

func (s *NPMService) updateProxyHostWithVariants(host *npmProxyHost, domains []string, mode NPMUpdateMode, certificateOverride any, successMessage string, sslUpdated bool, sslMessage string) (*NPMResult, error) {
	targetDomains := normalizeNPMDomains("", domains)
	if len(targetDomains) == 0 {
		return &NPMResult{
			Status:     "warning",
			Message:    "No domains provided for NPM sync",
			NPMUpdated: false,
		}, nil
	}

	var updated map[string]any
	var lastErr error
	for index, variant := range s.buildUpdatePayloadVariants(host, targetDomains, mode, certificateOverride) {
		lastErr = s.doJSON(http.MethodPut, fmt.Sprintf("/nginx/proxy-hosts/%v", host.ID), variant.data, &updated)
		if lastErr == nil {
			message := successMessage
			if index > 0 {
				message = fmt.Sprintf("%s using %s payload", successMessage, variant.name)
			}
			return &NPMResult{
				Status:         "success",
				Message:        message,
				NPMUpdated:     true,
				ProxyHostID:    host.ID,
				UpdatedDomains: targetDomains,
				SSLResult: &NPMSSLResult{
					Updated:       sslUpdated,
					CertificateID: updated["certificate_id"],
					Message:       sslMessage,
				},
			}, nil
		}

		if !strings.Contains(lastErr.Error(), "additional properties") {
			break
		}
	}

	if lastErr != nil {
		return &NPMResult{
			Status:         "error",
			Message:        lastErr.Error(),
			NPMUpdated:     false,
			ProxyHostID:    host.ID,
			UpdatedDomains: targetDomains,
			SSLResult: &NPMSSLResult{
				Updated: false,
				Message: lastErr.Error(),
			},
		}, nil
	}

	return nil, errors.New("NPM update failed without an explicit error")
}

func (s *NPMService) updateSpecificProxyHostDomains(host *npmProxyHost, domains []string, mode NPMUpdateMode) (*NPMResult, error) {
	successMessage := "NPM proxy host updated successfully"
	sslUpdated := mode == NPMUpdateModeReissueSSL
	sslMessage := "Existing SSL certificate kept"
	if sslUpdated {
		sslMessage = "SSL certificate provisioned successfully"
	}

	result, err := s.updateProxyHostWithVariants(host, domains, mode, nil, successMessage, sslUpdated, sslMessage)
	if err != nil {
		return nil, err
	}
	if mode != NPMUpdateModeReissueSSL || result == nil || result.Status == "success" {
		return result, nil
	}

	// Fallback for older / stricter NPM installs: request the certificate first, then bind it.
	certificateID, certErr := s.requestLetsEncryptCertificate(normalizeNPMDomains("", domains))
	if certErr != nil {
		result.Message = result.Message + "; certificate fallback failed: " + certErr.Error()
		if result.SSLResult != nil {
			result.SSLResult.Message = result.Message
		}
		return result, nil
	}

	fallbackResult, fallbackErr := s.updateProxyHostWithVariants(
		host,
		domains,
		NPMUpdateModeKeepSSL,
		certificateID,
		"NPM proxy host updated successfully with separately issued certificate",
		true,
		"SSL certificate provisioned successfully via certificate endpoint",
	)
	if fallbackErr != nil {
		return nil, fallbackErr
	}
	return fallbackResult, nil
}

func (s *NPMService) updateProxyHostDomains(primaryDomain string, boundDomains []string, mode NPMUpdateMode) (*NPMResult, error) {
	host, allDomains, err := s.GetProxyHostForDomains(primaryDomain, boundDomains)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return &NPMResult{
			Status:         "warning",
			Message:        "No NPM proxy host found, only database was updated",
			NPMUpdated:     false,
			UpdatedDomains: allDomains,
		}, nil
	}
	return s.updateSpecificProxyHostDomains(host, allDomains, mode)
}

func (s *NPMService) UpdateDomainsAndSSL(primaryDomain string, boundDomains []string) (*NPMResult, error) {
	return s.updateProxyHostDomains(primaryDomain, boundDomains, NPMUpdateModeReissueSSL)
}

func (s *NPMService) UpdateDomainsKeepSSL(primaryDomain string, boundDomains []string) (*NPMResult, error) {
	return s.updateProxyHostDomains(primaryDomain, boundDomains, NPMUpdateModeKeepSSL)
}

func (s *NPMService) UpdateProxyHostDomainsByID(proxyHostID uint, primaryDomain string, boundDomains []string) (*NPMResult, error) {
	host, err := s.GetProxyHostByID(proxyHostID)
	if err != nil {
		return nil, err
	}
	return s.updateSpecificProxyHostDomains(host, normalizeNPMDomains(primaryDomain, boundDomains), NPMUpdateModeReissueSSL)
}

func (s *NPMService) EnsureDomainOnProxyHost(primaryDomain string, boundDomains []string, targetDomain string) (*NPMResult, bool, error) {
	host, lookupDomains, err := s.GetProxyHostForDomains(primaryDomain, boundDomains)
	if err != nil {
		return nil, false, err
	}
	if host == nil {
		return &NPMResult{
			Status:         "warning",
			Message:        "No NPM proxy host found for tenant domains",
			NPMUpdated:     false,
			UpdatedDomains: lookupDomains,
		}, false, nil
	}

	if ProxyHostContainsDomain(host, targetDomain) {
		return &NPMResult{
			Status:         "success",
			Message:        "Target domain already exists in NPM proxy host",
			NPMUpdated:     false,
			ProxyHostID:    host.ID,
			UpdatedDomains: host.DomainNames,
			SSLResult: &NPMSSLResult{
				Updated: false,
				Message: "NPM unchanged",
			},
		}, true, nil
	}

	result, err := s.updateSpecificProxyHostDomains(host, appendUniqueDomain(host.DomainNames, targetDomain), NPMUpdateModeReissueSSL)
	return result, false, err
}
