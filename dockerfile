# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main .

# Stage 2: Run
FROM alpine:latest

# Install Redis and Docker client
RUN apk add --no-cache \
    redis \
    docker-cli

WORKDIR /app
COPY --from=builder /app/main .
COPY .env .

# Create directory for Redis data and config
RUN mkdir -p /data /etc/redis
COPY redis.conf /etc/redis/
COPY start.sh .
RUN chmod +x start.sh && \
    chown -R redis:redis /data /etc/redis

EXPOSE 4000 6379 2375

ENTRYPOINT ["./start.sh"]
