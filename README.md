# Open Storage

[![Travis branch](https://img.shields.io/travis/libopenstorage/openstorage/master.svg)](https://travis-ci.org/libopenstorage/openstorage)
[![Docker Pulls](https://img.shields.io/docker/pulls/openstorage/osd.svg)](https://hub.docker.com/r/openstorage/osd)
[![Go Report Card](https://goreportcard.com/badge/github.com/libopenstorage/openstorage)](https://goreportcard.com/report/github.com/libopenstorage/openstorage)

OpenStorage is a clustered implementation of the [Open Storage](https://github.com/libopenstorage/specs) specification and facilitates the provisioning of cloud native volumes for Kubernetes.  It allows you to run stateful services in Linux Containers in a multi host, multi zone and multi region environment.  It plugs into [CSI](https://landscape.cncf.io/selected=container-storage-interface-csi) and Docker volumes to provide storage to containers and plugs into Kubernetes to allow for dynamic and programmatic provisioning of volumes.

# What you get from using Open Storage

When you install openstorage on a Linux host, you will automatically get a stateful storage overlay that integrates with [CSI](https://landscape.cncf.io/selected=container-storage-interface-csi) or the Docker runtime and provide volumes that are usable across hosts that are running in different regions or even clouds.  It starts an Open Storage Daemon - `OSD` that supports any Linux container runtime that conforms to the [OCI](https://www.opencontainers.org/) spec.  From here, you can use Kubernetes to directly create cloud native storage volumes that are highly available cluster wide.

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
