HAS_SDKTEST := $(shell command -v sdk-test 2> /dev/null)
BRANCH	:= $(shell git rev-parse --abbrev-ref HEAD)

ifeq ($(TRAVIS_BRANCH), master)
MOCKSDKSERVERTAG := latest
else

ifeq ($(BRANCH), master)
MOCKSDKSERVERTAG := latest
else
MOCKSDKSERVERTAG := $(shell go run tools/sdkver/sdkver.go)
endif

endif

REGISTRY = openstorage
IMAGE_MOCKSDKSERVER := $(REGISTRY)/mock-sdk-server:$(MOCKSDKSERVERTAG)

ifndef TAGS
TAGS := daemon
endif

ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'vendor' | grep -v 'sanity')
endif

ifeq ($(BUILD_TYPE),debug)
BUILDFLAGS := -gcflags "-N -l"
endif

ifdef HAVE_BTRFS
TAGS+=btrfs_noversion have_btrfs
endif

ifdef HAVE_CHAINFS
TAGS+=have_chainfs
endif

ifndef PROTOC
PROTOC = protoc
endif

ifndef PROTOS_PATH
PROTOS_PATH = $(GOPATH)/src
endif

ifndef PROTOSRC_PATH
PROTOSRC_PATH = $(PROTOS_PATH)/github.com/libopenstorage/openstorage
endif

OSDSANITY:=cmd/osd-sanity/osd-sanity

.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	vendor-update \
	vendor-without-update \
	vendor-gomod \
	vendor \
	build \
	install \
	proto \
	lint \
	vet \
	fmt \
	packr \
	errcheck \
	pretest \
	test \
	hack-tests \
	docs \
	docker-build-osd-dev \
	docker-build \
	docker-test \
	docker-build-osd-internal \
	docker-build-osd \
	launch \
	launch-local-btrfs \
	install-flexvolume-plugin \
	$(OSDSANITY)-install \
	$(OSDSANITY)-clean \
	clean \
	generate \
	generate-mockfiles \
	e2e \
	verify \
	osd-tests \
	pr-lint \
	pr-verify \
	pr-unit-tests \
	sdk-check-version


all: build $(OSDSANITY)

# TOOLS build rules
#
$(GOPATH)/bin/golint:
	@echo "Installing missing $@ ..."
	GO111MODULE=off go get -u github.com/golang/lint/golint

$(GOPATH)/bin/errcheck:
	@echo "Installing missing $@ ..."
	GO111MODULE=off go get -u github.com/kisielk/errcheck

$(GOPATH)/bin/packr2:
	@echo "Installing missing $@ ..."
	GO111MODULE=off go get -u github.com/gobuffalo/packr/...

$(GOPATH)/bin/cover:
	@echo "Installing missing $@ ..."
	GO111MODULE=off go get -u golang.org/x/tools/cmd/cover

$(GOPATH)/bin/gotestcover:
	@echo "Installing missing $@ ..."
	GO111MODULE=off go get -u github.com/pierrre/gotestcover

$(GOPATH)/bin/contextcheck:
	@echo "Installing missing $@ ..."
	GO111MODULE=off go get -u github.com/sylvia7788/contextcheck

$(GOPATH)/bin/misspell:
	@echo "Installing missing $@ ..."
	GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell

$(GOPATH)/bin/gomock:
	GO111MODULE=off go get github.com/golang/mock/gomock

$(GOPATH)/bin/mockgen:
	GO111MODULE=off go get github.com/golang/mock/mockgen

# DEPS build rules
#

deps:
	GO111MODULE=off go get -d -v $(PKGS)

update-deps:
	GO111MODULE=off go get -d -v -u -f $(PKGS)

test-deps:
	GO111MODULE=off go get -d -v -t $(PKGS)

update-test-deps:
	GO111MODULE=off go get -tags "$(TAGS)" -d -v -t -u -f $(PKGS)

vendor-update:
	GOOS=linux GOARCH=amd64 go get -tags "daemon btrfs_noversion have_btrfs have_chainfs" -d -v -t -u -f $(PKGS)

vendor-gomod:
	GOOS=linux GOARCH=amd64 go mod tidy
	GOOS=linux GOARCH=amd64 go mod vendor

vendor: vendor-gomod

build: packr
	go build -tags "$(TAGS)" $(BUILDFLAGS) $(PKGS)

install: packr $(OSDSANITY)-install
	go install -gcflags="all=-N -l" -tags "$(TAGS)" $(PKGS)
	go install github.com/libopenstorage/openstorage/cmd/osd-token-generator

$(OSDSANITY):
	@$(MAKE) -C cmd/osd-sanity

$(OSDSANITY)-install:
	@$(MAKE) -C cmd/osd-sanity install

$(OSDSANITY)-clean:
	@$(MAKE) -C cmd/osd-sanity clean

docker-build-proto:
	docker build -t quay.io/openstorage/osd-proto --network=host -f Dockerfile.proto .

docker-proto:
	docker run \
		--privileged --rm \
		-v $(shell pwd):/go/src/github.com/libopenstorage/openstorage \
		-e "GOPATH=/go" \
		-e "DOCKER_PROTO=yes" \
		-e "PATH=/bin:/usr/bin:/usr/local/bin:/go/bin:/usr/local/go/bin" \
		quay.io/openstorage/osd-proto \
			make proto mockgen

proto: $(GOPATH)/bin/protoc-gen-go $(GOPATH)/bin/protoc-gen-grpc-gateway $(GOPATH)/bin/protoc-gen-swagger
ifndef DOCKER_PROTO
	$(error Do not run directly. Run 'make docker-proto' instead.)
endif

	@echo ">>> Generating protobuf definitions from api/api.proto"
	$(PROTOC) -I $(PROTOSRC_PATH) \
		-I /usr/local/include \
		-I $(PROTOS_PATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:. \
		$(PROTOSRC_PATH)/api/api.proto
	$(PROTOC) -I $(PROTOSRC_PATH) \
		-I /usr/local/include \
		-I $(PROTOS_PATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:. \
		$(PROTOSRC_PATH)/api/api.proto
	$(PROTOC) -I $(PROTOSRC_PATH) \
		-I /usr/local/include \
		-I $(PROTOS_PATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=logtostderr=true:$(PROTOSRC_PATH)/api/server/sdk \
		$(PROTOSRC_PATH)/api/api.proto
	@echo ">>> Upgrading swagger 2.0 to openapi 3.0"
	mv api/server/sdk/api/api.swagger.json api/server/sdk/api/20api.swagger.json
	swagger2openapi api/server/sdk/api/20api.swagger.json -o api/server/sdk/api/api.swagger.json
	rm -f api/server/sdk/api/20api.swagger.json
	@echo ">>> Generating grpc protobuf definitions from pkg/flexvolume/flexvolume.proto"
	$(PROTOC) -I/usr/local/include -I$(PROTOSRC_PATH) -I$(PROTOS_PATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. $(PROTOSRC_PATH)/pkg/flexvolume/flexvolume.proto
	$(PROTOC) -I/usr/local/include -I$(PROTOSRC_PATH) -I$(PROTOS_PATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. $(PROTOSRC_PATH)/pkg/flexvolume/flexvolume.proto
	@echo ">>> Generating protobuf definitions from pkg/jsonpb/testing/testing.proto"
	$(PROTOC) -I $(PROTOSRC_PATH) $(PROTOSRC_PATH)/pkg/jsonpb/testing/testing.proto --go_out=plugins=grpc:.
	@echo ">>> Updating SDK versions"
	go run tools/sdkver/sdkver.go --swagger api/server/sdk/api/api.swagger.json

lint: $(GOPATH)/bin/golint
	golint $(PKGS)

vet:
	@if [ $(shell go vet $(PKGS) | grep -v ".*contains sync.Mutex.*" | wc -l) != 0 ]; then\
		echo "go vet failed (ignore the errors associate with sync.Mutex";\
		exit 1;\
	fi
	
fmt:
	go fmt $(go list ./... | grep -v vendor) | grep -v "api.pb.go" | wc -l | grep "^0";

errcheck: $(GOPATH)/bin/errcheck
	errcheck -tags "$(TAGS)" $(PKGS)

contextcheck: $(GOPATH)/bin/contextcheck
	contextcheck $(PKGS)

misspell: $(GOPATH)/bin/misspell
	git ls-files | grep -v vendor | grep -v .pb.go | grep -v .js | grep -v .css | xargs misspell

pretest: lint vet errcheck

install-sdk-test:
ifndef HAS_SDKTEST
	@echo "Installing sdk-test"
	@-GO111MODULE=off go get -d -u github.com/libopenstorage/sdk-test 1>/dev/null 2>&1
	@(cd $(GOPATH)/src/github.com/libopenstorage/sdk-test/cmd/sdk-test && make install )
endif

test-sdk: install-sdk-test launch-sdk
	timeout 30 sh -c 'until curl --silent -X GET -d {} http://localhost:9110/v1/clusters/inspectcurrent | grep STATUS_OK; do sleep 1; done'
	sdk-test -ginkgo.noColor -ginkgo.noisySkippings=false -sdk.endpoint=localhost:9100 -sdk.cpg=$(GOPATH)/src/github.com/libopenstorage/sdk-test/cmd/sdk-test/cb.yaml

# TODO: Remove GODEBUG and fix test certs
test: packr
	GODEBUG=x509ignoreCN=0 go test -tags "$(TAGS)" $(TESTFLAGS) $(PKGS)
	
docs: $(GOPATH)/bin/gomock $(GOPATH)/bin/swagger $(GOPATH)/bin/mockgen
	go generate ./cmd/osd/main.go

packr: $(GOPATH)/bin/packr2
	packr2 clean
	packr2

generate-mockfiles: $(GOPATH)/bin/gomock $(GOPATH)/bin/swagger $(GOPATH)/bin/mockgen
	go generate $(PKGS)

generate: docs generate-mockfiles 

sdk: docker-proto docker-build-mock-sdk-server

docker-build-mock-sdk-server: packr
	rm -rf _tmp
	mkdir -p _tmp
	CGO_ENABLED=0 GOOS=linux go build \
				-a -ldflags '-extldflags "-static"' \
				-tags "$(TAGS)" \
				-o ./_tmp/osd \
				./cmd/osd
	docker build -t $(IMAGE_MOCKSDKSERVER) -f Dockerfile.sdk .
	rm -rf _tmp

docker-build-osd-dev-base:
	docker build -t quay.io/openstorage/osd-dev-base -f Dockerfile.osd-dev-base .

push-mock-sdk-server: docker-build-mock-sdk-server
	docker push $(IMAGE_MOCKSDKSERVER)

docker-build-osd-dev:
	# This image is local only and will not be pushed
	docker build -t openstorage/osd-dev -f Dockerfile.osd-dev .

docker-build: docker-build-osd-dev
	docker run \
		--privileged \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		openstorage/osd-dev \
			make build

docker-test: docker-build-osd-dev
	docker run \
		--privileged \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /mnt:/mnt \
		-e AWS_REGION \
		-e AWS_ZONE \
		-e AWS_INSTANCE_NAME \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e GOOGLE_APPLICATION_CREDENTIALS \
		-e GCE_INSTANCE_NAME \
		-e GCE_INSTANCE_ZONE \
		-e GCE_INSTANCE_PROJECT \
		-e AZURE_INSTANCE_NAME \
		-e AZURE_SUBSCRIPTION_ID \
		-e AZURE_RESOURCE_GROUP_NAME \
		-e AZURE_ENVIRONMENT \
		-e AZURE_TENANT_ID \
		-e AZURE_CLIENT_ID \
		-e AZURE_CLIENT_SECRET \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		-e "TESTFLAGS=$(TESTFLAGS)" \
		-e "GO111MODULE=auto" \
		openstorage/osd-dev \
			make test

docker-build-osd-internal:
	rm -rf _tmp
	mkdir -p _tmp
	go build -a -tags "$(TAGS)" -o _tmp/osd cmd/osd/main.go
	docker build -t quay.io/openstorage/osd -f Dockerfile.osd .

docker-build-osd: docker-build-osd-dev
	docker run \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		openstorage/osd-dev \
			make docker-build-osd-internal

launch-sdk-quick:
	@-docker stop sdk > /dev/null 2>&1
	docker run --rm --name sdk \
		-d -p 9110:9110 -p 9100:9100 \
		$(IMAGE_MOCKSDKSERVER)

launch-sdk: sdk launch-sdk-quick

launch: docker-build-osd
	docker run \
		--privileged \
		-d \
		-v $(shell pwd)/etc:/etc \
		-v /run/docker/plugins:/run/docker/plugins \
		-v /var/lib/osd/:/var/lib/osd/ \
		-p 9005:9005 \
		-p 9100:9100 \
		-p 9110:9110 \
		quay.io/openstorage/osd -d -f /etc/config/config.yaml

# must set HAVE_BTRFS
launch-local-btrfs: install
	sudo bash -x etc/btrfs/init.sh
	sudo $(shell which osd) -d -f etc/config/config_btrfs.yaml

install-flexvolume:
	go install -a -tags "$(TAGS)" github.com/libopenstorage/openstorage/pkg/flexvolume github.com/libopenstorage/openstorage/pkg/flexvolume/cmd/flexvolume

install-flexvolume-plugin: install-flexvolume
	sudo rm -rf /usr/libexec/kubernetes/kubelet/volume/exec-plugins/openstorage~openstorage
	sudo mkdir -p /usr/libexec/kubernetes/kubelet/volume/exec-plugins/openstorage~openstorage
	sudo chmod 777 /usr/libexec/kubernetes/kubelet/volume/exec-plugins/openstorage~openstorage
	cp $(GOPATH)/bin/flexvolume /usr/libexec/kubernetes/kubelet/volume/exec-plugins/openstorage~openstorage/openstorage

clean: $(OSDSANITY)-clean
	go clean -i $(PKGS)
	packr2 clean

# Generate test-coverage HTML report
# - note: the 'go test -coverprofile...' does append results, so we're merging individual pkgs in for-loop
coverage: packr $(GOPATH)/bin/cover $(GOPATH)/bin/gotestcover
	gotestcover -coverprofile=coverage.out $(PKGS)
	go tool cover -html=coverage.out -o coverage.html
	@echo "INFO: Summary of coverage"
	go tool cover -func=coverage.out
	@cp coverage.out coverage.html /mnt/ && \
	echo "INFO: libopenstorage coverage saved at /mnt/coverage.{html,out}"

docker-coverage: docker-build-osd-dev
	docker run \
		--privileged \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /mnt:/mnt \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		-e "TESTFLAGS=$(TESTFLAGS)" \
		openstorage/osd-dev \
			make coverage

docker-images: docker-build-proto docker-build-osd-dev-base
push-docker-images: docker-images
	docker push quay.io/openstorage/osd-dev-base
	docker push quay.io/openstorage/osd-proto

# This needs to be adjusted for each release branch according
# to the SDK Version.
# For master (until released), major should be 0 and patch should be 0.
# For release branches, major and minor should be frozen.
sdk-check-version:
	go run tools/sdkver/sdkver.go --check-major=0 --check-patch=0

mockgen:
	GO111MODULE=off go get github.com/golang/mock/gomock
	GO111MODULE=off go get github.com/golang/mock/mockgen
	mockgen -destination=api/mock/mock_storagepool.go -package=mock github.com/libopenstorage/openstorage/api OpenStoragePoolServer,OpenStoragePoolClient
	mockgen -destination=api/mock/mock_cluster.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageClusterServer,OpenStorageClusterClient
	mockgen -destination=api/mock/mock_node.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageNodeServer,OpenStorageNodeClient
	mockgen -destination=api/mock/mock_diags.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageDiagsServer,OpenStorageDiagsClient
	mockgen -destination=api/mock/mock_volume.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageVolumeServer,OpenStorageVolumeClient
	mockgen -destination=api/mock/mock_watch.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageWatchServer,OpenStorageWatchClient,OpenStorageWatch_WatchClient,OpenStorageWatch_WatchServer
	mockgen -destination=api/mock/mock_bucket.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageBucketServer,OpenStorageBucketClient
	mockgen -destination=api/mock/mock_cloud_backup.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageCloudBackupServer,OpenStorageCloudBackupClient
	mockgen -destination=cluster/mock/cluster.mock.go -package=mock github.com/libopenstorage/openstorage/cluster Cluster
	mockgen -destination=api/mock/mock_fstrim.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageFilesystemTrimServer,OpenStorageFilesystemTrimClient
	mockgen -destination=api/mock/mock_fscheck.go -package=mock github.com/libopenstorage/openstorage/api OpenStorageFilesystemCheckServer,OpenStorageFilesystemCheckClient
	mockgen -destination=api/server/mock/mock_schedops_k8s.go -package=mock github.com/portworx/sched-ops/k8s/core Ops
	mockgen -destination=volume/drivers/mock/driver.mock.go -package=mock github.com/libopenstorage/openstorage/volume VolumeDriver
	mockgen -destination=bucket/drivers/mock/bucket_driver.mock.go -package=mock github.com/libopenstorage/openstorage/bucket BucketDriver
	mockgen -destination=pkg/loadbalancer/mock/balancer.go -package=mock github.com/libopenstorage/openstorage/pkg/loadbalancer Balancer


osd-tests: install
	./hack/csi-sanity-test.sh
	./hack/docker-integration-test.sh

e2e: docker-build-osd
	cd test && ./run.bash

verify: vet sdk-check-version docker-test e2e

pr-verify: vet fmt sdk-check-version
	git-validation -run DCO,short-subject
	make docker-proto
	git diff $(find . -name "*.pb.*go" -o -name "api.swagger.json" | grep -v vendor) | wc -l | grep "^0"
	hack/check-api-version.sh
	git grep -rw GPL vendor | grep LICENSE | egrep -v "yaml.v2" | wc -l | grep "^0"
	hack/check-registered-rest.sh
	
pr-test: osd-tests docker-test e2e