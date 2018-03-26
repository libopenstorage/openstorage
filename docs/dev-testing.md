# Testing

openstorage supports has multiple ways to test your code. openstorage has
unit tests, csi-sanity, and osd-sanity to check the quality of the code.

## csi-sanity
csi-sanity is a program provided by the CSI community which assures that the
CSI API implementation is correct. This program is executed automatically by
unit tests when [csi/csisanity_test.go](csi/csisanity_test.go) is executed.

## osd-sanity
osd-sanity is a tool provided by openstorage which can be used to point to a
running instance of a openstorage implementation to make sure it behaves and
interacts according to the API.

## Unit tests and Golang mock


## Running tests

Docker is _required_ for the tests. To run the tests, run:

```
make docker-test # test within a docker container
```

## Testing manually

To test changes in the library, you could use the NFS driver. Here is how to
use it with Docker.

* Build using `make install`
* Start up your NFS server and expose a share called `/nfs`
* Run on the `osd` sample daemon in one window:

```
sudo $GOPATH/bin/osd -d -f etc/config/config.yaml
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