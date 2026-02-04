# --- 第一阶段：编译 (使用真实存在的 Go 版本) ---
FROM golang:1.24-bullseye AS builder

# 开启 CGO，因为你要编译/链接 C++ 或使用 cgo
ENV CGO_ENABLED=1
WORKDIR /src

# 安装 C++ 编译环境和依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    make g++ libcurl4-openssl-dev python3 python3-pip ca-certificates git && \
    rm -rf /var/lib/apt/lists/*

# 缓存 Go 依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝全量代码
COPY . .

# 编译 C++ 邮件模块
# 确保你的 Makefile 执行后产物在 mail/bin/send
RUN if [ -f mail/Makefile ]; then make -C mail all; fi

# 编译 Go 主程序
RUN go mod tidy && go build -o /out/noticat ./main.go

# --- 第二阶段：运行环境 ---
FROM debian:bullseye-slim

# 安装运行时的必要库（比如 Python 和 libcurl）
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates python3 python3-pip libcurl4 libstdc++6 python3-dev \
    gcc libc6-dev && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 从 builder 拷贝二进制产物
COPY --from=builder /out/noticat ./noticat

# 检查路径：确保编译出来的 send 确实在那个位置
# 如果你的 Makefile 产物位置不同，请修改这里
COPY --from=builder /src/mail/bin/send ./mail/bin/send
COPY --from=builder /src/scripts ./scripts

# 安装 Python 运行时依赖（如果在运行环境也需要跑脚本）
RUN if [ -f ./scripts/requirements.txt ]; then \
    pip3 install --no-cache-dir --upgrade pip && \
    pip3 install --no-cache-dir -r ./scripts/requirements.txt; \
    fi

# 暴露端口
EXPOSE 8080

# 启动！
ENTRYPOINT ["/app/noticat"]