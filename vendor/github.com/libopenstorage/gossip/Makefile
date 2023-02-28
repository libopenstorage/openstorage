ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'github.com/libopenstorage/gossip/vendor')
endif

.PHONY: dep

export GO15VENDOREXPERIMENT=1

vendor-update:
	GOOS=linux GOARCH=amd64 go get -tags "daemon btrfs_noversion have_btrfs have_chainfs" -d -v -t -u -f $(PKGS)

vendor-gomod:
	GOOS=linux GOARCH=amd64 go mod tidy
	GOOS=linux GOARCH=amd64 go mod vendor

vendor: vendor-gomod

all: test

test:
	for pkg in $(PKGS); \
	do \
		go test --timeout 1h -v -tags unittest -coverprofile=profile.out -covermode=atomic $(BUILD_OPTIONS) $${pkg} || exit 1; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done

docker-test:
	docker run \
	--privileged \
        --net=host \
	openstorage/osd-gossip \
		make test

docker-build-osd-gossip:
	docker build -t openstorage/osd-gossip -f Dockerfile.osd-gossip .

push-docker-images: docker-build-osd-gossip
	docker push openstorage/osd-gossip
