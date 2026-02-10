FROM acr.aishu.cn/library/golang:1.25.6 AS builder
WORKDIR /app
ENV GOPROXY='https://mirrors.huaweicloud.com/repository/goproxy/' CGO_ENABLED="0"

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked,id=gomod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked,id=gomod \
    --mount=type=bind,target=/app,readwrite \
    go install .

FROM scratch
COPY --from=builder /go/bin/build /build