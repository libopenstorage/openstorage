# openstorage
openstorage is an implementation of the Open Storage specification

### Building:

**Note**: libopenstorage only builds on Linux since it uses Linux-only APIs.

## Installing Dependencies

libopenstorage is written in the [Go](http://golang.org) programming language. If you haven't set up a Go development environment, please follow [these instructions](http://golang.org/doc/code.html) to install go tool and set up GOPATH. Ensure that your version of Go is at least 1.3. Note that the version of Go in package repositories of some operating systems is outdated, so please [download](https://golang.org/dl/) the latest version.

After setting up Go, you should be able to `go get` libopenstorage as expected (we use `-d` to only download):

```
$ go get -d github.com/libopenstorage/openstorage
```

We use `godep` so you will need to get that as well:

```
$ go get github.com/tools/godep
```

## Building from Source

At this point you can build cAdvisor from the source folder:

```
$GOPATH/src/github.com/libopenstorage/openstorage $ godep go build .
```

or run only unit tests:

```
$GOPATH/src/github.com/libopenstorage/openstorage $ godep go test ./... -test.short
```
## Updating to latest Source

To update the source folder and all dependencies:

```
$GOPATH/src/github.com/libopenstorage/openstorage $ go get -u all
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
