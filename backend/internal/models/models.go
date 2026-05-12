package models

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/vape-group/backend/config"
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

// Product 全局商品表
type Product struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	SKU                 string    `gorm:"type:varchar(255);uniqueIndex" json:"sku"`
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
	TenantID  uint      `gorm:"index" json:"tenant_id"`
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
	TenantID    uint      `gorm:"index" json:"tenant_id"`
	Name        string    `json:"name"`
	LogoURL     *string   `json:"logo_url"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tenant      *Tenant   `json:"tenant,omitempty"`
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
