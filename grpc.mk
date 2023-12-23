gen: deps gen-server

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

local-link:
	ln -s $(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/google api/google

gen-server: .pre-gen-server .gen-server
.pre-gen-server:
	mkdir -p pkg/

.gen-server:
	protoc \
    	-I=./api \
    	--grpc-gateway_out=logtostderr=true:./pkg/ \
    	--openapiv2_out ./api \
		--descriptor_set_out=./pkg/velez_api/api_descriptor.pb \
    	--go_out=./pkg/ \
    	--go-grpc_out=./pkg/ \
    	./api/grpc/*.proto
