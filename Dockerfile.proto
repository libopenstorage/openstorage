#
# Do not use directly, use `make docker-proto` instead
#
FROM fedora
MAINTAINER luis@portworx.com

ENV GOPATH=/go
RUN dnf -y install \
	golang-bin \
	python \
	python-pip \
	gem \
	npm \
	make \
	git \
	protobuf-compiler \
	protobuf-devel && dnf -y clean all && rm -rf /var/cache/yum
RUN pip install virtualenv
RUN gem install grpc && gem install grpc-tools
RUN go get -u github.com/golang/protobuf/protoc-gen-go && \
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger && \
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

# Lock protoc-gen-go version to v1.1.0.
#
# NOTE:
# The latest uses new Invoke api which needs updates on:
# $ go get -u google.golang.org/grpc/...
# $ go get -u golang.org/x/sys/unix/...
# $ govendor remove google.golang.org/grpc/...
# $ govendor add +external google.golang.org/grpc/...
# $ govendor update +external golang.org/x/sys/unix/...
# Which may apply to project depending on OpenStorage.
#
WORKDIR /go/src/github.com/golang/protobuf
RUN git checkout v1.1.0 && go install github.com/golang/protobuf/protoc-gen-go


# Lock protoc-gen-swagger to v1.4.1
# The swagger output in the latest version seems to be incorrect
# See: https://github.com/grpc-ecosystem/grpc-gateway/issues/688
#
WORKDIR /go/src/github.com/grpc-ecosystem/grpc-gateway
RUN git checkout v1.4.1 && \
    go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger && \
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

# Install swagger 2.0 to OpenApi 3.0 converter
RUN npm install -g swagger2openapi
# Finally, set working directory to the openstorage project
RUN mkdir -p /go/src/github.com/libopenstorage/openstorage
WORKDIR /go/src/github.com/libopenstorage/openstorage
