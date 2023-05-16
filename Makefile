PLUGIN_BUILD_SHA_SHORT := $(shell git rev-parse --short HEAD)
PLUGIN_BUILD_VERSION ?= $(shell cat APPS_PLUGIN_VERSION)-dev-$(PLUGIN_BUILD_SHA_SHORT)
PLUGIN_BUILD_DIRTY = $(shell git diff --quiet HEAD || echo "-dirty")
PLUGIN_BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
PLUGIN_BUILD_SHA = $(shell git rev-parse HEAD)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

LD_FLAGS = -X 'github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo.Date=$(PLUGIN_BUILD_DATE)' \
           -X 'github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo.SHA=$(PLUGIN_BUILD_SHA)$(PLUGIN_BUILD_DIRTY)' \
           -X 'github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo.Version=$(PLUGIN_BUILD_VERSION)'

GO_SOURCES = $(shell find ./cmd ./pkg -type f -name '*.go')
WORKING_DIR ?= $(shell pwd)

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

ARTIFACTS_DIR ?= ./artifacts
TANZU_PLUGIN_PUBLISH_PATH ?= $(ARTIFACTS_DIR)/published


# Add supported OS-ARCHITECTURE combinations here
ENVS ?= linux-amd64 windows-amd64 darwin-amd64 # linux-arm64 darwin-arm64

BUILD_JOBS := $(addprefix build-,${ENVS})
PUBLISH_JOBS := $(addprefix publish-,${ENVS})

include ./plugin-tooling.mk

.PHONY: all
all: test build ## Prepare and run the project tests

.PHONY: install
install: test## Install the plugin binaries to the local machine
	@# TODO avoid deleting an existing plugin once in place reinstalls are working again
	@tanzu plugin delete apps > /dev/null 2>&1 || true
	tanzu builder cli compile --version $(PLUGIN_BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/plugin --target local --artifacts ${ARTIFACTS_DIR}/${GOHOSTOS}/${GOHOSTARCH}/cli
	tanzu builder publish --type local --plugins "apps" --version $(PLUGIN_BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${GOHOSTOS}-${GOHOSTARCH}/discovery/standalone" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${GOHOSTOS}-${GOHOSTARCH}/distribution" --input-artifact-dir $(ARTIFACTS_DIR)
	tanzu plugin install apps --version $(PLUGIN_BUILD_VERSION) --local $(TANZU_PLUGIN_PUBLISH_PATH)/${GOHOSTOS}-${GOHOSTARCH}

.PHONY: build
build: $(BUILD_JOBS)
	tar -zcvf tanzu-apps-plugin-build.tar.gz -C $(ARTIFACTS_DIR) .

.PHONY: build-%
build-%: ## Build the plugin binaries for the given OS-ARCHITECTURE combination
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))
	tanzu builder cli compile --version $(PLUGIN_BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/plugin --artifacts ${ARTIFACTS_DIR}/${OS}/${ARCH}/cli --target ${OS}_${ARCH}

.PHONY: publish
publish: $(PUBLISH_JOBS) ## Generate the distributable plugin binaries packages
	tar -zcvf tanzu-apps-plugin.tar.gz -C $(TANZU_PLUGIN_PUBLISH_PATH) .

.PHONY: publish-oci
publish-oci: 
	tanzu builder publish --input-artifact-dir $(ARTIFACTS_DIR) --plugins "apps" --version "${PLUGIN_BUILD_VERSION}" --type oci --oci-discovery-image "${DISCOVERY_REPO}" --oci-distribution-image-repository "${DISTRIBUTION_REPO}"

.PHONY: publish-%
publish-%: ## Generate the dustributable plugin binaries packages for the given OS-ARCHITECTURE combination
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))
	tanzu builder publish --type local --plugins "apps" --version $(PLUGIN_BUILD_VERSION) --os-arch "${OS}-${ARCH}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH}/discovery/standalone" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH}/distribution" --input-artifact-dir $(ARTIFACTS_DIR)
	tar -zcvf tanzu-apps-plugin-${OS}-${ARCH}.tar.gz -C $(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH} .

docs: $(GO_SOURCES) ## Generate the plugin documentation
	@rm -rf docs/command-reference
	go run --ldflags "$(LD_FLAGS)" ./cmd/plugin/apps docs -d docs/command-reference

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
