FROM acr.aishu.cn/ict/builder:golang-1.24.0 AS goenv
ENV CGO_ENABLED=0
ENV GOPROXY 'https://mirrors.aliyun.com/goproxy,https://goproxy.aishu.cn'
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download


FROM goenv as builder
ARG WEB_URL=https://ftp-ict.aishu.cn/proton/proton-cli-web/MISSION/proton-cli-web.latest.tar.gz
RUN --mount=type=bind,target=/app,readwrite \
    curl -kfsSL ${WEB_URL} | tar -xzvC cmd/proton-cli/cmd/web && \
    go build -x -ldflags \
    " \
    -X devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version.gitCommit=$(git rev-parse HEAD) \
    -X devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version.gitVersion=0.0.0-master+gxxxxxx \
    -X devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version.gitTreeState=dirty \
    -X devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version.buildDate=1970-01-01T00:00:00Z \
    " \
    -o /proton-cli ./cmd/proton-cli

FROM scratch as bin
COPY --from=builder /proton-cli /proton-cli 