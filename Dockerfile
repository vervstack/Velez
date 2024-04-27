FROM --platform=$BUILDPLATFORM golang as builder

WORKDIR /app
COPY . .

RUN --mount=target=. \
        --mount=type=cache,target=/root/.cache/go-build \
        --mount=type=cache,target=/go/pkg \
        GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -o /deploy/server/service ./cmd/service/main.go && \
        cp -r config /deploy/server/config
FROM alpine

WORKDIR /app

COPY --from=builder /deploy/server/service service
COPY --from=builder /app/config/config.yaml ./config/config.yaml

EXPOSE 80

ENTRYPOINT ["./service"]