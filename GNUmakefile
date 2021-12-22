TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website

MODULE   = $(shell env GO111MODULE=on go list -m)
GIT_VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
VERSION = $(shell echo $(GIT_VERSION) | sed 's/^v//' | sed 's/-.*//')

PROVIDER_HOSTNAME=registry.upcloud.com
PROVIDER_NAMESPACE=upcloud
PROVIDER_TYPE=upcloud
PROVIDER_TARGET=$(shell go env GOOS)_$(shell go env GOARCH)
PROVIDER_PATH=~/.terraform.d/plugins/$(PROVIDER_HOSTNAME)/$(PROVIDER_NAMESPACE)/$(PROVIDER_TYPE)/$(VERSION)/$(PROVIDER_TARGET)

default: build

build: fmtcheck
	@mkdir -p $(PROVIDER_PATH)
	go build \
		-tags release \
		-ldflags '-X $(MODULE)/internal/config.Version=$(GIT_VERSION)' \
		-o $(PROVIDER_PATH)/terraform-provider-$(PROVIDER_NAMESPACE)_v$(VERSION)

build_0_12: fmtcheck
	go install

test: fmtcheck
	go test $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=300s -parallel=4 -count=1

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -parallel=8 -timeout 240m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -ge 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./aws"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

update-deps:
	go mod vendor

docs:
	tfplugindocs
.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile update-deps website website-test build_0_13 docs
