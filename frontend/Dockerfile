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
# 对于静态站点，建议使用更轻量的镜像或继续使用 bun 运行
FROM oven/bun:1-alpine AS runner

WORKDIR /app

# SvelteKit 默认构建产物在 .svelte-kit/output 或 build 目录
# 这里根据您的 package.json 逻辑复制构建出的静态资源
# COPY --from=builder /app/build ./build
COPY --from=builder /app/.svelte-kit/output ./output
COPY --from=builder /app/package.json ./

# 暴露端口（与您的原配置保持一致）
EXPOSE 5173

# 使用 bun 直接运行服务，或者继续使用 serve 插件
# 推荐做法：安装一个轻量的静态服务器
RUN bun add -g serve

# 启动服务
CMD ["serve", "-s", "output", "-l", "5173"]
