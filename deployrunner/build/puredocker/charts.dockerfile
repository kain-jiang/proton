ARG VERSION="0.1.0-alpha"
ARG REPOSITORY
ARG REGISTRY
ARG TAG

FROM acr.aishu.cn/public/ubuntu:22.04.20251014 AS runtime
FROM acr.aishu.cn/ict/builder:tools-helm-3.17 AS helm3
FROM acr.aishu.cn/ict/builder:tools-yq.latest AS yq

# stage1: 修改版本
FROM yq AS stage1
ARG TAG REPOSITORY REGISTRY
COPY ./charts /src

WORKDIR /work

RUN true \
    && cp -r /src/* ./ \
    && yq -i '.image.tag = "'${TAG}'"' ./deploy-installer/values.yaml \
    && yq -i '.image.repository = "'${REPOSITORY}'"' ./deploy-installer/values.yaml \
    && yq -i '.image.registry = "'${REGISTRY}'"' ./deploy-installer/values.yaml


# stage2: 打包Chart
FROM runtime AS stage2
ARG VERSION
COPY --from=helm3 /usr/bin/helm /usr/bin/helm
COPY --from=stage1 /work /src

WORKDIR /work

RUN true \
    && helm package --version ${VERSION} /src/*


# stage3: 生成制品
FROM scratch AS result
COPY --from=stage2 /work/*.tgz /
