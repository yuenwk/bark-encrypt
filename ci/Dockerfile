# --- Stage 1: Build ---
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY ../go.mod go.sum ./
RUN go mod download
COPY ci .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM alpine:latest
COPY --from=builder /server /server
EXPOSE 9090
CMD ["/server"]