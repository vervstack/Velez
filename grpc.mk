gen: deps gen-server

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

gen-server: .pre-gen-server .gen-server
.pre-gen-server:
	mkdir -p pkg/

.gen-server:
	protoc \
	--go_out=./pkg/ \
	--go-grpc_out=./pkg/ \
	--grpc-gateway_out ./pkg/ \
	-I=./api \
	--proto_path=. \
	./api/grpc/*.proto
