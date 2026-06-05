FROM golang:1.23.4-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata docker-cli docker-cli-compose
WORKDIR /app
COPY --from=build /out/bot /app/bot
VOLUME ["/data"]
ENV DATA_DIR=/data
ENTRYPOINT ["/app/bot"]
