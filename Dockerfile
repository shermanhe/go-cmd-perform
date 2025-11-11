# 第一阶段：构建阶段
FROM golang:1.23 AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

RUN go mod tidy

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -o perform_go -ldflags "-s -w" src/main.go

# 第二阶段：运行阶段
FROM debian:12-slim

# 设置工作目录
WORKDIR /app

# 复制编译好的二进制文件
COPY --from=builder /app/perform_go .