# containerd Makefile for Proton project
#
# This Makefile downloads official containerd release binaries and creates
# RPM/DEB packages using nfpm.
ALL_ARCH_LIST = amd64 arm64

# Default architecture and version
export VERSION	?= 1.4.2
export RELEASE	?= 3.proton
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

# Artifact
GITHUB      ?= https://github.com
GITHUB_REPO ?= opencontainers/runc
GITHUB_REL  ?= $(GITHUB)/$(GITHUB_REPO)/releases/download/v$(VERSION)/runc.$(ARCH)
BIN         ?= runc.$(ARCH)
RPM         ?= runc-$(VERSION)-$(RELEASE).$(OS_ARCH).rpm
SWR         ?= swr.cn-east-3.myhuaweicloud.com/kweaver-ai/proton/rpm/runc:$(VERSION)-$(RELEASE).$(OS_ARCH)

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: build
build: $(RPM)
build: ## Build containerd rpm package

$(RPM): nfpm.yaml $(BIN)
	chmod +x $(BIN)
	nfpm package --packager=rpm --target="$(@)"

$(BIN):
	curl --location --fail --output "$(@)" --show-error --silent "$(GITHUB_REL)"

##@ Release
.PHONY: push
push: ## Push rpm package to Huawei Cloud SWR
push: $(RPM)
	oras push $(SWR) $(RPM)


##@ Clean

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(foreach arch,$(ALL_ARCH_LIST),runc.$(arch)) *.rpm

##@ Info

.PHONY: info
info: ## Display build information
	@echo "Build Configuration:"
	@echo "  Architecture: $(ARCH) (OS: $(OS_ARCH), NFPM: $(NFPM_ARCH))"
	@echo "  Version:      $(VERSION)"
	@echo "  Release:      $(RELEASE)"
	@echo "  Source:       $(GITHUB_REL)"
	@echo "  SWR:          $(SWR)"
