# Multi-stage Dockerfile for NotiCat-Server
# Builder: compile Go server and C++ mail binary, install Python deps
FROM golang:1.25.5-bullseye AS builder

ENV CGO_ENABLED=1
WORKDIR /src

# Install build deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    make g++ libcurl4-openssl-dev python3 python3-pip ca-certificates git && \
    rm -rf /var/lib/apt/lists/*

# Copy go modules and download first (cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy project
COPY . .

# Install python dependencies for scripts
RUN if [ -f scripts/requirements.txt ]; then pip3 install -r scripts/requirements.txt; fi

# Build C++ mail binary
RUN if [ -f mail/Makefile ]; then make -C mail all; fi

# Build Go binary
RUN go build -o /out/noticat ./main.go

### Runtime image
FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates python3 python3-pip libcurl4 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy built Go binary and mail binary and scripts
COPY --from=builder /out/noticat ./noticat
COPY --from=builder /src/mail/bin/send ./mail/bin/send
COPY --from=builder /src/scripts ./scripts

# Expose default port (adjust via env/config)
EXPOSE 8080

ENV NOTICAT_SMTP_SERVER="163"

ENTRYPOINT ["/app/noticat"]
