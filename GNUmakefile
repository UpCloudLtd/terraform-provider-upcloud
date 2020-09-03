TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website

PROVIDER_REGISTRY=upcloud.com
PROVIDER_NAME=upcloud
PROVIDER_VERSION=0.1.0
PROVIDER_PATH=~/.terraform.d/plugins/$(PROVIDER_REGISTRY)/$(PROVIDER_NAME)/$(PROVIDER_NAME)/$(PROVIDER_VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)

default: build

build: fmtcheck
	go install

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 240m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
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

.PHONY: build test testacc vet fmt fmtcheck errcheck test-compile update-deps website website-test build_0_13



website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), getting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	ln -s ../../../ext/providers/$(PROVIDER_NAME)/website/$(PROVIDER_NAME).erb $(GOPATH)/src/$(WEBSITE_REPO)/content/source/layouts/$(PROVIDER_NAME).erb || true
	ln -s ../../../../ext/providers/$(PROVIDER_NAME)/website/docs $(GOPATH)/src/$(WEBSITE_REPO)/content/source/docs/providers/$(PROVIDER_NAME) || true
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PROVIDER_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), getting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	ln -s ../../../ext/providers/$(PROVIDER_NAME)/website/$(PROVIDER_NAME).erb $(GOPATH)/src/$(WEBSITE_REPO)/content/source/layouts/$(PROVIDER_NAME).erb || true
	ln -s ../../../../ext/providers/$(PROVIDER_NAME)/website/docs $(GOPATH)/src/$(WEBSITE_REPO)/content/source/docs/providers/$(PROVIDER_NAME) || true
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PROVIDER_NAME)

build_0_13: fmtcheck
	@mkdir -p $(PROVIDER_PATH)
	go build -o $(PROVIDER_PATH)/terraform-provider-$(PROVIDER_NAME)_v$(PROVIDER_VERSION)