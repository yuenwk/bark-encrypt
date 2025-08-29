# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY compose.yaml .
RUN CGO_ENABLED=0 GOOS=linux go build -o /myapp -ldflags="-s -w" ./cmd/myapp

FROM alpine:latest

COPY --from=builder /myapp /app/myapp

EXPOSE 9090

CMD ["./myapp"]