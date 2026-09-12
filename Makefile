BINARY := harness
IMAGE := ai-harness
PLATFORMS := linux/amd64,linux/arm64

.PHONY: build test vet fmt-check docker-build docker-buildx clean

build:
	CGO_ENABLED=0 go build -trimpath -o bin/$(BINARY) ./cmd/harness

test:
	go test ./...

vet:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "gofmt needs to be run on:"; gofmt -l .; exit 1)

# Single-platform image for the local host, loaded into the local docker daemon.
docker-build:
	docker build -t $(IMAGE):local .

# Multi-arch image (linux/amd64 + linux/arm64). Build only — does not push.
# Requires a buildx builder that supports multiple platforms. buildx can't
# `--load` a multi-platform result into the local daemon; the build still
# verifies both the Go binary and the image for each platform.
docker-buildx:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE):latest .

clean:
	rm -rf bin
