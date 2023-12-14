FROM golang as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /deploy/server/velez ./cmd/velez/main.go

FROM alpine

WORKDIR /app
COPY --from=builder ./deploy/server/ .
COPY --from=builder /app/config/ ./config/

EXPOSE 50051
EXPOSE 50052
ENTRYPOINT ["./velez"]