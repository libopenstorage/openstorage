ETCD := $(shell pgrep etcd)
CONSUL := $(shell pgrep consul)

all: test

deps:
	go get -d -v ./...

updatedeps:
	go get -d -v -u -f ./...

testdeps:
	go get -d -v -t ./...

updatetestdeps:
	go get -d -v -t -u -f ./...

build: deps
	go build ./...

install: deps
	go install ./...

lint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go' | grep -v '\.pb\.gw\.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	go vet ./...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck \
		github.com/portworx/kvdb \
		github.com/portworx/kvdb/common \
		github.com/portworx/kvdb/consul \
		github.com/portworx/kvdb/etcd/v2 \
		github.com/portworx/kvdb/etcd/v3 \
		github.com/portworx/kvdb/mem

pretest: errcheck lint vet

testcheck:
ifndef ETCD
	@echo "etcd not running. Unable to run tests."
	exit 1
endif
ifndef CONSUL
	@echo "consul not running. Unable to run tests."
	exit 1
endif

gotest: testdeps testcheck
	go test -v ./...

test: pretest gotest

clean:
	go clean -i ./...

.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	build \
	install \
	lint \
	vet \
	errcheck \
	pretest \
	gotest \
	test \
	clean
