HOSTNAME := terraform-provider-voipms
VERSION ?= 0.0.1-dev
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
GOBIN := $(shell go env GOPATH)/bin
PLUGIN_MIRROR := $(HOME)/.terraform.d/plugins/registry.terraform.io/vetal-ca-org/voipms/$(VERSION)/$(OS_ARCH)

default: fmt test

build:
	go build -o $(HOSTNAME) -v .

install: build
	mkdir -p "$(GOBIN)"
	cp $(HOSTNAME) "$(GOBIN)/$(HOSTNAME)"

# Install a versioned binary Terraform can find from another repo without
# Terraform Registry (filesystem mirror). Pin version = "0.0.1-dev" there.
install-plugin: build
	mkdir -p "$(PLUGIN_MIRROR)"
	cp $(HOSTNAME) "$(PLUGIN_MIRROR)/$(HOSTNAME)_v$(VERSION)"
	@echo "Installed $(PLUGIN_MIRROR)/$(HOSTNAME)_v$(VERSION)"

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=4 ./internal/...

testacc:
	TF_ACC=1 go test -v -cover -timeout 30m ./internal/provider/

generate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.24.0 generate -provider-name voipms

.PHONY: build install install-plugin fmt test testacc generate
