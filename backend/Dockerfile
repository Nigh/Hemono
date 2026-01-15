# ===== 构建阶段 =====
FROM golang:1.25-alpine AS builder

# 安装编译依赖
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# 优化缓存：先复制 mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制：明确输出到当前目录的根位置
RUN CGO_ENABLED=1 GOOS=linux go build \
    -o /app/hemono \
    -ldflags="-s -w" \
    .

# ===== 运行时镜像 =====
FROM alpine:3.20

# 安装运行时必要的库
RUN apk add --no-cache libc6-compat sqlite-libs ca-certificates tzdata

WORKDIR /app

# 从构建阶段显式复制
COPY --from=builder /app/hemono /app/hemono

# 赋予执行权限（防止权限丢失）
RUN chmod +x /app/hemono

# 数据目录
ENV PB_DATA_DIR=/pb_data
RUN mkdir -p ${PB_DATA_DIR}

EXPOSE 8090

ENTRYPOINT ["/app/hemono"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/pb_data"]
