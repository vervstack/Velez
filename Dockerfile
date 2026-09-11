FROM --platform=$BUILDPLATFORM oven/bun:1 AS webclient

WORKDIR /web

COPY pkg/web/Velez-UI/package.json pkg/web/Velez-UI/bun.lock ./
RUN bun install --frozen-lockfile

COPY pkg/web/Velez-UI/ ./
RUN bun run build && \
    mv dist /dist

FROM --platform=$BUILDPLATFORM golang:1.24.2 AS builder

WORKDIR /app

COPY --from=webclient /dist internal/transport/ui/dist

RUN --mount=target=. \
        --mount=type=cache,target=/root/.cache/go-build \
        --mount=type=cache,target=/go/pkg \
        GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -o /deploy/server/service ./cmd/service/main.go && \
    cp -r config /deploy/server/config

FROM alpine

# Service name for correct config parsing when setting environment variables
ENV VERV_NAME=velez
# Link to existing matreshka instance (by default points to instance inside docker local network)
ENV MATRESHKA_URL=matreshka
# Link to existing makosh instance (by default points to instance inside docker local network)
ENV MAKOSH_URL=makosh
ENV MATRESHKA_CONFIG_ENABLED=true

WORKDIR /app

COPY --from=builder /deploy/server/ .

EXPOSE 53890

RUN echo yes > is_in_container.txt

ENTRYPOINT ["./service"]