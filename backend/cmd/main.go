package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vape-group/backend/internal/middleware"
	"github.com/vape-group/backend/internal/models"
	"github.com/vape-group/backend/internal/api"
	"github.com/vape-group/backend/config"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库
	db, err := models.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.PlatformConfig{},
		&models.User{},
		&models.Product{},
		&models.TenantProductOverride{},
		&models.Category{},
		&models.Brand{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if err := models.EnsureDevTenants(db, cfg.DevTenantDomains); err != nil {
		log.Fatal("Failed to ensure development tenants:", err)
	}
	if err := models.EnsureSharedCategories(db); err != nil {
		log.Fatal("Failed to ensure shared categories:", err)
	}
	if err := models.EnsureProductSlugs(db, func(productID uint, baseName string) (string, error) {
		return api.GenerateProductSlugForModel(db, productID, baseName)
	}); err != nil {
		log.Fatal("Failed to ensure product slugs:", err)
	}

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatal("Failed to prepare upload directory:", err)
	}

	// 初始化Gin路由
	router := gin.Default()

	// 应用中间件
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.TenantMiddleware(db))

	// 设置API路由
	router.Static("/uploads", cfg.UploadDir)
	api.SetupRoutes(router, db, cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server running on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Server failed:", err)
	}
}
