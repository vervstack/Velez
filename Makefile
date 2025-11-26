include rscli.mk

.PHONY: codegen

# generates folders and installs dependencies
warmup:
	make .prepare-grpc-folders
	make .deps-grpc
	PROTOPACKPATH=proto_deps protopack mod download
# generates code on warm project
codegen:
	PROTOPACKPATH=proto_deps protopack generate
	cd pkg/web/@vervstack/velez && npm run build

lint:
	golangci-lint run ./...

test:
