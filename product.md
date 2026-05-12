# Vape Group 多租户外贸商城技术文档

## 项目概述

### 项目背景
Vape Group 是一个多租户外贸电商平台，维护一套统一的商品数据，但支持绑定多个域名，每个域名可以配置为独立的商城网站。平台专为台湾市场设计，所有界面和内容必须使用台湾繁体中文。每个租户（域名）可以有独立的品牌形象、产品展示和配置。

### 核心特性
- **多租户架构**：一套商品数据，支持多个域名独立运营
- **统一管理**：后台集中管理所有租户的数据和配置
- **灵活配置**：每个域名可自定义主题、品牌和产品展示
- **智能商品管理**：维护全局商品基础数据，后台根据租户生成定制化详情说明
- **SEO 优化**：每个网站独立支持搜索引擎优化

### 技术架构
- **前端**：Vue.js 3 + TypeScript + Vite（支持多租户主题切换）
- **后端**：Golang + Gin 框架（租户隔离中间件）
- **数据库**：MySQL 8.0+（支持租户数据隔离）
- **部署**：Docker + Nginx（域名路由）
- **版本控制**：Git

### 核心要求
- 全站繁体中文支持
- SEO 优化（每个租户独立）
- 响应式设计（桌面端 + 移动端）
- 高性能和安全性
- 租户数据隔离

## 技术栈详情

### 前端技术栈
- **Vue.js 3**：使用 Composition API，支持动态主题加载
- **TypeScript**：类型安全
- **Vite**：快速构建，支持多环境配置
- **Vue Router 4**：路由管理，支持租户路由隔离
- **Pinia**：状态管理，支持租户状态隔离
- **Axios**：HTTP 客户端，支持租户 API 路由
- **Element Plus**：UI 组件库，支持自定义主题
- **SEO 支持**：动态 meta 标签管理

### 后端技术栈
- **Golang 1.19+**：高性能后端
- **Gin**：Web 框架，支持中间件
- **GORM**：ORM，支持多租户查询
- **JWT**：用户认证，支持租户隔离
- **Viper**：配置管理
- **Zap**：日志，支持租户日志分离
- **Redis**：缓存，支持租户缓存隔离

### 数据库设计
- **MySQL 8.0**：关系型数据库
- **字符集**：utf8mb4
- **租户隔离**：通过 tenant_id 字段实现数据隔离

#### 核心表结构

**tenants（租户表）**
```sql
CREATE TABLE tenants (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    domain VARCHAR(255) UNIQUE NOT NULL COMMENT '域名',
    name VARCHAR(100) NOT NULL COMMENT '租户名称',
    theme_config JSON COMMENT '主题配置',
    seo_config JSON COMMENT 'SEO 配置',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**users（用户表）**
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_email_per_tenant (tenant_id, email),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

**products（商品表 - 全局基础数据）**
```sql
CREATE TABLE products (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    sku VARCHAR(100) UNIQUE NOT NULL COMMENT '商品SKU',
    base_name VARCHAR(255) NOT NULL COMMENT '基础商品名称',
    base_price DECIMAL(10,2) NOT NULL COMMENT '基础价格',
    base_stock_quantity INT NOT NULL DEFAULT 0 COMMENT '基础库存',
    base_images JSON COMMENT '基础图片URL数组',
    specifications JSON COMMENT '规格信息（全局）',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**tenant_product_overrides（租户商品覆盖表）**
```sql
CREATE TABLE tenant_product_overrides (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    custom_name VARCHAR(255) COMMENT '自定义商品名称',
    custom_description TEXT COMMENT '自定义商品描述',
    custom_price DECIMAL(10,2) COMMENT '自定义价格（覆盖基础价格）',
    custom_stock_quantity INT COMMENT '自定义库存（覆盖基础库存）',
    custom_images JSON COMMENT '自定义图片（覆盖基础图片）',
    seo_title VARCHAR(255) COMMENT 'SEO标题',
    seo_description TEXT COMMENT 'SEO描述',
    is_visible BOOLEAN DEFAULT TRUE COMMENT '是否在该租户显示',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    UNIQUE KEY unique_tenant_product (tenant_id, product_id)
);
```

## 商品管理策略

### 数据结构设计
- **全局商品数据**：存储SKU、基础名称、价格、库存、图片等核心数据
- **租户覆盖数据**：每个租户可以覆盖商品的名称、描述、价格、库存、图片等
- **智能合并**：前端展示时，优先使用租户覆盖数据，fallback 到全局数据

### 管理流程
1. **创建全局商品**：管理员在后台添加基础商品信息（SKU、基础价格、库存、规格等）
2. **租户定制**：为每个租户设置自定义名称、描述、SEO信息等
3. **自动生成**：系统可根据租户配置自动生成差异化的商品详情页
4. **库存同步**：全局库存可被租户覆盖，支持独立库存管理

### 优势
- **高效维护**：减少重复数据输入
- **灵活定制**：每个网站可有独特的商品展示
- **集中控制**：统一管理核心数据，避免数据不一致

**categories（分类表）**
```sql
CREATE TABLE categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    name VARCHAR(100) NOT NULL,
    parent_id BIGINT,
    sort_order INT DEFAULT 0,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (parent_id) REFERENCES categories(id)
);
```

**brands（品牌表）**
```sql
CREATE TABLE brands (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    name VARCHAR(100) NOT NULL,
    logo_url VARCHAR(255),
    description TEXT,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

**orders（订单表）**
```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status ENUM('pending', 'paid', 'shipped', 'completed', 'cancelled') DEFAULT 'pending',
    shipping_address TEXT,
    payment_method VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

**order_items（订单项表）**
```sql
CREATE TABLE order_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10,2) NOT NULL COMMENT '快照价格',
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
```

## API 接口设计

### 租户识别
所有 API 请求通过域名自动识别租户，后端中间件解析域名获取 tenant_id。

### 认证接口
- `POST /api/auth/register` - 用户注册（按租户）
- `POST /api/auth/login` - 用户登录（按租户）
- `POST /api/auth/logout` - 用户登出
- `GET /api/auth/me` - 获取当前用户信息

### 商品接口
- `GET /api/products` - 获取商品列表（租户内可见商品，合并全局和租户覆盖数据）
- `GET /api/products/{id}` - 获取商品详情（根据租户返回定制化名称、描述、价格等）
- `POST /api/products` - 创建全局商品（管理员）
- `PUT /api/products/{id}` - 更新全局商品基础数据（管理员）
- `DELETE /api/products/{id}` - 删除全局商品（管理员）
- `PUT /api/products/{id}/overrides` - 设置租户商品覆盖数据（管理员，按租户）
- `GET /api/products/{id}/overrides` - 获取租户商品覆盖数据（管理员）

### 分类和品牌接口
- `GET /api/categories` - 获取租户分类树
- `GET /api/brands` - 获取租户品牌列表

### 订单接口
- `POST /api/orders` - 创建订单（租户内）
- `GET /api/orders` - 获取用户订单列表
- `GET /api/orders/{id}` - 获取订单详情

### 后台管理接口
- `GET /api/admin/tenants` - 租户列表
- `POST /api/admin/tenants` - 创建租户
- `PUT /api/admin/tenants/{id}` - 更新租户配置
- `GET /api/admin/dashboard/{tenant_id}` - 租户仪表板数据
- `GET /api/admin/products` - 全局商品列表
- `POST /api/admin/products` - 创建全局商品
- `PUT /api/admin/products/{id}` - 更新全局商品
- `DELETE /api/admin/products/{id}` - 删除全局商品
- `GET /api/admin/products/{id}/overrides/{tenant_id}` - 获取特定租户的商品覆盖数据
- `PUT /api/admin/products/{id}/overrides/{tenant_id}` - 更新特定租户的商品覆盖数据（自定义名称、描述、价格等）

## 前端架构

### 多租户支持
- 通过域名动态加载租户配置
- 主题和样式动态切换
- 路由和组件按租户隔离

### 项目结构
```
src/
├── components/
│   ├── common/     # 通用组件
│   └── tenant/     # 租户特定组件
├── views/
│   ├── Home.vue
│   ├── Product.vue
│   ├── Cart.vue
│   └── Admin/
├── router/
├── store/
│   ├── tenant.ts   # 租户状态
│   ├── user.ts
│   └── cart.ts
├── api/
├── utils/
├── types/
└── themes/         # 租户主题文件
```

## 后端架构

### 租户中间件
```go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        domain := getDomainFromRequest(c)
        tenant, err := getTenantByDomain(domain)
        if err != nil {
            c.AbortWithStatusJSON(404, gin.H{"error": "Tenant not found"})
            return
        }
        c.Set("tenant_id", tenant.ID)
        c.Next()
    }
}
```

### 项目结构
```
backend/
├── cmd/
├── internal/
│   ├── api/
│   ├── models/
│   ├── service/
│   ├── repository/
│   ├── middleware/
│   │   └── tenant.go  # 租户中间件
│   └── config/
├── pkg/
├── config/
├── migrations/
└── docker/
```

## 部署架构

### Nginx 配置
```nginx
server {
    listen 80;
    server_name ~^(?<subdomain>.+)\.vape-group\.com$;
    
    location / {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Tenant-Domain $subdomain.vape-group.com;
    }
}
```

### Docker Compose
```yaml
version: '3.8'
services:
  backend:
    build: ./backend
    environment:
      - DB_HOST=db
  frontend:
    build: ./frontend
    environment:
      - API_URL=http://backend:8080
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=password
```

## 开发环境配置

### 租户配置
在开发环境中，通过环境变量或配置文件模拟多个租户：
```env
TENANT_DOMAINS=tenant1.localhost,tenant2.localhost
```

### 前端开发
```bash
npm install
npm run dev -- --host 0.0.0.0
```

### 后端开发
```bash
go mod tidy
go run cmd/main.go
```

## 安全考虑

### 租户隔离
- 数据层面：通过 tenant_id 强制隔离
- API 层面：中间件验证租户权限
- 文件存储：按租户分目录存储

### 其他安全措施
- HTTPS 强制使用
- API 密钥管理
- 敏感数据加密
- 定期安全审计

## 测试策略

### 多租户测试
- 每个租户独立测试数据隔离
- 跨租户访问控制测试
- 性能测试：多租户并发访问

此文档将根据开发进度持续更新。