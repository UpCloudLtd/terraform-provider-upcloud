TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website

PACKAGE=$(shell env GO111MODULE=on go list)
GIT_VERSION?=$(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
VERSION=$(shell echo $(GIT_VERSION) | sed 's/^v//' | sed 's/-.*//')

PROVIDER_HOSTNAME=registry.upcloud.com
PROVIDER_NAMESPACE=upcloud
PROVIDER_TYPE=upcloud
PROVIDER_TARGET=$(shell go env GOOS)_$(shell go env GOARCH)
PROVIDER_PATH=~/.terraform.d/plugins/$(PROVIDER_HOSTNAME)/$(PROVIDER_NAMESPACE)/$(PROVIDER_TYPE)/$(VERSION)/$(PROVIDER_TARGET)

TOOLS_DIR:=$(CURDIR)/.ci/bin

default: build

build: fmtcheck
	@mkdir -p $(PROVIDER_PATH)
	go build \
		-tags release \
		-ldflags '-X $(PACKAGE)/internal/config.Version=$(GIT_VERSION)' \
		-o $(PROVIDER_PATH)/terraform-provider-$(PROVIDER_NAMESPACE)_v$(VERSION)

generate:
	go generate ./...

test: fmtcheck
	@echo "Running unit tests (excluding acceptance tests in upcloud/)"
	TF_ACC=0 go test $$(go list ./... | grep -v 'vendor' | grep -vE '/upcloud($|/)') -v -timeout=240m -parallel=4 -count=1

testacc: fmtcheck
	@echo "Running acceptance tests in ./upcloud only..."
	TF_ACC=1 go test ./upcloud -v -parallel=8 -timeout=240m

testacc-kubernetes:
	@echo "Running acceptance tests in ./upcloud/kubernetes only..."
	TF_ACC=1 go test ./upcloud/kubernetes -v -parallel=4 -count=1 -timeout=150m

testacc-database:
	@echo "Running acceptance tests in ./upcloud/database only..."
	TF_ACC=1 go test ./upcloud/database -v -parallel=4 -count=1 -timeout=150m

testacc-objectstorage:
	@echo "Running acceptance tests in ./upcloud/objectstorage only..."
	TF_ACC=1 go test ./upcloud/objectstorage -v -parallel=4 -count=1 -timeout=150m

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

install-tools:
	cd .ci/tools && GOBIN=$(TOOLS_DIR) go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

docs: install-tools
	terraform fmt -recursive examples/
	$(TOOLS_DIR)/tfplugindocs
	./scripts/group-docs.sh

release-notes: CHANGELOG_HEADER = ^\#\# \[
release-notes: CHANGELOG_VERSION = $(subst v,,$(VERSION))
release-notes:
	@awk \
		'/${CHANGELOG_HEADER}${CHANGELOG_VERSION}/ { flag = 1; next } \
		/${CHANGELOG_HEADER}/ { if ( flag ) { exit; } } \
		flag { if ( n ) { print prev; } n++; prev = $$0 }' \
		CHANGELOG.md

.PHONY: build generate test testacc vet fmt fmtcheck errcheck test-compile update-deps website website-test build_0_12 install-tools docs release-notes
