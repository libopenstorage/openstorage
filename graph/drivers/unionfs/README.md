## Fuse based Union FS Graph Driver
### EXPERIMENTAL!  
Note that this is still in development and experimental.  Currently the following are known issues

1. Data modified on shared layers are not snap'd and therefore visible in other containers.
2. There are two heavy weight locks around accessing the UFS data structures which can be avoided.

### About
The `unionfs` graph driver leverages the kernel-userspace communication protocol `fuse` to implement the storage of graph layers.

It also uses an optimized way of computing the layer diffs and avoids using the NaiveDiff implementation.

To use this as the graphdriver in Docker with btrfs as the backend volume provider:

```
DOCKER_STORAGE_OPTIONS= -s unionfs --storage-opt unionfs.volume_driver=btrfs
```

or

```
docker daemon --storage-driver=unionfs --storage-opt unionfs.volume_driver=btrfs
```

### Building

Make sure you have `fuse` installed.

When building `OSD`, run:

```
HAVE_UNIONFS=1 EXPERIMENTAL_=1 make
```
