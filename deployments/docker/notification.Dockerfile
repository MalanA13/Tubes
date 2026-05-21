# Gunakan 1.25 sesuai permintaan go.mod kamu
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Karena context adalah root, Docker sekarang bisa lihat go.mod & go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh isi folder tubes-unit_test ke dalam container
COPY . .

# Build aplikasi. Karena main.go ada di dalam auth-service, kita arahkan path-nya
RUN go build -o main ./notification-service/main.go

# Stage 2: Final Image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]