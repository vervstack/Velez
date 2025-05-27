FROM --platform=$BUILDPLATFORM node:23-alpine3.20 AS webclient

WORKDIR /web

RUN --mount=type=bind,target=/web,rw \
# Step 1: Build the API lib
    cd /web/pkg/web/@vervstack/velez && \
    yarn && \
    yarn build && \
# Step 2: Install and build Vue app (now that web is built)
    cd /web/pkg/web/Velez-UI && \
    yarn && \
    yarn build && \
    mv dist /dist

FROM --platform=$BUILDPLATFORM golang:1.24.2 AS builder

WORKDIR /app

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