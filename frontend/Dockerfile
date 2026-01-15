# ===== 阶段 1: 构建阶段 =====
FROM oven/bun:1-alpine AS builder

WORKDIR /app

# 复制配置文件以利用 Docker 层缓存
COPY package.json bun.lock* ./

# 使用 bun 安装依赖，--frozen-lockfile 确保环境一致性 
RUN bun install --frozen-lockfile

# 复制其余源码并构建项目
COPY . .
RUN bun run build

# ===== 阶段 2: 运行时镜像 =====
FROM oven/bun:1-alpine AS runner

WORKDIR /app

# SvelteKit 默认构建产物在 build 目录
COPY --from=builder /app/build ./build

# 暴露端口
EXPOSE 5173

ENV PORT=5173
CMD ["bun", "./build/index.js"]
