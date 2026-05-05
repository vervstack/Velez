include rscli.mk

.PHONY: codegen

# generates folders and installs dependencies
warmup:
	make .prepare-grpc-folders
	make .deps-grpc
	PROTOPACKPATH=proto_deps protopack mod download
# generates code on warm project
codegen:
	moti g
	cd pkg/web/@vervstack/velez && npm run build

lint:
	golangci-lint run ./...
