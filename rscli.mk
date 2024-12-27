gen: gen-server-grpc
build-local-container:
	docker buildx build \
			--load \
			--platform linux/arm64 \
			-t velez:local .

### Grpc server generation
gen-server-grpc: .prepare-grpc-folders .download-grpc-deps .gen-server-grpc

.prepare-grpc-folders:
	mkdir -p pkg/web
	mkdir -p pkg/docs/api

.update-grpc-deps:
	EASYPPATH=proto_deps easyp mod update

.download-grpc-deps:
	EASYPPATH=proto_deps easyp mod download

.gen-server-grpc:
	EASYPPATH=proto_deps easyp generate