HAS_GRPCTOOLS := $(shell command -v grpc_tools_ruby_protoc)
APIDIR := "../../../../api"
all: grpc

grpc:
	grpc_tools_ruby_protoc -I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--ruby_out=. \
		--grpc_out=. \
		$(APIDIR)/api.proto

.PHONY: all grpc
