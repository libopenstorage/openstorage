TAGS := daemon btrfs_noversion
PKGS := $(shell go list ./... | grep -v 'github.com/libopenstorage/openstorage/vendor')

ifeq ($(BUILD_TYPE),debug)
BUILDFLAGS := -gcflags "-N -l"
endif

export GO15VENDOREXPERIMENT=1

all: test install

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v $(PKGS)

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f $(PKGS)

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t $(PKGS)

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f $(PKGS)

vendor:
	go get -v github.com/tools/godep
	rm -rf Godeps
	rm -rf vendor
	# TODO: when godep fixes downloading all tags, remove the custom package
	# https://github.com/tools/godep/issues/271
	godep save $(PKGS) github.com/docker/docker/pkg/chrootarchive
	rm -rf Godeps

build:
	go build -tags "$(TAGS)" $(BUILDFLAGS) $(PKGS)

install:
	go install -tags "$(TAGS)" $(PKGS)

lint:
	go get -v github.com/golang/lint/golint
	$(foreach pkg,$(PKGS),golint $(pkg);)

vet:
	go vet $(PKGS)

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck $(PKGS)

pretest: lint vet errcheck

test:
	go test -tags "$(TAGS)" $(PKGS)

docker-build:
	docker build -t openstorage/osd-dev -f Dockerfile.osd-dev .

docker-test: docker-build
	docker run \
		--privileged \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-v /var/run/docker.sock:/var/run/docker.sock \
		openstorage/osd-dev \
			make test

docker-build-osd-internal:
	rm -rf _tmp
	mkdir -p _tmp
	go build -a -tags "$(TAGS)" -o _tmp/osd cmd/osd/main.go
	docker build -t openstorage/osd -f Dockerfile.osd .

docker-build-osd: docker-build
	docker run -v /var/run/docker.sock:/var/run/docker.sock openstorage/osd-dev make docker-build-osd-internal

launch: docker-build-osd
	docker run \
		--privileged \
		-v $(shell pwd):/etc \
		-v /usr/share/docker/plugins:/usr/share/docker/plugins \
		-v /var/lib/osd/driver:/var/lib/osd/driver \
		-v /mnt:/mnt \
		openstorage/osd -d -f /etc/config.yaml

clean:
	go clean -i $(PKGS)

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
	docker-build-osd-internal \
	docker-build-osd \
	launch \
	clean
