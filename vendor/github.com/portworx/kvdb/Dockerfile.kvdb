FROM ubuntu
MAINTAINER support@portworx.com

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools \
    gcc \
    g++ \
    ca-certificates && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN apt-get update && \
  apt-get -y install unzip curl make git default-jre


RUN curl -s -L https://dl.google.com/go/go1.10.7.linux-amd64.tar.gz | tar -C /usr/local/ -xz  && \
  curl -s -L https://github.com/coreos/etcd/releases/download/v3.2.15/etcd-v3.2.15-linux-amd64.tar.gz -o /tmp/etcd-v3.2.15-linux-amd64.tar.gz  && \
  mkdir -p /tmp/test-etcd && tar xzvf /tmp/etcd-v3.2.15-linux-amd64.tar.gz -C /tmp/test-etcd --strip-components=1 && cp /tmp/test-etcd/etcd /usr/local/bin  && \
  curl -s -L https://releases.hashicorp.com/consul/1.0.0/consul_1.0.0_linux_amd64.zip -o /tmp/consul.zip && \
  mkdir -p /tmp/test-consul && unzip /tmp/consul.zip -d /tmp/test-consul && cp /tmp/test-consul/consul /usr/local/bin/ && \
  curl -s -L https://archive.apache.org/dist/zookeeper/zookeeper-3.4.13/zookeeper-3.4.13.tar.gz -o /tmp/zookeeper-3.4.13.tar.gz && \
  mkdir -p /tmp/test-zookeeper && tar -xvf /tmp/zookeeper-3.4.13.tar.gz -C /tmp/test-zookeeper --strip-components=1 && mkdir -p /data/zookeeper

ENV PATH /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV GOPATH /go
ENV GOROOT /usr/local/go

RUN mkdir -p /go/src/github.com/portworx/kvdb
ADD . /go/src/github.com/portworx/kvdb

RUN echo $'tickTime=2000 \n\
dataDir=/data/zookeeper \n\
clientPort=2181 \n\
initLimit=5 \n\
syncLimit=2 \n\
server.1=127.0.0.1:2888:3888' > /tmp/test-zookeeper/conf/zoo.cfg

WORKDIR /go/src/github.com/portworx/kvdb
