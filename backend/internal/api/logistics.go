package api

import (
	"crypto/md5"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vape-group/backend/config"
)

const (
	ecPayLogisticsStageURL = "https://logistics-stage.ecpay.com.tw/Express/map"
	ecPayLogisticsProdURL  = "https://logistics.ecpay.com.tw/Express/map"
	ecPayStoreListStageURL = "https://logistics-stage.ecpay.com.tw/Helper/GetStoreList"
	ecPayStoreListProdURL  = "https://logistics.ecpay.com.tw/Helper/GetStoreList"
)

type ecPayCvsMapRequest struct {
	ReturnURL string `json:"return_url"`
	Flow      string `json:"flow"`
}

type ecPayCvsMapResponse struct {
	Action string            `json:"action"`
	Method string            `json:"method"`
	Fields map[string]string `json:"fields"`
}

type convenienceStoreLocation struct {
	ID       string `json:"store_id"`
	Name     string `json:"store_name"`
	Address  string `json:"store_address"`
	Phone    string `json:"store_phone"`
	City     string `json:"city"`
	District string `json:"district"`
}

type ecPayStoreListItem struct {
	StoreID    string `json:"StoreId"`
	StoreName  string `json:"StoreName"`
	StoreAddr  string `json:"StoreAddr"`
	StorePhone string `json:"StorePhone"`
}

type ecPayStoreListGroup struct {
	CvsType   string             `json:"CvsType"`
	StoreInfo []ecPayStoreListItem `json:"StoreInfo"`
}

type ecPayStoreListResponse struct {
	RtnCode   int                    `json:"RtnCode"`
	RtnMsg    string                 `json:"RtnMsg"`
	StoreList []ecPayStoreListGroup  `json:"StoreList"`
}

type convenienceStoreCache struct {
	mu        sync.RWMutex
	items     []convenienceStoreLocation
	expiresAt time.Time
}

func newConvenienceStoreCache() *convenienceStoreCache {
	return &convenienceStoreCache{}
}

func (c *convenienceStoreCache) get() ([]convenienceStoreLocation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.items) == 0 || time.Now().After(c.expiresAt) {
		return nil, false
	}
	result := make([]convenienceStoreLocation, len(c.items))
	copy(result, c.items)
	return result, true
}

func (c *convenienceStoreCache) set(items []convenienceStoreLocation, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make([]convenienceStoreLocation, len(items))
	copy(c.items, items)
	c.expiresAt = time.Now().Add(ttl)
}

func GetECPayCvsMapConfigHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.ECPayLogisticsMerchantID == "" || cfg.ECPayLogisticsHashKey == "" || cfg.ECPayLogisticsHashIV == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ECPay logistics config is incomplete"})
			return
		}

		var payload ecPayCvsMapRequest
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		returnURL := strings.TrimSpace(payload.ReturnURL)
		if returnURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "return_url is required"})
			return
		}

		replyURL := buildECPayReplyURL(c)
		flow := strings.TrimSpace(payload.Flow)
		if flow == "" {
			flow = "cart"
		}

		extraData := fmt.Sprintf("%s|%s", flow, returnURL)
		fields := map[string]string{
			"MerchantID":      cfg.ECPayLogisticsMerchantID,
			"MerchantTradeNo": buildMerchantTradeNo(flow),
			"LogisticsType":   "CVS",
			"LogisticsSubType": strings.TrimSpace(cfg.ECPayLogisticsSubType),
			"IsCollection":    "N",
			"ServerReplyURL":  replyURL,
			"ExtraData":       extraData,
			"Device":          "0",
		}
		fields["CheckMacValue"] = buildECPayCheckMacValue(fields, cfg.ECPayLogisticsHashKey, cfg.ECPayLogisticsHashIV)

		c.JSON(http.StatusOK, ecPayCvsMapResponse{
			Action: resolveECPayMapURL(cfg),
			Method: http.MethodPost,
			Fields: fields,
		})
	}
}

func HandleECPayCvsSelectionCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		flow, returnURL := parseECPayExtraData(c.PostForm("ExtraData"))
		if returnURL == "" {
			c.String(http.StatusBadRequest, "missing return url")
			return
		}

		values := url.Values{}
		values.Set("flow", flow)
		values.Set("store_id", c.PostForm("CVSStoreID"))
		values.Set("store_name", c.PostForm("CVSStoreName"))
		values.Set("store_address", c.PostForm("CVSAddress"))
		values.Set("store_telephone", c.PostForm("CVSTelephone"))

		target, err := appendQuery(returnURL, values)
		if err != nil {
			c.String(http.StatusBadRequest, "invalid return url")
			return
		}

		c.Redirect(http.StatusFound, target)
	}
}

func GetECPayConvenienceStoresHandler(cfg *config.Config, cache *convenienceStoreCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.ECPayLogisticsMerchantID == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ECPay logistics config is incomplete"})
			return
		}

		items, err := loadConvenienceStores(cfg, cache)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch convenience stores"})
			return
		}

		c.JSON(http.StatusOK, items)
	}
}

func resolveECPayMapURL(cfg *config.Config) string {
	if cfg.ECPayLogisticsStage {
		return ecPayLogisticsStageURL
	}
	return ecPayLogisticsProdURL
}

func resolveECPayStoreListURL(cfg *config.Config) string {
	if cfg.ECPayLogisticsStage {
		return ecPayStoreListStageURL
	}
	return ecPayStoreListProdURL
}

func buildMerchantTradeNo(flow string) string {
	initial := "C"
	trimmedFlow := strings.TrimSpace(flow)
	if trimmedFlow != "" {
		initial = strings.ToUpper(trimmedFlow[:1])
	}

	sanitizedTime := regexp.MustCompile(`[^0-9]`).ReplaceAllString(time.Now().Format("20060102150405.000"), "")
	return fmt.Sprintf("VG%s%s", initial, sanitizedTime[:17])
}

func buildECPayReplyURL(c *gin.Context) string {
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.Request.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}

	return fmt.Sprintf("%s://%s/api/logistics/ecpay/callback", scheme, host)
}

func buildECPayCheckMacValue(fields map[string]string, hashKey string, hashIV string) string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if key == "CheckMacValue" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("HashKey=")
	builder.WriteString(hashKey)
	for _, key := range keys {
		builder.WriteString("&")
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(fields[key])
	}
	builder.WriteString("&HashIV=")
	builder.WriteString(hashIV)

	escaped := url.QueryEscape(builder.String())
	escaped = strings.ToLower(escaped)
	replacer := strings.NewReplacer(
		"%2d", "-",
		"%5f", "_",
		"%2e", ".",
		"%21", "!",
		"%2a", "*",
		"%28", "(",
		"%29", ")",
		"%20", "+",
	)
	normalized := replacer.Replace(escaped)

	sum := md5.Sum([]byte(normalized))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func parseECPayExtraData(value string) (string, string) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return "cart", ""
	}

	flow := strings.TrimSpace(parts[0])
	if flow == "" {
		flow = "cart"
	}

	return flow, strings.TrimSpace(parts[1])
}

func appendQuery(rawURL string, values url.Values) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	for key, items := range values {
		for _, item := range items {
			query.Set(key, item)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func loadConvenienceStores(cfg *config.Config, cache *convenienceStoreCache) ([]convenienceStoreLocation, error) {
	if items, ok := cache.get(); ok {
		return items, nil
	}

	form := url.Values{}
	form.Set("PlatformID", "")
	form.Set("MerchantID", cfg.ECPayLogisticsMerchantID)
	form.Set("CvsType", strings.TrimSpace(cfg.ECPayLogisticsSubType))
	if form.Get("CvsType") == "" {
		form.Set("CvsType", "UNIMART")
	}
	checkMacFields := map[string]string{
		"MerchantID": cfg.ECPayLogisticsMerchantID,
		"CvsType":    form.Get("CvsType"),
	}
	form.Set("CheckMacValue", buildECPayCheckMacValue(checkMacFields, cfg.ECPayLogisticsHashKey, cfg.ECPayLogisticsHashIV))

	request, err := http.NewRequest(http.MethodPost, resolveECPayStoreListURL(cfg), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return nil, fmt.Errorf("ecpay store list request failed: %s", strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	items, err := parseConvenienceStores(body)
	if err != nil {
		return nil, err
	}

	cache.set(items, 12*time.Hour)
	return items, nil
}

func parseConvenienceStores(body []byte) ([]convenienceStoreLocation, error) {
	var raw ecPayStoreListResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.RtnCode != 1 {
		return nil, fmt.Errorf("ecpay returned %d: %s", raw.RtnCode, strings.TrimSpace(raw.RtnMsg))
	}

	items := make([]convenienceStoreLocation, 0)
	for _, group := range raw.StoreList {
		for _, item := range group.StoreInfo {
			address := strings.TrimSpace(item.StoreAddr)
			if address == "" {
				continue
			}

			city, district := splitTaiwanAddress(address)
			items = append(items, convenienceStoreLocation{
				ID:       strings.TrimSpace(item.StoreID),
				Name:     strings.TrimSpace(item.StoreName),
				Address:  address,
				Phone:    strings.TrimSpace(item.StorePhone),
				City:     city,
				District: district,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if left.City != right.City {
			return left.City < right.City
		}
		if left.District != right.District {
			return left.District < right.District
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		leftID, leftErr := strconv.Atoi(left.ID)
		rightID, rightErr := strconv.Atoi(right.ID)
		if leftErr == nil && rightErr == nil {
			return leftID < rightID
		}
		return left.ID < right.ID
	})

	return items, nil
}

func splitTaiwanAddress(address string) (string, string) {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return "", ""
	}

	citySuffixes := []string{"市", "縣"}
	for _, suffix := range citySuffixes {
		if index := strings.Index(trimmed, suffix); index >= 0 {
			city := strings.TrimSpace(trimmed[:index+len(suffix)])
			rest := strings.TrimSpace(trimmed[index+len(suffix):])
			districtSuffixes := []string{"區", "鄉", "鎮", "市"}
			for _, districtSuffix := range districtSuffixes {
				if districtIndex := strings.Index(rest, districtSuffix); districtIndex >= 0 {
					district := strings.TrimSpace(rest[:districtIndex+len(districtSuffix)])
					return city, district
				}
			}
			return city, ""
		}
	}

	return "", ""
}
