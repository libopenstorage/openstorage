HAS_PACKR := $(shell command -v packr 2> /dev/null)
HAS_PROTOC_GEN_GRPC_GATEWAY := $(shell command -v protoc-gen-grpc-gateway 2> /dev/null)
HAS_PROTOC_GEN_SWAGGER := $(shell command -v protoc-gen-swagger 2> /dev/null)
HAS_PROTOC_GEN_GO := $(shell command -v protoc-gen-go 2> /dev/null)

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

export GO15VENDOREXPERIMENT=1

all: build $(OSDSANITY)

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v $(PKGS)

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f $(PKGS)

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t $(PKGS)

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -tags "$(TAGS)" -d -v -t -u -f $(PKGS)

vendor-update:
	GO15VENDOREXPERIMENT=0 GOOS=linux GOARCH=amd64 go get -tags "daemon btrfs_noversion have_btrfs have_chainfs" -d -v -t -u -f $(PKGS)

vendor-without-update:
	go get -v github.com/kardianos/govendor
	rm -rf vendor
	govendor init
	GOOS=linux GOARCH=amd64 govendor add +external
	GOOS=linux GOARCH=amd64 govendor update +vendor
	GOOS=linux GOARCH=amd64 govendor add +external
	GOOS=linux GOARCH=amd64 govendor update +vendor

vendor: vendor-update vendor-without-update

build: packr
	go build -tags "$(TAGS)" $(BUILDFLAGS) $(PKGS)

install: packr $(OSDSANITY)-install
	go install -tags "$(TAGS)" $(PKGS)

$(OSDSANITY):
	@$(MAKE) -C cmd/osd-sanity

$(OSDSANITY)-install:
	@$(MAKE) -C cmd/osd-sanity install

$(OSDSANITY)-clean:
	@$(MAKE) -C cmd/osd-sanity clean

proto:
ifndef HAS_PROTOC_GEN_GO
	@echo "Installing protoc-gen-go"
	go get -u github.com/golang/protobuf/protoc-gen-go
endif

ifndef HAS_PROTOC_GEN_GRPC_GATEWAY
	@echo "Installing protoc-gen-grpc-gateway"
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
endif

ifndef HAS_PROTOC_GEN_SWAGGER
	@echo "Installing protoc-gen-swagger"
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
endif

	@echo "Generating protobuf definitions from api/api.proto"
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
	@echo "Generating grpc protobuf definitions from pkg/flexvolume/flexvolume.proto"
	$(PROTOC) -I/usr/local/include -I$(PROTOSRC_PATH) -I$(PROTOS_PATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. $(PROTOSRC_PATH)/pkg/flexvolume/flexvolume.proto
	$(PROTOC) -I/usr/local/include -I$(PROTOSRC_PATH) -I$(PROTOS_PATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. $(PROTOSRC_PATH)/pkg/flexvolume/flexvolume.proto
	@echo "Generating protobuf definitions from pkg/jsonpb/testing/testing.proto"
	$(PROTOC) -I $(PROTOSRC_PATH) $(PROTOSRC_PATH)/pkg/jsonpb/testing/testing.proto --go_out=plugins=grpc:.
	@echo "Generating gRPC clients"
	$(MAKE) -C api/client/sdk

lint:
	go get -v github.com/golang/lint/golint
	golint $(PKGS)

vet:
	go vet $(PKGS)

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck -tags "$(TAGS)" $(PKGS)

pretest: lint vet errcheck

test: packr
	go test -tags "$(TAGS)" $(TESTFLAGS) $(PKGS)

docs:
	go generate ./cmd/osd/main.go
	swagger validate api/swagger/swagger.json

packr:
ifndef HAS_PACKR
	@echo "Installing packr to embed websites in golang"
	go get -u github.com/gobuffalo/packr/...
endif
	packr

generate-mockfiles:
	go generate $(PKGS)

generate: docs generate-mockfiles

sdk: proto docker-build-mock-sdk-server

docker-build-mock-sdk-server: packr
	rm -rf _tmp
	mkdir -p _tmp
	CGO_ENABLED=0 GOOS=linux go build \
				-a -ldflags '-extldflags "-static"' \
				-tags "$(TAGS)" \
				-o ./_tmp/osd \
				./cmd/osd
	docker build -t openstorage/mock-sdk-server -f Dockerfile.sdk .
	rm -rf _tmp

docker-build-osd-dev:
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
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		-e "TESTFLAGS=$(TESTFLAGS)" \
		openstorage/osd-dev \
			make test

docker-build-osd-internal:
	rm -rf _tmp
	mkdir -p _tmp
	go build -a -tags "$(TAGS)" -o _tmp/osd cmd/osd/main.go
	docker build -t openstorage/osd -f Dockerfile.osd .

docker-build-osd: docker-build-osd-dev
	docker run \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		openstorage/osd-dev \
			make docker-build-osd-internal

launch-sdk: sdk
	docker run \
		-d -p 9110:9110 -p 9100:9100 \
		openstorage/mock-sdk-server

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
		openstorage/osd -d -f /etc/config/config.yaml

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
	packr clean

.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	vendor-update \
	vendor-without-update \
	vendor \
	build \
	install \
	proto \
	lint \
	vet \
	errcheck \
	pretest \
	test \
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
	generate-mockfiles

$(GOPATH)/bin/cover:
	go get golang.org/x/tools/cmd/cover

$(GOPATH)/bin/gotestcover:
	go get github.com/pierrre/gotestcover

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
