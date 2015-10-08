ifeq ($(BUILD_TYPE),debug)
BUILDFLAGS := -gcflags "-N -l"
endif

all: test install

deps:
	go get -d -v ./...

update-deps:
	go get -d -v -u -f ./...

test-deps:
	go get -d -v -t ./...

update-test-deps:
	go get -d -v -t -u -f ./...

build: deps
	go build -tags daemon $(BUILDFLAGS) ./...

install: deps
	go install -tags daemon ./...

lint:
	go get -v github.com/golang/lint/golint
	golint ./...

vet:
	go vet ./...

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck ./...

pretest: lint vet errcheck

test: test-deps
	go test -tags daemon ./...

docker-build:
	docker build -t openstorage/osd .

docker-test: docker-build
	docker run openstorage/osd make test

clean:
	go clean -i ./...

.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	vendor \
	build \
	install \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	docker-build \
	docker-test \
	clean
