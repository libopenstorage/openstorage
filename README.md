# Open Storage

[![Travis branch](https://img.shields.io/travis/libopenstorage/openstorage/master.svg)](https://travis-ci.org/libopenstorage/openstorage)
[![Docker Pulls](https://img.shields.io/docker/pulls/openstorage/osd.svg)](https://hub.docker.com/r/openstorage/osd)
[![Go Report Card](https://goreportcard.com/badge/github.com/libopenstorage/openstorage)](https://goreportcard.com/report/github.com/libopenstorage/openstorage)

OpenStorage is a clustered implementation of the [Open Storage](https://github.com/libopenstorage/specs) specification and relies on the Docker runtime.  It allows you to run stateful services in Linux Containers in a multi-host environment.  It plugs into CSI and Docker volumes to provide storage to a container and plugs into Kubernetes and Mesosphere to operate in a clustered environment. 

# What you get from using Open Storage

When you install openstorage on a Linux host, you will automatically get a stateful storage layer that integrates with CSI or the Docker runtime and operates in a multi host environment.  It starts an Open Storage Daemon - `OSD` that currently supports CSI and DVDI and will support any Linux container runtime that conforms to the [OCI](https://www.opencontainers.org/).  

## Scheduler integration

OSD will work with any distributed scheduler that is compatible with the [CSI](https://github.com/container-storage-interface/spec) or [Docker remote API](https://docs.docker.com/engine/reference/api/docker_remote_api/)

![OSD with schedulers](https://i.imgur.com/YNCqiwY.png)

### CSI
[Container Storage Interface](https://github.com/container-storage-interface/spec) is the standard way for a container orchestrator such as Kubernetes or Mesosphere to communicate with a storage provider.  OSD provides a CSI implementation to provision storage volumes to a container on behalf of any third party OSD driver and ensures the volumes are available in a multi host environment.

### Docker Volumes

OSD integrates with [Docker Volumes](https://docs.docker.com/engine/extend/plugins_volume/) and provisions storage to a container on behalf of any third party OSD driver and ensures the volumes are available in a multi host environment. 

# Documents

* [Example using NFS](docs/example-nfs.md)
* [Development](docs/development.md)

# Licensing
openstorage is licensed under the Apache License, Version 2.0.  See [LICENSE](https://github.com/pblcache/pblcache/blob/master/LICENSE) for the full license text.