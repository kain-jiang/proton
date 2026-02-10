FROM acr.aishu.cn/public/ubuntu:22.04.20251014 AS runtime
FROM acr.aishu.cn/ict/builder:apps-deploy-service-py310.latest AS pyenv
FROM acr.aishu.cn/ict/builder:tools-helm-3.19 AS helm3


##### 2.deploy-service
FROM pyenv AS deploy-service-builder
RUN --mount=type=bind,target=/src,rw \
    set -x && \
    cd /src/src/deploy-service && \
    uv sync -p 3.10 && uv tree && \
    uv run pyinstaller \
        --name deploy-service \
        --onedir \
        --log-level WARN \
        --strip \
        --distpath /build/dist \
        ./main.py \
    && mkdir -p /build/dist/conf/default/cert/anyshare \
    && cp -f ./src/cert/anyshare/openssl.cnf /build/dist/conf/default/cert/anyshare/openssl.cnf

##### builder result #####

FROM runtime AS resultenv
COPY --from=helm3 /helm /usr/bin/helm
ENV TZ=Asia/Shanghai LANG=en_US.UTF-8 DM_SVC_PATH=/tmp/conf/

FROM resultenv AS build-result
COPY --from=deploy-service-builder /build/dist /app
WORKDIR /app

#### result list
FROM build-result AS deploy-service
ENTRYPOINT ["./deploy-service/deploy-service"]
