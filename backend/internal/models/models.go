package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/vape-group/backend/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Tenant 租户表
type Tenant struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Domain       string    `gorm:"type:varchar(255);uniqueIndex" json:"domain"`
	BoundDomains JSONArray `gorm:"type:json" json:"bound_domains"`
	Name         string    `json:"name"`
	ThemeConfig  JSONMap   `gorm:"type:json" json:"theme_config"`
	SEOConfig    JSONMap   `gorm:"type:json" json:"seo_config"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PlatformConfig 平台级配置表
type PlatformConfig struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	LineContactURL      string    `gorm:"type:varchar(1024)" json:"line_contact_url"`
	FaqHTML             string    `gorm:"type:longtext" json:"faq_html"`
	ShippingFee         float64   `json:"shipping_fee"`
	FreeShippingThreshold float64 `json:"free_shipping_threshold"`
	FeaturedCategoryIDs UIntArray `gorm:"type:json" json:"featured_category_ids"`
	FeaturedBrandIDs    UIntArray `gorm:"type:json" json:"featured_brand_ids"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// User 用户表
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     uint      `gorm:"index" json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Tenant       *Tenant   `json:"tenant,omitempty"`
}

// AdminUser 后台管理员表
type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex" json:"username"`
	PasswordHash string    `json:"-"`
	Name         string    `gorm:"type:varchar(255)" json:"name"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Product 全局商品表
type Product struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	SKU                 string    `gorm:"type:varchar(255);uniqueIndex" json:"sku"`
	Slug                string    `gorm:"type:varchar(255);uniqueIndex" json:"slug"`
	BaseName            string    `json:"base_name"`
	BasePrice           float64   `json:"base_price"`
	BaseStockQuantity   int       `json:"base_stock_quantity"`
	BaseImages          JSONArray `gorm:"type:json" json:"base_images"`
	DetailImages        JSONArray `gorm:"type:json" json:"detail_images"`
	Specifications      JSONMap   `gorm:"type:json" json:"specifications"`
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ProductWithOverride struct {
	Product
	CustomName          *string   `json:"custom_name,omitempty"`
	CustomDescription   *string   `json:"custom_description,omitempty"`
	CustomPrice         *float64  `json:"custom_price,omitempty"`
	CustomStockQuantity *int      `json:"custom_stock_quantity,omitempty"`
	CustomImages        JSONArray `json:"custom_images,omitempty"`
	CustomDetailImages  JSONArray `json:"custom_detail_images,omitempty"`
	SEOTitle            *string   `json:"seo_title,omitempty"`
	SEODescription      *string   `json:"seo_description,omitempty"`
	IsVisible           bool      `json:"is_visible"`
}

// TenantProductOverride 租户商品覆盖表
type TenantProductOverride struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	TenantID            uint      `gorm:"index" json:"tenant_id"`
	ProductID           uint      `gorm:"index" json:"product_id"`
	CustomName          *string   `json:"custom_name"`
	CustomDescription   *string   `json:"custom_description"`
	CustomPrice         *float64  `json:"custom_price"`
	CustomStockQuantity *int      `json:"custom_stock_quantity"`
	CustomImages        JSONArray `gorm:"type:json" json:"custom_images"`
	CustomDetailImages  JSONArray `gorm:"type:json" json:"custom_detail_images"`
	SEOTitle            *string   `json:"seo_title"`
	SEODescription      *string   `json:"seo_description"`
	IsVisible           bool      `gorm:"default:true" json:"is_visible"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Tenant              *Tenant   `json:"tenant,omitempty"`
	Product             *Product  `json:"product,omitempty"`
}

// Category 分类表
type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  *uint     `gorm:"index" json:"tenant_id"`
	Name      string    `json:"name"`
	ParentID  *uint     `json:"parent_id"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tenant    *Tenant   `json:"tenant,omitempty"`
}

// Brand 品牌表
type Brand struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	LogoURL     *string   `json:"logo_url"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Domain 域名管理表
type Domain struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DomainName    string    `gorm:"type:varchar(255);uniqueIndex" json:"domain_name"`
	Registrar     string    `gorm:"type:varchar(100)" json:"registrar"`
	ExpireDate    *time.Time `json:"expire_date"`
	DNSRecords    JSONArray `gorm:"type:json" json:"dns_records"`
	IsBlocked     bool      `gorm:"default:false" json:"is_blocked"`
	LastCheckIP   *string   `gorm:"type:varchar(45)" json:"last_check_ip"`
	LastCheckedAt *time.Time `json:"last_checked_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Order 订单表
type Order struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	TenantID         uint      `gorm:"index" json:"tenant_id"`
	UserID           uint      `gorm:"index" json:"user_id"`
	TotalAmount      float64   `json:"total_amount"`
	Status           string    `json:"status"` // pending, paid, shipped, completed, cancelled
	LineID           string    `json:"line_id"`
	Phone            string    `json:"phone"`
	ConvenienceStore string    `json:"convenience_store"`
	ShippingAddress  string    `json:"shipping_address"`
	PaymentMethod    string    `json:"payment_method"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Tenant           *Tenant   `json:"tenant,omitempty"`
	User             *User     `json:"user,omitempty"`
}

// OrderItem 订单项表
type OrderItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OrderID     uint      `gorm:"index" json:"order_id"`
	ProductID   uint      `json:"product_id"`
	VariantName string    `json:"variant_name"`
	VariantSKU  string    `json:"variant_sku"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	Order       *Order    `json:"order,omitempty"`
	Product     *Product  `json:"product,omitempty"`
}

// JSONMap 自定义JSON字段类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	return json.Unmarshal(bytes, &j)
}

// JSONArray 自定义JSON数组字段类型
type JSONArray []interface{}

func (j JSONArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	return json.Unmarshal(bytes, &j)
}

// UIntArray 自定义 uint 数组 JSON 字段类型
type UIntArray []uint

func (u UIntArray) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *UIntArray) Scan(value interface{}) error {
	if value == nil {
		*u = UIntArray{}
		return nil
	}

	switch typed := value.(type) {
	case []byte:
		if len(typed) == 0 {
			*u = UIntArray{}
			return nil
		}
		return json.Unmarshal(typed, u)
	case string:
		if typed == "" {
			*u = UIntArray{}
			return nil
		}
		return json.Unmarshal([]byte(typed), u)
	default:
		bytes, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		return json.Unmarshal(bytes, u)
	}
}

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.DBUser + ":" + cfg.DBPass + "@tcp(" + cfg.DBHost + ":" + cfg.DBPort + ")/" + cfg.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func EnsureDevTenants(db *gorm.DB, domains []string) error {
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}

		var count int64
		if err := db.Model(&Tenant{}).Where("domain = ?", domain).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		tenant := Tenant{
			Domain: domain,
			Name:   "Development Store (" + domain + ")",
			ThemeConfig: JSONMap{
				"source": "auto-seeded",
			},
			SEOConfig: JSONMap{
				"title": "Development Store",
			},
			IsActive: true,
		}
		if err := db.Create(&tenant).Error; err != nil {
			return err
		}
	}

	return nil
}

func EnsureAdminUser(db *gorm.DB, username, password, name string) error {
	username = strings.TrimSpace(strings.ToLower(username))
	password = strings.TrimSpace(password)
	name = strings.TrimSpace(name)

	if username == "" || password == "" {
		return nil
	}
	if name == "" {
		name = "Platform Admin"
	}

	var admin AdminUser
	err := db.Where("username = ?", username).Take(&admin).Error
	if err == nil {
		updates := map[string]any{}
		if strings.TrimSpace(admin.Username) == "" && username != "" {
			updates["username"] = username
		}
		if strings.TrimSpace(admin.Name) == "" && name != "" {
			updates["name"] = name
		}
		if !admin.IsActive {
			updates["is_active"] = true
		}
		if len(updates) > 0 {
			return db.Model(&admin).Updates(updates).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.Create(&AdminUser{
		Username:     username,
		PasswordHash: string(passwordHash),
		Name:         name,
		IsActive:     true,
	}).Error
}

func EnsureProductSlugs(db *gorm.DB, slugGenerator func(uint, string, string) (string, error)) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var products []Product
		if err := tx.Order("id asc").Find(&products).Error; err != nil {
			return err
		}

		for _, product := range products {
			if strings.TrimSpace(product.Slug) != "" {
				continue
			}

			slug, err := slugGenerator(product.ID, product.BaseName, product.SKU)
			if err != nil {
				return err
			}

			if err := tx.Model(&Product{}).Where("id = ?", product.ID).Update("slug", slug).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func EnsureSharedCategories(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var categories []Category
		if err := tx.Order("parent_id asc, sort_order asc, id asc").Find(&categories).Error; err != nil {
			return err
		}
		if len(categories) == 0 {
			return nil
		}

		sharedByID := make(map[uint]Category)
		legacyByID := make(map[uint]Category)
		for _, category := range categories {
			if category.TenantID == nil {
				sharedByID[category.ID] = category
				continue
			}
			legacyByID[category.ID] = category
		}
		if len(legacyByID) == 0 {
			return nil
		}

		sharedSignatureCache := make(map[uint]string)
		legacySignatureCache := make(map[uint]string)

		var buildLegacySignature func(Category) string
		buildLegacySignature = func(category Category) string {
			if signature, ok := legacySignatureCache[category.ID]; ok {
				return signature
			}

			signature := normalizeCategorySignatureSegment(category.Name)
			if category.ParentID != nil {
				if parent, ok := legacyByID[*category.ParentID]; ok {
					parentSignature := buildLegacySignature(parent)
					if parentSignature != "" {
						signature = parentSignature + "/" + signature
					}
				}
			}

			legacySignatureCache[category.ID] = signature
			return signature
		}

		var buildSharedSignature func(Category) string
		buildSharedSignature = func(category Category) string {
			if signature, ok := sharedSignatureCache[category.ID]; ok {
				return signature
			}

			signature := normalizeCategorySignatureSegment(category.Name)
			if category.ParentID != nil {
				if parent, ok := sharedByID[*category.ParentID]; ok {
					parentSignature := buildSharedSignature(parent)
					if parentSignature != "" {
						signature = parentSignature + "/" + signature
					}
				}
			}

			sharedSignatureCache[category.ID] = signature
			return signature
		}

		sharedBySignature := make(map[string]Category)
		for _, category := range categories {
			if category.TenantID != nil {
				continue
			}
			signature := buildSharedSignature(category)
			if signature == "" {
				continue
			}
			if _, exists := sharedBySignature[signature]; !exists {
				sharedBySignature[signature] = category
			}
		}

		legacyToShared := make(map[uint]uint)
		var ensureSharedCategory func(Category) (uint, error)
		ensureSharedCategory = func(category Category) (uint, error) {
			if sharedID, ok := legacyToShared[category.ID]; ok {
				return sharedID, nil
			}

			var parentSharedID *uint
			if category.ParentID != nil {
				if parent, ok := legacyByID[*category.ParentID]; ok {
					sharedID, err := ensureSharedCategory(parent)
					if err != nil {
						return 0, err
					}
					parentSharedID = &sharedID
				}
			}

			signature := buildLegacySignature(category)
			if shared, ok := sharedBySignature[signature]; ok {
				legacyToShared[category.ID] = shared.ID
				return shared.ID, nil
			}

			shared := Category{
				TenantID:  nil,
				Name:      category.Name,
				ParentID:  parentSharedID,
				SortOrder: category.SortOrder,
			}
			if err := tx.Create(&shared).Error; err != nil {
				return 0, err
			}

			sharedByID[shared.ID] = shared
			sharedSignatureCache[shared.ID] = signature
			sharedBySignature[signature] = shared
			legacyToShared[category.ID] = shared.ID
			return shared.ID, nil
		}

		for _, category := range categories {
			if category.TenantID == nil {
				continue
			}
			if _, err := ensureSharedCategory(category); err != nil {
				return err
			}
		}

		sharedNameByID := make(map[uint]string, len(sharedByID))
		for id, category := range sharedByID {
			sharedNameByID[id] = category.Name
		}

		var products []Product
		if err := tx.Find(&products).Error; err != nil {
			return err
		}
		for _, product := range products {
			if !remapProductCategory(product.Specifications, legacyToShared, sharedNameByID) {
				continue
			}
			if err := tx.Model(&product).Update("specifications", product.Specifications).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func normalizeCategorySignatureSegment(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func remapProductCategory(specs JSONMap, legacyToShared map[uint]uint, sharedNameByID map[uint]string) bool {
	if specs == nil {
		return false
	}

	changed := false
	categoryID, ok := jsonMapUint(specs, "categoryId")
	if ok {
		sharedID, exists := legacyToShared[categoryID]
		if exists && sharedID != categoryID {
			specs["categoryId"] = sharedID
			if name, found := sharedNameByID[sharedID]; found {
				specs["category"] = name
			}
			changed = true
		}
	}

	rawCategoryIDs, hasCategoryIDs := specs["categoryIds"]
	if hasCategoryIDs {
		normalized := make([]uint, 0)
		seen := make(map[uint]struct{})
		appendID := func(id uint) {
			if id == 0 {
				return
			}
			if _, exists := seen[id]; exists {
				return
			}
			seen[id] = struct{}{}
			normalized = append(normalized, id)
		}

		switch typed := rawCategoryIDs.(type) {
		case []interface{}:
			for _, item := range typed {
				var id uint
				switch value := item.(type) {
				case float64:
					if value > 0 {
						id = uint(value)
					}
				case int:
					if value > 0 {
						id = uint(value)
					}
				case uint:
					id = value
				}
				if id == 0 {
					continue
				}
				if sharedID, exists := legacyToShared[id]; exists && sharedID != id {
					id = sharedID
					changed = true
				}
				appendID(id)
			}
		case []uint:
			for _, id := range typed {
				if sharedID, exists := legacyToShared[id]; exists && sharedID != id {
					id = sharedID
					changed = true
				}
				appendID(id)
			}
		}

		specs["categoryIds"] = normalized
	}

	return changed
}

func jsonMapUint(source JSONMap, key string) (uint, bool) {
	if source == nil {
		return 0, false
	}

	value, exists := source[key]
	if !exists || value == nil {
		return 0, false
	}

	switch typed := value.(type) {
	case uint:
		if typed == 0 {
			return 0, false
		}
		return typed, true
	case uint64:
		if typed == 0 {
			return 0, false
		}
		return uint(typed), true
	case int:
		if typed <= 0 {
			return 0, false
		}
		return uint(typed), true
	case int64:
		if typed <= 0 {
			return 0, false
		}
		return uint(typed), true
	case float64:
		if typed <= 0 {
			return 0, false
		}
		return uint(typed), true
	case float32:
		if typed <= 0 {
			return 0, false
		}
		return uint(typed), true
	default:
		return 0, false
	}
}
