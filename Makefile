# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)
BUILDTAGS=

.PHONY: clean all fmt vet lint build test install static release
.DEFAULT: default

# VERSIONING: Z.Y.x[[-rc][+COMMIT_COUNT.COMMMIT]
GIT_DESCRIBE := $(shell git describe --match '${TAG_PREFIX}*' --tags --always --abbrev=0 --exact-match 2>/dev/null || \
					git describe --match '${TAG_PREFIX}*' --tags --always --abbrev=0)
GIT_TAG := $(shell echo ${GIT_DESCRIBE} | sed 's;^${TAG_PREFIX};;')
GIT_REVS := $(shell git rev-list ${GIT_DESCRIBE}..HEAD --count 2>/dev/null || echo 0)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell echo "${GIT_TAG}+${GIT_REVS}.git-${GIT_COMMIT}" | sed 's|\+0\..*$|||')

LDFLAGS = -ldflags="-X main.VERSION=${VERSION}"

all: clean build fmt lint test vet install

build:
	@echo "+ $@"
	@go build ${LDFLAGS} -tags "$(BUILDTAGS) cgo" .

static:
	@echo "+ $@"
	CGO_ENABLED=1 go build -tags "$(BUILDTAGS) cgo static_build" -ldflags "-w -extldflags -static" ${LDFLAGS} -o ghedgetrim .

fmt:
	@echo "+ $@"
	@gofmt -s -l . | grep -v vendor | tee /dev/stderr

lint:
	@echo "+ $@"
	@golint ./... | grep -v vendor | tee /dev/stderr

test: fmt lint vet
	@echo "+ $@"
	@go test -v ${LDFLAGS} -tags "$(BUILDTAGS) cgo" $(shell go list ./... | grep -v vendor)

vet:
	@echo "+ $@"
	@go vet $(shell go list ./... | grep -v vendor)

clean:
	@echo "+ $@"
	@rm -rf ghedgetrim build

install:
	@echo "+ $@"
	@go install .

release:
	@echo "+ $@"
	@mkdir -p build && rm -f build/SHA1SUMS
	@gox -osarch="linux/amd64 linux/arm darwin/amd64" ${LDFLAGS} -output="build/{{.Dir}}-{{.OS}}-{{.Arch}}"
	@find build -type f ! -name "SHA1SUMS" -print0 | xargs -0 shasum | sed 's/build\///g' >> build/SHA1SUMS
	@cat build/SHA1SUMS
