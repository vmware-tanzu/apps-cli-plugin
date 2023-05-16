PLUGIN_BUILD_SHA_SHORT := $(shell git rev-parse --short HEAD)
PLUGIN_BUILD_VERSION ?= $(shell cat APPS_PLUGIN_VERSION)-dev-$(PLUGIN_BUILD_SHA_SHORT)
PLUGIN_BUILD_DIRTY = $(shell git diff --quiet HEAD || echo "-dirty")
PLUGIN_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
PLUGIN_BUILD_SHA = $(shell git rev-parse HEAD)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

ifeq ($(strip $(PLUGIN_BUILD_VERSION)),)
PLUGIN_BUILD_VERSION = v0.0.0
endif
PLUGIN_LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo.Date=$(PLUGIN_BUILD_DATE)'
PLUGIN_LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo.SHA=$(PLUGIN_BUILD_SHA)'
PLUGIN_LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo.Version=$(PLUGIN_BUILD_VERSION)'

GO_SOURCES = $(shell find ./cmd ./pkg -type f -name '*.go')

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTROLLER_GEN ?= go run -mod=mod -modfile hack/go.mod sigs.k8s.io/controller-tools/cmd/controller-gen
DIEGEN ?= go run -modfile hack/go.mod -mod=mod dies.dev/diegen
GOIMPORTS ?= go run -mod=mod -modfile hack/go.mod golang.org/x/tools/cmd/goimports
WOKE ?= go run -mod=mod -modfile hack/go.mod github.com/get-woke/woke

# Add supported OS-ARCHITECTURE combinations here
PLUGIN_BUILD_OS_ARCH ?= linux-amd64 windows-amd64 darwin-amd64 # linux-arm64 darwin-arm64
PLUGIN_BUILD_TARGETS := $(addprefix plugin-build-,${PLUGIN_BUILD_OS_ARCH})
PLUGIN_BUNDLE_TARGETS := $(addprefix generate-plugin-bundle-,${PLUGIN_BUILD_OS_ARCH})

# Paths and Directory information
ROOT_DIR := $(shell git rev-parse --show-toplevel)

PLUGIN_DIR := ./cmd/plugin
PLUGIN_BINARY_ARTIFACTS_DIR := $(ROOT_DIR)/artifacts/plugins

PLUGIN_NAME ?= apps

# Repository specific configuration
TZBIN ?= tanzu
BUILDER_PLUGIN ?= $(TZBIN) builder
PUBLISHER ?= tzcli
VENDOR ?= vmware
PLUGIN_SCOPE_ASSOCIATION_FILE ?= ""

include ./plugin-tooling.mk

.PHONY: all
all: test build ## Prepare and run the project tests

.PHONY: install
install: test## Install the plugin binaries to the local machine
	@# TODO avoid deleting an existing plugin once in place reinstalls are working again
	@tanzu plugin delete apps > /dev/null 2>&1 || true
	$(BUILDER_PLUGIN) plugin build --path $(PLUGIN_DIR) --binary-artifacts $(PLUGIN_BINARY_ARTIFACTS_DIR) --version $(PLUGIN_BUILD_VERSION) --ldflags "$(PLUGIN_LD_FLAGS)" --os-arch local
	tanzu plugin install $(PLUGIN_NAME) --version $(PLUGIN_BUILD_VERSION) --local $(PLUGIN_BINARY_ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)

.PHONY: plugin-build
plugin-build: $(PLUGIN_BUILD_TARGETS) generate-plugin-bundle ## Build all plugin binaries for all supported os-arch

plugin-build-local: plugin-build-$(GOHOSTOS)-$(GOHOSTARCH) ## Build all plugin binaries for local platform

.PHONY: plugin-build-%
plugin-build-%: ## Build the plugin binaries for the given OS-ARCHITECTURE combination
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))
	$(BUILDER_PLUGIN) plugin build \
		--path $(PLUGIN_DIR) \
		--binary-artifacts $(PLUGIN_BINARY_ARTIFACTS_DIR) \
		--version $(PLUGIN_BUILD_VERSION) \
		--ldflags "$(PLUGIN_LD_FLAGS)" \
		--os-arch $(OS)_$(ARCH) \
		--match "$(PLUGIN_NAME)" \
		--plugin-scope-association-file $(PLUGIN_SCOPE_ASSOCIATION_FILE)

docs: $(GO_SOURCES) ## Generate the plugin documentation
	@rm -rf docs/command-reference
	go run --ldflags "$(PLUGIN_LD_FLAGS)" ./cmd/plugin/apps docs -d docs/command-reference

.PHONY: test
test: prepare## Run tests
	go test ./... -coverprofile=coverage.txt -covermode=atomic -timeout 3m -race

.PHONY: lint
lint: ## Run lint tools to identfy stylistic errors
	@echo "Scanning for inclusive terminology errors"
	@$(WOKE) . -c https://via.vmw.com/its-woke-rules

.PHONY: integration-test
integration-test:  ## Run integration test
	go test -v -timeout 10m github.com/vmware-tanzu/apps-cli-plugin/testing/e2e/... --tags=integration

.PHONY: prepare
prepare: generate fmt vet

.PHONY: fmt
fmt: ## Run go fmt against code
ifneq ($(OS),Windows_NT)
	$(GOIMPORTS) --local github.com/vmware-tanzu/apps-cli-plugin -w pkg/ cmd/
endif 

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: generate
generate: generate-internal fmt ## Generate code

.PHONY: generate-internal
generate-internal:
ifneq ($(OS),Windows_NT)
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."
	@echo "diegen die:headerFile=\"hack/boilerplate.go.txt\" paths=\"./...\""
	@$(DIEGEN) die:headerFile="hack/boilerplate.go.txt" paths="./..."
endif

vendor: go.mod go.sum $(GO_SOURCES) ## Generate the vendor directory
	go mod tidy
	go mod vendor

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

## --------------------------------------
## Helpers
## --------------------------------------

generate-plugin-bundle: $(PLUGIN_BUNDLE_TARGETS)
	cd $(PLUGIN_BINARY_ARTIFACTS_DIR) && tar -czvf ../tanzu-apps-plugin.tar.gz .

generate-plugin-bundle-%:
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))
	cd $(PLUGIN_BINARY_ARTIFACTS_DIR) && tar -czvf ../tanzu-apps-plugin-${OS}-${ARCH}.tar.gz ./${OS}/${ARCH}