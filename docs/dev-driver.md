# Driver development

There are two main ways to have an openstorage driver, one is by vendoring
openstorage with your repo, and the other is to include your driver in this repo.

Both models have their positives and negatives. By vendoring openstorage you can
include the api and the server on your repo, but are responsible for updating
the library. By including it in openstorage, it will be part of the openstorage
library and the community and you will try to keep it up to date.

## Creating a volume driver

Adding a driver is fairly straightforward:

1. Add your driver decleration in `volumes/drivers.go`
2. Add your driver `mydriver` implementation in the `volumes/drivers/mydriver` directory.  The driver must implement the `VolumeDriver` interface specified in [`volumes/volume.go`](https://github.com/libopenstorage/openstorage/blob/master/volume/volume.go).  This interface is an implementation of the specification available at [api.openstorage.org](http://api.openstorage.org/).
3. You're driver must be a `File Volume` driver or a `Block Volume` driver.  A `File Volume` driver will not implement a few low level primatives, such as `Format`, `Attach` and `Detach`.

Here is an example of `drivers.go`:

```
// To add a provider to openstorage, declare the provider here.
package drivers

import (
    "github.com/libopenstorage/openstorage/volume/drivers/aws"
    "github.com/libopenstorage/openstorage/volume/drivers/btrfs"
    "github.com/libopenstorage/openstorage/volume/drivers/nfs"
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

## Building a Docker image

OSD can run inside of Docker:

```
make docker-build-osd
```

This builds a Docker image called `openstorage/osd`.  You can then run the image:

```
make launch
```

### OSD on the Docker registry

Pre-built Docker images of the OSD are available at https://hub.docker.com/r/openstorage/osd/

## Using openstorage with systemd

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