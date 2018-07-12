HAS_GRPCTOOLS := $(shell command -v grpc_tools_ruby_protoc 2> /dev/null)
APIDIR := "../../../../api"
all: grpc

grpc:
ifndef HAS_GRPCTOOLS
	$(error "Please install grpc rools for ruby: gem install grpc && gem install grpc-tools")
endif
	grpc_tools_ruby_protoc -I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--ruby_out=. \
		--grpc_out=. \
		$(APIDIR)/api.proto
	grpc_tools_ruby_protoc -I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--ruby_out=. \
		--grpc_out=. \
		$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api/annotations.proto
	grpc_tools_ruby_protoc -I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--ruby_out=. \
		--grpc_out=. \
		$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api/http.proto

.PHONY: all grpc
