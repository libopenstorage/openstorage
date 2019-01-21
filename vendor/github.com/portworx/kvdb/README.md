## kvdb

[![Travis branch](https://img.shields.io/travis/portworx/kvdb/master.svg)](https://travis-ci.org/portworx/kvdb)
[![Go Report Card](https://goreportcard.com/badge/github.com/portworx/kvdb)](https://goreportcard.com/report/github.com/portworx/kvdb)
[![Code Coverage](https://codecov.io/gh/portworx/kvdb/branch/master/graph/badge.svg)](https://codecov.io/gh/portworx/kvdb)

Key Value Store abstraction library.

This library abstracts the caller from the specific key-value database implementation. This makes the underlying providers interchangable. Current supported implementations are:
* `Etcd v2`
* `Etcd v3`
* `Consul`
* `In-memory store`
