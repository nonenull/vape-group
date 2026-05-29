package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vape-group/backend/config"
	"gorm.io/gorm"
)

// SetupRoutes 设置所有API路由
func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	cacheSeconds, _ := strconv.Atoi(cfg.ProductListCacheSeconds)
	if cacheSeconds < 0 {
		cacheSeconds = 0
	}
	productListCache := newProductListCache(time.Duration(cacheSeconds) * time.Second)
	categoryDescendantCache := newCategoryDescendantCache()
	registerCatalogCaches(productListCache, categoryDescendantCache)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/__tenant_host_check", GetTenantHostCheckHandler())
	router.GET("/api/tenant/current", GetCurrentTenantHandler())
	router.GET("/api/platform-config", GetPlatformConfigHandler(db))
	// 认证相关路由
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", RegisterHandler(db))
		authGroup.POST("/login", LoginHandler(db))
		authGroup.POST("/logout", LogoutHandler())
		authGroup.GET("/me", GetCurrentUserHandler())
	}

	// 商品相关路由
	productGroup := router.Group("/api/products")
	{
		productGroup.GET("", GetProductsHandler(db, productListCache, categoryDescendantCache))
		productGroup.GET("/:id", GetProductDetailHandler(db))
		productGroup.POST("", CreateProductHandler(db))
		productGroup.PUT("/:id", UpdateProductHandler(db))
		productGroup.DELETE("/:id", DeleteProductHandler(db))
		productGroup.PUT("/:id/overrides", SetProductOverridesHandler(db))
		productGroup.GET("/:id/overrides", GetProductOverridesHandler(db))
	}

	// 分类相关路由
	categoryGroup := router.Group("/api/categories")
	{
		categoryGroup.GET("", GetCategoriesHandler(db))
		categoryGroup.POST("", CreateCategoryHandler(db))
		categoryGroup.PUT("/:id", UpdateCategoryHandler(db))
		categoryGroup.DELETE("/:id", DeleteCategoryHandler(db))
	}

	// 品牌相关路由
	brandGroup := router.Group("/api/brands")
	{
		brandGroup.GET("", GetBrandsHandler(db))
		brandGroup.POST("", CreateBrandHandler(db))
		brandGroup.PUT("/:id", UpdateBrandHandler(db))
		brandGroup.DELETE("/:id", DeleteBrandHandler(db))
	}

	// 订单相关路由
	orderGroup := router.Group("/api/orders")
	{
		orderGroup.POST("", CreateOrderHandler(db))
		orderGroup.GET("", GetOrdersHandler(db))
		orderGroup.GET("/:id", GetOrderDetailHandler(db))
		orderGroup.PUT("/:id/status", UpdateOrderStatusHandler(db))
	}

	// 后台管理路由
	adminGroup := router.Group("/api/admin")
	{
		adminGroup.GET("/tenants", GetTenantsHandler(db))
		adminGroup.POST("/tenants", CreateTenantHandler(db))
		adminGroup.PUT("/tenants/:id", UpdateTenantHandler(db))
		adminGroup.DELETE("/tenants/:id", DeleteTenantHandler(db))
		adminGroup.GET("/platform-config", GetAdminPlatformConfigHandler(db))
		adminGroup.PUT("/platform-config", UpdateAdminPlatformConfigHandler(db))
		adminGroup.GET("/dashboard", GetDashboardHandler(db))

		adminGroup.GET("/products", GetAllProductsHandler(db))
		adminGroup.POST("/products", CreateProductAdminHandler(db))
		adminGroup.PUT("/products/bulk-update", BulkUpdateProductsAdminHandler(db))
		adminGroup.POST("/products/bulk-generate-custom-names", BulkGenerateProductOverrideNamesHandler(db, cfg))
		adminGroup.PUT("/products/:id", UpdateProductAdminHandler(db))
		adminGroup.DELETE("/products/:id", DeleteProductAdminHandler(db))
		adminGroup.POST("/uploads/images", UploadImageHandler(cfg))
		adminGroup.GET("/products/:id/overrides/:tenant_id", GetProductOverrideHandler(db))
		adminGroup.PUT("/products/:id/overrides/:tenant_id", UpdateProductOverrideHandler(db))

		adminGroup.GET("/categories", GetAdminCategoriesHandler(db))
		adminGroup.POST("/categories", CreateAdminCategoryHandler(db))
		adminGroup.PUT("/categories/:category_id", UpdateAdminCategoryHandler(db))
		adminGroup.DELETE("/categories/:category_id", DeleteAdminCategoryHandler(db))

		adminGroup.GET("/brands", GetAdminBrandsHandler(db))
		adminGroup.POST("/brands", CreateAdminBrandHandler(db))
		adminGroup.PUT("/brands/:id", UpdateAdminBrandHandler(db))
		adminGroup.DELETE("/brands/:id", DeleteAdminBrandHandler(db))
		adminGroup.GET("/domains", GetDomainsHandler(db))
		adminGroup.POST("/domains", CreateDomainHandler(db))
		adminGroup.POST("/domains/sync", SyncDomainsFromGoDaddyHandler(db, cfg))
		adminGroup.POST("/domains/check-dns", CheckDomainDNSStatusHandler(db, cfg))
		adminGroup.GET("/domains/:id/dns", GetDomainDNSRecordsHandler(db, cfg))
		adminGroup.PUT("/domains/:id/dns", UpdateDomainDNSRecordsHandler(db, cfg))
		adminGroup.PUT("/domains/:id", UpdateDomainHandler(db))
		adminGroup.DELETE("/domains/:id", DeleteDomainHandler(db))
		adminGroup.GET("/orders", GetAdminOrdersHandler(db))
	}
}
