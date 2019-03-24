# tested with make 4.2.1

GO := /usr/bin/go
BUILDAH := sudo /usr/bin/buildah
GIT := /usr/bin/git

BINARY := ecgoserver
TARGET := $(CURDIR)/$(BINARY)

VERSION := v0.2.0
GIT_COMMIT := $(shell $(GIT) rev-parse --short HEAD)
LDFLAGS := -ldflags '-s -w -X main.versionNumber=$(VERSION) -X main.gitCommit=${GIT_COMMIT}'

APP_NAME := ecgo
CONTAINER := $(APP_NAME)-scratch-container-1
IMAGE_NAME := jostho/$(APP_NAME):$(VERSION)
IMAGE_BINARY_PATH := /bin/$(BINARY)
PORT := 8000

check:
	$(GO) version
	/usr/bin/buildah version | head -1
	$(GIT) --version

clean:
	rm -f $(TARGET)

build:
	$(GO) build $(LDFLAGS) -o $(TARGET) ecgoserver.go

build-static:
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(TARGET) ecgoserver.go

build-image:
	$(BUILDAH) from --name $(CONTAINER) scratch
	$(BUILDAH) copy $(CONTAINER) $(TARGET) $(IMAGE_BINARY_PATH)
	$(BUILDAH) config \
		--entrypoint '[ "$(IMAGE_BINARY_PATH)" ]' \
		--created-by buildah -p $(PORT) \
		-l Name=$(APP_NAME) -l Version=$(VERSION) -l Commit=$(GIT_COMMIT) \
		$(CONTAINER)
	$(BUILDAH) commit --rm $(CONTAINER) $(IMAGE_NAME)

clean-image:
	$(BUILDAH) rmi $(IMAGE_NAME)

image: clean build-static build-image

.PHONY: check clean build build-static
.PHONY: build-image clean-image image
