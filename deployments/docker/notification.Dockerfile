FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
COPY vendor/ vendor/

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -mod=vendor -o main ./cmd/notification

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/main .

EXPOSE 8086

CMD ["./main"]