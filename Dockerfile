FROM golang as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /deploy/server/velez ./cmd/velez/main.go

FROM alpine

WORKDIR /app

COPY --from=builder ./deploy/server/ .
COPY --from=builder /app/config/ ./config/

ENV VELEZ_CUSTOM_PASS_TO_KEY="/tmp/velez/private.key"

EXPOSE 53890

VOLUME /var/run/docker.sock

ENTRYPOINT ["./velez"]