# Frontend

商城前端已经迁移为 `Nuxt 3`，用于服务端渲染以下关键页面：

- `/`
- `/products`
- `/products/category/:category`
- `/products/:id`
- `/about`

`/cart` 和 `/checkout` 保留客户端互动，但运行在同一套 Nuxt 应用内。

## 开发

```sh
npm install
npm run dev
```

默认开发端口为 `3000`。

## 构建

```sh
npm run build
```

构建结果由 Nuxt 产出到 `.output/`，生产容器直接运行 Nitro server。
