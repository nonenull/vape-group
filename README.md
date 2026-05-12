# Vape Group 多租户外贸商城

这是一个多租户外贸电商平台项目，维护一套统一的商品数据，但支持绑定多个域名，每个域名可以配置为独立的商城网站。

## 项目结构

```
vape-group/
├── frontend/           # 商城前端 (Vue.js 3 + TypeScript)
├── admin/              # 后台管理前端 (Vue.js 3 + TypeScript)
├── backend/            # 后端API (Golang + Gin)
├── product.md          # 技术文档
└── docker-compose.yml  # Docker编排配置
```

## 快速开始

### 前置条件
- Node.js 18+
- Go 1.19+
- MySQL 8.0+
- Docker & Docker Compose (可选)

### 开发环境设置

#### 1. 后端设置
```bash
cd backend

# 安装依赖
go mod tidy

# 配置环境变量 (创建 .env 文件)
PORT=8080
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=password
DB_NAME=vape_group
JWT_SECRET=your-secret-key

# 运行迁移和启动服务
go run cmd/main.go
```

#### 2. 商城前端设置
```bash
cd frontend

# 安装依赖
npm install

# 开发环境运行
npm run dev

# 访问: http://localhost:5173
```

#### 3. 后台管理前端设置
```bash
cd admin

# 安装依赖
npm install

# 开发环境运行
npm run dev

# 访问: http://localhost:5174
```

### Docker 部署

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## API 文档

详见 [product.md](./product.md) 中的 API 接口设计章节

### 主要 API 端点

#### 商城接口
- `GET /api/products` - 获取商品列表
- `GET /api/products/{id}` - 获取商品详情
- `POST /api/orders` - 创建订单
- `GET /api/orders` - 获取订单列表

#### 后台管理接口
- `GET /api/admin/tenants` - 获取租户列表
- `GET /api/admin/products` - 获取全局商品列表
- `PUT /api/admin/products/{id}/overrides/{tenant_id}` - 更新租户商品覆盖数据

## 开发指南

### 多租户识别
系统通过请求的域名自动识别租户。在开发环境中，可以通过修改 hosts 文件模拟多个域名：

```
127.0.0.1 tenant1.localhost
127.0.0.1 tenant2.localhost
```

### 数据库初始化
系统会自动创建所有必要的表。首次启动时确保数据库存在：

```sql
CREATE DATABASE IF NOT EXISTS vape_group CHARACTER SET utf8mb4;
```

### 代码规范
- 前端：遵循 Vue.js 3 Composition API 风格
- 后端：遵循 Go 的标准代码规范

## 项目进度

- [x] 项目架构设计
- [x] 数据库模型定义
- [x] 后端基础框架
- [x] 前端项目创建
- [ ] 用户认证模块
- [ ] 商品管理模块
- [ ] 订单管理模块
- [ ] 支付集成
- [ ] 单元测试
- [ ] 性能优化
- [ ] 上线部署

## 许可证

MIT
