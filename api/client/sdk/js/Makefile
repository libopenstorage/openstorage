HAS_NPM := $(shell command -v npm 2> /dev/null)
JSPROTOC := ./node_modules/.bin/grpc_tools_node_protoc
APIDIR := "../../../../api"
all: grpc

grpc:
ifndef HAS_NPM
	$(error "Please install npm")
endif
	if [ ! -x $(JSPROTOC) ] ; then npm install grpc-tools ; fi
	$(JSPROTOC) \
		--js_out=import_style=commonjs,binary:. \
		--grpc_out=. \
		--plugin=proto-gen-grpc=./node_modules/grpc-tools/bin/grpc_tools_node_protoc_plugin \
		-I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		api.proto
	$(JSPROTOC) \
		--js_out=import_style=commonjs,binary:. \
		--grpc_out=. \
		--plugin=proto-gen-grpc=./node_modules/grpc-tools/bin/grpc_tools_node_protoc_plugin \
		-I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api/annotations.proto
	$(JSPROTOC) \
		--js_out=import_style=commonjs,binary:. \
		--grpc_out=. \
		--plugin=proto-gen-grpc=./node_modules/grpc-tools/bin/grpc_tools_node_protoc_plugin \
		-I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api/http.proto

.PHONY: all grpc
