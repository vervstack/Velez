FROM alpine:3.24.2

ENV VERV_NAME=velez
ENV MATRESHKA_URL=matreshka
ENV MAKOSH_URL=makosh
ENV MATRESHKA_CONFIG_ENABLED=true

WORKDIR /app

COPY service /app/service
COPY config /app/config
COPY migrations /app/migrations

EXPOSE 53890

ENTRYPOINT ["./service"]
