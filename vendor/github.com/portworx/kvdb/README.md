# kvdb

[![GoDoc](https://godoc.org/github.com/portworx/kvdb?status.png)](https://godoc.org/github.com/portworx/kvdb)
[![Travis branch](https://img.shields.io/travis/portworx/kvdb/master.svg)](https://travis-ci.org/portworx/kvdb)
[![Go Report Card](https://goreportcard.com/badge/github.com/portworx/kvdb)](https://goreportcard.com/report/github.com/portworx/kvdb)
[![Code Coverage](https://codecov.io/gh/portworx/kvdb/branch/master/graph/badge.svg)](https://codecov.io/gh/portworx/kvdb)

Key Value Store abstraction library.

The kvdb library abstracts the caller from the specific key-value database implementation. The main goal of the kvdb library is to provide simple APIs to deal with only keys and values, and abstract away the intricate details of a specific key value stores. It also provides support for complex APIs like Snapshot, Watch and Lock which are built using the basic APIs.

### Supported key value stores

* `Etcd v2`
* `Etcd v3`
* `Consul`
* `In-memory store` (local to the node)
* `Bolt DB` (local to the node)
* `Zookeeper`

### Usage

The kvdb library is easy to use and requires you to create a new instance of the Kvdb object

```
package main

import (
  "github.com/portworx/kvdb"
  "github.com/portworx/kvdb/etcd/v3"
  "github.com/libopenstorage/openstorage/pkg/dbg"
)

func getKvdb(
  kvdbName string, // Use one of the kv store implementation names
  basePath string, // The path under which all the keys will be created by this kv instance
  discoveryEndpoints []string,  // A list of kv store endpoints
  options map[string]string, // Options that need to be passed to the kv store
  panicHandler kvdb.FatalErrorCB, // A callback function to execute when the library needs to panic
) (kvdb.Kvdb, error) {

	kv, err := kvdb.New(
		kvdbName,
		basePath,
		discoveryEndpoints,
		options,
		panicHandler,
	)
  return kv, err

}

type A struct {
   a1 string
   a2 int
}

func main() {

  // An example kvdb using etcd v3 as a key value store
  kv, err := getKvdb(
    v3.Name,
    "root/",
    []{"127.0.0.1:2379"},
    nil,
    dbg.Panicf,
  )
  if err != nil {
    fmt.Println("Failed to create a kvdb instance: ", err)
    return
  }

  // Put a key value pair foo=bar
  a := &A{"bar", 1}
  _, err = kv.Put("foo", &a, 0)
  if err != nil {
    fmt.Println("Failed to put a key in kvdb: ", err)
    return
  }

  // Get a key
  value := A{}
  _, err = kv.GetVal("foo", &value)
  if err != nil {
    fmt.Println("Failed to get a key from kvdb: ", err)
    return
  }
}

```

### Contributing

We are always looking for contributions from the open source community. Send out a PR and we will review it.


### Sign your work

The sign-off is a simple line at the end of the explanation for the
patch, which certifies that you wrote it or otherwise have the right to
pass it on as an open-source patch.  The rules are pretty simple: if you
can certify the below (from
[developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe@gmail.com>

using your real name (sorry, no pseudonyms or anonymous contributions.)

You can add the sign off when creating the git commit via `git commit -s`.

### License

kvdb library is licensed under the Apache License 2.0
