# Current Build Status
[![Build Status](http://emma.openstorage.org/buildStatus/icon?job=Openstorage&style=plastic)](http://emma.openstorage.org/job/Openstorage/)

# About Open Storage

Openstorage is a clustered implementation of the [Open Storage](https://github.com/libopenstorage/specs) specification and relies on the Docker runtime.  It allows you to run stateful services in Docker in a multi-host environment.  It plugs into Docker volumes to provide storage to a container and plugs into Swarm to operate in a clustered environment. 

# What you get from using Open Storage

When you install openstorage on a Linux host, you will automatically get a stateful storage layer that integrates with the Docker runtime and operates in a multi host environment.  It starts an Open Storage Daemon - `OSD` that currently supports Docker and will support any Linux container runtime that conforms to the [OCI](https://www.opencontainers.org/).  This daemon integrates with Docker volumes and provisions storage to a container on behalf of any third party OSD driver and ensures the volumes are available in a multi host environment with a scheduler that can support the openstorage plugin API.

### An example usage

The diagram below shows OSD integrated with Docker and Swarm to allow for provisioning of storage to containers in a multi node environment.

![OSD - Docker - Swarm integration](http://i.imgur.com/W2pXsqb.png)

There are default drivers built-in for NFS, AWS and BTRFS.  By using openstorage, you can get container granular, stateful storage provisioning to Linux containers with the backends supported by openstorage.  We are working with the storage ecosystem to add more drivers for various storage providers.

Providers that support a multi-node environment, such as AWS or NFS to name a few, can provide highly available storage to linux containers across multiple hosts.

## Installing Dependencies

libopenstorage is written in the [Go](http://golang.org) programming language. If you haven't set up a Go development environment, please follow [these instructions](http://golang.org/doc/code.html) to install `golang` and set up GOPATH. Your version of Go must be at least 1.5 - we use the golang 1.5 vendor experiment https://golang.org/s/go15vendor. Note that the version of Go in package repositories of some operating systems is outdated, so please [download](https://golang.org/dl/) the latest version.

After setting up Go, you should be able to `go get` libopenstorage as expected (we use `-d` to only download):

```
$ go get -d github.com/libopenstorage/openstorage/...
```

## Building from Source

At this point you can build openstorage from the source folder:

```
$GOPATH/src/github.com/libopenstorage/openstorage $ make install
```

or run only unit tests:

```
$GOPATH/src/github.com/libopenstorage/openstorage $ make test
```

## Starting OSD

OSD is both the openstorage daemon and the CLI.  When run as a daemon, the OSD is ready to receive RESTful commands to operate on volumes and attach them to a Docker container.  It works with the [Docker volumes plugin interface](https://github.com/docker/docker/blob/e5af7a0e869c0a66f8ab30d3a90280843b9999e0/docs/extend/plugins_volume.md) will communicate with Docker version 1.7 and later.  When this daemon is running, Docker will automatically commincate with the daemon to manage a container's volumes.

To start the OSD in daemon mode:
```
osd -d -f config.yaml
```
where, config.yaml is the daemon's configuiration file and it's format is explained [below](https://github.com/libopenstorage/openstorage/blob/master/README.md#osd-config-file).

To use the OSD cli, see the CLI help menu:
```
NAME:
   osd - Open Storage CLI

USAGE:
   osd [global options] command [command options] [arguments...]

VERSION:
   0.3

COMMANDS:
   driver, d    Manage drivers
   aws, v       Manage aws volumes
   nfs, v       Manage nfs volumes
   help, h      Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --json, -j                                   output in json
   --daemon, -d                                 Start OSD in daemon mode
   --driver [--driver option --driver option]   driver name and options: name=btrfs,root_vol=/var/openstorage/btrfs
   --file, -f                                   file to read the OSD configuration from.
   --help, -h                                   show help
   --version, -v                                print the version
```

## OSD config file

The OSD daemon loads a YAML configuration file that tells the daemon what drivers to load and the driver specific attributes.  Here is an example of config.yaml:

```
osd:
  drivers:
      nfs:
        server: "171.30.0.20"
        path: "/nfs"
      btrfs:
      aws:
        aws_access_key_id: your_aws_access_key_id
        aws_secret_access_key: your_aws_secret_access_key
```

The above example initializes the `OSD` with three drivers: NFS, BTRFS and AWS.  Each have their own configuration sections.

## Adding your driver

Adding a driver is fairly straightforward:

1. Add your driver decleration in `drivers.go`


2. Add your driver `mydriver` implementation in the `drivers/mydriver` directory.  The driver must implement the `VolumeDriver` interface specified in [`volumes/volume.go`](https://github.com/libopenstorage/openstorage/blob/master/volume/volume.go).  This interface is an implementation of the specification available [here] (http://api.openstorage.org/).

3. You're driver must be a `File Volume` driver or a `Block Volume` driver.  A `File Volume` driver will not implement a few low level primatives, such as `Format`, `Attach` and `Detach`.


Here is an example of `drivers.go`:

```
// To add a provider to openstorage, declare the provider here.
package main

import (
    "github.com/libopenstorage/openstorage/drivers/aws"
    "github.com/libopenstorage/openstorage/drivers/btrfs"
    "github.com/libopenstorage/openstorage/drivers/nfs"
    "github.com/libopenstorage/openstorage/volume"
)           
            
type Driver struct {
    providerType volume.ProviderType
    name         string
}       
            
var (       
    providers = []Driver{
        // AWS provider. This provisions storage from EBS.
        {providerType: volume.Block,
            name: aws.Name},
        // NFS provider. This provisions storage from an NFS server.
        {providerType: volume.File,
            name: nfs.Name},
        // BTRFS provider. This provisions storage from local btrfs fs.
        {providerType: volume.File,
            name: btrfs.Name},
    }
)
```

That's pretty much it.  At this point, when you start the OSD, your driver will be loaded.

## Testing

```
make test # test on your local machine
make docker-test # test within a docker container
```

## Updating to latest Source

To update the source folder and all dependencies:

```
$GOPATH/src/github.com/libopenstorage/openstorage $ make update-test-deps
```

However note that all dependencies are vendored in the vendor directory, so this is not necessary in general as long as you have `GO15VENDOREXPERIMENT` set:

```
export GO15VENDOREXPERIMENT=1
```

## Building a Docker image

OSD can run inside of Docker:

```
make docker-build-osd
```

This builds a Docker image called `openstorage/osd`.  You can then run the image:

```
make launch
```

#### Using openstorage with systemd

```service
[Unit]
Description=Open Storage

[Service]
CPUQuota=200%
MemoryLimit=1536M
ExecStart=/usr/local/bin/osd
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

# Contributing

The specification and code is licensed under the Apache 2.0 license found in 
the `LICENSE` file of this repository.  
