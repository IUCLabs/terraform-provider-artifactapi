.PHONY: build install test lint fmt clean tidy

BINARY := terraform-provider-artifactapi
VERSION ?= 0.0.1
OS_ARCH := linux_amd64
INSTALL_DIR := ~/.terraform.d/plugins/git.unkin.net/unkin/artifactapi/$(VERSION)/$(OS_ARCH)

GO_VERSION_REQUIRED := 1.23
GO_VERSION_ACTUAL := $(shell go version | sed 's/go version go\([0-9]*\.[0-9]*\).*/\1/')

check-go:
	@if [ "$$(printf '%s\n%s' "$(GO_VERSION_REQUIRED)" "$(GO_VERSION_ACTUAL)" | sort -V | head -1)" != "$(GO_VERSION_REQUIRED)" ]; then \
		echo "ERROR: Go >= $(GO_VERSION_REQUIRED) required, found $(GO_VERSION_ACTUAL)"; exit 1; \
	fi

build: check-go tidy
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/

test: check-go
	go test -race -count=1 ./...

lint: check-go
	go vet ./...

fmt: check-go
	gofmt -w .

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy
