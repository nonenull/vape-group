package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
	"github.com/vape-group/backend/config"
	"github.com/vape-group/backend/internal/models"
	"gorm.io/gorm"
)

type tenantPayload struct {
	Domain         string   `json:"domain"`
	BoundDomains   []string `json:"bound_domains"`
	Name           string   `json:"name"`
	IsActive       bool     `json:"is_active"`
	Theme          string   `json:"theme"`
	HomeTemplate   string   `json:"home_template"`
	HomeModuleOrder []string `json:"home_module_order"`
	PrimaryBrandID *uint    `json:"primary_brand_id"`
	PreviewImage   string   `json:"preview_image"`
	LogoImage      string   `json:"logo_image"`
	AccentColor    string   `json:"accent_color"`
	AccentStrongColor string `json:"accent_strong_color"`
	SurfaceColor   string   `json:"surface_color"`
	PageBgColor    string   `json:"page_bg_color"`
	CardBgColor    string   `json:"card_bg_color"`
	TextColor      string   `json:"text_color"`
	MutedTextColor string   `json:"muted_text_color"`
	BorderColor    string   `json:"border_color"`
	HeroBgColor    string   `json:"hero_bg_color"`
	TagBgColor     string   `json:"tag_bg_color"`
	HeroTitle      string   `json:"hero_title"`
	Tagline        string   `json:"tagline"`
	Announcement   string   `json:"announcement"`
	SupportText    string   `json:"support_text"`
	SEOTitle       string   `json:"seo_title"`
	SEODescription string   `json:"seo_description"`
}

type tenantResponse struct {
	ID             uint     `json:"id"`
	Domain         string   `json:"domain"`
	BoundDomains   []string `json:"bound_domains"`
	Name           string   `json:"name"`
	IsActive       bool     `json:"is_active"`
	Theme          string   `json:"theme"`
	HomeTemplate   string   `json:"home_template"`
	HomeModuleOrder []string `json:"home_module_order"`
	PrimaryBrandID *uint    `json:"primary_brand_id"`
	PreviewImage   string   `json:"preview_image"`
	LogoImage      string   `json:"logo_image"`
	AccentColor    string   `json:"accent_color"`
	AccentStrongColor string `json:"accent_strong_color"`
	SurfaceColor   string   `json:"surface_color"`
	PageBgColor    string   `json:"page_bg_color"`
	CardBgColor    string   `json:"card_bg_color"`
	TextColor      string   `json:"text_color"`
	MutedTextColor string   `json:"muted_text_color"`
	BorderColor    string   `json:"border_color"`
	HeroBgColor    string   `json:"hero_bg_color"`
	TagBgColor     string   `json:"tag_bg_color"`
	HeroTitle      string   `json:"hero_title"`
	Tagline        string   `json:"tagline"`
	Announcement   string   `json:"announcement"`
	SupportText    string   `json:"support_text"`
	SEOTitle       string   `json:"seo_title"`
	SEODescription string   `json:"seo_description"`
}

type platformConfigPayload struct {
	LineContactURL      string `json:"line_contact_url"`
	FeaturedCategoryIDs []uint `json:"featured_category_ids"`
	FeaturedBrandIDs    []uint `json:"featured_brand_ids"`
}

type platformConfigResponse struct {
	ID                  uint   `json:"id"`
	LineContactURL      string `json:"line_contact_url"`
	FeaturedCategoryIDs []uint `json:"featured_category_ids"`
	FeaturedBrandIDs    []uint `json:"featured_brand_ids"`
}

type productPayload struct {
	SKU               string   `json:"sku"`
	BaseName          string   `json:"base_name"`
	BasePrice         float64  `json:"base_price"`
	BaseStockQuantity int      `json:"base_stock_quantity"`
	Category          string   `json:"category"`
	CategoryID        *uint    `json:"category_id"`
	Brand             string   `json:"brand"`
	BrandID           *uint    `json:"brand_id"`
	PreviewImage      string   `json:"preview_image"`
	Gallery           []string `json:"gallery"`
	DetailImages      []string `json:"detail_images"`
	Status            string   `json:"status"`
	IsActive          bool     `json:"is_active"`
	Description       string   `json:"description"`
	LongDescription   string   `json:"long_description"`
	Badge             string   `json:"badge"`
	Rating            float64  `json:"rating"`
	Reviews           int      `json:"reviews"`
	Flavors           []string `json:"flavors"`
	Variants          []productVariantPayload `json:"variants"`
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
		LineContactURL:      "",
		FeaturedCategoryIDs: models.UIntArray{},
		FeaturedBrandIDs:    models.UIntArray{},
	}
	if err := db.Create(&config).Error; err != nil {
		return config, err
	}
	return config, nil
}

func platformConfigToResponse(config models.PlatformConfig) platformConfigResponse {
	return platformConfigResponse{
		ID:                  config.ID,
		LineContactURL:      config.LineContactURL,
		FeaturedCategoryIDs: []uint(config.FeaturedCategoryIDs),
		FeaturedBrandIDs:    []uint(config.FeaturedBrandIDs),
	}
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
	ID                  uint                   `json:"id"`
	SKU                 string                 `json:"sku"`
	Slug                string                 `json:"slug"`
	BaseName            string                 `json:"base_name"`
	BasePrice           float64                `json:"base_price"`
	BaseStockQuantity   int                    `json:"base_stock_quantity"`
	BaseImages          []string               `json:"base_images"`
	DetailImages        []string               `json:"detail_images"`
	Specifications      map[string]interface{} `json:"specifications"`
	IsActive            bool                   `json:"is_active"`
	Category            string                 `json:"category"`
	CategoryID          *uint                  `json:"category_id,omitempty"`
	Brand               string                 `json:"brand"`
	BrandID             *uint                  `json:"brand_id,omitempty"`
	PreviewImage        string                 `json:"preview_image"`
	Gallery             []string               `json:"gallery"`
	Status              string                 `json:"status"`
	Description         string                 `json:"description"`
	LongDescription     string                 `json:"long_description"`
	Badge               string                 `json:"badge"`
	Rating              float64                `json:"rating"`
	Reviews             int                    `json:"reviews"`
	Flavors             []string               `json:"flavors"`
	Variants            []productVariantPayload `json:"variants"`
	OptionGroups        []productOptionGroupPayload `json:"option_groups"`
	SkuVariants         []productSkuVariantPayload `json:"sku_variants"`
	CustomName          *string                `json:"custom_name,omitempty"`
	CustomDescription   *string                `json:"custom_description,omitempty"`
	CustomPrice         *float64               `json:"custom_price,omitempty"`
	CustomStockQuantity *int                   `json:"custom_stock_quantity,omitempty"`
	CustomImages        []string               `json:"custom_images,omitempty"`
	CustomDetailImages  []string               `json:"custom_detail_images,omitempty"`
	SEOTitle            *string                `json:"seo_title,omitempty"`
	SEODescription      *string                `json:"seo_description,omitempty"`
	IsVisible           bool                   `json:"is_visible"`
	CreatedAt           interface{}            `json:"created_at"`
	UpdatedAt           interface{}            `json:"updated_at"`
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
	VariantName string  `json:"variant_name"`
	VariantSKU  string  `json:"variant_sku"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
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

type uploadImageResponse struct {
	URL string `json:"url"`
}

type bulkProductUpdatePayload struct {
	ProductIDs []uint   `json:"product_ids"`
	Status     *string  `json:"status"`
	IsActive   *bool    `json:"is_active"`
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
	}
	return nil
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

func tenantToResponse(tenant models.Tenant) tenantResponse {
	return tenantResponse{
		ID:             tenant.ID,
		Domain:         tenant.Domain,
		BoundDomains:   jsonArrayToStrings(tenant.BoundDomains),
		Name:           tenant.Name,
		IsActive:       tenant.IsActive,
		Theme:          jsonString(tenant.ThemeConfig, "theme", ""),
		HomeTemplate:   jsonString(tenant.ThemeConfig, "homeTemplate", ""),
		HomeModuleOrder: jsonStringSlice(tenant.ThemeConfig, "homeModuleOrder", []string{}),
		PrimaryBrandID: jsonUint(tenant.ThemeConfig, "primaryBrandId"),
		PreviewImage:   jsonString(tenant.ThemeConfig, "previewImage", ""),
		LogoImage:      jsonString(tenant.ThemeConfig, "logoImage", ""),
		AccentColor:    jsonString(tenant.ThemeConfig, "accentColor", ""),
		AccentStrongColor: jsonString(tenant.ThemeConfig, "accentStrongColor", ""),
		SurfaceColor:   jsonString(tenant.ThemeConfig, "surfaceColor", ""),
		PageBgColor:    jsonString(tenant.ThemeConfig, "pageBgColor", ""),
		CardBgColor:    jsonString(tenant.ThemeConfig, "cardBgColor", ""),
		TextColor:      jsonString(tenant.ThemeConfig, "textColor", ""),
		MutedTextColor: jsonString(tenant.ThemeConfig, "mutedTextColor", ""),
		BorderColor:    jsonString(tenant.ThemeConfig, "borderColor", ""),
		HeroBgColor:    jsonString(tenant.ThemeConfig, "heroBgColor", ""),
		TagBgColor:     jsonString(tenant.ThemeConfig, "tagBgColor", ""),
		HeroTitle:      jsonString(tenant.ThemeConfig, "heroTitle", ""),
		Tagline:        jsonString(tenant.ThemeConfig, "tagline", ""),
		Announcement:   jsonString(tenant.ThemeConfig, "announcement", ""),
		SupportText:    jsonString(tenant.ThemeConfig, "supportText", ""),
		SEOTitle:       jsonString(tenant.SEOConfig, "title", ""),
		SEODescription: jsonString(tenant.SEOConfig, "description", ""),
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
		"theme":        payload.Theme,
		"homeTemplate": payload.HomeTemplate,
		"homeModuleOrder": payload.HomeModuleOrder,
		"primaryBrandId": payload.PrimaryBrandID,
		"previewImage": payload.PreviewImage,
		"logoImage":    payload.LogoImage,
		"accentColor":  payload.AccentColor,
		"accentStrongColor": payload.AccentStrongColor,
		"surfaceColor": payload.SurfaceColor,
		"pageBgColor":  payload.PageBgColor,
		"cardBgColor":  payload.CardBgColor,
		"textColor":    payload.TextColor,
		"mutedTextColor": payload.MutedTextColor,
		"borderColor":  payload.BorderColor,
		"heroBgColor":  payload.HeroBgColor,
		"tagBgColor":   payload.TagBgColor,
		"heroTitle":    payload.HeroTitle,
		"tagline":      payload.Tagline,
		"announcement": payload.Announcement,
		"supportText":  payload.SupportText,
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
		CategoryID:        jsonUint(specs, "categoryId"),
		Brand:             jsonString(specs, "brand", ""),
		BrandID:           jsonUint(specs, "brandId"),
		PreviewImage:      jsonString(specs, "previewImage", ""),
		Gallery:           jsonStringSlice(specs, "gallery", jsonArrayToStrings(product.BaseImages)),
		Status:            jsonString(specs, "status", "草稿"),
		Description:       jsonString(specs, "description", ""),
		LongDescription:   jsonString(specs, "longDescription", ""),
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

	model.SKU = payload.SKU
	model.BaseName = payload.BaseName
	model.BasePrice = payload.BasePrice
	model.BaseStockQuantity = payload.BaseStockQuantity
	model.IsActive = payload.IsActive
	model.BaseImages = stringSliceToJSONArray(gallery)
	model.DetailImages = stringSliceToJSONArray(payload.DetailImages)
	model.Specifications = models.JSONMap{
		"category":        payload.Category,
		"categoryId":      payload.CategoryID,
		"brand":           payload.Brand,
		"brandId":         payload.BrandID,
		"previewImage":    payload.PreviewImage,
		"gallery":         stringSliceToJSONArray(gallery),
		"detailImages":    stringSliceToJSONArray(payload.DetailImages),
		"status":          payload.Status,
		"description":     payload.Description,
		"longDescription": payload.LongDescription,
		"badge":           payload.Badge,
		"rating":          payload.Rating,
		"reviews":         payload.Reviews,
		"flavors":         stringSliceToJSONArray(flavors),
		"variants":        variantSliceToJSONArray(variants),
		"optionGroups":    optionGroupSliceToJSONArray(optionGroups),
		"skuVariants":     skuVariantSliceToJSONArray(skuVariants),
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

func loadTenantOverrideMap(db *gorm.DB, tenantID uint) (map[uint]*models.TenantProductOverride, error) {
	var overrides []models.TenantProductOverride
	if err := db.Where("tenant_id = ?", tenantID).Find(&overrides).Error; err != nil {
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

func orderToResponse(order models.Order, items []models.OrderItem) orderResponse {
	result := make([]orderItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, orderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
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

func GetProductsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageInt, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limitInt, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if pageInt < 1 {
			pageInt = 1
		}
		if limitInt < 1 {
			limitInt = 20
		}
		offset := (pageInt - 1) * limitInt
		tenantID := c.GetUint("tenant_id")

		var products []models.Product
		if err := db.Where("is_active = ?", true).Order("id desc").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		overrideMap, err := loadTenantOverrideMap(db, tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant overrides"})
			return
		}

		visibleProducts := buildVisibleProductResponses(products, overrideMap)
		total := len(visibleProducts)
		if offset > total {
			offset = total
		}
		end := offset + limitInt
		if end > total {
			end = total
		}
		result := visibleProducts[offset:end]

		c.JSON(http.StatusOK, gin.H{
			"data":  result,
			"total": total,
			"page":  pageInt,
			"limit": limitInt,
		})
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

		var childCount int64
		if err := db.Model(&models.Category{}).Where("tenant_id IS NULL AND parent_id = ?", categoryID).Count(&childCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to inspect category tree"})
			return
		}
		if childCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please delete child categories first"})
			return
		}

		if err := db.Where("tenant_id IS NULL AND id = ?", categoryID).Delete(&models.Category{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
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
			selectedVariantName := strings.TrimSpace(item.VariantName)
			selectedVariantSKU := strings.TrimSpace(item.VariantSKU)
			if len(skuVariants) > 0 {
				if selectedVariantSKU == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Product variant is required"})
					return
				}
				matched := false
				for _, variant := range skuVariants {
					if variant.SKU != selectedVariantSKU {
						continue
					}
					parts := make([]string, 0, len(variant.Selections))
					for _, group := range jsonProductOptionGroups(productSpecs, "optionGroups") {
						if value, ok := variant.Selections[group.Name]; ok {
							parts = append(parts, group.Name+"："+value)
						}
					}
					selectedVariantName = strings.Join(parts, " / ")
					selectedVariantSKU = variant.SKU
					if variant.Price != nil {
						price = *variant.Price
					}
					if variant.Stock != nil && item.Quantity > *variant.Stock {
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
			Status:           "pending",
			LineID:           payload.LineID,
			Phone:            payload.Phone,
			ConvenienceStore: payload.ConvenienceStore,
			ShippingAddress:  payload.ShippingAddress,
			PaymentMethod:    payload.PaymentMethod,
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&order).Error; err != nil {
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

		c.JSON(http.StatusCreated, orderToResponse(order, orderItems))
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
			result = append(result, orderToResponse(order, items))
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

		c.JSON(http.StatusOK, orderToResponse(order, items))
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

func CreateTenantHandler(db *gorm.DB) gin.HandlerFunc {
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

		c.JSON(http.StatusCreated, tenantToResponse(tenant))
	}
}

func UpdateTenantHandler(db *gorm.DB) gin.HandlerFunc {
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

		tenant = tenantPayloadToModel(payload, &tenant)
		if err := db.Save(&tenant).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
			return
		}

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

		c.JSON(http.StatusOK, override)
	}
}
