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
Go provides built-in tools to update the vendor and add packages. All the packages name and versions
are defined in go.mod. Checksums are in go.sum.

To vendor the packages:
```
# clean up the go.mod require sections with only the required packages
go mod tidy
# download the packages (relevant go files) into the vendor folder
go mod vendor

# --- or ----
make vendor
```

There are multiple ways to adding a new package. First is to run the command below. This automatically updates go.mod and go.sum.
```
go get -u some/package@version
```
After adding, make sure to vendor the package as well.

Second way is to manually update the go.mod file and run vendor command. 

Note: In some cases, you might need a pesudo version instead of a tagged one. 
Example:
```
# via go get
go get golang.org/x/sys@v0.0.0-20190726090000-fde4db37ae7a
# via manual modifying go.mod
require golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
```
For reference: https://jfrog.com/blog/go-big-with-pseudo-versions-and-gocenter/


## Updating vendor with govendor (Old)

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
