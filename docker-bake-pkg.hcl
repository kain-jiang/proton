variable "TARGETARCH" {
    default = "amd64"
}

variable "VERSION" {
    default = "0.0.1-local"
}


target "system" {
    dockerfile = "package-proton/ci/Dockerfile.system"
    target = "result"
    output = [ ".buildx/pkgs" ]
    args = {
        TARGETARCH = "${TARGETARCH}"
        VERSION = "${VERSION}"
    }
}


target "system-image" {
    dockerfile = "package-proton/ci/Dockerfile.system"
    target = "image-result"
    output = [ ".buildx/pkgs" ]
    args = {
        TARGETARCH = "${TARGETARCH}"
        VERSION = "${VERSION}"
    }
}


target "studio" {
    dockerfile = "package-proton/ci/Dockerfile.studio"
    target = "result"
    output = [ ".buildx/pkgs" ]
    args = {
        TARGETARCH = "${TARGETARCH}"
        VERSION = "${VERSION}"
    }
}


target "studio-image" {
    dockerfile = "package-proton/ci/Dockerfile.studio"
    target = "image-result"
    output = [ ".buildx/pkgs" ]
    args = {
        TARGETARCH = "${TARGETARCH}"
        VERSION = "${VERSION}"
    }
}


group "default" {
    targets = [
        "system",
        "system-image",
        "studio",
        "studio-image"
    ]
}