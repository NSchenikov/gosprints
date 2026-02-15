FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o task-service cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/task-service .
EXPOSE 50051 8081
CMD ["./task-service"]