# 构建阶段
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates

# 容器内默认 GOPROXY 为 proxy.golang.org，国内网络通常不可达，
# 显式设置为本机一致的国内镜像，避免 go mod download 失败。
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOSUMDB=off

# 引入本地 gocommon（go.mod 中 replace 指向 ../gocommon），使共享的 Consul 注册等逻辑
# 的最新改动进入镜像。构建上下文为仓库根目录，故此处从根拷贝。
COPY article-service/ ./article-service/
COPY gocommon/ ./gocommon/

# 版本号（docker build --build-arg GIT_VERSION=xxx 注入，默认 dev）
ARG GIT_VERSION=dev

WORKDIR /app/article-service

# 拉取依赖（本地 replace 的 gocommon 不会被下载，其余走 GOPROXY）
RUN go mod download

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X main.Version=${GIT_VERSION}" \
    -o /app/article-service-bin ./cmd/server

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# 从构建阶段复制二进制文件与配置
COPY --from=builder /app/article-service-bin ./article-service
COPY --from=builder /app/article-service/config.yaml .

# 创建非 root 用户
RUN adduser -D -g '' appuser
# 封面上传目录（运行用户可写）
RUN mkdir -p /app/uploads && chown -R appuser /app/uploads
USER appuser

# 暴露端口（8082=HTTP，9002=gRPC，9092=metrics）
EXPOSE 8082 9002 9092

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# 启动命令
CMD ["./article-service"]
