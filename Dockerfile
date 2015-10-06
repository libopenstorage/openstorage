FROM golang:1.5.1
MAINTAINER gou@portworx.com

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    asciidoc \
    autoconf \
    automake \
    e2fslibs-dev \
    libacl1-dev \
    libattr1-dev \
    libblkid-dev \
    liblzo2-dev \
    pkg-config \
    uuid-dev \
    xmlto \
    zlib1g-dev

RUN \
  git clone git://git.kernel.org/pub/scm/linux/kernel/git/kdave/btrfs-progs.git /tmp/btrfs-progs && \
  /tmp/btrfs-progs/autogen.sh && \
  cd /tmp/btrfs-progs && \
  ./configure && \
  make && \
  make install && \
  rm -rf /tmp/btrfs-progs

RUN go get github.com/tools/godep
RUN mkdir -p /go/src/github.com/libopenstorage/openstorage/Godeps
WORKDIR /go/src/github.com/libopenstorage/openstorage
ADD Godeps/ /go/src/github.com/libopenstorage/openstorage/Godeps/
RUN godep restore
ADD . /go/src/github.com/libopenstorage/openstorage/
RUN go build -tags daemon -o /bin/osd
CMD ["/bin/osd"]
