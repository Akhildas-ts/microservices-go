FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/api-gateway ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/api-gateway /app/api-gateway

COPY --from=builder /src/pkg/config/envs/dev.env /app/pkg/config/envs/dev.env

USER app

EXPOSE 3000

# Probe the dedicated /health endpoint registered in cmd/main.go.
# The previous probe hit "/" which is unrouted and returns 404, making
# wget exit non-zero and the container flip to unhealthy.
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1

ENTRYPOINT ["/app/api-gateway"]