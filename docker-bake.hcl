variable "TARGETARCH" {
    default = "amd64"
}

variable "TAG" {
    default = "0.0.1-localtag"
}

variable "VERSION" {
    default = "0.0.1-local"
}



variable "REPOSITORY_BASE" {
    default = "opensource"
}

variable "REGISTRY" {
    default = "acr.aishu.cn"
}


target "deploy-service_image" {
    context = "deploy-service"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/deploy-service:${TAG}"
    ]
}


target "deploy-service_chart" {
    context = "deploy-service"
    dockerfile = "Dockerfile.chart"
    target = "result"
    output = [ ".buildx/charts" ]
    args = {
        VERSION = "${VERSION}"
        TAG = "${TAG}"
        REPOSITORY = "${REPOSITORY_BASE}/deploy-service"
        REGISTRY = "${REGISTRY}"
    }
}

group "deploy-service" {
    targets = [
        "deploy-service_image",
        "deploy-service_chart"
    ]
}


target "deploy-web-static" {
    context = "deploy-web/deploy-web-static"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/deploy-web-static:${TAG}"
    ]
}

target "deploy-web-service" {
    context = "deploy-web/deploy-web-service"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/deploy-web-service:${TAG}"
    ]
}


target "deploy-web_chart" {
    context = "deploy-web"
    dockerfile = "Dockerfile.chart"
    target = "result"
    output = [ ".buildx/charts" ]
    args = {
        VERSION = "${VERSION}"
        TAG = "${TAG}"
        REPOSITORY_SERVICE = "${REPOSITORY_BASE}/deploy-web-service"
        REPOSITORY_STATIC = "${REPOSITORY_BASE}/deploy-web-static"
        REGISTRY = "${REGISTRY}"
    }
}

group "deploy-web" {
    targets = [
        "deploy-web-static",
        "deploy-web-service",
        "deploy-web_chart"
    ]
}


target "component-manage_image" {
    context = "component-manage"
    target = "build-result"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/component-manage:${TAG}"
    ]
}

target "component-manage_chart" {
    context = "component-manage"
    dockerfile = "Dockerfile.chart"
    target = "result"
    output = [ ".buildx/charts" ]
    args = {
        VERSION = "${VERSION}"
        TAG = "${TAG}"
        REPOSITORY = "${REPOSITORY_BASE}/component-manage"
        REGISTRY = "${REGISTRY}"
    }
}

group "component-manage" {
    targets = [
        "component-manage_image",
        "component-manage_chart"
    ]
}


// component-manage-src
target "_deployrunner-common" {
    dockerfile = "build/puredocker/installer.dockerfile"
    context = "deployrunner"
    // FIXME: azure docker version unsupport moby/buildkit contextx
    // contexts = {
    //     "ctx-component-manage-src" = "./component-manage"
    //     "ctx-static-web" = "target:deploy-web-static"
    // }
    args = {
        TARGETARCH = "${TARGETARCH}"
    }
}


target "deployrunner_core-installer" {
    inherits = [ "_deployrunner-common" ]
    target = "result-core-installer"
    output = [ ".buildx" ]
    args = {
        TAG = "${VERSION}"
    }
}

target "deployrunner_binaries" {
    inherits = [ "_deployrunner-common" ]
    target = "result-binary"
    output = [ ".buildx" ]
}

target "deployrunner_deploy-installer" {
    inherits = [ "_deployrunner-common" ]
    target = "result-installer"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/deploy-installer:${TAG}"
    ]
}

target "deployrunner_deploy-builder" {
    inherits = [ "_deployrunner-common" ]
    target = "result-builder"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/deploy-builder:${TAG}"
    ]
}

target "deployrunner_chart" {
    context = "deployrunner"
    dockerfile = "build/puredocker/charts.dockerfile"
    target = "result"
    output = [ ".buildx/charts" ]
    args = {
        VERSION = "${VERSION}"
        TAG = "${TAG}"
        REPOSITORY = "${REPOSITORY_BASE}/deploy-installer"
        REGISTRY = "${REGISTRY}"
    }
}

group "deployrunner" {
    targets = [
        "deployrunner_core-installer",
        "deployrunner_deploy-installer",
        "deployrunner_deploy-builder",
        "deployrunner_binaries",
        "deployrunner_chart"
    ]
}



target "communication_controller" {
    context = "communications"
    dockerfile = "Dockerfile"
    target = "controller"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/communication-controller:${TAG}"
    ]
}
target "communication_controller118" {
    context = "communications"
    dockerfile = "Dockerfile"
    target = "controller118"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/communication-controller118:${TAG}"
    ]
}
target "communication_controller120" {
    context = "communications"
    dockerfile = "Dockerfile"
    target = "controller120"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/communication-controller120:${TAG}"
    ]
}
target "communication_runtime" {
    context = "communications"
    dockerfile = "Dockerfile"
    target = "runtime"
    tags = [
        "${REGISTRY}/${REPOSITORY_BASE}/communication-runtime:${TAG}"
    ]
}

target "communication_charts" {
    context = "communications"
    dockerfile = "Dockerfile.chart"
    target = "result"
    output = [ ".buildx/charts" ]
    args = {
        VERSION = "${VERSION}"
        TAG = "${TAG}"
        REPOSITORY = "${REPOSITORY_BASE}/communication-controller"
        REPOSITORY_118 = "${REPOSITORY_BASE}/communication-controller118"
        REPOSITORY_120 = "${REPOSITORY_BASE}/communication-controller120"
        REPOSITORY_RT = "${REPOSITORY_BASE}/communication-runtime"
        REGISTRY = "${REGISTRY}"
    }
}

group "communications" {
    targets = [
        "communication_controller",
        "communication_controller118",
        "communication_controller120",
        "communication_runtime",
        "communication_charts"
    ]
}


group "default" {
    targets = [
        "deploy-service",
        "component-manage",
        "deployrunner",
        "communications",
        "deploy-web"
    ]
}