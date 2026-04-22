# syntax=docker/dockerfile:1.7

FROM golang:1.22-bookworm AS builder

WORKDIR /src/backend
ARG TARGETARCH=amd64
ARG GOPROXY=https://goproxy.cn,direct
COPY backend/go.mod backend/go.sum ./
ENV GOPROXY=${GOPROXY}
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod tidy && go mod download
COPY backend/ ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags='-s -w' -o /out/goadmin ./cmd/server

FROM alpine:3.20 AS runtime

# Use a mirror that is typically reachable from macOS Docker Desktop environments.
RUN sed -i 's#dl-cdn.alpinelinux.org#mirrors.aliyun.com#g' /etc/apk/repositories \
    && apk add --no-cache bash ca-certificates tzdata \
    && addgroup -S goadmin \
    && adduser -S -G goadmin goadmin

WORKDIR /app/backend
COPY --from=builder /out/goadmin /usr/local/bin/goadmin
COPY --from=builder /src/backend/config ./config
COPY --from=builder /src/backend/core/auth/casbin/model ./core/auth/casbin/model
COPY --from=builder /src/backend/core/auth/casbin/adapter/policy.csv ./core/auth/casbin/adapter/policy.csv
COPY --from=builder /src/backend/scripts ./scripts
RUN chmod +x ./scripts/*.sh \
    && chown -R goadmin:goadmin /app/backend

ENV GOADMIN_CONFIG_DIR=/app/backend/config \
    GOADMIN_ENV=dev \
    APP_ENV=dev \
    GOADMIN_SERVER_HTTP_HOST=0.0.0.0 \
    GOADMIN_SERVER_HTTP_PORT=8080

USER goadmin

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=10s CMD ["bash", "/app/backend/scripts/healthcheck.sh"]

ENTRYPOINT ["/usr/local/bin/goadmin"]
