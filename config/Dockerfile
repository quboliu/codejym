FROM node:20-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
COPY frontend/tsconfig*.json ./
COPY frontend/vite.config.ts ./
COPY frontend/index.html ./
COPY frontend/public ./public
COPY frontend/src ./src
RUN npm install && npm run build

FROM golang:1.24-alpine AS backend-build
WORKDIR /app/backend
# 安装 git（go mod download 需要）
RUN apk add --no-cache git
# 先拷贝 go.mod 和 go.sum
COPY backend/go.mod backend/go.sum ./
# 下载依赖
RUN go mod download
# 拷贝所有源代码
COPY backend/ .
# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app -u 1001
COPY --from=backend-build /app/server /app/server
COPY --from=frontend-build /app/frontend/dist /app/frontend
RUN mkdir -p /data && chown -R app:app /data /app
USER app
ENV DATA_DIR=/data
ENV FRONTEND_DIR=/app/frontend
EXPOSE 8080
VOLUME ["/data"]
CMD ["/app/server", "-addr", ":8080"]
