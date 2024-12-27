gen: gen-server-grpc
build-local-container:
	docker buildx build \
			--load \
			--platform linux/arm64 \
			-t velez:local .

### Grpc server generation
gen-server-grpc: .prepare-grpc-folders .deps-grpc .gen-server-grpc

.prepare-grpc-folders:
	mkdir -p pkg/web
	mkdir -p pkg/docs/api

.deps-grpc:
	EASYPPATH=proto_deps easyp mod update

.gen-server-grpc:
	EASYPPATH=proto_deps easyp generate