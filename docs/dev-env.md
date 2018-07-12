# Setup Development Environment

## Installing Dependencies

libopenstorage is written in the [Go](http://golang.org) programming language. If you haven't set up a Go development environment, please follow [these instructions](http://golang.org/doc/code.html) to install `golang` and set up GOPATH.

* Install [swagger](https://github.com/go-swagger/go-swagger)

```
go get -u github.com/go-swagger/go-swagger/cmd/swagger
```

* Install [Golang mock](https://github.com/golang/mock)

```
go get -u github.com/golang/mock/gomock
go get -u github.com/golang/mock/mockgen
```

The install also requires `jq` so depending on your distro you'll have to install that too.