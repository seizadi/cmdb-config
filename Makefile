PROJECT_ROOT    := github.com/seizadi/template-engine
BUILD_PATH      := bin
DOCKERFILE_PATH := $(CURDIR)/docker

# configuration for image names
USERNAME       := $(USER)
GIT_COMMIT     := $(shell git describe --dirty=-unsupported --always || echo pre-commit)
IMAGE_VERSION  ?= $(USERNAME)-dev-$(GIT_COMMIT)
IMAGE_REGISTRY ?= soheileizadi

# configuration for server binary and image
SERVER_BINARY     := $(BUILD_PATH)/server
SERVER_PATH       := $(PROJECT_ROOT)/cmd/server
SERVER_IMAGE      := $(IMAGE_REGISTRY)/template-engine
SERVER_DOCKERFILE := $(DOCKERFILE_PATH)/Dockerfile
# Placeholder. modify as defined conventions.
DB_VERSION        := 3
SRV_VERSION       := $(shell git describe --tags)
API_VERSION       := v1

# configuration for the protobuf gentool
SRCROOT_ON_HOST      := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
SRCROOT_IN_CONTAINER := /go/src/$(PROJECT_ROOT)
DOCKER_RUNNER        := docker run --rm
DOCKER_RUNNER        += -v $(SRCROOT_ON_HOST):$(SRCROOT_IN_CONTAINER)
DOCKER_GENERATOR     := infoblox/atlas-gentool:latest
GENERATOR            := $(DOCKER_RUNNER) $(DOCKER_GENERATOR)

# configuration for the database
DATABASE_HOST ?= localhost:5432

# configuration for building on host machine
GO_CACHE       := -pkgdir $(BUILD_PATH)/go-cache
GO_BUILD_FLAGS ?= $(GO_CACHE) -i -v
GO_TEST_FLAGS  ?= -v -cover
GO_PACKAGES    := $(shell go list ./... | grep -v vendor)

.PHONY: fmt
fmt:
	@go fmt $(GO_PACKAGES)

.PHONY: test
test: fmt
	@go test $(GO_TEST_FLAGS) $(GO_PACKAGES)

.PHONY: vendor
vendor:
	@dep ensure -vendor-only

.PHONY: vendor-update
vendor-update:
	@dep ensure

.PHONY: build
build:
	@go build -o bin/template-engine ./cmd/template-engine/*.go

.PHONY: protobuf
protobuf:
	@mkdir -p pkg/pb
	@curl https://raw.githubusercontent.com/seizadi/cmdb/master/pkg/pb/cmdb.proto > pkg/pb/cmdb.proto
	$(GENERATOR) \
	--go_out=plugins=grpc:. \
	--grpc-gateway_out=logtostderr=true:. \
	--gorm_out="engine=postgres:." \
	--swagger_out="atlas_patch=true:." \
	--atlas-query-validate_out=. \
	--atlas-validate_out="." \
	--validate_out="lang=go:." 	$(PROJECT_ROOT)/pkg/pb/cmdb.proto