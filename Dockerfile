FROM golang:1.25.5-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
  && addgroup -S app \
  && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/server /app/server

RUN mkdir -p /app/data \
  && chown -R app:app /app

USER app

ENV APP_ENV=prod \
    LOG_LEVEL=INFO \
    HTTP_PORT=8080 \
    DB_PATH=/app/data/app.db

EXPOSE 8080
ENTRYPOINT ["/app/server"]


