PROJECT      := github.com/devzero-inc/pulumi-provider-devzero
PROVIDER     := pulumi-resource-devzero
WORKING_DIR  := $(shell pwd)
GOPATH       := $(shell go env GOPATH)
SCHEMA_FILE  := schema.json
SDK_NODEJS   := sdk/nodejs
SDK_PYTHON   := sdk/python
SDK_GO       := sdk/go

# VERSION is injected into the provider binary via ldflags. Falls back to
# the latest git tag (stripped of a leading `v`), or "dev" when no tags exist.
GIT_TAG      := $(shell git describe --tags --abbrev=0 2>/dev/null)
VERSION      ?= $(if $(GIT_TAG),$(patsubst v%,%,$(GIT_TAG)),dev)
LDFLAGS      := -ldflags "-X main.version=$(VERSION)"

# Proto sync settings — override with: make proto SERVICES_DIR=/path/to/services
SERVICES_DIR ?= ../services

SOURCE_PROTO_DIR       = dakr/proto/api/v1
SOURCE_GEN_PB_DIR      = dakr/gen/api/v1
SOURCE_GEN_CONNECT_DIR = dakr/gen/api/v1/apiv1connect

TARGET_PROTO_DIR       = internal/proto/api/v1
TARGET_GEN_PB_DIR      = internal/gen/api/v1
TARGET_GEN_CONNECT_DIR = internal/gen/api/v1/apiv1connect

PROTO_FILES       = common.proto instance.proto k8s.proto recommendation.proto cluster.proto profiling.proto
GEN_PB_FILES      = common.pb.go instance.pb.go k8s.pb.go recommendation.pb.go cluster.pb.go cluster_grpc.pb.go profiling.pb.go
GEN_CONNECT_FILES = k8s.connect.go recommendation.connect.go cluster.connect.go

OLD_IMPORT = github.com/devzero-inc/services/dakr/gen/api/v1
NEW_IMPORT = github.com/devzero-inc/pulumi-provider-devzero/internal/gen/api/v1

.PHONY: default
default: build

# -----------------------------------------------------------------------
# Proto sync
# Run `make proto` whenever proto/generated files change in services repo.
# -----------------------------------------------------------------------

.PHONY: proto
proto:
	@echo "Copying .proto files from $(SERVICES_DIR)/$(SOURCE_PROTO_DIR) ..."
	@for f in $(PROTO_FILES); do \
		if [ -f $(SERVICES_DIR)/$(SOURCE_PROTO_DIR)/$$f ]; then \
			cp $(SERVICES_DIR)/$(SOURCE_PROTO_DIR)/$$f $(TARGET_PROTO_DIR)/; \
		else \
			echo "Error: missing $(SERVICES_DIR)/$(SOURCE_PROTO_DIR)/$$f"; \
			exit 1; \
		fi; \
	done
	@echo "Done."

	@echo "Copying generated .pb.go files from $(SERVICES_DIR)/$(SOURCE_GEN_PB_DIR) ..."
	@for f in $(GEN_PB_FILES); do \
		if [ -f $(SERVICES_DIR)/$(SOURCE_GEN_PB_DIR)/$$f ]; then \
			cp $(SERVICES_DIR)/$(SOURCE_GEN_PB_DIR)/$$f $(TARGET_GEN_PB_DIR)/; \
		else \
			echo "Error: missing $(SERVICES_DIR)/$(SOURCE_GEN_PB_DIR)/$$f"; \
			exit 1; \
		fi; \
	done
	@echo "Done."

	@echo "Copying generated connect .go files from $(SERVICES_DIR)/$(SOURCE_GEN_CONNECT_DIR) ..."
	@for f in $(GEN_CONNECT_FILES); do \
		if [ -f $(SERVICES_DIR)/$(SOURCE_GEN_CONNECT_DIR)/$$f ]; then \
			cp $(SERVICES_DIR)/$(SOURCE_GEN_CONNECT_DIR)/$$f $(TARGET_GEN_CONNECT_DIR)/; \
		else \
			echo "Error: missing $(SERVICES_DIR)/$(SOURCE_GEN_CONNECT_DIR)/$$f"; \
			exit 1; \
		fi; \
	done
	@echo "Done."

	@echo "Rewriting import paths ..."
	@for f in $(TARGET_GEN_PB_DIR)/*.go $(TARGET_GEN_CONNECT_DIR)/*.go; do \
		if [ "$$(uname)" = "Darwin" ]; then \
			sed -i '' "s|$(OLD_IMPORT)|$(NEW_IMPORT)|g" $$f; \
		else \
			sed -i "s|$(OLD_IMPORT)|$(NEW_IMPORT)|g" $$f; \
		fi; \
	done
	@echo "Import path rewrite complete."

# -----------------------------------------------------------------------
# Build
# -----------------------------------------------------------------------

.PHONY: build
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(PROVIDER) ./provider/cmd/$(PROVIDER)/...

.PHONY: install
install: build
	cp bin/$(PROVIDER) $(GOPATH)/bin/$(PROVIDER)

# -----------------------------------------------------------------------
# Schema & SDK generation
# Requires Pulumi CLI: brew install pulumi
# -----------------------------------------------------------------------

.PHONY: gen-schema
gen-schema: build
	@echo "Extracting schema from provider binary..."
	pulumi package get-schema ./bin/$(PROVIDER) > /tmp/devzero-generated.json
	@echo "Merging resources+types into $(SCHEMA_FILE)..."
	python3 -c "\
import json; \
gen=json.load(open('/tmp/devzero-generated.json')); \
ex=json.load(open('$(SCHEMA_FILE)')); \
ex['resources']=gen.get('resources',{}); \
ex['types']=gen.get('types',{}); \
ex['functions']=gen.get('functions',{}); \
json.dump(ex, open('$(SCHEMA_FILE)','w'), indent=4); \
print(f'Merged: {len(ex[\"resources\"])} resources, {len(ex[\"types\"])} types, {len(ex[\"functions\"])} functions')"
	@echo "Applying enum patches..."
	python3 scripts/patch-schema-enums.py

.PHONY: gen-sdk
gen-sdk: gen-schema
	@echo "Generating TypeScript SDK..."
	pulumi package gen-sdk --language nodejs --out sdk $(SCHEMA_FILE)
	@echo "Generating Python SDK..."
	pulumi package gen-sdk --language python --out sdk $(SCHEMA_FILE)
	@echo "Generating Go SDK..."
	pulumi package gen-sdk --language go     --out sdk $(SCHEMA_FILE)
	@echo "All SDKs generated."

.PHONY: sdk
sdk: gen-sdk

# -----------------------------------------------------------------------
# Test
# -----------------------------------------------------------------------

.PHONY: test
test:
	go test ./... -v -count=1

# -----------------------------------------------------------------------
# Tidy
# -----------------------------------------------------------------------

.PHONY: tidy
tidy:
	go mod tidy

# -----------------------------------------------------------------------
# Clean
# -----------------------------------------------------------------------

.PHONY: clean
clean:
	rm -rf bin/ sdk/
