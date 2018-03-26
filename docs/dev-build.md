# Building the source

## Building from Source

From the root of the tree type:

```
$ make install
```

If you have changed any APIs or interfaces you may need to regenerate software
using the following command:

```
$ make generate
```

This will regenerate the swagger json file and any Golang mock files.

## Updating vendor

If you are lucky enough to need to update a vendor package you will need to
use [govendor](https://github.com/kardianos/govendor) to update it.

First install `govendor`:

```
go get -u github.com/kardianos/govendor
```

Next `go get` your package. Govendor pulls the version of the package from
your `$GOPATH`.

```
go get -u some/package
```

Lastly, you can now use govendor from the top of the openstorage tree to
either `add` or `update` a depency:

```
govendor add some/package/...
```

Make sure to include the `/...` after your package. You may also need to replace `add` with `update` if the package is already known.

### Logrus issue

As of yet, there are a lot of packages which use `logrus` with older `Sirupsen/logrus`
instead of `sirupsen/logrus`. For this reason, we provide a simple script
which fixes any of these issues. Just run the script from the top of the
openstorage tree:

```
./hack/force-github.com-sirupsen-logrus.sh
```
