package service

import (
	"github.com/vape-group/backend/internal/models"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// GetProductForTenant 获取租户可见的商品（合并全局和覆盖数据）
func (s *ProductService) GetProductForTenant(productID uint, tenantID uint) (*models.ProductWithOverride, error) {
	// 获取全局商品
	var product models.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return nil, err
	}

	// 获取租户覆盖数据
	var override models.TenantProductOverride
	s.db.Where("product_id = ? AND tenant_id = ?", productID, tenantID).First(&override)

	// 合并数据
	result := &models.ProductWithOverride{
		Product: product,
	}

	if override.ID != 0 {
		if override.CustomName != nil {
			result.BaseName = *override.CustomName
		}
		if override.CustomPrice != nil {
			result.BasePrice = *override.CustomPrice
		}
		if override.CustomStockQuantity != nil {
			result.BaseStockQuantity = *override.CustomStockQuantity
		}
		if len(override.CustomImages) > 0 {
			result.BaseImages = override.CustomImages
		}
	}

	return result, nil
}

// GetProductsForTenant 获取租户可见的商品列表
func (s *ProductService) GetProductsForTenant(tenantID uint, page, limit int) ([]models.ProductWithOverride, int64, error) {
	var products []models.Product
	var total int64

	offset := (page - 1) * limit

	// 查询全局商品
	s.db.Model(&models.Product{}).Where("is_active = ?", true).Count(&total)
	s.db.Where("is_active = ?", true).Offset(offset).Limit(limit).Find(&products)

	// 获取所有覆盖数据
	var overrides []models.TenantProductOverride
	s.db.Where("tenant_id = ?", tenantID).Find(&overrides)

	// 构建覆盖数据映射
	overrideMap := make(map[uint]*models.TenantProductOverride)
	for i := range overrides {
		overrideMap[overrides[i].ProductID] = &overrides[i]
	}

	// 合并数据
	var results []models.ProductWithOverride
	for _, p := range products {
		result := models.ProductWithOverride{
			Product: p,
		}
		if override, ok := overrideMap[p.ID]; ok {
			if override.CustomName != nil {
				result.BaseName = *override.CustomName
			}
			if override.CustomPrice != nil {
				result.BasePrice = *override.CustomPrice
			}
			if override.CustomStockQuantity != nil {
				result.BaseStockQuantity = *override.CustomStockQuantity
			}
			if len(override.CustomImages) > 0 {
				result.BaseImages = override.CustomImages
			}
		}
		results = append(results, result)
	}

	return results, total, nil
}

// CreateProduct 创建全局商品
func (s *ProductService) CreateProduct(product *models.Product) error {
	return s.db.Create(product).Error
}

// UpdateProduct 更新全局商品
func (s *ProductService) UpdateProduct(id uint, updates map[string]interface{}) error {
	return s.db.Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error
}

// SetProductOverride 设置租户商品覆盖数据
func (s *ProductService) SetProductOverride(tenantID uint, productID uint, override *models.TenantProductOverride) error {
	override.TenantID = tenantID
	override.ProductID = productID

	// 使用 upsert
	return s.db.Save(override).Error
}
