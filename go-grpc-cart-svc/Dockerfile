# =========================================================
# Stage 1: Builder — full Go toolchain, compiles the binary.
# =========================================================
FROM golang:1.23-alpine AS builder

# git is needed for some `go mod download` operations.
RUN apk add --no-cache git

WORKDIR /src

# Copy go.mod / go.sum first so this layer is cached as long
# as dependencies don't change. Rebuilds become MUCH faster.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source.
COPY . .

# Produce a fully static binary:
#   CGO_ENABLED=0   -> no libc, runs in any minimal image
#   -trimpath       -> reproducible: strip local file paths
#   -ldflags "-s -w"-> strip debug info, shrinks binary ~30%
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/app ./cmd

# =========================================================
# Stage 2: Runtime — tiny final image, only what's needed.
# =========================================================
FROM alpine:3.20

# ca-certificates -> required if the service makes outbound HTTPS calls.
# tzdata          -> avoids "unknown timezone" issues if you set TZ.
# wget            -> used to download grpc_health_probe below.
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S app -G app

# Install grpc_health_probe for the HEALTHCHECK directive.
# TARGETARCH is auto-set by Docker BuildKit (amd64/arm64/etc.) so the
# right binary is fetched on Intel Macs, Apple Silicon, and CI runners.
ARG TARGETARCH
ARG GRPC_HEALTH_PROBE_VERSION=v0.4.24
RUN wget -qO /bin/grpc_health_probe \
    "https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-${TARGETARCH}" \
 && chmod +x /bin/grpc_health_probe

WORKDIR /app

# The compiled binary from stage 1.
COPY --from=builder /out/app /app/app

# Viper expects this file at this path. Compose env vars will
# OVERRIDE whatever values are inside it (because AutomaticEnv()).
COPY --from=builder /src/pkg/config/envs/dev.env /app/pkg/config/envs/dev.env

# Run as non-root.
USER app

# CHANGE THIS PER SERVICE ↓↓↓
EXPOSE 50051

# Real gRPC healthcheck — speaks the standard grpc.health.v1 protocol
# against the Health server registered in cmd/main.go.
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD /bin/grpc_health_probe -addr=localhost:50051 || exit 1

ENTRYPOINT ["/app/app"]