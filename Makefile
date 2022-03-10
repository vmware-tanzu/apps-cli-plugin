BUILD_VERSION ?= v0.0.0-dev
BUILD_DIRTY = $(shell git diff --quiet HEAD || echo "-dirty")
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
BUILD_SHA = $(shell git rev-parse HEAD)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

# Values taken from https://github.com/vmware-tanzu/tanzu-framework/blob/main/pkg/v1/cli/buildvar.go#L12
LD_FLAGS = -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Date=$(BUILD_DATE)' \
           -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.SHA=$(BUILD_SHA)$(BUILD_DIRTY)' \
           -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Version=$(BUILD_VERSION)'

GO_SOURCES = $(shell find ./cmd ./pkg -type f -name '*.go')
WORKING_DIR ?= $(shell pwd)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTROLLER_GEN ?= go run -mod=mod -modfile hack/go.mod sigs.k8s.io/controller-tools/cmd/controller-gen
GOIMPORTS ?= go run -mod=mod -modfile hack/go.mod golang.org/x/tools/cmd/goimports

ARTIFACTS_DIR ?= ./artifacts
TANZU_PLUGIN_PUBLISH_PATH ?= $(ARTIFACTS_DIR)/published


# Add supported OS-ARCHITECTURE combinations here
ENVS ?= linux-amd64 windows-amd64 darwin-amd64 # linux-arm64 darwin-arm64

BUILD_JOBS := $(addprefix build-,${ENVS})
PUBLISH_JOBS := $(addprefix publish-,${ENVS})

.PHONY: all
all: build

.PHONY: install
install: publish-local
	@# TODO avoid deleting an existing plugin once in place reinstalls are working again
	@tanzu plugin delete apps > /dev/null 2>&1 || true
	tanzu plugin install apps --version $(BUILD_VERSION) --local $(TANZU_PLUGIN_PUBLISH_PATH)/${GOHOSTOS}-${GOHOSTARCH}

.PHONY: build-local
build-local: prepare test
	@echo BUILD_VERSION: $(BUILD_VERSION)
	tanzu builder cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/plugin --target local --artifacts ${ARTIFACTS_DIR}/${GOHOSTOS}/${GOHOSTARCH}/cli

.PHONY: build
build: prepare test $(BUILD_JOBS)

.PHONY: build-%
build-%:
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))
	tanzu builder cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/plugin --artifacts ${ARTIFACTS_DIR}/${OS}/${ARCH}/cli --target ${OS}_${ARCH}

.PHONY: publish
publish: build $(PUBLISH_JOBS)
	tar -zcvf tanzu-apps-plugin.tar.gz -C $(TANZU_PLUGIN_PUBLISH_PATH) .

.PHONY: publish-local
publish-local: build-local
	tanzu builder publish --type local --plugins "apps" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${GOHOSTOS}-${GOHOSTARCH}/discovery/standalone" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${GOHOSTOS}-${GOHOSTARCH}/distribution" --input-artifact-dir $(ARTIFACTS_DIR)
	tar -zcvf tanzu-apps-plugin.tar.gz -C $(TANZU_PLUGIN_PUBLISH_PATH)/${GOHOSTOS}-${GOHOSTARCH} .

.PHONY: publish-%
publish-%:
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))
	tanzu builder publish --type local --plugins "apps" --version $(BUILD_VERSION) --os-arch "${OS}-${ARCH}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH}/discovery/standalone" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH}/distribution" --input-artifact-dir $(ARTIFACTS_DIR)
	tar -zcvf tanzu-apps-plugin-${OS}-${ARCH}.tar.gz -C $(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH} .

docs: $(GO_SOURCES)
	@rm -rf docs/command-reference
	go run --ldflags "$(LD_FLAGS)" ./cmd/plugin/apps docs -d docs/command-reference

.PHONY: test
test: generate fmt vet ## Run tests
	go test ./... -coverprofile=coverage.txt -covermode=atomic -timeout 30s -race

.PHONY: prepare
prepare: generate fmt vet

# Run go fmt against code
.PHONY: fmt
fmt:
	$(GOIMPORTS) --local github.com/vmware-tanzu/apps-cli-plugin -w pkg/ cmd/

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

.PHONY: generate
generate: generate-internal fmt ## Generate code

.PHONY: generate-internal
generate-internal:
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

vendor: go.mod go.sum $(GO_SOURCES)
	go mod tidy
	go mod vendor

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
