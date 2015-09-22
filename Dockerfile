FROM golang:latest
RUN apt-get update

# Install supporting packages required for btrfs-progs
RUN apt-get install -y asciidoc xmlto --no-install-recommends
RUN apt-get install -y uuid-dev libattr1-dev zlib1g-dev libacl1-dev e2fslibs-dev libblkid-dev liblzo2-dev
RUN apt-get install -y autoconf pkg-config

# Clone btrfs-progs and build it
RUN git clone git://git.kernel.org/pub/scm/linux/kernel/git/kdave/btrfs-progs.git ~/btrfs-progs
RUN /root/btrfs-progs/autogen.sh
RUN cd /root/btrfs-progs && \
  ./configure && make && \
  make install

# Fetch openstorage and dependencies.
RUN go get -d github.com/libopenstorage/openstorage
RUN go get github.com/tools/godep

# Upload current openstorage source
COPY . /go/src/github.com/libopenstorage/openstorage

# Build openstorage and install it.
RUN cd $GOPATH/src/github.com/libopenstorage/openstorage && godep restore
RUN cd $GOPATH/src/github.com/libopenstorage/openstorage && make openstorage \
    && cp osd /bin/osd
CMD ["/bin/osd"]
