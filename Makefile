PROJECT      := github.com/devzero-inc/pulumi-provider-devzero
PROVIDER     := pulumi-resource-devzero
WORKING_DIR  := $(shell pwd)
GOPATH       := $(shell go env GOPATH)

# Proto sync settings — override with: make proto SERVICES_DIR=/path/to/services
SERVICES_DIR ?= ../services

SOURCE_PROTO_DIR       = dakr/proto/api/v1
SOURCE_GEN_PB_DIR      = dakr/gen/api/v1
SOURCE_GEN_CONNECT_DIR = dakr/gen/api/v1/apiv1connect

TARGET_PROTO_DIR       = internal/proto/api/v1
TARGET_GEN_PB_DIR      = internal/gen/api/v1
TARGET_GEN_CONNECT_DIR = internal/gen/api/v1/apiv1connect

PROTO_FILES       = common.proto instance.proto k8s.proto recommendation.proto
GEN_PB_FILES      = common.pb.go instance.pb.go k8s.pb.go recommendation.pb.go
GEN_CONNECT_FILES = k8s.connect.go recommendation.connect.go

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
	go build -o bin/$(PROVIDER) ./provider/cmd/$(PROVIDER)/...

.PHONY: install
install: build
	cp bin/$(PROVIDER) $(GOPATH)/bin/$(PROVIDER)

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
	rm -rf bin/
