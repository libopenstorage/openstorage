## Fuse Graph Driver

The `fuse` graph driver leverages the kernel-userspace communication protocol to implement the storage of graph layers.

It also uses an optimized way of computing the layer diffs and avoids using the NaiveDiff implementation.

To use this as the graphdriver in Docker with aws as the backend volume provider:

```
DOCKER_STORAGE_OPTIONS= -s unionfs
```
