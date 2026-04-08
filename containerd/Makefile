# containerd Makefile for Proton project
#
# This Makefile downloads official containerd release binaries and creates
# RPM/DEB packages using nfpm.

# Default architecture and version
export VERSION	?= 2.2.2
export RELEASE	?= 1.proton
export ARCH   	?= amd64
export PLATFORM	?= linux

# Convert architecture names
ifeq ($(ARCH),amd64)
	OS_ARCH   := x86_64
	NFPM_ARCH := $(ARCH)
else ifeq ($(ARCH),arm64)
	OS_ARCH   := aarch64
	NFPM_ARCH := $(ARCH)
else
	OS_ARCH   := $(ARCH)
	NFPM_ARCH := $(ARCH)
endif


# Output directory
OUTPUT_DIR := output

# Build directory
BUILD_DIR := build/$(PLATFORM)-$(ARCH)
# .unpacked-success
UNPACKED_SUCCESS := $(BUILD_DIR)/.unpacked-success-$(ARCH)

# Artifact
TARBALL ?= https://localhost/mirror/github.com/containerd/containerd/releases/download/v$(VERSION)/containerd-$(VERSION)-linux-$(ARCH).tar.gz
RPM		?= containerd-$(VERSION)-$(RELEASE).$(OS_ARCH).rpm
SWR		?= swr.cn-east-3.myhuaweicloud.com/kweaver-ai/proton/rpm/containerd:$(VERSION)-$(RELEASE).$(OS_ARCH)

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: build
build: $(OUTPUT_DIR)/$(RPM)
build: ## Build containerd rpm package

$(OUTPUT_DIR)/$(RPM): $(OUTPUT_DIR) $(UNPACKED_SUCCESS)
# 	VERSION=$(VERSION) RELEASE=$(RELEASE) ARCH=$(ARCH) nfpm package --packager=rpm --target="$(OUTPUT_DIR)/$(RPM)"
	nfpm package --packager=rpm --target="$(OUTPUT_DIR)/$(RPM)"

$(UNPACKED_SUCCESS): $(BUILD_DIR)
	curl --location --fail --silent $(TARBALL) | tar --extract --gzip --directory $(BUILD_DIR)
	touch $(UNPACKED_SUCCESS)

$(OUTPUT_DIR):
	mkdir --parent $(OUTPUT_DIR)
$(BUILD_DIR):
	mkdir --parent $(BUILD_DIR)

##@ Release
.PHONY: release
push: ## Push rpm packages to Huawei Cloud SWR
push: $(OUTPUT_DIR)/$(RPM)
	oras push $(SWR) $(OUTPUT_DIR)/$(RPM)

##@ Clean

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf $(OUTPUT_DIR) $(BUILD_DIR)

##@ Info

.PHONY: info
info: ## Display build information
	@echo "Build Configuration:"
	@echo "  Architecture: $(ARCH) (OS: $(OS_ARCH), NFPM: $(NFPM_ARCH))"
	@echo "  Version:      $(VERSION)"
	@echo "  Release:      $(RELEASE)"
	@echo "  Source:       $(TARBALL)"
