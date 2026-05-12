FROM golang:alpine AS builder

WORKDIR /app

# Karena kita menggunakan SQLite (CGO), kita butuh gcc dan musl-dev
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build aplikasi courier dengan CGO_ENABLED=1
RUN CGO_ENABLED=1 GOOS=linux go build -o courier ./cmd/courier

# Stage 2: Minimal image
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/courier .

EXPOSE 8085

CMD ["./courier"]
