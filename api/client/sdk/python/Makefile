HAS_VIRTUALENV := $(shell command -v virtualenv 2> /dev/null)
APIDIR := "../../../../api"
VEDIR := "build"
all: grpc

grpc:
ifndef HAS_VIRTUALENV
	$(error "Please install virtualenv for python: pip install virtualenv")
endif
	virtualenv $(VEDIR)
	bash -c "source $(VEDIR)/bin/activate && \
	pip install grpcio grpcio-tools && \
	python -m grpc_tools.protoc \
		-I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--python_out=. \
		--grpc_python_out=. \
		$(APIDIR)/api.proto"
	bash -c "source $(VEDIR)/bin/activate && \
	python -m grpc_tools.protoc \
		-I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--python_out=. \
		--grpc_python_out=. \
		$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api/annotations.proto"
	bash -c "source $(VEDIR)/bin/activate && \
	python -m grpc_tools.protoc \
		-I$(APIDIR) \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--python_out=. \
		--grpc_python_out=. \
		$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api/http.proto"

.PHONY: all grpc

