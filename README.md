# Open Storage
test
[![Travis branch](https://img.shields.io/travis/libopenstorage/openstorage/master.svg)](https://travis-ci.org/libopenstorage/openstorage)
[![Docker Pulls](https://img.shields.io/docker/pulls/openstorage/osd.svg)](https://hub.docker.com/r/openstorage/osd)
[![Go Report Card](https://goreportcard.com/badge/github.com/libopenstorage/openstorage)](https://goreportcard.com/report/github.com/libopenstorage/openstorage)

OpenStorage is an API abstraction layer providing support for multiple public APIs, including the [OpenStorage SDK](https://libopenstorage.github.io), [CSI](https://github.com/container-storage-interface/spec), and the [Docker Volume API](https://docs.docker.com/engine/reference/api/docker_remote_api/). Developers using OpenStorage for their storage systems can expect it to work seamlessly with any of the supported public APIs. These implementations provide users with the ability to run stateful services in Linux containers on multiple hosts.

OpenStoage makes it simple for developers to write a single implementation which supports many methods of control:

![openstorage](docs/images/openstorage.png)

Not only does OpenStorage allow storage developers to integrated their storage system with container orchestrations systems,
but also enables applications developers to use the OpenStorage SDK to manage and expose the latest storage features to their
clients.

## Supported Control APIs

### CSI
[Container Storage Interface](https://github.com/container-storage-interface/spec) is the standard way for a container orchestrator such as Kubernetes or Mesosphere to communicate with a storage provider.  OSD provides a CSI implementation to provision storage volumes to a container on behalf of any third party OSD driver and ensures the volumes are available in a multi host environment.

### Docker Volumes
OSD integrates with [Docker Volumes](https://docs.docker.com/engine/extend/plugins_volume/) and provisions storage to a container on behalf of any third party OSD driver and ensures the volumes are available in a multi host environment.

### OpenStorage SDK
CSI and Docker Volumes API provide a very generic storage control model, but with the [OpenStorage SDK](https://libopenstorage.github.io), applications can take control and utilize the latest features of a storage system. For example, with the OpenStorage SDK, applications can control their volumes backups, schedules, etc.

# Documents

* [Example using NFS](docs/example-nfs.md)
* [Development](docs/development.md)

# Licensing
openstorage is licensed under the Apache License, Version 2.0.  See [LICENSE](https://github.com/pblcache/pblcache/blob/master/LICENSE) for the full license text.
