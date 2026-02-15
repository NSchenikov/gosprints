FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notification-service cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/notification-service .
EXPOSE 8082
CMD ["./notification-service"]