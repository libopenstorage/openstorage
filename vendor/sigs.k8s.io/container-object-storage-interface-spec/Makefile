all: build

########################################################################
##                             GOLANG                                 ##
########################################################################

# If GOPATH isn't defined then set its default location.
ifeq (,$(strip $(GOPATH)))
GOPATH := $(HOME)/go
else
# If GOPATH is already set then update GOPATH to be its own
# first element.
GOPATH := $(word 1,$(subst :, ,$(GOPATH)))
endif
export GOPATH


########################################################################
##                             PROTOC                                 ##
########################################################################

# Only set PROTOC_VER if it has an empty value.
ifeq (,$(strip $(PROTOC_VER)))
PROTOC_VER := 3.9.1
endif

PROTOC_OS := $(shell uname -s)
ifeq (Darwin,$(PROTOC_OS))
PROTOC_OS := osx
endif

PROTOC_ARCH := $(shell uname -m)
ifeq (i386,$(PROTOC_ARCH))
PROTOC_ARCH := x86_32
endif

PROTOC := ./protoc
PROTOC_ZIP := protoc-$(PROTOC_VER)-$(PROTOC_OS)-$(PROTOC_ARCH).zip
PROTOC_URL := https://github.com/google/protobuf/releases/download/v$(PROTOC_VER)/$(PROTOC_ZIP)
PROTOC_TMP_DIR := .protoc
PROTOC_TMP_BIN := $(PROTOC_TMP_DIR)/bin/protoc

$(PROTOC):
	-mkdir -p "$(PROTOC_TMP_DIR)" && \
	  curl -L $(PROTOC_URL) -o "$(PROTOC_TMP_DIR)/$(PROTOC_ZIP)" && \
	  unzip "$(PROTOC_TMP_DIR)/$(PROTOC_ZIP)" -d "$(PROTOC_TMP_DIR)" && \
	  chmod 0755 "$(PROTOC_TMP_BIN)" && \
	  cp -f "$(PROTOC_TMP_BIN)" "$@"
	stat "$@" > /dev/null 2>&1


########################################################################
##                          PROTOC-GEN-GO                             ##
########################################################################

# This is the recipe for getting and installing the go plug-in
# for protoc
PROTOC_GEN_GO_PKG := github.com/golang/protobuf/protoc-gen-go
PROTOC_GEN_GO := protoc-gen-go
$(PROTOC_GEN_GO): PROTOBUF_PKG := $(dir $(PROTOC_GEN_GO_PKG))
$(PROTOC_GEN_GO): PROTOBUF_VERSION := v1.3.2
$(PROTOC_GEN_GO):
	mkdir -p $(dir $(GOPATH)/src/$(PROTOBUF_PKG))
	test -d $(GOPATH)/src/$(PROTOBUF_PKG)/.git || git clone https://$(PROTOBUF_PKG) $(GOPATH)/src/$(PROTOBUF_PKG)
	(cd $(GOPATH)/src/$(PROTOBUF_PKG) && \
		(test "$$(git describe --tags | head -1)" = "$(PROTOBUF_VERSION)" || \
			(git fetch && git checkout tags/$(PROTOBUF_VERSION))))
	(cd $(GOPATH)/src/$(PROTOBUF_PKG) && go get -v -d $$(go list -f '{{ .ImportPath }}' ./...)) && \
	go build -o "$@" $(PROTOC_GEN_GO_PKG)


########################################################################
##                          PROTOC-GEN-GO-JSON                        ##
########################################################################

# This is the recipe for getting and installing the json plug-in
# for protoc-gen-go
PROTOC_GEN_GO_JSON_PKG := github.com/mitchellh/protoc-gen-go-json
PROTOC_GEN_GO_JSON := protoc-gen-go-json
$(PROTOC_GEN_GO_JSON): PROTOC_GEN_GO_JSON_VERSION := v1.0.0
$(PROTOC_GEN_GO_JSON):
	mkdir -p $(dir $(GOPATH)/src/$(PROTOC_GEN_GO_JSON_PKG))
	test -d $(GOPATH)/src/$(PROTOC_GEN_GO_JSON_PKG)/.git || git clone https://$(PROTOC_GEN_GO_JSON_PKG) $(GOPATH)/src/$(PROTOC_GEN_GO_JSON_PKG)
	(cd $(GOPATH)/src/$(PROTOC_GEN_GO_JSON_PKG) && \
		(test "$$(git describe --tags | head -1)" = "$(PROTOC_GEN_GO_JSON_VERSION)" || \
			(git fetch && git checkout tags/$(PROTOC_GEN_GO_JSON_VERSION))))
	(cd $(GOPATH)/src/$(PROTOC_GEN_GO_JSON_PKG) && go get -v -d $$(go list -f '{{ .ImportPath }}' ./...)) && \
	go build -o "$@" $(PROTOC_GEN_GO_JSON_PKG)


########################################################################
##                          GEN-PROTO-GO                              ##
########################################################################

# This is the recipe for getting and installing the gen-proto pkg
# This is a dependency of grpc-go and must be installed before
# installing grpc-go.
GENPROTO_GO_SRC := github.com/googleapis/go-genproto
GENPROTO_GO_PKG := google.golang.org/genproto
GENPROTO_BUILD_GO := genproto-build-go
$(GENPROTO_BUILD_GO): GENPROTO_VERSION := 24fa4b261c55da65468f2abfdae2b024eef27dfb
$(GENPROTO_BUILD_GO):
	mkdir -p $(dir $(GOPATH)/src/$(GENPROTO_GO_PKG))
	test -d $(GOPATH)/src/$(GENPROTO_GO_PKG)/.git || git clone https://$(GENPROTO_GO_SRC) $(GOPATH)/src/$(GENPROTO_GO_PKG)
	(cd $(GOPATH)/src/$(GENPROTO_GO_PKG) && \
			(git fetch && git checkout $(GENPROTO_VERSION)))
	(cd $(GOPATH)/src/$(GENPROTO_GO_PKG) && go get -v -d $$(go list -f '{{ .ImportPath }}' ./...))


########################################################################
##                          GRPC-GO                                   ##
########################################################################

# This is the recipe for getting and installing the grpc go
GRPC_GO_SRC := github.com/grpc/grpc-go
GRPC_GO_PKG := google.golang.org/grpc
GRPC_BUILD_GO := grpc-build-go
$(GRPC_BUILD_GO): GRPC_VERSION := v1.26.0
$(GRPC_BUILD_GO):
	mkdir -p $(dir $(GOPATH)/src/$(GRPC_GO_PKG))
	test -d $(GOPATH)/src/$(GRPC_GO_PKG)/.git || git clone https://$(GRPC_GO_SRC) $(GOPATH)/src/$(GRPC_GO_PKG)
	(cd $(GOPATH)/src/$(GRPC_GO_PKG) && \
		(test "$$(git describe --tags | head -1)" = "$(GRPC_VERSION)" || \
			(git fetch && git checkout tags/$(GRPC_VERSION))))
	(cd $(GOPATH)/src/$(GRPC_GO_PKG) && go get -v -d $$(go list -f '{{ .ImportPath }}' ./...) && \
		go build -o "$@" $(GRPC_GO_PKG))


########################################################################
##                          PROTOC-GEN-GO-FAKE                        ##
########################################################################

# This is the recipe for getting and installing the grpc go
PROTOC_GEN_GO_FAKE_SRC := ./hack/fake-gen
PROTOC_GEN_GO_FAKE := protoc-gen-gofake
$(PROTOC_GEN_GO_FAKE):
	go build -o $(PROTOC_GEN_GO_FAKE) $(PROTOC_GEN_GO_FAKE_SRC)


########################################################################
##                              PATH                                  ##
########################################################################

# Update PATH with the current directory. This enables the protoc
# binary to discover the protoc-gen-go binary, built inside this
# directory.
export PATH := $(shell pwd):$(PATH)


########################################################################
##                              BUILD                                 ##
########################################################################
COSI_PROTO := ./cosi.proto
COSI_SPEC := spec.md
COSI_PKG_ROOT := sigs.k8s.io/container-object-storage-interface-spec
COSI_PKG_SUB := .
COSI_BUILD := $(COSI_PKG_SUB)/.build
COSI_GO := $(COSI_PKG_SUB)/cosi.pb.go
COSI_GO_JSON := $(COSI_PKG_SUB)/cosi.pb.json.go
COSI_GO_FAKE := $(COSI_PKG_SUB)/fake/cosi.pb.fake.go
COSI_A := cosi.a
COSI_GO_TMP := $(COSI_BUILD)/$(COSI_PKG_ROOT)/cosi.pb.go
COSI_GO_JSON_TMP := $(COSI_BUILD)/$(COSI_PKG_ROOT)/cosi.pb.json.go
COSI_GO_FAKE_TMP := $(COSI_BUILD)/fake/$(COSI_PKG_ROOT)/cosi.pb.fake.go

# This recipe generates the go language bindings to a temp area.
$(COSI_GO_TMP): HERE := $(shell pwd)
$(COSI_GO_TMP): PTYPES_PKG := github.com/golang/protobuf/ptypes
$(COSI_GO_TMP): GO_OUT := plugins=grpc
$(COSI_GO_TMP): GO_OUT := $(GO_OUT),Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor
$(COSI_GO_TMP): GO_OUT := $(GO_OUT),Mgoogle/protobuf/wrappers.proto=$(PTYPES_PKG)/wrappers
$(COSI_GO_TMP): GO_OUT := $(GO_OUT):"$(HERE)/$(COSI_BUILD)"
$(COSI_GO_TMP): GO_JSON_OUT := emit_defaults
$(COSI_GO_TMP): GO_JSON_OUT := $(GO_JSON_OUT):"$(HERE)/$(COSI_BUILD)"
$(COSI_GO_TMP): GO_FAKE_OUT := emit_defaults
$(COSI_GO_TMP): GO_FAKE_OUT := $(GO_FAKE_OUT),packagePath=sigs.k8s.io/container-object-storage-interface-spec
$(COSI_GO_TMP): GO_FAKE_OUT := $(GO_FAKE_OUT):"$(HERE)/$(COSI_BUILD)"/fake
$(COSI_GO_TMP): INCLUDE := -I$(GOPATH)/src -I$(HERE)/$(PROTOC_TMP_DIR)/include
$(COSI_GO_TMP): $(COSI_PROTO) | $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_JSON) $(PROTOC_GEN_GO_FAKE)
	@mkdir -p "$(@D)"
	@mkdir -p "$(COSI_BUILD)/fake"
	(cd "$(GOPATH)/src" && \
		$(HERE)/$(PROTOC) $(INCLUDE) --go_out=$(GO_OUT) --go-json_out=$(GO_JSON_OUT) --gofake_out=$(GO_FAKE_OUT) "$(COSI_PKG_ROOT)/$(<F)")

# The temp language bindings are compared to the ones that are
# versioned. If they are different then it means the language
# bindings were not updated prior to being committed.
$(COSI_GO): $(COSI_GO_TMP)
ifeq (true,$(TRAVIS))
	diff "$@" "$?"
else
	@mkdir -p "$(@D)"
	diff "$@" "$?" > /dev/null 2>&1 || cp -f "$?" "$@"
endif

# The temp language bindings are compared to the ones that are
# versioned. If they are different then it means the language
# bindings were not updated prior to being committed.
$(COSI_GO_JSON): $(COSI_GO_JSON_TMP)
ifeq (true,$(TRAVIS))
	diff "$@" "$?"
else
	@mkdir -p "$(@D)"
	diff "$@" "$?" > /dev/null 2>&1 || cp -f "$?" "$@"
endif

# The temp language bindings are compared to the ones that are
# versioned. If they are different then it means the language
# bindings were not updated prior to being committed.
$(COSI_GO_FAKE): $(COSI_GO_FAKE_TMP)
ifeq (true,$(TRAVIS))
	diff "$@" "$?"
else
	@mkdir -p "$(@D)"
	diff "$@" "$?" > /dev/null 2>&1 || cp -f "$?" "$@"
endif

# This recipe builds the Go archive from the sources in three steps:
#
#   1. Go get any missing dependencies.
#   2. Cache the packages.
#   3. Build the archive file.
$(COSI_A): $(COSI_GO) $(COSI_GO_JSON) $(COSI_GO_FAKE) $(GENPROTO_BUILD_GO) $(GRPC_BUILD_GO)
	go get -v -d ./...
	go install ./$(COSI_PKG_SUB)
	go build -o "$@" ./$(COSI_PKG_SUB)

generate:
	echo "// Code generated by make; DO NOT EDIT." > "$(COSI_PROTO)"
	cat $(COSI_SPEC) | sed -n -e '/```protobuf$$/,/^```$$/ p' | sed '/^```/d' >> "$(COSI_PROTO)"

build: generate $(COSI_A)

clean:
	go clean -i ./...
	rm -rf "$(COSI_PROTO)" "$(COSI_A)" "$(COSI_GO)" "$(COSI_GO_JSON)" "$(COSI_BUILD)"

clobber: clean
	rm -fr "$(PROTOC)" "$(PROTOC_TMP_DIR)" "$(PROTOC_GEN_GO)" "$(PROTOC_GEN_GO_JSON)" "$(PROTOC_GEN_GO_FAKE)"

.PHONY: clean clobber $(GRPC_BUILD_GO) $(GENPROTO_BUILD_GO)
