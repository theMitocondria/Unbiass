# Stage 1: Build
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Stage 2: Run
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/main .

# Set environment variables via Docker, not .env file directly
EXPOSE 4000 
ENTRYPOINT ["./main"]
