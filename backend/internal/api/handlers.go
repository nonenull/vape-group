package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
	"github.com/vape-group/backend/config"
	"github.com/vape-group/backend/internal/middleware"
	"github.com/vape-group/backend/internal/models"
	"github.com/vape-group/backend/internal/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type tenantPayload struct {
	Domain            string              `json:"domain"`
	BoundDomains      []string            `json:"bound_domains"`
	NPMProxyHostID    *uint               `json:"npm_proxy_host_id"`
	Name              string              `json:"name"`
	IsActive          bool                `json:"is_active"`
	Theme             string              `json:"theme"`
	HomeTemplate      string              `json:"home_template"`
	HomeModuleOrder   []string            `json:"home_module_order"`
	HomeBanner        homeBannerConfig    `json:"home_banner"`
	HomeSections      []homeSectionConfig `json:"home_sections"`
	PrimaryBrandID    *uint               `json:"primary_brand_id"`
	PreviewImage      string              `json:"preview_image"`
	LogoImage         string              `json:"logo_image"`
	AccentColor       string              `json:"accent_color"`
	AccentStrongColor string              `json:"accent_strong_color"`
	SurfaceColor      string              `json:"surface_color"`
	PageBgColor       string              `json:"page_bg_color"`
	CardBgColor       string              `json:"card_bg_color"`
	TextColor         string              `json:"text_color"`
	MutedTextColor    string              `json:"muted_text_color"`
	BorderColor       string              `json:"border_color"`
	HeroBgColor       string              `json:"hero_bg_color"`
	TagBgColor        string              `json:"tag_bg_color"`
	HeroTitle         string              `json:"hero_title"`
	Tagline           string              `json:"tagline"`
	Announcement      string              `json:"announcement"`
	SupportText       string              `json:"support_text"`
	SEOTitle          string              `json:"seo_title"`
	SEODescription    string              `json:"seo_description"`
}

type tenantResponse struct {
	ID                uint                `json:"id"`
	Domain            string              `json:"domain"`
	BoundDomains      []string            `json:"bound_domains"`
	NPMProxyHostID    *uint               `json:"npm_proxy_host_id"`
	Name              string              `json:"name"`
	IsActive          bool                `json:"is_active"`
	Theme             string              `json:"theme"`
	HomeTemplate      string              `json:"home_template"`
	HomeModuleOrder   []string            `json:"home_module_order"`
	HomeBanner        homeBannerConfig    `json:"home_banner"`
	HomeSections      []homeSectionConfig `json:"home_sections"`
	PrimaryBrandID    *uint               `json:"primary_brand_id"`
	PreviewImage      string              `json:"preview_image"`
	LogoImage         string              `json:"logo_image"`
	AccentColor       string              `json:"accent_color"`
	AccentStrongColor string              `json:"accent_strong_color"`
	SurfaceColor      string              `json:"surface_color"`
	PageBgColor       string              `json:"page_bg_color"`
	CardBgColor       string              `json:"card_bg_color"`
	TextColor         string              `json:"text_color"`
	MutedTextColor    string              `json:"muted_text_color"`
	BorderColor       string              `json:"border_color"`
	HeroBgColor       string              `json:"hero_bg_color"`
	TagBgColor        string              `json:"tag_bg_color"`
	HeroTitle         string              `json:"hero_title"`
	Tagline           string              `json:"tagline"`
	Announcement      string              `json:"announcement"`
	SupportText       string              `json:"support_text"`
	SEOTitle          string              `json:"seo_title"`
	SEODescription    string              `json:"seo_description"`
}

type tenantDomainPayload struct {
	Domain string `json:"domain"`
}

type tenantDomainOperationResponse struct {
	TenantResponse tenantResponse      `json:"tenant"`
	NPMResult      *service.NPMResult  `json:"npm_result,omitempty"`
	GSCResults     []service.GSCResult `json:"gsc_results,omitempty"`
}

type homeBannerConfig struct {
	Enabled    bool   `json:"enabled"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Image      string `json:"image"`
	Link       string `json:"link"`
	ButtonText string `json:"button_text"`
}

type homeSectionConfig struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Title   string `json:"title"`
	Limit   int    `json:"limit"`
}

type platformConfigPayload struct {
	LineContactURL        string  `json:"line_contact_url"`
	FaqHTML               string  `json:"faq_html"`
	ShippingFee           float64 `json:"shipping_fee"`
	FreeShippingThreshold float64 `json:"free_shipping_threshold"`
	FeaturedCategoryIDs   []uint  `json:"featured_category_ids"`
	FeaturedBrandIDs      []uint  `json:"featured_brand_ids"`
}

type platformConfigResponse struct {
	ID                    uint    `json:"id"`
	LineContactURL        string  `json:"line_contact_url"`
	FaqHTML               string  `json:"faq_html"`
	ShippingFee           float64 `json:"shipping_fee"`
	FreeShippingThreshold float64 `json:"free_shipping_threshold"`
	FeaturedCategoryIDs   []uint  `json:"featured_category_ids"`
	FeaturedBrandIDs      []uint  `json:"featured_brand_ids"`
}

type domainPayload struct {
	DomainName string  `json:"domain_name"`
	Registrar  string  `json:"registrar"`
	ExpireDate *string `json:"expire_date"`
}

type dnsRecordPayload struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

type domainSyncPayload struct {
	IP *string `json:"ip"`
}

type domainResponse struct {
	ID            uint               `json:"id"`
	DomainName    string             `json:"domain_name"`
	Registrar     string             `json:"registrar"`
	ExpireDate    *string            `json:"expire_date,omitempty"`
	DNSRecords    []dnsRecordPayload `json:"dns_records"`
	IsBlocked     bool               `json:"is_blocked"`
	LastCheckIP   *string            `json:"last_check_ip,omitempty"`
	LastCheckedAt *time.Time         `json:"last_checked_at,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type productPayload struct {
	SKU               string                      `json:"sku"`
	BaseName          string                      `json:"base_name"`
	BasePrice         float64                     `json:"base_price"`
	BaseStockQuantity int                         `json:"base_stock_quantity"`
	Category          string                      `json:"category"`
	CategoryID        *uint                       `json:"category_id"`
	CategoryIDs       []uint                      `json:"category_ids"`
	Brand             string                      `json:"brand"`
	BrandID           *uint                       `json:"brand_id"`
	PreviewImage      string                      `json:"preview_image"`
	Gallery           []string                    `json:"gallery"`
	DetailImages      []string                    `json:"detail_images"`
	Status            string                      `json:"status"`
	IsActive          bool                        `json:"is_active"`
	Description       string                      `json:"description"`
	LongDescription   string                      `json:"long_description"`
	SpecificationHTML string                      `json:"specification_html"`
	Badge             string                      `json:"badge"`
	Rating            float64                     `json:"rating"`
	Reviews           int                         `json:"reviews"`
	Flavors           []string                    `json:"flavors"`
	Variants          []productVariantPayload     `json:"variants"`
	OptionGroups      []productOptionGroupPayload `json:"option_groups"`
	SkuVariants       []productSkuVariantPayload  `json:"sku_variants"`
}

func getOrCreatePlatformConfig(db *gorm.DB) (models.PlatformConfig, error) {
	var config models.PlatformConfig
	err := db.First(&config).Error
	if err == nil {
		return config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return config, err
	}

	config = models.PlatformConfig{
		LineContactURL:        "",
		FaqHTML:               "",
		ShippingFee:           90,
		FreeShippingThreshold: 1200,
		FeaturedCategoryIDs:   models.UIntArray{},
		FeaturedBrandIDs:      models.UIntArray{},
	}
	if err := db.Create(&config).Error; err != nil {
		return config, err
	}
	return config, nil
}

func platformConfigToResponse(config models.PlatformConfig) platformConfigResponse {
	return platformConfigResponse{
		ID:                    config.ID,
		LineContactURL:        config.LineContactURL,
		FaqHTML:               config.FaqHTML,
		ShippingFee:           config.ShippingFee,
		FreeShippingThreshold: config.FreeShippingThreshold,
		FeaturedCategoryIDs:   []uint(config.FeaturedCategoryIDs),
		FeaturedBrandIDs:      []uint(config.FeaturedBrandIDs),
	}
}

func domainToResponse(domain models.Domain) domainResponse {
	var expireDate *string
	if domain.ExpireDate != nil {
		formatted := domain.ExpireDate.Format("2006-01-02")
		expireDate = &formatted
	}

	dnsRecords := make([]dnsRecordPayload, 0, len(domain.DNSRecords))
	for _, item := range domain.DNSRecords {
		recordMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ttl := 600
		switch value := recordMap["ttl"].(type) {
		case float64:
			ttl = int(value)
		case int:
			ttl = value
		}
		dnsRecords = append(dnsRecords, dnsRecordPayload{
			Type: firstNonEmptyString(fmt.Sprintf("%v", recordMap["type"])),
			Name: firstNonEmptyString(fmt.Sprintf("%v", recordMap["name"])),
			Data: firstNonEmptyString(fmt.Sprintf("%v", recordMap["data"])),
			TTL:  ttl,
		})
	}

	return domainResponse{
		ID:            domain.ID,
		DomainName:    domain.DomainName,
		Registrar:     domain.Registrar,
		ExpireDate:    expireDate,
		DNSRecords:    dnsRecords,
		IsBlocked:     domain.IsBlocked,
		LastCheckIP:   domain.LastCheckIP,
		LastCheckedAt: domain.LastCheckedAt,
		CreatedAt:     domain.CreatedAt,
		UpdatedAt:     domain.UpdatedAt,
	}
}

func parseOptionalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func dnsRecordsToJSONArray(records []dnsRecordPayload) models.JSONArray {
	result := make(models.JSONArray, 0, len(records))
	for _, record := range records {
		result = append(result, models.JSONMap{
			"type": record.Type,
			"name": record.Name,
			"data": record.Data,
			"ttl":  record.TTL,
		})
	}
	return result
}

func godaddyRequest[T any](cfg *config.Config, method, path string, body io.Reader) (T, error) {
	var result T
	if strings.TrimSpace(cfg.GoDaddyAPIKey) == "" || strings.TrimSpace(cfg.GoDaddyAPISecret) == "" {
		return result, errors.New("GoDaddy API credentials are not configured")
	}

	req, err := http.NewRequest(method, strings.TrimRight(cfg.GoDaddyAPIBaseURL, "/")+path, body)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "sso-key "+cfg.GoDaddyAPIKey+":"+cfg.GoDaddyAPISecret)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		payload, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("GoDaddy API request failed: %s", strings.TrimSpace(string(payload)))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil && !errors.Is(err, io.EOF) {
		return result, err
	}
	return result, nil
}

func resolveDomainIP(serverAddr, domainName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, "udp", serverAddr)
		},
	}

	ips, err := resolver.LookupHost(ctx, domainName)
	if err != nil || len(ips) == 0 {
		return "", err
	}
	return ips[0], nil
}

type productVariantPayload struct {
	Name string `json:"name"`
	SKU  string `json:"sku"`
}

type productOptionGroupPayload struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type productSkuVariantPayload struct {
	SKU        string            `json:"sku"`
	Price      *float64          `json:"price"`
	Stock      *int              `json:"stock"`
	Selections map[string]string `json:"selections"`
}

type productResponse struct {
	ID                  uint                        `json:"id"`
	SKU                 string                      `json:"sku"`
	Slug                string                      `json:"slug"`
	BaseName            string                      `json:"base_name"`
	BasePrice           float64                     `json:"base_price"`
	BaseStockQuantity   int                         `json:"base_stock_quantity"`
	BaseImages          []string                    `json:"base_images"`
	DetailImages        []string                    `json:"detail_images"`
	Specifications      map[string]interface{}      `json:"specifications"`
	IsActive            bool                        `json:"is_active"`
	Category            string                      `json:"category"`
	CategoryID          *uint                       `json:"category_id,omitempty"`
	CategoryIDs         []uint                      `json:"category_ids"`
	Brand               string                      `json:"brand"`
	BrandID             *uint                       `json:"brand_id,omitempty"`
	PreviewImage        string                      `json:"preview_image"`
	Gallery             []string                    `json:"gallery"`
	Status              string                      `json:"status"`
	Description         string                      `json:"description"`
	LongDescription     string                      `json:"long_description"`
	SpecificationHTML   string                      `json:"specification_html"`
	Badge               string                      `json:"badge"`
	Rating              float64                     `json:"rating"`
	Reviews             int                         `json:"reviews"`
	Flavors             []string                    `json:"flavors"`
	Variants            []productVariantPayload     `json:"variants"`
	OptionGroups        []productOptionGroupPayload `json:"option_groups"`
	SkuVariants         []productSkuVariantPayload  `json:"sku_variants"`
	CustomName          *string                     `json:"custom_name,omitempty"`
	CustomDescription   *string                     `json:"custom_description,omitempty"`
	CustomPrice         *float64                    `json:"custom_price,omitempty"`
	CustomStockQuantity *int                        `json:"custom_stock_quantity,omitempty"`
	CustomImages        []string                    `json:"custom_images,omitempty"`
	CustomDetailImages  []string                    `json:"custom_detail_images,omitempty"`
	SEOTitle            *string                     `json:"seo_title,omitempty"`
	SEODescription      *string                     `json:"seo_description,omitempty"`
	IsVisible           bool                        `json:"is_visible"`
	CreatedAt           interface{}                 `json:"created_at"`
	UpdatedAt           interface{}                 `json:"updated_at"`
}

type productOverridePayload struct {
	CustomName          string   `json:"custom_name"`
	CustomDescription   string   `json:"custom_description"`
	CustomPrice         *float64 `json:"custom_price"`
	CustomStockQuantity *int     `json:"custom_stock_quantity"`
	CustomImages        []string `json:"custom_images"`
	CustomDetailImages  []string `json:"custom_detail_images"`
	SEOTitle            string   `json:"seo_title"`
	SEODescription      string   `json:"seo_description"`
	IsVisible           bool     `json:"is_visible"`
}

type generateOverrideDraftPayload struct {
	Instruction string `json:"instruction"`
}

type generatedOverrideDraftResponse struct {
	CustomName        string `json:"custom_name"`
	CustomDescription string `json:"custom_description"`
	SEOTitle          string `json:"seo_title"`
	SEODescription    string `json:"seo_description"`
}

type bulkGenerateOverrideNamesPayload struct {
	TenantID    uint   `json:"tenant_id"`
	ProductIDs  []uint `json:"product_ids"`
	Instruction string `json:"instruction"`
}

type generatedOverrideNameResponse struct {
	ProductID  uint   `json:"product_id"`
	TenantID   uint   `json:"tenant_id"`
	CustomName string `json:"custom_name"`
}

type deepSeekChatCompletionRequest struct {
	Model          string                          `json:"model"`
	Messages       []deepSeekChatCompletionMessage `json:"messages"`
	ResponseFormat map[string]string               `json:"response_format,omitempty"`
	Temperature    float64                         `json:"temperature,omitempty"`
}

type deepSeekChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type orderItemPayload struct {
	ProductID   uint   `json:"product_id"`
	VariantName string `json:"variant_name"`
	VariantSKU  string `json:"variant_sku"`
	Quantity    int    `json:"quantity"`
}

type createOrderPayload struct {
	Items            []orderItemPayload `json:"items"`
	LineID           string             `json:"line_id"`
	Phone            string             `json:"phone"`
	ConvenienceStore string             `json:"convenience_store"`
	ShippingAddress  string             `json:"shipping_address"`
	PaymentMethod    string             `json:"payment_method"`
}

type orderItemResponse struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	Name        string  `json:"name"`
	VariantName string  `json:"variant_name"`
	VariantSKU  string  `json:"variant_sku"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type productListResponse struct {
	Data  []productResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

type productListCacheEntry struct {
	value     productListResponse
	expiresAt time.Time
}

type productListCache struct {
	ttl   time.Duration
	mutex sync.RWMutex
	items map[string]productListCacheEntry
}

func newProductListCache(ttl time.Duration) *productListCache {
	return &productListCache{
		ttl:   ttl,
		items: make(map[string]productListCacheEntry),
	}
}

func (c *productListCache) get(key string) (productListResponse, bool) {
	if c == nil || c.ttl <= 0 {
		return productListResponse{}, false
	}

	c.mutex.RLock()
	entry, exists := c.items[key]
	c.mutex.RUnlock()
	if !exists {
		return productListResponse{}, false
	}
	if time.Now().After(entry.expiresAt) {
		c.mutex.Lock()
		delete(c.items, key)
		c.mutex.Unlock()
		return productListResponse{}, false
	}
	return entry.value, true
}

func (c *productListCache) set(key string, value productListResponse) {
	if c == nil || c.ttl <= 0 {
		return
	}

	c.mutex.Lock()
	c.items[key] = productListCacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mutex.Unlock()
}

func (c *productListCache) clear() {
	if c == nil {
		return
	}

	c.mutex.Lock()
	c.items = make(map[string]productListCacheEntry)
	c.mutex.Unlock()
}

type categoryDescendantCache struct {
	mutex sync.RWMutex
	items map[uint]map[uint]struct{}
}

var sharedProductListCache *productListCache
var sharedCategoryDescendantCache *categoryDescendantCache

func newCategoryDescendantCache() *categoryDescendantCache {
	return &categoryDescendantCache{
		items: make(map[uint]map[uint]struct{}),
	}
}

func registerCatalogCaches(productCache *productListCache, categoryCache *categoryDescendantCache) {
	sharedProductListCache = productCache
	sharedCategoryDescendantCache = categoryCache
}

func invalidateCatalogCaches() {
	if sharedProductListCache != nil {
		sharedProductListCache.clear()
	}
}

func invalidateCategoryCaches() {
	invalidateCatalogCaches()
	if sharedCategoryDescendantCache != nil {
		sharedCategoryDescendantCache.clear()
	}
}

func (c *categoryDescendantCache) get(categoryID uint) (map[uint]struct{}, bool) {
	if c == nil {
		return nil, false
	}
	c.mutex.RLock()
	value, exists := c.items[categoryID]
	c.mutex.RUnlock()
	return value, exists
}

func (c *categoryDescendantCache) set(categoryID uint, value map[uint]struct{}) {
	if c == nil {
		return
	}
	c.mutex.Lock()
	c.items[categoryID] = value
	c.mutex.Unlock()
}

func (c *categoryDescendantCache) clear() {
	if c == nil {
		return
	}
	c.mutex.Lock()
	c.items = make(map[uint]map[uint]struct{})
	c.mutex.Unlock()
}

func buildProductListCacheKey(tenantID uint, page, limit int, keyword, category, brand, sortBy string) string {
	return strings.Join([]string{
		strconv.FormatUint(uint64(tenantID), 10),
		strconv.Itoa(page),
		strconv.Itoa(limit),
		strings.TrimSpace(keyword),
		strings.TrimSpace(category),
		strings.TrimSpace(brand),
		strings.TrimSpace(sortBy),
	}, "|")
}

type orderResponse struct {
	ID               uint                `json:"id"`
	TenantID         uint                `json:"tenant_id"`
	UserID           uint                `json:"user_id"`
	TotalAmount      float64             `json:"total_amount"`
	Status           string              `json:"status"`
	LineID           string              `json:"line_id"`
	Phone            string              `json:"phone"`
	ConvenienceStore string              `json:"convenience_store"`
	ShippingAddress  string              `json:"shipping_address"`
	PaymentMethod    string              `json:"payment_method"`
	Items            []orderItemResponse `json:"items"`
	CreatedAt        interface{}         `json:"created_at"`
	UpdatedAt        interface{}         `json:"updated_at"`
}

type categoryPayload struct {
	Name      string `json:"name"`
	ParentID  *uint  `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

type categoryResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ParentID  *uint  `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

const defaultUncategorizedCategoryName = "未分類"
const defaultUncategorizedCategoryCode = "uncategorized"

type brandPayload struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}

type brandResponse struct {
	ID          uint   `json:"id"`
	TenantID    uint   `json:"tenant_id"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}

type adminLoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type adminUserResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	IsActive    bool   `json:"is_active"`
	LastLoginAt string `json:"last_login_at,omitempty"`
}

type adminLoginResponse struct {
	Token string            `json:"token"`
	User  adminUserResponse `json:"user"`
}

type uploadImageResponse struct {
	URL string `json:"url"`
}

type bulkProductUpdatePayload struct {
	ProductIDs []uint  `json:"product_ids"`
	Status     *string `json:"status"`
	IsActive   *bool   `json:"is_active"`
}

var uploadFilenameUnsafeChars = regexp.MustCompile(`[^a-z0-9]+`)

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func sanitizeUploadFilename(originalName string) string {
	baseName := strings.TrimSpace(strings.TrimSuffix(originalName, filepath.Ext(originalName)))
	if baseName == "" {
		return "image"
	}

	segments := make([]string, 0, len(baseName))
	var asciiBuffer strings.Builder
	addASCIIBuffer := func() {
		if asciiBuffer.Len() == 0 {
			return
		}
		segments = append(segments, asciiBuffer.String())
		asciiBuffer.Reset()
	}

	pinyinArgs := pinyin.NewArgs()
	pinyinArgs.Style = pinyin.Normal

	for _, r := range baseName {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if r <= unicode.MaxASCII {
				asciiBuffer.WriteRune(unicode.ToLower(r))
				continue
			}

			addASCIIBuffer()
			syllables := pinyin.Pinyin(string(r), pinyinArgs)
			if len(syllables) > 0 && len(syllables[0]) > 0 {
				segments = append(segments, strings.ToLower(syllables[0][0]))
			}
		default:
			addASCIIBuffer()
		}
	}

	addASCIIBuffer()
	baseName = strings.Join(segments, "-")
	baseName = uploadFilenameUnsafeChars.ReplaceAllString(baseName, "-")
	baseName = strings.Trim(baseName, "-")
	if baseName == "" {
		return "image"
	}
	return baseName
}

func sanitizeSlug(value string) string {
	result := sanitizeUploadFilename(value)
	if result == "" || result == "image" {
		return ""
	}
	return result
}

func generateProductSlug(db *gorm.DB, productID uint, baseName string, sku string) (string, error) {
	baseSlug := sanitizeSlug(baseName)
	if baseSlug == "" {
		baseSlug = sanitizeSlug(sku)
	}
	if baseSlug == "" {
		if productID > 0 {
			baseSlug = "product-" + strconv.FormatUint(uint64(productID), 10)
		} else {
			baseSlug = "product"
		}
	}
	slug := baseSlug
	suffix := 2

	for {
		var count int64
		query := db.Model(&models.Product{}).Where("slug = ?", slug)
		if productID > 0 {
			query = query.Where("id <> ?", productID)
		}
		if err := query.Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		slug = baseSlug + "-" + strconv.Itoa(suffix)
		suffix++
	}
}

func GenerateProductSlugForModel(db *gorm.DB, productID uint, baseName string, sku string) (string, error) {
	return generateProductSlug(db, productID, baseName, sku)
}

func normalizeDomain(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		parts := strings.SplitN(value, "://", 2)
		value = parts[1]
	}
	if strings.Contains(value, "/") {
		value = strings.SplitN(value, "/", 2)[0]
	}
	if strings.Contains(value, ":") {
		value = strings.SplitN(value, ":", 2)[0]
	}
	return value
}

func normalizeDomainList(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		domain := normalizeDomain(value)
		if domain == "" {
			continue
		}
		if _, exists := seen[domain]; exists {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	return result
}

func jsonString(source models.JSONMap, key, fallback string) string {
	if source == nil {
		return fallback
	}
	if value, ok := source[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

func jsonFloat(source models.JSONMap, key string, fallback float64) float64 {
	if source == nil {
		return fallback
	}
	switch value := source[key].(type) {
	case float64:
		return value
	case int:
		return float64(value)
	}
	return fallback
}

func jsonInt(source models.JSONMap, key string, fallback int) int {
	if source == nil {
		return fallback
	}
	switch value := source[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	}
	return fallback
}

func jsonUint(source models.JSONMap, key string) *uint {
	if source == nil {
		return nil
	}
	switch value := source[key].(type) {
	case float64:
		if value <= 0 {
			return nil
		}
		converted := uint(value)
		return &converted
	case int:
		if value <= 0 {
			return nil
		}
		converted := uint(value)
		return &converted
	case uint:
		if value == 0 {
			return nil
		}
		converted := value
		return &converted
	case *uint:
		if value == nil || *value == 0 {
			return nil
		}
		converted := *value
		return &converted
	case *int:
		if value == nil || *value <= 0 {
			return nil
		}
		converted := uint(*value)
		return &converted
	}
	return nil
}

func jsonUintSlice(source models.JSONMap, key string, fallback []uint) []uint {
	if source == nil {
		return fallback
	}

	raw, exists := source[key]
	if !exists || raw == nil {
		return fallback
	}

	var values []uint
	switch typed := raw.(type) {
	case []interface{}:
		values = make([]uint, 0, len(typed))
		for _, item := range typed {
			switch value := item.(type) {
			case float64:
				if value > 0 {
					values = append(values, uint(value))
				}
			case int:
				if value > 0 {
					values = append(values, uint(value))
				}
			case uint:
				if value > 0 {
					values = append(values, value)
				}
			}
		}
	case []uint:
		values = append(values, typed...)
	default:
		return fallback
	}

	if len(values) == 0 {
		return fallback
	}

	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	if len(result) == 0 {
		return fallback
	}
	return result
}

func jsonStringSlice(source models.JSONMap, key string, fallback []string) []string {
	if source == nil {
		return fallback
	}
	raw, ok := source[key].([]interface{})
	if !ok {
		return fallback
	}

	result := make([]string, 0, len(raw))
	for _, item := range raw {
		if value, ok := item.(string); ok && value != "" {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func jsonBool(source models.JSONMap, key string, fallback bool) bool {
	if source == nil {
		return fallback
	}
	switch value := source[key].(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	return fallback
}

func jsonMap(source models.JSONMap, key string) models.JSONMap {
	if source == nil {
		return nil
	}
	raw, exists := source[key]
	if !exists || raw == nil {
		return nil
	}
	if typed, ok := raw.(map[string]interface{}); ok {
		return models.JSONMap(typed)
	}
	if typed, ok := raw.(models.JSONMap); ok {
		return typed
	}
	return nil
}

func jsonSectionConfigs(source models.JSONMap, key string) []homeSectionConfig {
	if source == nil {
		return []homeSectionConfig{}
	}
	raw, ok := source[key].([]interface{})
	if !ok {
		return []homeSectionConfig{}
	}

	result := make([]homeSectionConfig, 0, len(raw))
	for index, item := range raw {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		config := homeSectionConfig{
			ID:      firstNonEmptyString(jsonString(models.JSONMap(entry), "id", "")),
			Type:    strings.TrimSpace(jsonString(models.JSONMap(entry), "type", "")),
			Enabled: jsonBool(models.JSONMap(entry), "enabled", true),
			Title:   strings.TrimSpace(jsonString(models.JSONMap(entry), "title", "")),
			Limit:   jsonInt(models.JSONMap(entry), "limit", 0),
		}

		if config.Type == "" {
			continue
		}
		if config.ID == "" {
			config.ID = fmt.Sprintf("%s-%d", config.Type, index+1)
		}
		if config.Limit < 0 {
			config.Limit = 0
		}
		result = append(result, config)
	}
	return result
}

func normalizeHomeBannerConfig(input homeBannerConfig) homeBannerConfig {
	return homeBannerConfig{
		Enabled:    input.Enabled,
		Title:      strings.TrimSpace(input.Title),
		Subtitle:   strings.TrimSpace(input.Subtitle),
		Image:      strings.TrimSpace(input.Image),
		Link:       strings.TrimSpace(input.Link),
		ButtonText: strings.TrimSpace(input.ButtonText),
	}
}

func normalizeHomeSectionConfigs(sections []homeSectionConfig) []homeSectionConfig {
	result := make([]homeSectionConfig, 0, len(sections))
	seen := make(map[string]struct{}, len(sections))
	for index, section := range sections {
		normalized := homeSectionConfig{
			ID:      strings.TrimSpace(section.ID),
			Type:    strings.TrimSpace(section.Type),
			Enabled: section.Enabled,
			Title:   strings.TrimSpace(section.Title),
			Limit:   section.Limit,
		}
		if normalized.Type == "" {
			continue
		}
		if normalized.ID == "" {
			normalized.ID = fmt.Sprintf("%s-%d", normalized.Type, index+1)
		}
		if _, exists := seen[normalized.ID]; exists {
			continue
		}
		seen[normalized.ID] = struct{}{}
		if normalized.Limit < 0 {
			normalized.Limit = 0
		}
		result = append(result, normalized)
	}
	return result
}

func normalizeProductVariants(input []productVariantPayload) []productVariantPayload {
	result := make([]productVariantPayload, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, variant := range input {
		name := strings.TrimSpace(variant.Name)
		sku := strings.TrimSpace(variant.SKU)
		if name == "" || sku == "" {
			continue
		}
		if _, exists := seen[sku]; exists {
			continue
		}
		seen[sku] = struct{}{}
		result = append(result, productVariantPayload{
			Name: name,
			SKU:  sku,
		})
	}
	return result
}

func normalizeProductOptionGroups(input []productOptionGroupPayload) []productOptionGroupPayload {
	result := make([]productOptionGroupPayload, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, group := range input {
		name := strings.TrimSpace(group.Name)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		valuesSeen := make(map[string]struct{}, len(group.Values))
		values := make([]string, 0, len(group.Values))
		for _, value := range group.Values {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			if _, exists := valuesSeen[trimmed]; exists {
				continue
			}
			valuesSeen[trimmed] = struct{}{}
			values = append(values, trimmed)
		}
		if len(values) == 0 {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, productOptionGroupPayload{Name: name, Values: values})
	}
	return result
}

func normalizeProductSkuVariants(input []productSkuVariantPayload, groups []productOptionGroupPayload) []productSkuVariantPayload {
	validGroupValues := make(map[string]map[string]struct{}, len(groups))
	for _, group := range groups {
		valueSet := make(map[string]struct{}, len(group.Values))
		for _, value := range group.Values {
			valueSet[value] = struct{}{}
		}
		validGroupValues[group.Name] = valueSet
	}

	result := make([]productSkuVariantPayload, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, variant := range input {
		sku := strings.TrimSpace(variant.SKU)
		if sku == "" {
			continue
		}
		if _, exists := seen[sku]; exists {
			continue
		}
		selections := make(map[string]string)
		valid := true
		for groupName, groupValues := range validGroupValues {
			rawValue, ok := variant.Selections[groupName]
			if !ok {
				valid = false
				break
			}
			value := strings.TrimSpace(rawValue)
			if _, exists := groupValues[value]; !exists {
				valid = false
				break
			}
			selections[groupName] = value
		}
		if !valid || len(selections) != len(validGroupValues) {
			continue
		}
		seen[sku] = struct{}{}
		result = append(result, productSkuVariantPayload{
			SKU:        sku,
			Price:      variant.Price,
			Stock:      variant.Stock,
			Selections: selections,
		})
	}
	return result
}

func variantSliceToJSONArray(values []productVariantPayload) models.JSONArray {
	result := make(models.JSONArray, 0, len(values))
	for _, value := range values {
		result = append(result, models.JSONMap{
			"name": value.Name,
			"sku":  value.SKU,
		})
	}
	return result
}

func optionGroupSliceToJSONArray(values []productOptionGroupPayload) models.JSONArray {
	result := make(models.JSONArray, 0, len(values))
	for _, value := range values {
		result = append(result, models.JSONMap{
			"name":   value.Name,
			"values": stringSliceToJSONArray(value.Values),
		})
	}
	return result
}

func skuVariantSliceToJSONArray(values []productSkuVariantPayload) models.JSONArray {
	result := make(models.JSONArray, 0, len(values))
	for _, value := range values {
		entry := models.JSONMap{
			"sku":        value.SKU,
			"selections": models.JSONMap{},
		}
		if value.Price != nil {
			entry["price"] = *value.Price
		}
		if value.Stock != nil {
			entry["stock"] = *value.Stock
		}
		selections := models.JSONMap{}
		for key, selection := range value.Selections {
			selections[key] = selection
		}
		entry["selections"] = selections
		result = append(result, entry)
	}
	return result
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func limitRunes(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max <= 0 {
		return ""
	}

	runes := []rune(trimmed)
	if len(runes) <= max {
		return trimmed
	}
	return strings.TrimSpace(string(runes[:max]))
}

func jsonStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func jsonArrayToStringSlice(values models.JSONArray) []string {
	result := make([]string, 0, len(values))
	for _, item := range values {
		text := strings.TrimSpace(fmt.Sprintf("%v", item))
		if text == "" {
			continue
		}
		result = append(result, text)
	}
	return result
}

func buildOverrideGenerationPrompt(product models.Product, tenant models.Tenant, override models.TenantProductOverride, instruction string) string {
	specs := product.Specifications
	baseName := firstNonEmptyString(
		jsonStringValue(override.CustomName),
		product.BaseName,
		jsonString(specs, "name", ""),
	)
	description := firstNonEmptyString(
		jsonStringValue(override.CustomDescription),
		jsonString(specs, "description", ""),
		jsonString(specs, "longDescription", ""),
	)
	brand := jsonString(specs, "brand", "")
	category := jsonString(specs, "category", "")
	status := jsonString(specs, "status", "")
	flavors := strings.Join(jsonStringSlice(specs, "flavors", nil), "、")
	gallery := strings.Join(jsonArrayToStringSlice(product.BaseImages), "\n")
	detailImages := strings.Join(jsonArrayToStringSlice(product.DetailImages), "\n")
	currentSEOTitle := firstNonEmptyString(jsonStringValue(override.SEOTitle), jsonString(specs, "seoTitle", ""))
	currentSEODescription := firstNonEmptyString(jsonStringValue(override.SEODescription), jsonString(specs, "seoDescription", ""))
	tenantTitle := jsonString(tenant.SEOConfig, "title", "")
	tenantDescription := jsonString(tenant.SEOConfig, "description", "")

	return strings.TrimSpace(fmt.Sprintf(`
你是电商内容编辑，请为指定租户生成“租户商品覆写”文案。

输出要求：
1. 只返回 JSON，不要输出 markdown、说明文字或代码块。
2. JSON 字段固定为：custom_name、custom_description、seo_title、seo_description。
3. custom_name 保持自然、适合商品标题，尽量不超过 60 个字。
4. custom_description 写成面向消费者的简洁介绍，突出卖点、口味/适用场景，不要虚构功效，建议 80-220 个字。
5. seo_title 适合搜索结果，尽量不超过 65 个字。
6. seo_description 适合搜索摘要，尽量不超过 160 个字。
7. 文案语言使用繁體中文，并尽量贴近当前租户品牌语气。
8. 不要编造价格、库存、配送、赠品、法律合规承诺；如果资料不足，就基于已有信息做稳妥表达。

租户信息：
- 租户名称：%s
- 主域名：%s
- 租户 SEO 标题：%s
- 租户 SEO 描述：%s

商品基础信息：
- 商品 ID：%d
- SKU：%s
- 商品名称：%s
- 品牌：%s
- 分类：%s
- 状态：%s
- 基础售价：%.2f
- 基础库存：%d
- 商品简介：%s
- 商品长描述：%s
- 口味：%s
- 规格 HTML：%s
- 主图列表：%s
- 详情图列表：%s

当前覆写（如有）：
- custom_name：%s
- custom_description：%s
- seo_title：%s
- seo_description：%s

额外要求：
%s
`, tenant.Name, tenant.Domain, tenantTitle, tenantDescription, product.ID, product.SKU, baseName, brand, category, status, product.BasePrice, product.BaseStockQuantity, jsonString(specs, "description", description), jsonString(specs, "longDescription", ""), flavors, jsonString(specs, "specificationHtml", ""), gallery, detailImages, jsonStringValue(override.CustomName), jsonStringValue(override.CustomDescription), currentSEOTitle, currentSEODescription, firstNonEmptyString(instruction, "若无额外要求，请输出最适合该租户站点的默认文案。")))
}

func buildOverrideNameGenerationPrompt(product models.Product, tenant models.Tenant, override models.TenantProductOverride, instruction string) string {
	specs := product.Specifications
	baseName := firstNonEmptyString(
		jsonStringValue(override.CustomName),
		product.BaseName,
		jsonString(specs, "name", ""),
	)
	brand := jsonString(specs, "brand", "")
	category := jsonString(specs, "category", "")
	status := jsonString(specs, "status", "")
	flavors := strings.Join(jsonStringSlice(specs, "flavors", nil), "、")
	description := firstNonEmptyString(
		jsonStringValue(override.CustomDescription),
		jsonString(specs, "description", ""),
		jsonString(specs, "longDescription", ""),
	)
	tenantTitle := jsonString(tenant.SEOConfig, "title", "")

	return strings.TrimSpace(fmt.Sprintf(`
你是电商命名编辑，请为指定租户生成商品“自訂商品名稱”。

输出要求：
1. 只返回 JSON，不要输出 markdown、说明文字或代码块。
2. JSON 字段固定为：custom_name。
3. custom_name 使用繁體中文，适合前台商品卡片与详情页标题。
4. 名称要基于已有商品信息做优化，不要虚构不存在的规格、赠品、疗效、认证或承诺。
5. 尽量不超过 60 个字。
6. 如果原名称已经很适合，也可以只做轻微优化。

租户信息：
- 租户名称：%s
- 主域名：%s
- 租户 SEO 标题：%s

商品信息：
- 商品 ID：%d
- SKU：%s
- 当前名称：%s
- 品牌：%s
- 分类：%s
- 状态：%s
- 商品简介：%s
- 口味：%s

额外要求：
%s
`, tenant.Name, tenant.Domain, tenantTitle, product.ID, product.SKU, baseName, brand, category, status, description, flavors, firstNonEmptyString(instruction, "请输出最适合该租户站点风格的商品名称。")))
}

func generateOverrideDraftWithDeepSeek(cfg *config.Config, prompt string) (generatedOverrideDraftResponse, error) {
	if strings.TrimSpace(cfg.DeepSeekAPIKey) == "" {
		return generatedOverrideDraftResponse{}, errors.New("DEEPSEEK_API_KEY is not configured")
	}

	requestPayload := deepSeekChatCompletionRequest{
		Model: cfg.DeepSeekModel,
		Messages: []deepSeekChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a precise ecommerce copywriter. Return valid JSON only.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Temperature: 0.9,
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return generatedOverrideDraftResponse{}, err
	}

	endpoint := strings.TrimRight(cfg.DeepSeekBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return generatedOverrideDraftResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.DeepSeekAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return generatedOverrideDraftResponse{}, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return generatedOverrideDraftResponse{}, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var errorPayload deepSeekChatCompletionResponse
		if json.Unmarshal(responseBody, &errorPayload) == nil && errorPayload.Error != nil && strings.TrimSpace(errorPayload.Error.Message) != "" {
			return generatedOverrideDraftResponse{}, errors.New(errorPayload.Error.Message)
		}
		return generatedOverrideDraftResponse{}, fmt.Errorf("deepseek request failed with status %d", resp.StatusCode)
	}

	var completion deepSeekChatCompletionResponse
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return generatedOverrideDraftResponse{}, err
	}
	if len(completion.Choices) == 0 {
		return generatedOverrideDraftResponse{}, errors.New("deepseek returned no choices")
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result generatedOverrideDraftResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return generatedOverrideDraftResponse{}, err
	}

	result.CustomName = limitRunes(result.CustomName, 60)
	result.CustomDescription = limitRunes(result.CustomDescription, 260)
	result.SEOTitle = limitRunes(result.SEOTitle, 70)
	result.SEODescription = limitRunes(result.SEODescription, 180)

	return result, nil
}

func generateOverrideNameWithDeepSeek(cfg *config.Config, prompt string) (string, error) {
	if strings.TrimSpace(cfg.DeepSeekAPIKey) == "" {
		return "", errors.New("DEEPSEEK_API_KEY is not configured")
	}

	requestPayload := deepSeekChatCompletionRequest{
		Model: cfg.DeepSeekModel,
		Messages: []deepSeekChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a precise ecommerce naming editor. Return valid JSON only.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Temperature: 0.8,
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(cfg.DeepSeekBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.DeepSeekAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var errorPayload deepSeekChatCompletionResponse
		if json.Unmarshal(responseBody, &errorPayload) == nil && errorPayload.Error != nil && strings.TrimSpace(errorPayload.Error.Message) != "" {
			return "", errors.New(errorPayload.Error.Message)
		}
		return "", fmt.Errorf("deepseek request failed with status %d", resp.StatusCode)
	}

	var completion deepSeekChatCompletionResponse
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return "", err
	}
	if len(completion.Choices) == 0 {
		return "", errors.New("deepseek returned no choices")
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result struct {
		CustomName string `json:"custom_name"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return "", err
	}

	return limitRunes(result.CustomName, 60), nil
}

func jsonProductVariants(source models.JSONMap, key string) []productVariantPayload {
	if source == nil {
		return []productVariantPayload{}
	}
	raw, ok := source[key].([]interface{})
	if !ok {
		return []productVariantPayload{}
	}

	result := make([]productVariantPayload, 0, len(raw))
	for _, item := range raw {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		sku, _ := entry["sku"].(string)
		name = strings.TrimSpace(name)
		sku = strings.TrimSpace(sku)
		if name == "" || sku == "" {
			continue
		}
		result = append(result, productVariantPayload{Name: name, SKU: sku})
	}
	return result
}

func jsonProductOptionGroups(source models.JSONMap, key string) []productOptionGroupPayload {
	if source == nil {
		return []productOptionGroupPayload{}
	}
	raw, ok := source[key].([]interface{})
	if !ok {
		return []productOptionGroupPayload{}
	}

	result := make([]productOptionGroupPayload, 0, len(raw))
	for _, item := range raw {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		values := jsonStringSlice(models.JSONMap(entry), "values", nil)
		name = strings.TrimSpace(name)
		if name == "" || len(values) == 0 {
			continue
		}
		result = append(result, productOptionGroupPayload{Name: name, Values: values})
	}
	return result
}

func jsonProductSkuVariants(source models.JSONMap, key string) []productSkuVariantPayload {
	if source == nil {
		return []productSkuVariantPayload{}
	}
	raw, ok := source[key].([]interface{})
	if !ok {
		return []productSkuVariantPayload{}
	}

	result := make([]productSkuVariantPayload, 0, len(raw))
	for _, item := range raw {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		sku, _ := entry["sku"].(string)
		sku = strings.TrimSpace(sku)
		if sku == "" {
			continue
		}
		var price *float64
		switch value := entry["price"].(type) {
		case float64:
			price = &value
		case int:
			converted := float64(value)
			price = &converted
		}
		var stock *int
		switch value := entry["stock"].(type) {
		case float64:
			converted := int(value)
			stock = &converted
		case int:
			converted := value
			stock = &converted
		}
		selections := make(map[string]string)
		if rawSelections, ok := entry["selections"].(map[string]interface{}); ok {
			for key, value := range rawSelections {
				if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
					selections[key] = strings.TrimSpace(text)
				}
			}
		}
		if len(selections) == 0 {
			continue
		}
		result = append(result, productSkuVariantPayload{
			SKU:        sku,
			Price:      price,
			Stock:      stock,
			Selections: selections,
		})
	}
	return result
}

func generateUploadFilename(originalName string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	extension := strings.ToLower(filepath.Ext(originalName))
	switch extension {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg":
	default:
		extension = ".bin"
	}

	safeName := sanitizeUploadFilename(originalName)
	randomSuffix := hex.EncodeToString(randomBytes[:4])
	return safeName + "-" + randomSuffix + extension, nil
}

func UploadImageHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing upload file"})
			return
		}

		contentType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only image uploads are allowed"})
			return
		}

		filename, err := generateUploadFilename(file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare upload filename"})
			return
		}

		subDir := filepath.Join(cfg.UploadDir, "products")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare upload directory"})
			return
		}

		targetPath := filepath.Join(subDir, filename)
		if err := c.SaveUploadedFile(file, targetPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded image"})
			return
		}

		c.JSON(http.StatusOK, uploadImageResponse{
			URL: "/uploads/products/" + filename,
		})
	}
}

func stringSliceToJSONArray(values []string) models.JSONArray {
	result := make(models.JSONArray, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
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

func buildVariantDisplayName(groups []productOptionGroupPayload, selections map[string]string) string {
	parts := make([]string, 0, len(selections))
	for _, group := range groups {
		if value, ok := selections[group.Name]; ok && strings.TrimSpace(value) != "" {
			parts = append(parts, group.Name+"："+strings.TrimSpace(value))
		}
	}
	return strings.Join(parts, " / ")
}

func normalizedVariantLabel(value string) string {
	replacer := strings.NewReplacer("：", ":", "／", "/", " ", "")
	return strings.ToLower(strings.TrimSpace(replacer.Replace(value)))
}

func tenantToResponse(tenant models.Tenant) tenantResponse {
	homeBannerMap := jsonMap(tenant.ThemeConfig, "homeBanner")
	homeBanner := homeBannerConfig{
		Enabled:    jsonBool(homeBannerMap, "enabled", false),
		Title:      jsonString(homeBannerMap, "title", ""),
		Subtitle:   jsonString(homeBannerMap, "subtitle", ""),
		Image:      jsonString(homeBannerMap, "image", ""),
		Link:       jsonString(homeBannerMap, "link", ""),
		ButtonText: jsonString(homeBannerMap, "buttonText", ""),
	}

	return tenantResponse{
		ID:                tenant.ID,
		Domain:            tenant.Domain,
		BoundDomains:      jsonArrayToStrings(tenant.BoundDomains),
		NPMProxyHostID:    jsonUint(tenant.ThemeConfig, "npmProxyHostId"),
		Name:              tenant.Name,
		IsActive:          tenant.IsActive,
		Theme:             jsonString(tenant.ThemeConfig, "theme", ""),
		HomeTemplate:      jsonString(tenant.ThemeConfig, "homeTemplate", ""),
		HomeModuleOrder:   jsonStringSlice(tenant.ThemeConfig, "homeModuleOrder", []string{}),
		HomeBanner:        homeBanner,
		HomeSections:      jsonSectionConfigs(tenant.ThemeConfig, "homeSections"),
		PrimaryBrandID:    jsonUint(tenant.ThemeConfig, "primaryBrandId"),
		PreviewImage:      jsonString(tenant.ThemeConfig, "previewImage", ""),
		LogoImage:         jsonString(tenant.ThemeConfig, "logoImage", ""),
		AccentColor:       jsonString(tenant.ThemeConfig, "accentColor", ""),
		AccentStrongColor: jsonString(tenant.ThemeConfig, "accentStrongColor", ""),
		SurfaceColor:      jsonString(tenant.ThemeConfig, "surfaceColor", ""),
		PageBgColor:       jsonString(tenant.ThemeConfig, "pageBgColor", ""),
		CardBgColor:       jsonString(tenant.ThemeConfig, "cardBgColor", ""),
		TextColor:         jsonString(tenant.ThemeConfig, "textColor", ""),
		MutedTextColor:    jsonString(tenant.ThemeConfig, "mutedTextColor", ""),
		BorderColor:       jsonString(tenant.ThemeConfig, "borderColor", ""),
		HeroBgColor:       jsonString(tenant.ThemeConfig, "heroBgColor", ""),
		TagBgColor:        jsonString(tenant.ThemeConfig, "tagBgColor", ""),
		HeroTitle:         jsonString(tenant.ThemeConfig, "heroTitle", ""),
		Tagline:           jsonString(tenant.ThemeConfig, "tagline", ""),
		Announcement:      jsonString(tenant.ThemeConfig, "announcement", ""),
		SupportText:       jsonString(tenant.ThemeConfig, "supportText", ""),
		SEOTitle:          jsonString(tenant.SEOConfig, "title", ""),
		SEODescription:    jsonString(tenant.SEOConfig, "description", ""),
	}
}

func tenantPayloadToModel(payload tenantPayload, existing *models.Tenant) models.Tenant {
	model := models.Tenant{}
	if existing != nil {
		model = *existing
	}

	model.Domain = normalizeDomain(payload.Domain)
	boundDomains := normalizeDomainList(payload.BoundDomains)
	filteredBoundDomains := make([]string, 0, len(boundDomains))
	for _, domain := range boundDomains {
		if domain == model.Domain {
			continue
		}
		filteredBoundDomains = append(filteredBoundDomains, domain)
	}
	model.BoundDomains = stringSliceToJSONArray(filteredBoundDomains)
	model.Name = payload.Name
	model.IsActive = payload.IsActive
	model.ThemeConfig = models.JSONMap{
		"theme":           payload.Theme,
		"homeTemplate":    payload.HomeTemplate,
		"homeModuleOrder": payload.HomeModuleOrder,
		"npmProxyHostId":  payload.NPMProxyHostID,
		"homeBanner": models.JSONMap{
			"enabled":    payload.HomeBanner.Enabled,
			"title":      payload.HomeBanner.Title,
			"subtitle":   payload.HomeBanner.Subtitle,
			"image":      payload.HomeBanner.Image,
			"link":       payload.HomeBanner.Link,
			"buttonText": payload.HomeBanner.ButtonText,
		},
		"homeSections":      normalizeHomeSectionConfigs(payload.HomeSections),
		"primaryBrandId":    payload.PrimaryBrandID,
		"previewImage":      payload.PreviewImage,
		"logoImage":         payload.LogoImage,
		"accentColor":       payload.AccentColor,
		"accentStrongColor": payload.AccentStrongColor,
		"surfaceColor":      payload.SurfaceColor,
		"pageBgColor":       payload.PageBgColor,
		"cardBgColor":       payload.CardBgColor,
		"textColor":         payload.TextColor,
		"mutedTextColor":    payload.MutedTextColor,
		"borderColor":       payload.BorderColor,
		"heroBgColor":       payload.HeroBgColor,
		"tagBgColor":        payload.TagBgColor,
		"heroTitle":         payload.HeroTitle,
		"tagline":           payload.Tagline,
		"announcement":      payload.Announcement,
		"supportText":       payload.SupportText,
	}
	model.SEOConfig = models.JSONMap{
		"title":       payload.SEOTitle,
		"description": payload.SEODescription,
	}
	return model
}

func productToResponse(product models.Product, override *models.TenantProductOverride) productResponse {
	specs := product.Specifications
	if specs == nil {
		specs = models.JSONMap{}
	}

	categoryID := jsonUint(specs, "categoryId")
	categoryIDs := jsonUintSlice(specs, "categoryIds", nil)
	if len(categoryIDs) == 0 && categoryID != nil && *categoryID > 0 {
		categoryIDs = []uint{*categoryID}
	}

	response := productResponse{
		ID:                product.ID,
		SKU:               product.SKU,
		Slug:              product.Slug,
		BaseName:          product.BaseName,
		BasePrice:         product.BasePrice,
		BaseStockQuantity: product.BaseStockQuantity,
		BaseImages:        jsonArrayToStrings(product.BaseImages),
		DetailImages:      jsonArrayToStrings(product.DetailImages),
		Specifications:    specs,
		IsActive:          product.IsActive,
		Category:          jsonString(specs, "category", ""),
		CategoryID:        categoryID,
		CategoryIDs:       categoryIDs,
		Brand:             jsonString(specs, "brand", ""),
		BrandID:           jsonUint(specs, "brandId"),
		PreviewImage:      jsonString(specs, "previewImage", ""),
		Gallery:           jsonStringSlice(specs, "gallery", jsonArrayToStrings(product.BaseImages)),
		Status:            jsonString(specs, "status", "草稿"),
		Description:       jsonString(specs, "description", ""),
		LongDescription:   jsonString(specs, "longDescription", ""),
		SpecificationHTML: jsonString(specs, "specificationHtml", ""),
		Badge:             jsonString(specs, "badge", ""),
		Rating:            jsonFloat(specs, "rating", 4.5),
		Reviews:           jsonInt(specs, "reviews", 0),
		Flavors:           jsonStringSlice(specs, "flavors", []string{}),
		Variants:          jsonProductVariants(specs, "variants"),
		OptionGroups:      jsonProductOptionGroups(specs, "optionGroups"),
		SkuVariants:       jsonProductSkuVariants(specs, "skuVariants"),
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
		IsVisible:         true,
	}

	if response.PreviewImage == "" && len(response.BaseImages) > 0 {
		response.PreviewImage = response.BaseImages[0]
	}
	if len(response.Gallery) == 0 && response.PreviewImage != "" {
		response.Gallery = []string{response.PreviewImage}
	}
	if len(response.Flavors) == 0 && len(response.Variants) > 0 {
		response.Flavors = make([]string, 0, len(response.Variants))
		for _, variant := range response.Variants {
			response.Flavors = append(response.Flavors, variant.Name)
		}
	}
	if len(response.OptionGroups) == 0 && len(response.Variants) > 0 {
		values := make([]string, 0, len(response.Variants))
		skuVariants := make([]productSkuVariantPayload, 0, len(response.Variants))
		for _, variant := range response.Variants {
			values = append(values, variant.Name)
			skuVariants = append(skuVariants, productSkuVariantPayload{
				SKU: variant.SKU,
				Selections: map[string]string{
					"口味": variant.Name,
				},
			})
		}
		response.OptionGroups = []productOptionGroupPayload{{Name: "口味", Values: values}}
		response.SkuVariants = skuVariants
	}

	if override != nil {
		response.CustomName = override.CustomName
		response.CustomDescription = override.CustomDescription
		response.CustomPrice = override.CustomPrice
		response.CustomStockQuantity = override.CustomStockQuantity
		response.CustomImages = jsonArrayToStrings(override.CustomImages)
		response.CustomDetailImages = jsonArrayToStrings(override.CustomDetailImages)
		response.SEOTitle = override.SEOTitle
		response.SEODescription = override.SEODescription
		response.IsVisible = override.IsVisible
	}

	return response
}

func productPayloadToModel(payload productPayload, existing *models.Product) models.Product {
	model := models.Product{}
	if existing != nil {
		model = *existing
	}

	gallery := payload.Gallery
	if len(gallery) == 0 && payload.PreviewImage != "" {
		gallery = []string{payload.PreviewImage}
	}
	variants := normalizeProductVariants(payload.Variants)
	optionGroups := normalizeProductOptionGroups(payload.OptionGroups)
	skuVariants := normalizeProductSkuVariants(payload.SkuVariants, optionGroups)
	if len(optionGroups) == 0 && len(variants) > 0 {
		values := make([]string, 0, len(variants))
		for _, variant := range variants {
			values = append(values, variant.Name)
		}
		optionGroups = []productOptionGroupPayload{{Name: "口味", Values: values}}
		skuVariants = make([]productSkuVariantPayload, 0, len(variants))
		for _, variant := range variants {
			skuVariants = append(skuVariants, productSkuVariantPayload{
				SKU: variant.SKU,
				Selections: map[string]string{
					"口味": variant.Name,
				},
			})
		}
	}
	flavors := payload.Flavors
	if len(variants) > 0 {
		flavors = make([]string, 0, len(variants))
		for _, variant := range variants {
			flavors = append(flavors, variant.Name)
		}
	}

	categoryIDs := sanitizeFeaturedCategoryIDs(payload.CategoryIDs)
	if payload.CategoryID != nil && *payload.CategoryID > 0 {
		hasPrimary := false
		for _, id := range categoryIDs {
			if id == *payload.CategoryID {
				hasPrimary = true
				break
			}
		}
		if !hasPrimary {
			categoryIDs = append(models.UIntArray{*payload.CategoryID}, categoryIDs...)
		}
	}
	if len(categoryIDs) == 0 && payload.CategoryID != nil && *payload.CategoryID > 0 {
		categoryIDs = models.UIntArray{*payload.CategoryID}
	}

	model.SKU = payload.SKU
	model.BaseName = payload.BaseName
	model.BasePrice = payload.BasePrice
	model.BaseStockQuantity = payload.BaseStockQuantity
	model.IsActive = payload.IsActive
	model.BaseImages = stringSliceToJSONArray(gallery)
	model.DetailImages = stringSliceToJSONArray(payload.DetailImages)
	model.Specifications = models.JSONMap{
		"category":          payload.Category,
		"categoryId":        payload.CategoryID,
		"categoryIds":       categoryIDs,
		"brand":             payload.Brand,
		"brandId":           payload.BrandID,
		"previewImage":      payload.PreviewImage,
		"gallery":           stringSliceToJSONArray(gallery),
		"detailImages":      stringSliceToJSONArray(payload.DetailImages),
		"status":            payload.Status,
		"description":       payload.Description,
		"longDescription":   payload.LongDescription,
		"specificationHtml": payload.SpecificationHTML,
		"badge":             payload.Badge,
		"rating":            payload.Rating,
		"reviews":           payload.Reviews,
		"flavors":           stringSliceToJSONArray(flavors),
		"variants":          variantSliceToJSONArray(variants),
		"optionGroups":      optionGroupSliceToJSONArray(optionGroups),
		"skuVariants":       skuVariantSliceToJSONArray(skuVariants),
	}
	return model
}

func categoryToResponse(category models.Category) categoryResponse {
	return categoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		ParentID:  category.ParentID,
		SortOrder: category.SortOrder,
	}
}

func ensureDefaultSharedCategory(tx *gorm.DB) (models.Category, error) {
	var category models.Category
	if err := tx.
		Where("tenant_id IS NULL").
		Where("code = ? OR name IN ?", defaultUncategorizedCategoryCode, []string{defaultUncategorizedCategoryName, "未分类"}).
		Order("id asc").
		First(&category).Error; err == nil {
		return category, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Category{}, err
	}

	if err := tx.Exec(
		`INSERT INTO categories (code, name, parent_id, seo_title, seo_description, created_at, updated_at, tenant_id, sort_order)
		 VALUES (?, ?, NULL, ?, '', NOW(3), NOW(3), NULL, 0)`,
		defaultUncategorizedCategoryCode,
		defaultUncategorizedCategoryName,
		defaultUncategorizedCategoryName,
	).Error; err != nil {
		return models.Category{}, err
	}

	if err := tx.
		Where("tenant_id IS NULL AND code = ?", defaultUncategorizedCategoryCode).
		First(&category).Error; err != nil {
		return models.Category{}, err
	}

	return category, nil
}

func brandToResponse(brand models.Brand) brandResponse {
	response := brandResponse{
		ID:   brand.ID,
		Name: brand.Name,
	}
	if brand.LogoURL != nil {
		response.LogoURL = *brand.LogoURL
	}
	if brand.Description != nil {
		response.Description = *brand.Description
	}
	return response
}

func getUintParam(c *gin.Context, key string) (uint, error) {
	value, err := strconv.ParseUint(c.Param(key), 10, 64)
	return uint(value), err
}

func requestBaseURL(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.GetHeader("X-Tenant-Domain")
	}
	if host == "" {
		host = c.Request.Host
	}

	return scheme + "://" + host
}

func loadTenantOverrideMapForProducts(db *gorm.DB, tenantID uint, productIDs []uint) (map[uint]*models.TenantProductOverride, error) {
	if len(productIDs) == 0 {
		return map[uint]*models.TenantProductOverride{}, nil
	}

	var overrides []models.TenantProductOverride
	if err := db.Where("tenant_id = ? AND product_id IN ?", tenantID, productIDs).Find(&overrides).Error; err != nil {
		return nil, err
	}

	overrideMap := make(map[uint]*models.TenantProductOverride, len(overrides))
	for i := range overrides {
		overrideMap[overrides[i].ProductID] = &overrides[i]
	}

	return overrideMap, nil
}

func isProductPublished(product models.Product) bool {
	specs := product.Specifications
	status := strings.TrimSpace(jsonString(specs, "status", "草稿"))
	return product.IsActive && status == "上架中"
}

func isProductVisibleForTenant(product models.Product, override *models.TenantProductOverride) bool {
	if !isProductPublished(product) {
		return false
	}
	return override == nil || override.IsVisible
}

func buildVisibleProductResponses(products []models.Product, overrideMap map[uint]*models.TenantProductOverride) []productResponse {
	result := make([]productResponse, 0, len(products))
	for _, product := range products {
		override := overrideMap[product.ID]
		if !isProductVisibleForTenant(product, override) {
			continue
		}
		result = append(result, productToResponse(product, override))
	}
	return result
}

func loadSharedCategoryDescendantIDs(db *gorm.DB, cache *categoryDescendantCache, categoryID uint) (map[uint]struct{}, error) {
	if cached, ok := cache.get(categoryID); ok {
		return cached, nil
	}

	var categories []models.Category
	if err := db.Where("tenant_id IS NULL").Find(&categories).Error; err != nil {
		return nil, err
	}

	childrenByParent := make(map[uint][]uint)
	for _, category := range categories {
		if category.ParentID == nil {
			continue
		}
		childrenByParent[*category.ParentID] = append(childrenByParent[*category.ParentID], category.ID)
	}

	result := map[uint]struct{}{categoryID: {}}
	queue := []uint{categoryID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range childrenByParent[current] {
			if _, exists := result[childID]; exists {
				continue
			}
			result[childID] = struct{}{}
			queue = append(queue, childID)
		}
	}

	cache.set(categoryID, result)
	return result, nil
}

func productResponseCategoryIDs(product productResponse) []uint {
	if len(product.CategoryIDs) > 0 {
		return product.CategoryIDs
	}
	if product.CategoryID != nil && *product.CategoryID > 0 {
		return []uint{*product.CategoryID}
	}
	return nil
}

func productLastModified(product models.Product, override *models.TenantProductOverride) string {
	updatedAt := product.UpdatedAt
	if override != nil && override.UpdatedAt.After(updatedAt) {
		updatedAt = override.UpdatedAt
	}
	if updatedAt.IsZero() {
		return ""
	}
	return updatedAt.Format("2006-01-02T15:04:05Z07:00")
}

func loadOrderProductNameMap(db *gorm.DB, items []models.OrderItem) (map[uint]string, error) {
	productIDs := make([]uint, 0, len(items))
	for _, item := range items {
		if item.ProductID > 0 {
			productIDs = append(productIDs, item.ProductID)
		}
	}

	productNameByID := make(map[uint]string, len(productIDs))
	if len(productIDs) == 0 {
		return productNameByID, nil
	}

	var products []models.Product
	if err := db.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, err
	}
	for _, product := range products {
		productNameByID[product.ID] = firstNonEmptyString(product.BaseName, product.SKU)
	}

	return productNameByID, nil
}

func validateTenantDomains(db *gorm.DB, tenantID uint, primaryDomain string, boundDomains []string) error {
	if primaryDomain == "" {
		return errors.New("primary domain is required")
	}

	var tenants []models.Tenant
	if err := db.Find(&tenants).Error; err != nil {
		return err
	}

	candidateDomains := append([]string{primaryDomain}, boundDomains...)
	for _, tenant := range tenants {
		if tenant.ID == tenantID {
			continue
		}

		existing := append([]string{normalizeDomain(tenant.Domain)}, jsonArrayToStrings(tenant.BoundDomains)...)
		existingSet := make(map[string]struct{}, len(existing))
		for _, domain := range existing {
			existingSet[normalizeDomain(domain)] = struct{}{}
		}

		for _, candidate := range candidateDomains {
			if _, exists := existingSet[candidate]; exists {
				return errors.New("domain conflict: " + candidate)
			}
		}
	}

	return nil
}

func ensureDomainExists(db *gorm.DB, domain string) error {
	normalized := normalizeDomain(domain)
	if normalized == "" {
		return errors.New("domain is required")
	}

	var count int64
	if err := db.Model(&models.Domain{}).Where("LOWER(domain_name) = ?", normalized).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("domain not found in domain registry: " + normalized)
	}
	return nil
}

func addUniqueDomain(values []string, domain string) []string {
	normalized := normalizeDomain(domain)
	if normalized == "" {
		return normalizeDomainList(values)
	}

	result := normalizeDomainList(values)
	for _, item := range result {
		if item == normalized {
			return result
		}
	}
	return append(result, normalized)
}

func removeDomainFromList(values []string, domain string) ([]string, bool) {
	normalized := normalizeDomain(domain)
	if normalized == "" {
		return normalizeDomainList(values), false
	}

	result := make([]string, 0, len(values))
	found := false
	for _, item := range normalizeDomainList(values) {
		if item == normalized {
			found = true
			continue
		}
		result = append(result, item)
	}
	return result, found
}

func diffAddedDomains(previous []string, current []string) []string {
	previousSet := make(map[string]struct{}, len(previous))
	for _, domain := range normalizeDomainList(previous) {
		previousSet[domain] = struct{}{}
	}

	result := make([]string, 0, len(current))
	for _, domain := range normalizeDomainList(current) {
		if _, exists := previousSet[domain]; exists {
			continue
		}
		result = append(result, domain)
	}
	return result
}

func syncDomainsToGSC(cfg *config.Config, domains []string) []service.GSCResult {
	gscService, err := service.NewGSCService(cfg)
	if err != nil {
		return []service.GSCResult{{
			Status:  "skipped",
			Message: err.Error(),
		}}
	}
	return gscService.EnsureSites(domains)
}

func syncTenantDomainsToNPM(cfg *config.Config, primaryDomain string, boundDomains []string) (*service.NPMResult, error) {
	npmService, err := service.NewNPMService(cfg)
	if err != nil {
		return nil, err
	}
	return npmService.UpdateDomainsAndSSL(primaryDomain, boundDomains)
}

func syncTenantDomainsToNPMByProxyHostID(cfg *config.Config, proxyHostID uint, primaryDomain string, boundDomains []string) (*service.NPMResult, error) {
	npmService, err := service.NewNPMService(cfg)
	if err != nil {
		return nil, err
	}
	return npmService.UpdateProxyHostDomainsByID(proxyHostID, primaryDomain, boundDomains)
}

func ensureSuccessfulNPMResult(result *service.NPMResult) error {
	if result == nil {
		return errors.New("NPM sync returned an empty response")
	}
	if result.Status == "success" && result.NPMUpdated {
		return nil
	}

	message := strings.TrimSpace(result.Message)
	if message == "" {
		message = "NPM sync failed"
	}

	if result.ProxyHostID != nil {
		message += fmt.Sprintf(" (proxy_host_id=%v)", result.ProxyHostID)
	}
	if len(result.UpdatedDomains) > 0 {
		message += fmt.Sprintf(" [domains=%s]", strings.Join(result.UpdatedDomains, ", "))
	}
	return errors.New(message)
}

func orderToResponse(order models.Order, items []models.OrderItem, productNameByID map[uint]string) orderResponse {
	result := make([]orderItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, orderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			Name:        firstNonEmptyString(productNameByID[item.ProductID], fmt.Sprintf("商品 #%d", item.ProductID)),
			VariantName: item.VariantName,
			VariantSKU:  item.VariantSKU,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	return orderResponse{
		ID:               order.ID,
		TenantID:         order.TenantID,
		UserID:           order.UserID,
		TotalAmount:      order.TotalAmount,
		Status:           order.Status,
		LineID:           order.LineID,
		Phone:            order.Phone,
		ConvenienceStore: order.ConvenienceStore,
		ShippingAddress:  order.ShippingAddress,
		PaymentMethod:    order.PaymentMethod,
		Items:            result,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

// ===== 认证处理器 =====

func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Register endpoint"})
	}
}

func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Login endpoint"})
	}
}

func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Logout endpoint"})
	}
}

func GetCurrentUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get current user endpoint"})
	}
}

func adminUserToResponse(user models.AdminUser) adminUserResponse {
	response := adminUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		IsActive: user.IsActive,
	}
	if user.LastLoginAt != nil {
		response.LastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}
	return response
}

func AdminLoginHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload adminLoginPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login payload"})
			return
		}

		username := strings.TrimSpace(strings.ToLower(payload.Username))
		password := strings.TrimSpace(payload.Password)
		if username == "" || password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
			return
		}

		var adminUser models.AdminUser
		if err := db.Where("username = ?", username).Take(&adminUser).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		if !adminUser.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin account is disabled"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(adminUser.PasswordHash), []byte(password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		now := time.Now()
		token, err := middleware.CreateAdminToken(adminUser.ID, cfg.JWTSecret, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create auth token"})
			return
		}
		_ = db.Model(&adminUser).Update("last_login_at", now).Error
		adminUser.LastLoginAt = &now

		c.JSON(http.StatusOK, adminLoginResponse{
			Token: token,
			User:  adminUserToResponse(adminUser),
		})
	}
}

func AdminLogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func GetCurrentAdminUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("admin_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		adminUser, ok := value.(models.AdminUser)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin context invalid"})
			return
		}

		c.JSON(http.StatusOK, adminUserToResponse(adminUser))
	}
}

func GetCurrentTenantHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("tenant")
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}

		tenant, ok := value.(models.Tenant)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant context invalid"})
			return
		}

		c.JSON(http.StatusOK, tenantToResponse(tenant))
	}
}

func GetPlatformConfigHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		config, err := getOrCreatePlatformConfig(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load platform config"})
			return
		}

		c.JSON(http.StatusOK, platformConfigToResponse(config))
	}
}

func GetTenantHostCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		primaryDomain := c.GetString("tenant_primary_domain")
		matchedDomain := c.GetString("tenant_matched_domain")
		if primaryDomain != "" && matchedDomain != "" && primaryDomain != matchedDomain {
			scheme := c.GetHeader("X-Forwarded-Proto")
			if scheme == "" {
				scheme = "http"
			}
			c.Header("X-Primary-Domain", scheme+"://"+primaryDomain)
		}
		c.Status(http.StatusNoContent)
	}
}

// ===== 商品处理器 =====

func GetProductsHandler(db *gorm.DB, cache *productListCache, categoryCache *categoryDescendantCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageInt, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limitInt, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if pageInt < 1 {
			pageInt = 1
		}
		if limitInt < 1 {
			limitInt = 20
		}
		if limitInt > 100 {
			limitInt = 100
		}
		offset := (pageInt - 1) * limitInt
		tenantID := c.GetUint("tenant_id")
		keyword := strings.TrimSpace(c.Query("keyword"))
		brandFilter := strings.TrimSpace(c.Query("brand"))
		sortBy := strings.TrimSpace(c.DefaultQuery("sort", "default"))
		categoryQuery := strings.TrimSpace(c.Query("category"))
		categoryIDInt, _ := strconv.ParseUint(categoryQuery, 10, 64)
		cacheKey := buildProductListCacheKey(tenantID, pageInt, limitInt, keyword, categoryQuery, brandFilter, sortBy)
		if cached, ok := cache.get(cacheKey); ok {
			c.JSON(http.StatusOK, cached)
			return
		}

		query := db.Model(&models.Product{}).
			Joins("LEFT JOIN tenant_product_overrides tpo ON tpo.product_id = products.id AND tpo.tenant_id = ?", tenantID).
			Where("is_active = ?", true).
			Where("JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.status')) = ?", "上架中").
			Where("(tpo.id IS NULL OR tpo.is_visible = ?)", true)

		if keyword != "" {
			likeKeyword := "%" + strings.ToLower(keyword) + "%"
			query = query.Where(`
				LOWER(COALESCE(tpo.custom_name, base_name)) LIKE ? OR
				LOWER(sku) LIKE ? OR
				LOWER(slug) LIKE ? OR
				LOWER(COALESCE(tpo.custom_description, JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.description')))) LIKE ? OR
				LOWER(JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.category'))) LIKE ? OR
				LOWER(JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.brand'))) LIKE ?
			`, likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
		}

		if brandFilter != "" {
			query = query.Where("LOWER(JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.brand'))) = ?", strings.ToLower(brandFilter))
		}

		if categoryIDInt > 0 {
			categoryFilterIDs, err := loadSharedCategoryDescendantIDs(db, categoryCache, uint(categoryIDInt))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load category tree"})
				return
			}

			categoryIDs := make([]uint, 0, len(categoryFilterIDs))
			for id := range categoryFilterIDs {
				categoryIDs = append(categoryIDs, id)
			}
			sort.Slice(categoryIDs, func(i, j int) bool { return categoryIDs[i] < categoryIDs[j] })

			jsonConditions := make([]string, 0, len(categoryIDs))
			args := make([]interface{}, 0, len(categoryIDs)+1)
			if len(categoryIDs) > 0 {
				args = append(args, categoryIDs)
			}
			for _, id := range categoryIDs {
				jsonConditions = append(jsonConditions, "JSON_CONTAINS(JSON_EXTRACT(specifications, '$.categoryIds'), CAST(? AS JSON))")
				args = append(args, strconv.FormatUint(uint64(id), 10))
			}

			if len(jsonConditions) > 0 {
				query = query.Where(
					"(CAST(JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.categoryId')) AS UNSIGNED) IN ? OR "+strings.Join(jsonConditions, " OR ")+")",
					args...,
				)
			} else {
				query = query.Where("CAST(JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.categoryId')) AS UNSIGNED) = ?", categoryIDInt)
			}
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
			return
		}

		sortOrder := "id desc"
		switch sortBy {
		case "price-asc":
			sortOrder = "COALESCE(tpo.custom_price, base_price) asc, id desc"
		case "price-desc":
			sortOrder = "COALESCE(tpo.custom_price, base_price) desc, id desc"
		case "rating":
			sortOrder = "CAST(COALESCE(JSON_UNQUOTE(JSON_EXTRACT(specifications, '$.rating')), '0') AS DECIMAL(10,2)) desc, id desc"
		}

		var products []models.Product
		if err := query.Order(sortOrder).Offset(offset).Limit(limitInt).Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		productIDs := make([]uint, 0, len(products))
		for _, product := range products {
			productIDs = append(productIDs, product.ID)
		}

		overrideMap, err := loadTenantOverrideMapForProducts(db, tenantID, productIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant overrides"})
			return
		}

		result := buildVisibleProductResponses(products, overrideMap)
		response := productListResponse{
			Data:  result,
			Total: total,
			Page:  pageInt,
			Limit: limitInt,
		}
		cache.set(cacheKey, response)

		c.JSON(http.StatusOK, response)
	}
}

func GetProductDetailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		var product models.Product
		if err := db.Where("id = ? AND is_active = ?", productID, true).First(&product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		if !isProductPublished(product) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		tenantID := c.GetUint("tenant_id")
		var override models.TenantProductOverride
		if err := db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).First(&override).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, productToResponse(product, nil))
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product override"})
			return
		}

		if !isProductVisibleForTenant(product, &override) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, productToResponse(product, &override))
	}
}

func CreateProductHandler(db *gorm.DB) gin.HandlerFunc {
	return CreateProductAdminHandler(db)
}

func UpdateProductHandler(db *gorm.DB) gin.HandlerFunc {
	return UpdateProductAdminHandler(db)
}

func DeleteProductHandler(db *gorm.DB) gin.HandlerFunc {
	return DeleteProductAdminHandler(db)
}

func SetProductOverridesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		var payload productOverridePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid override payload"})
			return
		}

		tenantID := c.GetUint("tenant_id")
		var override models.TenantProductOverride
		err = db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).First(&override).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load override"})
			return
		}

		override.TenantID = tenantID
		override.ProductID = productID
		override.CustomName = stringPtr(payload.CustomName)
		override.CustomDescription = stringPtr(payload.CustomDescription)
		override.CustomPrice = payload.CustomPrice
		override.CustomStockQuantity = payload.CustomStockQuantity
		override.CustomImages = stringSliceToJSONArray(payload.CustomImages)
		override.CustomDetailImages = stringSliceToJSONArray(payload.CustomDetailImages)
		override.SEOTitle = stringPtr(payload.SEOTitle)
		override.SEODescription = stringPtr(payload.SEODescription)
		override.IsVisible = payload.IsVisible

		if override.ID == 0 {
			if err := db.Create(&override).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create override"})
				return
			}
		} else if err := db.Save(&override).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update override"})
			return
		}
		invalidateCatalogCaches()

		c.JSON(http.StatusOK, override)
	}
}

func GetProductOverridesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		tenantID := c.GetUint("tenant_id")
		var override models.TenantProductOverride
		if err := db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).First(&override).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch override"})
			return
		}

		c.JSON(http.StatusOK, override)
	}
}

// ===== 分类处理器 =====

func GetCategoriesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []models.Category
		db.Where("tenant_id IS NULL").Order("parent_id asc, sort_order asc, id asc").Find(&categories)

		result := make([]categoryResponse, 0, len(categories))
		for _, category := range categories {
			result = append(result, categoryToResponse(category))
		}

		c.JSON(http.StatusOK, result)
	}
}

func CreateCategoryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload categoryPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category payload"})
			return
		}

		category := models.Category{
			TenantID:  nil,
			Name:      payload.Name,
			ParentID:  payload.ParentID,
			SortOrder: payload.SortOrder,
		}
		if err := db.Create(&category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
			return
		}
		invalidateCategoryCaches()

		c.JSON(http.StatusCreated, categoryToResponse(category))
	}
}

func UpdateCategoryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramKey := c.GetString("param_key")
		if paramKey == "" {
			paramKey = "id"
		}
		categoryID, err := getUintParam(c, paramKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category id"})
			return
		}

		var payload categoryPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category payload"})
			return
		}

		var category models.Category
		if err := db.Where("tenant_id IS NULL AND id = ?", categoryID).First(&category).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		category.Name = payload.Name
		category.ParentID = payload.ParentID
		category.SortOrder = payload.SortOrder

		if err := db.Save(&category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
			return
		}
		invalidateCategoryCaches()

		c.JSON(http.StatusOK, categoryToResponse(category))
	}
}

func DeleteCategoryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramKey := c.GetString("param_key")
		if paramKey == "" {
			paramKey = "id"
		}
		categoryID, err := getUintParam(c, paramKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category id"})
			return
		}

		var category models.Category
		if err := db.Where("tenant_id IS NULL AND id = ?", categoryID).First(&category).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		var childCount int64
		if err := db.Model(&models.Category{}).Where("tenant_id IS NULL AND parent_id = ?", categoryID).Count(&childCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to inspect category tree"})
			return
		}
		if childCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please delete child categories first"})
			return
		}

		defaultCategory, err := ensureDefaultSharedCategory(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure default uncategorized category"})
			return
		}
		if category.ID == defaultCategory.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Default uncategorized category cannot be deleted"})
			return
		}

		var reassignedProductCount int64
		if err := db.Transaction(func(tx *gorm.DB) error {
			defaultCategory, err := ensureDefaultSharedCategory(tx)
			if err != nil {
				return err
			}

			var categories []models.Category
			if err := tx.Where("tenant_id IS NULL").Find(&categories).Error; err != nil {
				return err
			}
			categoryNameByID := make(map[uint]string, len(categories))
			for _, item := range categories {
				categoryNameByID[item.ID] = item.Name
			}

			var products []models.Product
			if err := tx.Find(&products).Error; err != nil {
				return err
			}

			for _, product := range products {
				specs := product.Specifications
				if specs == nil {
					specs = models.JSONMap{}
				}

				changed := false
				currentCategoryID := jsonUint(specs, "categoryId")
				categoryIDs := jsonUintSlice(specs, "categoryIds", nil)

				filteredCategoryIDs := make([]uint, 0, len(categoryIDs))
				for _, id := range categoryIDs {
					if id == categoryID {
						changed = true
						continue
					}
					filteredCategoryIDs = append(filteredCategoryIDs, id)
				}

				if currentCategoryID != nil && *currentCategoryID == categoryID {
					changed = true
					if len(filteredCategoryIDs) > 0 {
						specs["categoryId"] = filteredCategoryIDs[0]
						if name, exists := categoryNameByID[filteredCategoryIDs[0]]; exists {
							specs["category"] = name
						}
					} else {
						specs["categoryId"] = defaultCategory.ID
						specs["category"] = defaultCategory.Name
						filteredCategoryIDs = append(filteredCategoryIDs, defaultCategory.ID)
					}
				}

				if len(filteredCategoryIDs) == 0 && currentCategoryID == nil {
					continue
				}

				if !changed {
					continue
				}

				if nextCategoryPtr := jsonUint(specs, "categoryId"); nextCategoryPtr != nil {
					if *nextCategoryPtr == defaultCategory.ID {
						specs["category"] = defaultCategory.Name
					} else if name, exists := categoryNameByID[*nextCategoryPtr]; exists {
						specs["category"] = name
					}
				}

				specs["categoryIds"] = models.UIntArray(filteredCategoryIDs)
				reassignedProductCount++

				if err := tx.Model(&models.Product{}).
					Where("id = ?", product.ID).
					Updates(map[string]interface{}{
						"specifications": specs,
					}).Error; err != nil {
					return err
				}
			}

			if err := tx.Where("tenant_id IS NULL AND id = ?", categoryID).Delete(&models.Category{}).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to delete category",
				"detail": err.Error(),
			})
			return
		}
		invalidateCategoryCaches()

		c.JSON(http.StatusOK, gin.H{
			"success":                  true,
			"reassigned_product_count": reassignedProductCount,
			"default_category_id":      defaultCategory.ID,
			"default_category_name":    defaultCategory.Name,
		})
	}
}

// ===== 品牌处理器 =====

func GetBrandsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var brands []models.Brand
		db.Order("id desc").Find(&brands)

		result := make([]brandResponse, 0, len(brands))
		for _, brand := range brands {
			result = append(result, brandToResponse(brand))
		}

		c.JSON(http.StatusOK, result)
	}
}

func CreateBrandHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload brandPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand payload"})
			return
		}

		brand := models.Brand{
			Name: payload.Name,
		}
		if payload.LogoURL != "" {
			brand.LogoURL = stringPtr(payload.LogoURL)
		}
		if payload.Description != "" {
			brand.Description = stringPtr(payload.Description)
		}

		if err := db.Create(&brand).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create brand"})
			return
		}

		c.JSON(http.StatusCreated, brandToResponse(brand))
	}
}

func UpdateBrandHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramKey := c.GetString("param_key")
		if paramKey == "" {
			paramKey = "id"
		}
		brandID, err := getUintParam(c, paramKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand id"})
			return
		}

		var payload brandPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand payload"})
			return
		}

		var brand models.Brand
		if err := db.First(&brand, brandID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
			return
		}

		brand.Name = payload.Name
		brand.LogoURL = stringPtr(payload.LogoURL)
		brand.Description = stringPtr(payload.Description)

		if err := db.Save(&brand).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand"})
			return
		}

		c.JSON(http.StatusOK, brandToResponse(brand))
	}
}

func DeleteBrandHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramKey := c.GetString("param_key")
		if paramKey == "" {
			paramKey = "id"
		}
		brandID, err := getUintParam(c, paramKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand id"})
			return
		}

		if err := db.Delete(&models.Brand{}, brandID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete brand"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func withTenantContext(tenantID uint, c *gin.Context, next func()) {
	c.Set("tenant_id", tenantID)
	next()
}

func GetAdminCategoriesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		GetCategoriesHandler(db)(c)
	}
}

func CreateAdminCategoryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		CreateCategoryHandler(db)(c)
	}
}

func UpdateAdminCategoryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("param_key", "category_id")
		UpdateCategoryHandler(db)(c)
	}
}

func DeleteAdminCategoryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("param_key", "category_id")
		DeleteCategoryHandler(db)(c)
	}
}

func GetAdminBrandsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		GetBrandsHandler(db)(c)
	}
}

func CreateAdminBrandHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		CreateBrandHandler(db)(c)
	}
}

func UpdateAdminBrandHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		UpdateBrandHandler(db)(c)
	}
}

func DeleteAdminBrandHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		DeleteBrandHandler(db)(c)
	}
}

// ===== 订单处理器 =====

func CreateOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetUint("tenant_id")
		if tenantID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context is required"})
			return
		}

		var payload createOrderPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order payload"})
			return
		}
		if len(payload.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Order items are required"})
			return
		}
		if payload.Phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is required"})
			return
		}
		if payload.LineID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Line ID is required"})
			return
		}
		if payload.ConvenienceStore == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "711 store is required"})
			return
		}
		if payload.ShippingAddress == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Shipping address is required"})
			return
		}

		orderItems := make([]models.OrderItem, 0, len(payload.Items))
		totalAmount := 0.0

		for _, item := range payload.Items {
			if item.ProductID == 0 || item.Quantity <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order item"})
				return
			}

			var product models.Product
			if err := db.First(&product, item.ProductID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
				return
			}

			price := product.BasePrice
			productSpecs := product.Specifications
			skuVariants := jsonProductSkuVariants(productSpecs, "skuVariants")
			optionGroups := jsonProductOptionGroups(productSpecs, "optionGroups")
			selectedVariantName := strings.TrimSpace(item.VariantName)
			selectedVariantSKU := strings.TrimSpace(item.VariantSKU)
			if len(skuVariants) > 0 {
				if selectedVariantSKU == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Product variant is required"})
					return
				}
				matched := false
				for _, variant := range skuVariants {
					resolvedVariantName := buildVariantDisplayName(optionGroups, variant.Selections)
					matchesBySKU := variant.SKU == selectedVariantSKU
					matchesByName := normalizedVariantLabel(resolvedVariantName) != "" &&
						normalizedVariantLabel(resolvedVariantName) == normalizedVariantLabel(selectedVariantName)
					if !matchesBySKU && !matchesByName {
						continue
					}
					selectedVariantName = resolvedVariantName
					selectedVariantSKU = variant.SKU
					if variant.Price != nil {
						price = *variant.Price
					}
					availableStock := product.BaseStockQuantity
					if variant.Stock != nil {
						availableStock = *variant.Stock
					}
					if item.Quantity > availableStock {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Selected variant stock is insufficient"})
						return
					}
					matched = true
					break
				}
				if !matched {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product variant"})
					return
				}
			} else if item.Quantity > product.BaseStockQuantity {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Product stock is insufficient"})
				return
			}

			var override models.TenantProductOverride
			err := db.Where("tenant_id = ? AND product_id = ?", tenantID, item.ProductID).Take(&override).Error
			if err == nil {
				if override.IsVisible == false {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Product is not available for this tenant"})
					return
				}
				if override.CustomPrice != nil {
					price = *override.CustomPrice
				}
			} else if err != nil && err != gorm.ErrRecordNotFound {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve product pricing"})
				return
			}

			orderItems = append(orderItems, models.OrderItem{
				ProductID:   item.ProductID,
				VariantName: selectedVariantName,
				VariantSKU:  selectedVariantSKU,
				Quantity:    item.Quantity,
				Price:       price,
			})
			totalAmount += price * float64(item.Quantity)
		}

		order := models.Order{
			TenantID:         tenantID,
			TotalAmount:      totalAmount,
			Status:           "已下单",
			LineID:           payload.LineID,
			Phone:            payload.Phone,
			ConvenienceStore: payload.ConvenienceStore,
			ShippingAddress:  payload.ShippingAddress,
			PaymentMethod:    payload.PaymentMethod,
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Omit("UserID").Create(&order).Error; err != nil {
				return err
			}
			for index := range orderItems {
				orderItems[index].OrderID = order.ID
				if err := tx.Create(&orderItems[index]).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		productNameByID, err := loadOrderProductNameMap(db, orderItems)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve order product names"})
			return
		}

		c.JSON(http.StatusCreated, orderToResponse(order, orderItems, productNameByID))
	}
}

func GetOrdersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetUint("tenant_id")
		if tenantID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context is required"})
			return
		}

		var orders []models.Order
		if err := db.Where("tenant_id = ?", tenantID).Order("id desc").Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
			return
		}

		result := make([]orderResponse, 0, len(orders))
		for _, order := range orders {
			var items []models.OrderItem
			if err := db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
				return
			}
			productNameByID, err := loadOrderProductNameMap(db, items)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve order product names"})
				return
			}
			result = append(result, orderToResponse(order, items, productNameByID))
		}

		c.JSON(http.StatusOK, result)
	}
}

func GetOrderDetailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetUint("tenant_id")
		orderID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order id"})
			return
		}

		var order models.Order
		if err := db.Where("tenant_id = ? AND id = ?", tenantID, orderID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		var items []models.OrderItem
		if err := db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
			return
		}

		productNameByID, err := loadOrderProductNameMap(db, items)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve order product names"})
			return
		}

		c.JSON(http.StatusOK, orderToResponse(order, items, productNameByID))
	}
}

func GetAdminOrdersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orders []models.Order
		if err := db.Order("id desc").Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
			return
		}

		result := make([]orderResponse, 0, len(orders))
		for _, order := range orders {
			var items []models.OrderItem
			if err := db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
				return
			}
			productNameByID, err := loadOrderProductNameMap(db, items)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve order product names"})
				return
			}
			result = append(result, orderToResponse(order, items, productNameByID))
		}

		c.JSON(http.StatusOK, result)
	}
}

func GetDomainsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := strings.TrimSpace(c.Query("search"))
		query := db.Model(&models.Domain{})
		if search != "" {
			like := "%" + search + "%"
			query = query.Where("domain_name LIKE ? OR registrar LIKE ?", like, like)
		}

		var domains []models.Domain
		if err := query.Order("id desc").Find(&domains).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch domains"})
			return
		}

		result := make([]domainResponse, 0, len(domains))
		for _, domain := range domains {
			result = append(result, domainToResponse(domain))
		}

		c.JSON(http.StatusOK, result)
	}
}

func CreateDomainHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload domainPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain payload"})
			return
		}

		domainName := strings.TrimSpace(strings.ToLower(payload.DomainName))
		if domainName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain name is required"})
			return
		}

		var existingCount int64
		if err := db.Model(&models.Domain{}).Where("domain_name = ?", domainName).Count(&existingCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate domain uniqueness"})
			return
		}
		if existingCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain already exists"})
			return
		}

		expireDate, err := parseOptionalDate(payload.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expire date format, expected YYYY-MM-DD"})
			return
		}

		domain := models.Domain{
			DomainName: domainName,
			Registrar:  firstNonEmptyString(payload.Registrar, "manual"),
			ExpireDate: expireDate,
			DNSRecords: models.JSONArray{},
		}

		if err := db.Create(&domain).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create domain"})
			return
		}

		c.JSON(http.StatusCreated, domainToResponse(domain))
	}
}

func UpdateDomainHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		domainID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain id"})
			return
		}

		var payload domainPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain payload"})
			return
		}

		var domain models.Domain
		if err := db.First(&domain, domainID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}

		domainName := strings.TrimSpace(strings.ToLower(payload.DomainName))
		if domainName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain name is required"})
			return
		}

		var existingCount int64
		if err := db.Model(&models.Domain{}).
			Where("domain_name = ? AND id <> ?", domainName, domainID).
			Count(&existingCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate domain uniqueness"})
			return
		}
		if existingCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain already exists"})
			return
		}

		expireDate, err := parseOptionalDate(payload.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expire date format, expected YYYY-MM-DD"})
			return
		}

		domain.DomainName = domainName
		domain.Registrar = firstNonEmptyString(payload.Registrar, "manual")
		domain.ExpireDate = expireDate

		if err := db.Save(&domain).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update domain"})
			return
		}

		c.JSON(http.StatusOK, domainToResponse(domain))
	}
}

func DeleteDomainHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		domainID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain id"})
			return
		}

		if err := db.Delete(&models.Domain{}, domainID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func SyncDomainsFromGoDaddyHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload domainSyncPayload
		_ = c.ShouldBindJSON(&payload)

		var remoteDomains []struct {
			Domain  string `json:"domain"`
			Expires string `json:"expires"`
		}
		var err error
		remoteDomains, err = godaddyRequest[[]struct {
			Domain  string `json:"domain"`
			Expires string `json:"expires"`
		}](cfg, http.MethodGet, "/v1/domains", nil)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		created := 0
		updated := 0
		for _, item := range remoteDomains {
			domainName := strings.TrimSpace(strings.ToLower(item.Domain))
			if domainName == "" {
				continue
			}

			var domain models.Domain
			findErr := db.Where("domain_name = ?", domainName).First(&domain).Error
			if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync domains"})
				return
			}

			var expireDate *time.Time
			if strings.TrimSpace(item.Expires) != "" {
				parsed, parseErr := time.Parse(time.RFC3339, item.Expires)
				if parseErr == nil {
					expireDate = &parsed
				}
			}

			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				domain = models.Domain{
					DomainName: domainName,
					Registrar:  "godaddy",
					ExpireDate: expireDate,
					DNSRecords: models.JSONArray{},
				}
				if err := db.Create(&domain).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create synced domain"})
					return
				}
				created++
				continue
			}

			domain.Registrar = "godaddy"
			domain.ExpireDate = expireDate
			if err := db.Save(&domain).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update synced domain"})
				return
			}
			updated++
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"created": created,
			"updated": updated,
			"total":   created + updated,
		})
	}
}

func GetDomainDNSRecordsHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		domainID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain id"})
			return
		}

		var domain models.Domain
		if err := db.First(&domain, domainID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}

		records, err := godaddyRequest[[]dnsRecordPayload](cfg, http.MethodGet, "/v1/domains/"+domain.DomainName+"/records", nil)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		domain.DNSRecords = dnsRecordsToJSONArray(records)
		_ = db.Save(&domain).Error
		c.JSON(http.StatusOK, records)
	}
}

func UpdateDomainDNSRecordsHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		domainID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain id"})
			return
		}

		var records []dnsRecordPayload
		if err := c.ShouldBindJSON(&records); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid DNS records payload"})
			return
		}

		var domain models.Domain
		if err := db.First(&domain, domainID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}

		payload, err := json.Marshal(records)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode DNS records"})
			return
		}

		if _, err := godaddyRequest[[]map[string]interface{}](cfg, http.MethodPut, "/v1/domains/"+domain.DomainName+"/records", bytes.NewReader(payload)); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		domain.DNSRecords = dnsRecordsToJSONArray(records)
		if err := db.Save(&domain).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist DNS records"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func CheckDomainDNSStatusHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var domains []models.Domain
		if err := db.Find(&domains).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch domains"})
			return
		}

		checked := 0
		blocked := 0
		for _, domain := range domains {
			ip, err := resolveDomainIP(cfg.DNSCheckServer, domain.DomainName)
			now := time.Now()
			if err == nil && ip != "" {
				domain.LastCheckIP = &ip
				domain.IsBlocked = ip == cfg.DNSBlockedIP
				if domain.IsBlocked {
					blocked++
				}
			}
			domain.LastCheckedAt = &now
			if saveErr := db.Save(&domain).Error; saveErr == nil {
				checked++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"checked": checked,
			"blocked": blocked,
			"total":   len(domains),
		})
	}
}

func UpdateOrderStatusHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Update order status endpoint"})
	}
}

// ===== 后台管理处理器 =====

func GetTenantsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenants []models.Tenant
		if err := db.Order("id desc").Find(&tenants).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
			return
		}

		result := make([]tenantResponse, 0, len(tenants))
		for _, tenant := range tenants {
			result = append(result, tenantToResponse(tenant))
		}
		c.JSON(http.StatusOK, result)
	}
}

func CreateTenantHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload tenantPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant payload"})
			return
		}
		normalizedPrimary := normalizeDomain(payload.Domain)
		normalizedBoundDomains := normalizeDomainList(payload.BoundDomains)
		if err := validateTenantDomains(db, 0, normalizedPrimary, normalizedBoundDomains); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tenant := tenantPayloadToModel(payload, nil)
		if err := db.Create(&tenant).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
			return
		}

		addedDomains := append([]string{tenant.Domain}, jsonArrayToStrings(tenant.BoundDomains)...)
		_ = syncDomainsToGSC(cfg, addedDomains)

		c.JSON(http.StatusCreated, tenantToResponse(tenant))
	}
}

func UpdateTenantHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		var payload tenantPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant payload"})
			return
		}
		normalizedPrimary := normalizeDomain(payload.Domain)
		normalizedBoundDomains := normalizeDomainList(payload.BoundDomains)
		if err := validateTenantDomains(db, tenantID, normalizedPrimary, normalizedBoundDomains); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		previousDomains := append([]string{normalizeDomain(tenant.Domain)}, jsonArrayToStrings(tenant.BoundDomains)...)

		tenant = tenantPayloadToModel(payload, &tenant)
		if err := db.Save(&tenant).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
			return
		}
		currentDomains := append([]string{normalizeDomain(tenant.Domain)}, jsonArrayToStrings(tenant.BoundDomains)...)
		_ = syncDomainsToGSC(cfg, diffAddedDomains(previousDomains, currentDomains))

		c.JSON(http.StatusOK, tenantToResponse(tenant))
	}
}

func DeleteTenantHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		if err := db.Where("tenant_id = ?", tenantID).Delete(&models.TenantProductOverride{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant overrides"})
			return
		}
		if err := db.Where("id = ?", tenantID).Delete(&models.Tenant{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func AddTenantDomainHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		var payload tenantDomainPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain payload"})
			return
		}

		domain := normalizeDomain(payload.Domain)
		if domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
			return
		}
		if err := ensureDomainExists(db, domain); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}

		currentPrimary := normalizeDomain(tenant.Domain)
		currentBound := jsonArrayToStrings(tenant.BoundDomains)
		if domain == currentPrimary {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is already the primary domain"})
			return
		}
		for _, item := range currentBound {
			if normalizeDomain(item) == domain {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is already bound to the tenant"})
				return
			}
		}

		updatedBound := addUniqueDomain(currentBound, domain)
		if err := validateTenantDomains(db, tenantID, currentPrimary, updatedBound); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tenant.BoundDomains = stringSliceToJSONArray(updatedBound)
		if err := db.Save(&tenant).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bind tenant domain"})
			return
		}

		npmProxyHostID := jsonUint(tenant.ThemeConfig, "npmProxyHostId")
		var npmResult *service.NPMResult
		var syncErr error
		if npmProxyHostID != nil && *npmProxyHostID > 0 {
			npmResult, syncErr = syncTenantDomainsToNPMByProxyHostID(cfg, *npmProxyHostID, tenant.Domain, jsonArrayToStrings(tenant.BoundDomains))
		} else {
			npmResult, syncErr = syncTenantDomainsToNPM(cfg, tenant.Domain, jsonArrayToStrings(tenant.BoundDomains))
		}
		if syncErr != nil {
			npmResult = &service.NPMResult{
				Status:     "error",
				Message:    syncErr.Error(),
				NPMUpdated: false,
			}
		}
		gscResults := syncDomainsToGSC(cfg, []string{domain})

		c.JSON(http.StatusOK, tenantDomainOperationResponse{
			TenantResponse: tenantToResponse(tenant),
			NPMResult:      npmResult,
			GSCResults:     gscResults,
		})
	}
}

func RemoveTenantDomainHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		domain := normalizeDomain(c.Param("domain"))
		if domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}

		if domain == normalizeDomain(tenant.Domain) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Primary domain cannot be removed directly"})
			return
		}

		updatedBound, found := removeDomainFromList(jsonArrayToStrings(tenant.BoundDomains), domain)
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bound domain not found"})
			return
		}
		if err := validateTenantDomains(db, tenantID, normalizeDomain(tenant.Domain), updatedBound); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tenant.BoundDomains = stringSliceToJSONArray(updatedBound)
		if err := db.Save(&tenant).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove tenant domain"})
			return
		}

		npmResult, syncErr := syncTenantDomainsToNPM(cfg, tenant.Domain, jsonArrayToStrings(tenant.BoundDomains))
		if syncErr != nil {
			npmResult = &service.NPMResult{
				Status:     "error",
				Message:    syncErr.Error(),
				NPMUpdated: false,
			}
		}

		c.JSON(http.StatusOK, tenantDomainOperationResponse{
			TenantResponse: tenantToResponse(tenant),
			NPMResult:      npmResult,
		})
	}
}

func SetTenantPrimaryDomainHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		domain := normalizeDomain(c.Param("domain"))
		if domain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
			return
		}
		if err := ensureDomainExists(db, domain); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}

		currentPrimary := normalizeDomain(tenant.Domain)
		if domain == currentPrimary {
			c.JSON(http.StatusOK, tenantToResponse(tenant))
			return
		}

		currentBound := jsonArrayToStrings(tenant.BoundDomains)
		updatedBound, found := removeDomainFromList(currentBound, domain)
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bound domain not found"})
			return
		}
		if currentPrimary != "" {
			updatedBound = addUniqueDomain(updatedBound, currentPrimary)
		}
		if err := validateTenantDomains(db, tenantID, domain, updatedBound); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tenant.Domain = domain
		tenant.BoundDomains = stringSliceToJSONArray(updatedBound)
		if err := db.Save(&tenant).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant primary domain"})
			return
		}

		c.JSON(http.StatusOK, tenantDomainOperationResponse{
			TenantResponse: tenantToResponse(tenant),
			GSCResults:     syncDomainsToGSC(cfg, []string{domain}),
		})
	}
}

func GetAdminPlatformConfigHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		config, err := getOrCreatePlatformConfig(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch platform config"})
			return
		}

		c.JSON(http.StatusOK, platformConfigToResponse(config))
	}
}

func UpdateAdminPlatformConfigHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload platformConfigPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform config payload"})
			return
		}

		config, err := getOrCreatePlatformConfig(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load platform config"})
			return
		}

		config.LineContactURL = strings.TrimSpace(payload.LineContactURL)
		config.FaqHTML = strings.TrimSpace(payload.FaqHTML)
		config.ShippingFee = payload.ShippingFee
		config.FreeShippingThreshold = payload.FreeShippingThreshold
		config.FeaturedCategoryIDs = sanitizeFeaturedCategoryIDs(payload.FeaturedCategoryIDs)
		config.FeaturedBrandIDs = sanitizeFeaturedCategoryIDs(payload.FeaturedBrandIDs)
		if err := db.Save(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update platform config"})
			return
		}

		c.JSON(http.StatusOK, platformConfigToResponse(config))
	}
}

func sanitizeFeaturedCategoryIDs(input []uint) models.UIntArray {
	if len(input) == 0 {
		return models.UIntArray{}
	}

	result := make(models.UIntArray, 0, len(input))
	seen := make(map[uint]struct{}, len(input))
	for _, id := range input {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func GetDashboardHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get dashboard endpoint"})
	}
}

func GetAllProductsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product
		if err := db.Order("id desc").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		result := make([]productResponse, 0, len(products))
		for _, product := range products {
			result = append(result, productToResponse(product, nil))
		}
		c.JSON(http.StatusOK, result)
	}
}

func CreateProductAdminHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload productPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product payload"})
			return
		}

		product := productPayloadToModel(payload, nil)
		slug, err := generateProductSlug(db, 0, product.BaseName, product.SKU)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate product slug"})
			return
		}
		product.Slug = slug
		if err := db.Create(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}
		invalidateCatalogCaches()

		c.JSON(http.StatusCreated, productToResponse(product, nil))
	}
}

func UpdateProductAdminHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		var payload productPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product payload"})
			return
		}

		var product models.Product
		if err := db.First(&product, productID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		product = productPayloadToModel(payload, &product)
		slug, err := generateProductSlug(db, product.ID, product.BaseName, product.SKU)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate product slug"})
			return
		}
		product.Slug = slug
		if err := db.Save(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}
		invalidateCatalogCaches()

		c.JSON(http.StatusOK, productToResponse(product, nil))
	}
}

func BulkUpdateProductsAdminHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload bulkProductUpdatePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bulk update payload"})
			return
		}

		if len(payload.ProductIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No product ids provided"})
			return
		}

		normalizedStatus := ""
		if payload.Status != nil {
			normalizedStatus = strings.TrimSpace(*payload.Status)
		}
		if payload.Status == nil && payload.IsActive == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		var products []models.Product
		if err := db.Where("id IN ?", payload.ProductIDs).Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load products for bulk update"})
			return
		}

		if len(products) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Products not found"})
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, product := range products {
				if payload.IsActive != nil {
					product.IsActive = *payload.IsActive
				}

				specs := product.Specifications
				if specs == nil {
					specs = models.JSONMap{}
				}
				if payload.Status != nil {
					specs["status"] = normalizedStatus
				}
				product.Specifications = specs

				if err := tx.Save(&product).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk update products"})
			return
		}
		invalidateCatalogCaches()

		if err := db.Where("id IN ?", payload.ProductIDs).Order("id desc").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated products"})
			return
		}

		result := make([]productResponse, 0, len(products))
		for _, product := range products {
			result = append(result, productToResponse(product, nil))
		}

		c.JSON(http.StatusOK, result)
	}
}

func DeleteProductAdminHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		if err := db.Where("product_id = ?", productID).Delete(&models.TenantProductOverride{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product overrides"})
			return
		}
		if err := db.Where("id = ?", productID).Delete(&models.Product{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}
		invalidateCatalogCaches()

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func GetProductOverrideHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		tenantID, err := getUintParam(c, "tenant_id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		var override models.TenantProductOverride
		if err := db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).First(&override).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product override"})
			return
		}

		c.JSON(http.StatusOK, override)
	}
}

func UpdateProductOverrideHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		tenantID, err := getUintParam(c, "tenant_id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tenant"})
			return
		}

		var product models.Product
		if err := db.First(&product, productID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load product"})
			return
		}

		var payload productOverridePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid override payload"})
			return
		}

		var override models.TenantProductOverride
		err = db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).First(&override).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load override"})
			return
		}

		override.TenantID = tenantID
		override.ProductID = productID
		override.CustomName = stringPtr(payload.CustomName)
		override.CustomDescription = stringPtr(payload.CustomDescription)
		override.CustomPrice = payload.CustomPrice
		override.CustomStockQuantity = payload.CustomStockQuantity
		override.CustomImages = stringSliceToJSONArray(payload.CustomImages)
		override.CustomDetailImages = stringSliceToJSONArray(payload.CustomDetailImages)
		override.SEOTitle = stringPtr(payload.SEOTitle)
		override.SEODescription = stringPtr(payload.SEODescription)
		override.IsVisible = payload.IsVisible

		if override.ID == 0 {
			if err := db.Create(&override).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product override"})
				return
			}
		} else if err := db.Save(&override).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product override"})
			return
		}
		invalidateCatalogCaches()

		c.JSON(http.StatusOK, override)
	}
}

func GenerateProductOverrideDraftHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, err := getUintParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
			return
		}

		tenantID, err := getUintParam(c, "tenant_id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant id"})
			return
		}

		if strings.TrimSpace(cfg.DeepSeekAPIKey) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "DeepSeek API key is not configured"})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tenant"})
			return
		}

		var product models.Product
		if err := db.First(&product, productID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load product"})
			return
		}

		var payload generateOverrideDraftPayload
		if err := c.ShouldBindJSON(&payload); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid generate payload"})
			return
		}

		var override models.TenantProductOverride
		err = db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).First(&override).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load override"})
			return
		}

		prompt := buildOverrideGenerationPrompt(product, tenant, override, payload.Instruction)
		result, err := generateOverrideDraftWithDeepSeek(cfg, prompt)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to generate override draft: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func BulkGenerateProductOverrideNamesHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(cfg.DeepSeekAPIKey) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "DeepSeek API key is not configured"})
			return
		}

		var payload bulkGenerateOverrideNamesPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bulk generate payload"})
			return
		}

		if payload.TenantID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant id is required"})
			return
		}
		if len(payload.ProductIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ids are required"})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, payload.TenantID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tenant"})
			return
		}

		var products []models.Product
		if err := db.Where("id IN ?", payload.ProductIDs).Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load products"})
			return
		}
		if len(products) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Products not found"})
			return
		}

		productByID := make(map[uint]models.Product, len(products))
		for _, product := range products {
			productByID[product.ID] = product
		}

		var overrides []models.TenantProductOverride
		if err := db.Where("tenant_id = ? AND product_id IN ?", payload.TenantID, payload.ProductIDs).Find(&overrides).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load overrides"})
			return
		}

		overrideByProductID := make(map[uint]models.TenantProductOverride, len(overrides))
		for _, override := range overrides {
			overrideByProductID[override.ProductID] = override
		}

		results := make([]generatedOverrideNameResponse, 0, len(payload.ProductIDs))

		err := db.Transaction(func(tx *gorm.DB) error {
			for _, productID := range payload.ProductIDs {
				product, exists := productByID[productID]
				if !exists {
					continue
				}

				override := overrideByProductID[productID]
				prompt := buildOverrideNameGenerationPrompt(product, tenant, override, payload.Instruction)
				customName, err := generateOverrideNameWithDeepSeek(cfg, prompt)
				if err != nil {
					return err
				}
				if strings.TrimSpace(customName) == "" {
					continue
				}

				override.TenantID = payload.TenantID
				override.ProductID = productID
				override.CustomName = stringPtr(customName)
				if override.ID == 0 {
					override.IsVisible = true
					if err := tx.Create(&override).Error; err != nil {
						return err
					}
				} else if err := tx.Save(&override).Error; err != nil {
					return err
				}

				results = append(results, generatedOverrideNameResponse{
					ProductID:  productID,
					TenantID:   payload.TenantID,
					CustomName: customName,
				})
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to bulk generate custom names: " + err.Error()})
			return
		}

		invalidateCatalogCaches()
		c.JSON(http.StatusOK, results)
	}
}
