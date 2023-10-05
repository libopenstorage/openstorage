![version](https://img.shields.io/badge/status-pre--alpha-lightgrey)

# Container Object Storage Interface Spec

This repository hosts the gRPC API for the Container Object Storage Interface (COSI) standard. The interfaces defined in the [gRPC specification](cosi.proto) are meant to be the common interface for object storage provisioning and management across various object storage vendors.

For more information about the COSI effort, visit our [documentation](https://container-object-storage-interface.github.io/docs).

## Why another standard?

Kubernetes abstracts file/block storage via the CSI standard. The primitives for file/block storage do not extend well to object storage. Here is the **_extremely_** concise and incomplete list of reasons why:

 - Unit of provisioned storage - Bucket instead of filesystem mount or block device.
 - Access is over the network instead of local POSIX calls.
 - No common protocol for consumption across various implementations of object storage.
 - Management policies and primitives - for instance, mounting and unmounting do not apply to object storage.

The existing primitives in CSI do not apply to objectstorage. Thus the need for a new standard to automate the management of objectstorage.

## Developer Guide

All API definitions **_MUST_** satisfy the following requirements:

<!-- - Must be backwards compatible -->
 - Must be in-sync with the API definitions in [sigs.k8s.io/container-object-storage-interface-api](https://github.com/kubernetes-sigs/container-object-storage-interface-api)

### Build and Test

1. `cosi.proto` is generated from the specification defined in `spec.md`

2. In order to update the API, make changes to `spec.md`. Then, generate `cosi.proto` using:

```sh
# generates cosi.proto
make generate
```

3. Clean and Build

```sh
# cleans up old build files
make clobber
# builds the go bindings
make
```

4. Do it all in 1 step:

```
# generates cosi.proto and builds the go bindings
make all
```

## References

- [Documentation](https://container-object-storage-interface.github.io/)
- [Deployment Guide](https://container-object-storage-interface.github.io/docs/deployment-guide)
- [Weekly Meetings](https://container-object-storage-interface.github.io/docs/community/weekly-meetings)
- [Roadmap](https://github.com/orgs/kubernetes-sigs/projects/8)

## Community, discussion, contribution, and support

You can reach the maintainers of this project at:

- [#sig-storage-cosi](https://kubernetes.slack.com/messages/sig-storage-cosi) slack channel
- [container-object-storage-interface](https://groups.google.com/g/container-object-storage-interface-wg?pli=1) mailing list

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
