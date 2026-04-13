GOARCH_x86_64  = amd64
GOARCH_aarch64 = arm64

VERSION    := 1.1.8
RELEASE    := 120.1.proton
SWR        := swr.cn-east-3.myhuaweicloud.com/kweaver-ai/proton/rpm/ecms

ARCH_LIST        := x86_64 aarch64
BUILD_TARGETS    := $(addprefix build-,$(ARCH_LIST))
BUILD_DIRS       := $(addprefix linux-,$(ARCH_LIST))
PUSH_TARGETS     := $(addprefix push-,$(ARCH_LIST))

# build
build: $(BUILD_TARGETS)

$(BUILD_TARGETS): build-%: linux-%/.unpacked-success nfpm.yaml
	VERSION=$(VERSION) PRERELEASE=$(PRERELEASE) RELEASE=$(RELEASE) ARCH=$(*) GOARCH=$(GOARCH_$(*)) nfpm package --packager=rpm

$(foreach arch,$(ARCH_LIST),linux-$(arch)/.unpacked-success): linux-%/.unpacked-success: linux-% ecms-1.1.8-120.el7.%.rpm
	rpm2cpio ecms-1.1.8-120.el7.$(*).rpm | cpio -idmvD "linux-$(*)"
	touch "$(@)"

$(addprefix linux-,$(ARCH_LIST)):
	mkdir --parents $(@)



# release
push: $(PUSH_TARGETS)

$(PUSH_TARGETS): push-%: build-%
	oras push $(SWR):$(VERSION)-$(RELEASE).$(*) ecms-$(VERSION)-$(RELEASE).$(*).rpm

clean:
	@rm --force --recursive $(BUILD_DIRS)
