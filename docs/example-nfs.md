# Example: NFS backend for CSI and Docker

The diagram below shows OSD integrated with Docker and Kubernetes to allow for provisioning of storage to containers in a multi node environment using CSI.

![OSD - Kubernetes integration](https://i.imgur.com/ktcVm10.png)

There are default drivers built-in for NFS, AWS and BTRFS.  By using openstorage, you can get container granular, stateful storage provisioning to Linux containers with the backends supported by openstorage.  We are working with the storage ecosystem to add more drivers for various storage providers.

Providers that support a multi-node environment, such as AWS or NFS to name a few, can provide highly available storage to linux containers across multiple hosts.

## Starting OSD

OSD is both the openstorage daemon and the CLI.  When run as a daemon, the OSD is ready to receive RESTful commands to operate on volumes and attach them to a Docker container.  It works with the [Docker volumes plugin interface](https://github.com/docker/docker/blob/e5af7a0e869c0a66f8ab30d3a90280843b9999e0/docs/extend/plugins_volume.md) will communicate with Docker version 1.7 and later.  When this daemon is running, Docker will automatically communicate with the daemon to manage a container's volumes.

* Build using `make install`
* Start up your NFS server and expose a share called `/nfs`
* Run on the `osd` sample daemon in one window:

To start the OSD in daemon mode:

```
sudo $GOPATH/bin/osd -d -f etc/config/config.yaml
```

where, config.yaml is the daemon's configuiration file and its format is explained [below](https://github.com/libopenstorage/openstorage/blob/master/README.md#osd-config-file).

To have OSD persist the volume mapping across restarts, you must use an external key value database such as [etcd](https://coreos.com/etcd/docs/latest/docker_guide.html) or [consul](https://www.consul.io/intro/getting-started/install.html).  The URL of your key value database must be passed into the OSD using the `--kvdb` option.  For example:

```
$GOPATH/bin/osd -d -f etc/config/config.yaml -k etcd-kv://localhost:4001
```

### Testing with Docker

Assuming you are using the NFS driver, to create a volume with a default size of 1GB and attach it to a Docker container, you can do the following
```
$ docker volume create -d nfs
9ccb7280-918b-464f-8a34-34e73e9214d2
$ docker run -v 9ccb7280-918b-464f-8a34-34e73e9214d2:/root --volume-driver=nfs -ti busybox
```

### Testing with CSI

To test with csi, use the `csc` tool from [rexray/gocsi](https://github.com/rexray/gocsi).

* Install `csc`

```
go install github.com/rexray/gocsi/csc
```

* Run `csc` against your driver. Here is an example of running against the NFS example above using v0.2.0 of CSI:

```
csc controller list-volumes -v 0.2.0 --endpoint /var/lib/osd/driver/nfs-csi.sock
```