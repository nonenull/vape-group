FROM golang:1.19-alpine AS builder

WORKDIR /app

# 复制go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o backend ./cmd/main.go

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建镜像复制二进制文件
COPY --from=builder /app/backend .

# 暴露端口
EXPOSE 8080

CMD ["./backend"]
