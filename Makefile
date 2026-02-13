BINARY_NAME=readwise-mcp-server
BUILD_DIR=build
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')
GOFLAGS=-trimpath
LDFLAGS=-s -w

.PHONY: build test lint clean container deploy

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/readwise-mcp/

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -rf $(BUILD_DIR)
	go clean -testcache

container:
	podman build -t $(BINARY_NAME) -f Containerfile .

deploy:
	kustomize build deploy/ | kubectl --kubeconfig ~/.kube/home-config apply -f -
