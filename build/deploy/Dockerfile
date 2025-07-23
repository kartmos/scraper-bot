# Stage 1 — сборка
FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache curl
RUN apk add --no-cache curl make

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o bot ./cmd/app
RUN make yt-dlp

# Stage 2 — минимальный образ
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bot .
COPY --from=builder /app/bin/yt-dlp /app/bin/yt-dlp
COPY --from=builder /app/config/config.yaml /app/config/
COPY --from=builder /app/asserts /app/asserts

RUN mkdir -p /app/downloads
RUN mkdir -p /app/logs
RUN touch bot.log

CMD ["./bot"]