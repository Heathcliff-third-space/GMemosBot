# 构建阶段 - 使用官方Go镜像编译程序
FROM golang:1.24.1-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /telegram-bot

# 最终阶段 - 使用精简的Alpine镜像
FROM alpine:3.18

# 安装CA证书(用于HTTPS请求)
RUN apk --no-cache add ca-certificates

# 从构建阶段复制编译好的二进制文件
RUN mkdir /app
COPY --from=builder /telegram-bot /app/telegram-bot

# 创建非root用户
RUN adduser -D -g '' botuser && chown -R /app botuser
USER botuser

# 设置入口点
ENTRYPOINT ["/app/telegram-bot"]