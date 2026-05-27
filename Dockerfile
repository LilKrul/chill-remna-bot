# --- сборка ---
# Версия зафиксирована под toolchain go1.23.4 в go.mod (без докачивания тулчейна).
FROM golang:1.23.4-alpine AS build
WORKDIR /src

# Кэшируем зависимости отдельным слоем.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
# CGO выключен: modernc.org/sqlite — чистый Go, бинарь статический.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

# --- финальный образ ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -