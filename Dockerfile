FROM golang:1.5.1
MAINTAINER gou@portworx.com

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools \
    ca-certificates
ENV GO15VENDOREXPERIMENT 1
RUN mkdir -p /go/src/github.com/libopenstorage/openstorage
ADD . /go/src/github.com/libopenstorage/openstorage/
WORKDIR /go/src/github.com/libopenstorage/openstorage
RUN make install
CMD ["/go/bin/osd"]
