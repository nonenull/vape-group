# Vape Group 开发指南

## 项目进展概览

### ✅ 已完成
1. **项目架构设计**
   - 多租户系统设计
   - 数据库模型定义（SQL脚本）
   - API接口设计规范
   
2. **后端基础框架**
   - Go模块初始化（go.mod）
   - 项目目录结构搭建
   - 数据库模型定义（models.go）
   - 中间件实现（租户识别、CORS）
   - 路由定义和处理器框架
   - 配置管理系统
   - 商品服务层（ProductService）

3. **前端项目初始化**
   - Vue.js 3 项目创建（frontend/）
   - Vue.js 3 项目创建（admin/）
   - API客户端配置
   - Pinia stores（租户管理、商品管理）

4. **部署配置**
   - Docker Compose 编排配置
   - Nginx 反向代理配置（多租户域名路由）
   - Dockerfile（后端、前端、后台）
   - 环境变量配置示例

5. **文档**
   - 技术文档（product.md）
   - README文档
   - .gitignore配置

### 🔄 进行中
- npm install 安装前端依赖（预计5-10分钟）

### ⏳ 待完成
1. **核心功能实现**
   - [ ] 用户认证（注册、登录、JWT）
   - [ ] 商品管理（增删改查、多租户覆盖）
   - [ ] 分类和品牌管理
   - [ ] 购物车功能
   - [ ] 订单管理
   - [ ] 支付集成

2. **前端页面**
   - [ ] 首页
   - [ ] 商品列表页
   - [ ] 商品详情页
   - [ ] 购物车
   - [ ] 结算页面
   - [ ] 后台管理界面

3. **测试和优化**
   - [ ] 单元测试
   - [ ] 集成测试
   - [ ] 性能测试
   - [ ] 安全审计

## 快速开始

### 前置条件
```bash
# 检查 Node.js 版本
node --version  # 应为 18+

# 检查 Go 版本
go version      # 应为 1.19+

# 检查 Docker
docker --version
```

### 后端开发

```bash
cd backend

# 下载依赖
go mod download

# 创建 .env 文件（复制 .env.example）
cp .env.example .env

# 配置数据库连接
# 编辑 .env，修改以下内容：
# DB_HOST=localhost
# DB_PORT=3306
# DB_USER=root
# DB_PASS=your_password
# DB_NAME=vape_group

# 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS vape_group CHARACTER SET utf8mb4;"

# 运行服务
go run cmd/main.go

# 服务将在 http://localhost:8080 启动
```

### 前端开发（等待 npm install 完成后）

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问 http://localhost:5173
```

### 后台管理前端开发（等待 npm install 完成后）

```bash
cd admin

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问 http://localhost:5174
```

## 多租户本地开发

### 方法1：使用 localhost:port

- 商城：http://localhost:5173
- 后台：http://localhost:5174

### 方法2：使用 hosts 配置（推荐）

编辑 `/etc/hosts`（Linux/Mac）或 `C:\Windows\System32\drivers\etc\hosts`（Windows）：

```
127.0.0.1 tenant1.localhost
127.0.0.1 tenant2.localhost
127.0.0.1 admin.localhost
```

然后访问：
- 租户1商城：http://tenant1.localhost:5173
- 租户2商城：http://tenant2.localhost:5173
- 后台管理：http://admin.localhost:5174

## 目录结构详解

```
vape-group/
├── backend/                    # Golang后端
│   ├── cmd/
│   │   └── main.go            # 应用入口
│   ├── internal/
│   │   ├── api/               # API路由和处理器
│   │   ├── models/            # 数据模型
│   │   ├── service/           # 业务逻辑
│   │   ├── repository/        # 数据访问层
│   │   └── middleware/        # 中间件
│   ├── config/                # 配置管理
│   ├── migrations/            # 数据库迁移
│   ├── Dockerfile             # Docker配置
│   └── go.mod                 # Go模块定义
│
├── frontend/                  # 商城前端 (Vue.js 3)
│   ├── src/
│   │   ├── api/              # API调用
│   │   ├── stores/           # Pinia 状态管理
│   │   ├── views/            # 页面组件
│   │   ├── components/       # 可复用组件
│   │   ├── router/           # 路由配置
│   │   └── main.ts           # 应用入口
│   ├── package.json
│   └── Dockerfile
│
├── admin/                     # 后台管理前端 (Vue.js 3)
│   └── (结构同frontend/)
│
├── docker-compose.yml         # Docker编排
├── nginx.conf                 # Nginx配置
├── README.md                  # 项目说明
├── product.md                 # 技术文档
└── .gitignore
```

## 主要技术点

### 多租户识别
```
请求流程：
1. 请求到达 Nginx
2. Nginx 根据域名路由到后端
3. 后端中间件从 Host 头识别租户
4. 将 tenant_id 注入到请求上下文
5. 所有操作自动按租户隔离
```

### 商品数据管理
```
商品查询流程：
1. 获取全局商品基础数据（products表）
2. 查询租户覆盖数据（tenant_product_overrides表）
3. 合并数据（优先使用覆盖数据）
4. 返回给前端
```

### API认证
```
认证流程：
1. 用户登录，获取JWT token
2. 前端请求时在 Authorization header 中携带 token
3. 后端验证 token 并提取用户信息
4. 未认证请求返回 401
```

## 常见问题

### Q: 如何切换不同的租户?
A: 修改 hosts 文件，或在开发时改变请求的 Host 头。系统会自动根据 Host 识别租户。

### Q: 商品价格在不同网站显示不同?
A: 在后台管理中设置租户覆盖数据，价格、库存等都可以为不同租户定制。

### Q: 如何运行测试?
A: 目前暂未实现。后续会添加单元测试和集成测试。

## 下一步计划

### 第一阶段（本周）
- [x] 项目架构
- [ ] 实现用户认证模块
- [ ] 实现商品查询接口
- [ ] 实现基础前端页面

### 第二阶段（第二周）
- [ ] 实现订单管理
- [ ] 实现后台管理界面
- [ ] 实现购物车功能

### 第三阶段（第三周）
- [ ] 支付集成
- [ ] 测试和优化
- [ ] 部署上线

## 联系和支持

有问题请：
1. 查看 product.md 技术文档
2. 检查代码注释
3. 运行测试用例
4. 提交 Issue

---

祝开发愉快！🚀
