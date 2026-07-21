FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /src/assets \
    && CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/chat-api \
    ./cmd/main.go

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/assets \
    && chown -R app:app /app

WORKDIR /app

COPY --from=builder --chown=app:app /out/chat-api ./chat-api
COPY --from=builder --chown=app:app /src/assets ./assets

USER app

EXPOSE 8888

ENTRYPOINT ["/app/chat-api"]
