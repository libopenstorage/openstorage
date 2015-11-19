## What is BUSE?
BUSE is a block driver in user space.  It leverages `NBD` to provide block volume access to a container.  In the back, it writes data out to a local file.

### Using BUSE
Declare `buse` as a driver in your OSD config file as such:
```
---
osd:
  cluster:
    nodeid: "1"
    clusterid: "deadbeeef"
  drivers:
    buse:
```

BUSE relies on NBD to export block devices.  Therefore, remember to `modprobe nbd`.

## Roadmap
BUSE will add support for the [Docker Graph driver](https://github.com/docker/docker/tree/master/daemon/graphdriver).  You will be able to store Docker images on a distributed multi host Docker cluster and have updates made available from within the local cluster.

Future revs of BUSE will include replication to a host based on the cluster API.  It will log transactions to a local database and synchronously replicate blocks to a peer BUSE driver.

With this version of BUSE you should be able to use Docker in a multi host - shared nothing storage cluster.
