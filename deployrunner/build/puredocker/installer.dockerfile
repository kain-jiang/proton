# https://docs.docker.com/reference/dockerfile/#automatic-platform-args-in-the-global-scope
ARG TARGETARCH

FROM acr.aishu.cn/ict/builder:apps-deploy-installer-product.latest AS runtime
FROM acr.aishu.cn/ict/builder:golang-1.25.5n AS goenv

# FIXME: use contexts
# FROM scratch AS ctx-component-manage-src

FROM scratch AS static-src
# https://docs.docker.com/build/bake/reference/#targetcontexts
COPY --from=ctx-static-web /AnyShare/DeployWebStatic /

FROM goenv AS builder 
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app/deployrunner
COPY go.mod go.sum ./

RUN --mount=type=bind,target=/app/component-manage,from=ctx-component-manage-src \
    go mod download

# start build
ARG LDFLAGS="-s -w"
ARG TARGETARCH

RUN --mount=type=cache,id=dp_installer,target=/root/.cache/go-build \
    --mount=type=bind,target=/app/component-manage,from=ctx-component-manage-src \
    --mount=type=bind,target=/app/deployrunner \
    --mount=type=bind,target=/app/deployrunner/cmd/core-installer/static,from=static-src \
    GOARCH=${TARGETARCH} go build -ldflags "${LDFLAGS}" -o /build/${TARGETARCH}/bin/ ./cmd/*

FROM runtime AS tar-installer
ARG TAG TARGETARCH
WORKDIR /work
COPY --from=builder /build/${TARGETARCH}/bin/core-installer /work
RUN tar -czvf /work/core-installer-${TARGETARCH}-${TAG}.tgz core-installer

### build result
FROM runtime AS result-builder
ARG TARGETARCH
COPY --from=builder /build/${TARGETARCH}/bin/builder /usr/sbin/
ENTRYPOINT ["/usr/sbin/builder"]

FROM runtime AS result-installer
ARG TARGETARCH
COPY --from=builder /build/${TARGETARCH}/bin/installer /usr/sbin/
COPY --from=builder /build/${TARGETARCH}/bin/server /usr/sbin/
COPY sql-ddl /sql-ddl
COPY config.devel /
ENTRYPOINT ["/usr/sbin/server"]


FROM scratch AS result-core-installer
ARG TARGETARCH TAG
COPY --from=tar-installer /work/core-installer-${TARGETARCH}-${TAG}.tgz /

FROM scratch AS result-binary
ARG TARGETARCH
COPY --from=builder /build/${TARGETARCH} /
