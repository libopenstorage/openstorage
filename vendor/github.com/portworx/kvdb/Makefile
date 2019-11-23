ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v test)
endif

all: test

deps:
	dep ensure -v

updatedeps:
	dep ensure -update -v

build: deps
	go build ./...

install: deps
	go install ./...

lint:
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go' | grep -v '\.pb\.gw\.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet:
	go vet ./...

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck \
		github.com/portworx/kvdb \
		github.com/portworx/kvdb/common \
		github.com/portworx/kvdb/consul \
		github.com/portworx/kvdb/mem \
		github.com/portworx/kvdb/wrappers \
		github.com/portworx/kvdb/zookeeper

pretest: deps errcheck lint vet

gotest:
	for pkg in $(PKGS); \
		do \
			go test --timeout 1h -v -tags unittest -coverprofile=profile.out -covermode=atomic $(BUILD_OPTIONS) $${pkg} || exit 1; \
			if [ -f profile.out ]; then \
				cat profile.out >> coverage.txt; \
				rm profile.out; \
			fi; \
		done

test: pretest gotest

docker-build-kvdb-dev:
	docker build -t portworx/kvdb:test_container -f $(GOPATH)/src/github.com/portworx/kvdb/Dockerfile.kvdb .

docker-test:
	docker run --rm \
		-v $(GOPATH)/src/github.com/portworx/kvdb:/go/src/github.com/portworx/kvdb \
		portworx/kvdb:test_container \
		make gotest

clean:
	go clean -i ./...

.PHONY: \
	all \
	deps \
	updatedeps \
	build \
	install \
	lint \
	vet \
	errcheck \
	pretest \
	gotest \
	test \
	clean
