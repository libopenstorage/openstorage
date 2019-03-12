FROM golang:1.10.4
MAINTAINER gou@portworx.com

EXPOSE 9005
RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools \
    ca-certificates && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
# Get docker binary
RUN \
  curl -sSL https://get.docker.com/builds/Linux/x86_64/docker-1.10.3 -o /bin/docker && \
  chmod +x /bin/docker
# Get all required build-tools, then clean up GOPATH
RUN go get -u	\
	github.com/golang/lint/golint					\
	github.com/kisielk/errcheck					\
	github.com/golang/protobuf/protoc-gen-go			\
	github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway	\
	github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger	\
	github.com/kardianos/govendor					\
	github.com/gobuffalo/packr/...					\
	golang.org/x/tools/cmd/cover					\
	github.com/pierrre/gotestcover		&& \
		\
	rm -fr /go/src/* /go/pkg/*		&& \
	mkdir -p /go/src/github.com/libopenstorage/openstorage
