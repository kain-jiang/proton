# `offline-package app` 功能设计

本文档定义 `proton-cli offline-package app` 子功能，用于把应用级发布清单导出为可搬运的离线应用包，或将离线应用包导入到目标环境的镜像仓库和 Chart 仓库。

当前目录中的 [kweaver-core.yaml](./kweaver-core.yaml) 是一个 `VersionSet` 示例，其核心结构如下：

- `source`
  - 定义 Helm 仓库名称和地址
- `dependencies`
  - 定义依赖产品及其清单位置
  - 当前功能不处理该字段
- `releases`
  - 定义当前产品包含的每个 Helm Chart 及版本

## 1. 目标

`offline-package app` 解决的是“应用包离线分发”问题，不直接构建完整的 Proton 节点安装介质。

与现有 `offline-package build/install` 的差异：

- `build/install`
  - 面向整套 Proton 基础设施离线安装
  - 输出内容包含二进制、RPM、镜像、Chart、安装脚本
- `app export/import`
  - 面向应用发布内容迁移
  - 输出内容只包含应用清单、Chart 和镜像

目标能力：

- 根据应用清单自动收集所需 Chart
- 根据 Chart 内容自动提取所需镜像
- 将收集结果打成一个标准 tar 包
- 将 tar 包导入到目标镜像仓库和 Chart 仓库

非目标：

- 不负责部署应用
- 不负责安装目标仓库本身
- 不在导出包中保留完整 registry 地址
- 不把 RPM、二进制、系统依赖纳入应用包
- 不处理 `dependencies`

## 2. 命令设计

### 2.1 导出命令

```bash
proton-cli offline-package app export \
  --manifest ./kweaver-core.yaml \
  --output ./offline-app-package.tar \
  --platform linux/amd64,linux/arm64 \
  --override-registry mirror.example.com \
  --ignore-missing-images
```

等价短参数形式：

```bash
proton-cli offline-package app export \
  -f ./kweaver-core.yaml \
  -o ./offline-app-package.tar \
  --platform linux/amd64,linux/arm64 \
  --override-registry mirror.example.com \
  --ignore-missing-images
```

参数说明：

- `--manifest`, `-f`
  - 应用清单文件路径
  - 支持本地文件路径
  - 支持 `http://` 或 `https://` URL
- `--output`, `-o`
  - 导出 tar 包路径
  - 默认值建议为 `offline-app-package.tar`
- `--platform`
  - 目标平台列表
  - 支持单平台或逗号分隔的多平台
  - 当前支持 `linux/amd64`、`linux/arm64`
  - 同时兼容 `x86_64`、`aarch64`、`amd64`、`arm64`
  - 最终统一归一化为 `linux/amd64`、`linux/arm64`
  - 示例：`--platform linux/amd64`
  - 示例：`--platform linux/amd64,linux/arm64`
- `--ignore-missing-images`
  - 是否忽略无法拉取的镜像
  - 默认值为 `false`
  - `false` 时任一镜像拉取失败直接报错退出
  - `true` 时记录 warning，继续导出剩余镜像，并在 `manifest.yaml` 中补充失败镜像信息
- `--override-registry`
  - 导出时使用指定值重写 chart values 中的 `.image.registry`
  - 默认值为空，表示直接使用 chart 中识别出的原始镜像地址
  - 重写后保留原始 `image.repository` 和 `image.tag`
  - 示例：`docker.io/library/nginx:1.27.0` 重写为 `mirror.example.com/library/nginx:1.27.0`

### 2.2 导入命令

```bash
proton-cli offline-package app import \
  --input ./offline-app-package.tar \
  --auto \
  --force
```

或显式指定目标仓库：

```bash
proton-cli offline-package app import \
  --input ./offline-app-package.tar \
  --registry registry.hello.com \
  --registry-username username \
  --registry-password password \
  --registry-plain-http=false \
  --force \
  --chartmuseum-url https://chartmuseum.hello.com \
  --chartmuseum-username username \
  --chartmuseum-password password
```

参数说明：

- `--input`, `-i`
  - 本地离线包路径
  - 只接受本地文件路径，不接受 URL
- `--registry`
  - 目标镜像仓库地址
  - 例如 `registry.hello.com`
  - 未指定且启用 `--auto` 时，自动从当前 Proton 集群配置读取
- `--registry-username`
  - 镜像仓库用户名
  - 未指定且启用 `--auto` 时，自动从当前 Proton 集群配置读取
- `--registry-password`
  - 镜像仓库密码
  - 未指定且启用 `--auto` 时，自动从当前 Proton 集群配置读取
- `--registry-plain-http`
  - 是否允许以 HTTP 明文方式访问镜像仓库
  - 默认值建议为 `false`
  - 启用 `--auto` 且当前镜像仓库类型为外部 OCI 时，会自动继承集群配置中的 `plain_http`
- `--force`
  - 是否覆盖 ChartMuseum 中已存在的同版本 chart
  - 默认值为 `false`
  - `false` 时遇到已存在 chart 直接报错退出
  - `true` 时使用覆盖方式上传 chart
- `--auto`
  - 是否自动从当前 Proton 集群配置补全导入目标
  - 自动补全 `registry`、registry 认证、`chartmuseum-url`、ChartMuseum 认证
  - 若同时显式传入同名参数，以显式参数为准
  - 若当前集群 chart 仓库不是 ChartMuseum，则直接报错
- `--chartmuseum-url`
  - 目标 ChartMuseum 地址
  - 未指定且启用 `--auto` 时，自动从当前 Proton 集群配置读取
- `--chartmuseum-username`
  - ChartMuseum 用户名
  - 未指定且启用 `--auto` 时，自动从当前 Proton 集群配置读取
- `--chartmuseum-password`
  - ChartMuseum 密码
  - 未指定且启用 `--auto` 时，自动从当前 Proton 集群配置读取

## 3. 输入清单格式

### 3.1 顶层结构

当前设计基于 `VersionSet` 清单，至少需要以下字段：

```yaml
apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: kweaver-core
version: 0.5.0
source:
  helmRepoName: kweaver
  helmRepoUrl: https://kweaver-ai.github.io/helm-repo/
dependencies:
  - product: isf
    version: 0.5.0
    manifest: ./isf.yaml
    optional: true
    defaultEnabled: true
releases:
  deploy-web:
    chart: deploy-web
    version: 0.3.0
```

字段约束：

- `kind` 必须为 `VersionSet`
- `product` 必须存在
- `version` 必须存在
- `source.helmRepoUrl` 必须存在
- `releases` 必须存在且不能为空

### 3.2 `source`

`source` 用于定位 Chart 来源：

- `helmRepoName`
  - Helm 仓库名称
  - 用于临时添加 repo 或生成下载来源标识
- `helmRepoUrl`
  - Helm 仓库地址
  - 用于下载 `releases` 中声明的 Chart

### 3.3 `dependencies`

`dependencies` 表示当前产品依赖的其他产品清单。

当前版本明确不处理 `dependencies`，行为如下：

- 不递归读取依赖清单
- 不根据 `optional` 或 `defaultEnabled` 参与筛选
- 不把依赖产品中的 Chart 和镜像纳入导出结果

实现要求：

- 解析清单时可以保留该字段内容
- 导出逻辑必须忽略该字段
- 命令输出中应明确提示“当前仅处理主清单 releases，dependencies 已忽略”

### 3.4 `releases`

`releases` 是导出功能的核心输入。每个 release 至少包含：

- `chart`
  - Chart 名称
- `version`
  - Chart 版本

导出时应将每个 release 转换为一个待下载的 Chart 包：

- 下载目标：`<chart>-<version>.tgz`
- 来源：`source.helmRepoUrl`

## 4. 导出流程

### 4.1 总体流程

`export` 执行步骤：

1. 读取并解析清单
2. 归一化平台参数
3. 忽略 `dependencies`，仅收集主清单 `releases` 中涉及的 Chart
4. 下载所有 Chart 到临时目录
5. 解包或读取每个 Chart，提取镜像定义
6. 按 `--platform` 选择镜像 manifest，并拉取到本地 OCI Layout
7. 将所有已识别镜像及其导出结果写入 `manifest.yaml`
8. 组装 `charts/`、`images/`、`manifest.yaml`
9. 打包为 tar 文件

### 4.2 Chart 收集规则

Chart 来源仅取自清单中的 `source` 和 `releases`：

- 仓库地址来自 `source.helmRepoUrl`
- Chart 列表来自 `releases[*].chart`
- Chart 版本来自 `releases[*].version`

输出时所有 Chart 平铺保存到 `charts/` 目录，不保留 release 名称层级。

示例：

- `releases.deploy-web.chart = deploy-web`
- `releases.deploy-web.version = 0.3.0`
- 导出结果文件名为 `charts/deploy-web-0.3.0.tgz`

### 4.3 镜像提取规则

镜像不直接来自清单，而是从每个 Chart 的默认 values 中提取。

首版支持以下两种常见结构。

#### 单镜像结构

```yaml
image:
  registry: docker.io
  repository: library/nginx
  tag: 1.27.0
```

识别规则：

- `image.registry`
- `image.repository`
- `image.tag`

拼装后的完整镜像引用：

```text
docker.io/library/nginx:1.27.0
```

#### 多镜像结构

```yaml
image:
  registry: docker.io
  controller:
    repository: bitnami/nginx
    tag: 1.0.0
  sidecar:
    repository: bitnami/os-shell
    tag: 1.0.0
```

识别规则：

- `image.registry`
- `image.<name>.repository`
- `image.<name>.tag`

拼装后的完整镜像引用：

- `docker.io/bitnami/nginx:1.0.0`
- `docker.io/bitnami/os-shell:1.0.0`

### 4.4 镜像提取边界

首版建议只处理以下情况：

- `image.repository` + `image.tag`
- `image.<name>.repository` + `image.<name>.tag`
- 共用顶层 `image.registry`

首版暂不处理：

- `global.imageRegistry`
- `digest`
- `repository` 内已包含 registry 且与 `image.registry` 冲突
- 使用模板函数动态拼接出的镜像地址
- 镜像字段不在 `image` 根节点下的自定义结构

对于无法识别的 Chart：

- 输出 warning
- 在最终汇总中列出未识别项
- 是否终止执行可以作为后续策略参数控制

对于已识别但拉取失败的镜像：

- 默认直接失败
- 若指定 `--ignore-missing-images`，则输出 warning 并继续处理其他镜像
- 失败镜像必须写入导出包的 `manifest.yaml`

对于启用了 `--override-registry` 的导出：

- 行为上等价于导出时将 values 中的 `.image.registry` 替换为指定值
- `manifest.yaml` 中应保留原始识别结果 `source`
- `manifest.yaml` 中应记录本次使用的 `overrideRegistry`
- 同时补充实际拉取地址 `pullSource`

### 4.5 镜像去重

镜像拉取前需要按完整镜像引用去重。

去重维度：

- registry
- repository
- tag

相同镜像只拉取一次，只在 OCI Layout 中保存一份。

### 4.6 平台选择规则

- 当 `--platform` 只包含一个平台时：
  - 如果远端 tag 是单架构 manifest，则要求其平台与目标平台匹配
  - 如果远端 tag 是多架构 index，则只选择匹配平台的那个 manifest 导出
- 当 `--platform` 包含多个平台时：
  - 对每个平台分别选择匹配的 manifest
  - 导出包中的镜像 tag 会重新组织为一个只包含所选平台的 OCI index
  - 当前推荐使用 `linux/amd64,linux/arm64` 作为双架构导出形式
- 若指定平台在远端 tag 中不存在：
  - 默认直接失败
  - 若指定 `--ignore-missing-images`，则该镜像记录 warning 并跳过

## 5. 导出包格式

导出产物是一个 tar 包，例如：

```text
offline-app-package.tar
```

内部目录结构如下：

```text
offline-app-package.tar
├── manifest.yaml
├── images/
│   └── oci-layout...
└── charts/
    ├── deploy-web-0.3.0.tgz
    └── studio-web-0.3.0.tgz
```

### 5.1 `manifest.yaml`

要求：

- 保留原始清单内容
- 补充平台信息
- 补充本次识别出的完整镜像列表
- 如果有镜像拉取失败，补充失败原因

建议增加字段：

```yaml
platform: linux/amd64
```

如果不希望改动原始 `VersionSet` 结构，也可以增加一个打包元数据段，例如：

```yaml
offlinePackage:
  platforms:
    - linux/amd64
    - linux/arm64
  overrideRegistry: mirror.example.com
  exportedAt: 2026-03-31T00:00:00Z
  images:
    - source: docker.io/library/nginx:1.27.0
      pullSource: mirror.example.com/library/nginx:1.27.0
      repository: library/nginx
      tag: 1.27.0
      localRef: library/nginx:1.27.0
      requestedPlatforms:
        - linux/amd64
        - linux/arm64
      exportedPlatforms:
        - linux/amd64
      exported: true
    - source: acr.example.com/demo/api:1.0.0
      repository: demo/api
      tag: 1.0.0
      localRef: demo/api:1.0.0
      requestedPlatforms:
        - linux/amd64
        - linux/arm64
      exported: false
      error: EOF
  imageErrors:
    - image acr.example.com/demo/api:1.0.0 skipped: EOF
```

说明：

- `images` 记录的是“已识别出的完整镜像列表”，而不只是成功导出的镜像
- `requestedPlatforms` 记录该镜像本次导出请求的平台列表
- `exportedPlatforms` 记录该镜像实际导出的平台列表
- `pullSource` 记录该镜像本次实际拉取时使用的镜像地址
- `exported: true` 表示该镜像已进入离线包内的 OCI Layout
- `exported: false` 表示该镜像识别成功，但本次未成功导出
- `imageErrors` 用于汇总导出阶段被忽略的镜像错误

### 5.2 `images/`

镜像以 OCI Layout 格式保存。

要求：

- 所有镜像统一存放在一个 OCI Layout 中
- 文件名或 tag 名中只保留 `<repository>:<tag>`
- 不保留 registry 作为离线包中的命名空间前缀

示例：

- 原镜像：`docker.io/bitnami/nginx:1.0.0`
- 离线包中的命名：`bitnami/nginx:1.0.0`

说明：

- “不保留 registry” 仅指离线包内部的镜像命名
- 实际导入时仍需要重新打上目标 registry 前缀

### 5.3 `charts/`

要求：

- 所有 Chart 平铺保存
- 文件名统一为 `<chart>-<version>.tgz`
- 不保留 Helm 仓库目录结构

## 6. 导入流程

### 6.1 总体流程

`import` 执行步骤：

1. 校验输入 tar 包存在且可读
2. 解压到临时目录
3. 校验 `manifest.yaml`、`charts/`、`images/` 结构完整
4. 登录目标镜像仓库
5. 将 OCI Layout 中的镜像逐个推送到目标 registry
6. 将 `charts/` 中的 Chart 逐个上传到 ChartMuseum
7. 输出导入结果摘要

### 6.2 镜像导入规则

导入时将离线包中的镜像重新命名为：

```text
<registry>/<repository>:<tag>
```

示例：

- 离线包镜像名：`bitnami/nginx:1.0.0`
- 目标仓库：`registry.hello.com`
- 推送目标：`registry.hello.com/bitnami/nginx:1.0.0`

镜像仓库行为要求：

- 支持用户名密码认证
- 支持 HTTPS
- 允许通过 `--registry-plain-http` 显式切换到 HTTP

### 6.3 Chart 导入规则

`charts/` 目录下每个 `.tgz` 文件都应上传到 `ChartMuseum`。

要求：

- 支持 Basic Auth
- 失败时输出明确的 Chart 文件名和 HTTP 响应
- 已存在版本的处理策略可通过 `--force` 控制

当前策略：

- 默认 `--force=false`，如果远端已存在相同 chart+version，则报错并停止
- 当 `--force=true` 时，允许覆盖远端已存在 chart

## 7. 错误处理

### 7.1 导出阶段

以下情况应直接失败：

- 清单文件不存在
- 清单 URL 下载失败
- 清单 YAML 解析失败
- `kind` 不是 `VersionSet`
- `source.helmRepoUrl` 缺失
- `releases` 为空
- Chart 下载失败
- 镜像拉取失败
- tar 包写入失败

补充说明：

- 当指定 `--ignore-missing-images=true` 时，“镜像拉取失败”不再直接终止，而是转为 warning，并记录到 `manifest.yaml`

以下情况可先 warning，后续再决定是否提升为错误：

- Chart 未识别出任何镜像
- 某个镜像字段结构不符合约定

### 7.2 导入阶段

以下情况应直接失败：

- 输入 tar 包不存在
- tar 包损坏
- 缺失 `manifest.yaml`
- 缺失 `charts/` 或 `images/`
- 镜像仓库认证失败
- ChartMuseum 认证失败
- 任一镜像上传失败
- 任一 Chart 上传失败

## 8. 日志与输出

建议命令输出分阶段日志，至少包含：

- 读取清单
- 解析 release
- 下载 chart
- 提取镜像
- 拉取镜像
- 打包完成
- 解压离线包
- 推送镜像
- 上传 chart

建议最终输出摘要：

```text
export completed
- platform: linux/amd64
- charts: 28
- images: 63
- output: ./offline-app-package.tar
```

```text
import completed
- registry: registry.hello.com
- chartmuseum: https://chartmuseum.hello.com
- charts imported: 28
- images imported: 63
```

## 9. 与现有 `offline-package` 的关系

建议实现为现有 `offline-package` 的子命令树：

```text
proton-cli offline-package
├── plan
├── build
├── install
└── app
    ├── export
    └── import
```

原因：

- 概念上同属于“离线包”能力
- 可复用已有 tar 打包、解包和 OCI 制品处理逻辑
- 用户心智一致

建议代码组织：

- `cmd/proton-cli/cmd/offline_package/app/`
  - 放置文档、样例清单和测试数据
- `cmd/proton-cli/cmd/offline_package/app.go`
  - 注册 `app` 根命令
- `cmd/proton-cli/cmd/offline_package/app_export.go`
  - 导出命令
- `cmd/proton-cli/cmd/offline_package/app_import.go`
  - 导入命令
- `cmd/proton-cli/cmd/offline_package/app_manifest.go`
  - `VersionSet` 解析
- `cmd/proton-cli/cmd/offline_package/app_chart.go`
  - Chart 下载与解析
- `cmd/proton-cli/cmd/offline_package/app_image.go`
  - 镜像提取、去重、导入导出

## 10. 首版实现建议

为了降低复杂度，首版建议按下面范围落地：

- 支持本地文件和 HTTP(S) 清单输入
- 只处理主清单 `releases`
- 从 `source.helmRepoUrl` 下载 Chart
- 仅支持本文定义的两种镜像字段结构
- 使用 OCI Layout 保存镜像
- 支持推送到 OCI Registry
- 支持上传到 ChartMuseum

首版暂不实现：

- `dependencies` 处理与递归展开
- 更复杂的 values 模板求值
- 增量导入
- 并发拉取和并发推送控制
- SBOM、签名、校验和

## 11. 后续可扩展项

- 支持更多镜像字段模式
- 支持 Helm dependency build 后再分析子 Chart
- 支持导出校验和文件
- 支持导入前 dry-run
- 支持导入覆盖策略
- 支持导出包元数据版本号
