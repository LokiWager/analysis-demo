SHELL:=/bin/sh
.PHONY: build\
		test fmt vet clean \
		mod_update

export GO111MODULE=on

# Path Related
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
RELEASE_DIR := ${MKFILE_DIR}bin
GO_PATH := $(shell go env | grep GOPATH | awk -F '"' '{print $$2}')
INTEGRATION_TEST_PATH := build/test

VERSION_TAG=$(shell git describe --match 'v[0-9]*' --dirty='.m' --always --tags)
GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M)

TARGET_BUILDER=${RELEASE_DIR}/analysis

default: build

# Build binary to ./
build:
	@echo "build"
	cd ${MKFILE_DIR} && \
	CGO_ENABLED=0 go build -ldflags '-X main.versionTag=${VERSION_TAG} -X main.versionGitCommit=${GIT_COMMIT} -X main.versionBuildTime=${BUILD_TIME}' -gcflags=all="-N -l" -o -o ${TARGET_BUILDER} ${MKFILE_DIR}cmd/analysis

test:
	cd ${MKFILE_DIR}
	go mod tidy
	git diff --exit-code go.mod go.sum
	go mod verify
	go test -v -gcflags "all=-l" ${MKFILE_DIR}pkg/... ${MKFILE_DIR}cmd/... ${TEST_FLAGS}


clean:
	rm -rf ${RELEASE_DIR}
	rm -rf ${MKFILE_DIR}build/cache
	rm -rf ${MKFILE_DIR}build/bin


fmt:
	cd ${MKFILE_DIR} && go fmt ./...

vet:
	cd ${MKFILE_DIR} && go vet ./...

mod_update:
	cd ${MKFILE_DIR} && go get -u
