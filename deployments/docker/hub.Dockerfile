FROM golang:alpine AS builder

WORKDIR /app

# Karena kita menggunakan SQLite (CGO), kita butuh gcc dan musl-dev
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
COPY vendor/ vendor/

COPY . .

# Build aplikasi hub dengan CGO_ENABLED=1
RUN CGO_ENABLED=1 GOOS=linux go build -mod=vendor -o hub ./cmd/hub

# Stage 2: Minimal image
FROM alpine:latest

WORKDIR /app

# Tambahkan library dasar jika diperlukan oleh alpine
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/hub .

EXPOSE 8084

CMD ["./hub"]
