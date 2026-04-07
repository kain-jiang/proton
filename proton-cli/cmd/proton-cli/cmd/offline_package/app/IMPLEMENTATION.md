# `offline-package app` 实现方案

本文档对应 [FEATURE.md](./FEATURE.md) 的完整需求，说明当前代码的实现结构、关键数据流、落地约束，以及与需求文档之间的对应关系。

## 1. 范围与目标

`offline-package app` 的目标是：

- 从 `VersionSet` 清单导出应用级离线包
- 离线包中包含：
  - `manifest.yaml`
  - `charts/`
  - `images/` OCI Layout
- 将离线包重新导入到目标 registry 与 ChartMuseum

当前实现严格限定在 `proton-cli/cmd/proton-cli/cmd/offline_package` 目录内，不复用其他目录的现成 CR 推送或 chart 处理逻辑。

## 2. 代码组织

当前实现文件：

- `../main.go`
  - 注册 `app` 子命令
- `../app.go`
  - 注册 `export` / `import`
- `../app_types.go`
  - 数据结构定义
- `../app_manifest.go`
  - manifest 解析、平台归一化、导出 manifest 组装
- `../app_export.go`
  - 导出主流程、chart 下载、镜像提取、镜像导出
- `../app_import.go`
  - 导入主流程、包解压、镜像推送、chart 上传
- `../app_oras.go`
  - ORAS 认证 client
- `../app_test.go`
  - 单元测试

## 3. 命令设计落地

### 3.1 `export`

当前支持：

- `--manifest`, `-f`
- `--output`, `-o`
- `--platform`
- `--ignore-missing-images`
- `--disable-dependencies`
- `--override-registry`

实际语义：

- `--platform`
  - 支持单平台或逗号分隔的平台列表
  - 当前支持：
    - `linux/amd64`
    - `linux/arm64`
- `--ignore-missing-images`
  - 某些镜像无法拉取时，记录 warning 并继续
- `--disable-dependencies`
  - 关闭 `dependencies` 递归解析
  - 仅导出根清单中的 `releases`
- `--override-registry`
  - 等价于导出时重写 chart values 中的 `.image.registry`

### 3.2 `import`

当前支持：

- `--input`, `-i`
- `--auto`
- `--registry`
- `--registry-username`
- `--registry-password`
- `--registry-plain-http`
- `--chartmuseum-url`
- `--chartmuseum-username`
- `--chartmuseum-password`
- `--force`

实际语义：

- `--auto`
  - 通过当前 kubeconfig 连接 Kubernetes
  - 读取当前 Proton 集群保存的 cluster config
  - 自动补全 registry、registry 认证、ChartMuseum 地址、ChartMuseum 认证
  - 若显式参数已提供，则显式参数优先
  - 若当前 chart 仓库不是 ChartMuseum，则直接报错
- `--registry-plain-http`
  - 控制推送镜像时对目标 registry 使用 HTTP 还是 HTTPS
  - `--auto` 下仅在外部 OCI registry 场景继承 `plain_http`
- `--force`
  - 覆盖 ChartMuseum 中已存在的 chart

## 4. `VersionSet` 清单解析

输入清单按 `VersionSet` 解析，当前关注字段：

- `kind`
- `product`
- `version`
- `source.helmRepoUrl`
- `releases`
- `dependencies`

约束：

- `kind` 必须是 `VersionSet`
- `product` 必填
- `version` 必填
- `source.helmRepoUrl` 必填
- `releases` 非空
- `releases[*].chart` 必填
- `releases[*].version` 必填

`dependencies` 当前处理方式：

- 递归读取并校验依赖清单
- 支持本地路径与 `http://` / `https://` URL
- 当父清单来自 URL 时，依赖清单允许使用相对 URL 路径
- 使用已解析后的清单绝对位置去重，避免循环依赖导致重复读取
- 当启用 `--disable-dependencies` 时，只加载根清单，不展开依赖

## 5. Chart 收集与下载

导出时会遍历所有已解析清单，并使用各自的：

- `source.helmRepoUrl`
- `releases[*].chart`
- `releases[*].version`

实现流程：

1. 递归读取主清单及其依赖清单
2. 按每份清单的 `source.helmRepoUrl` 分组下载 Helm repo index
3. 在对应 index 中定位每个 chart+version
4. 下载为 `<chart>-<version>.tgz`
5. 平铺保存到临时 `charts/`

当前不处理：

- subchart 依赖递归下载

## 6. 镜像提取规则

镜像来自 chart 默认 values，而不是 `VersionSet` 顶层。

当前支持的 values 结构：

### 6.1 单镜像

```yaml
image:
  registry: docker.io
  repository: library/nginx
  tag: 1.27.0
```

### 6.2 多镜像

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

当前不支持：

- `global.imageRegistry`
- `digest`
- 模板渲染后的动态镜像地址
- `image` 根节点以外的自定义结构

无法识别的 chart：

- 只告警
- 不直接终止导出

## 7. 镜像命名规则

这是当前实现里最重要的约束之一。

在 values 中：

- `.image.registry` 被视为“完整 registry 地址”
- `.image.repository` 被视为“真正的 repository”

因此导出时有两套镜像名：

### 7.1 原始镜像地址 `source`

原始拉取地址由三部分拼接：

```text
<image.registry>/<image.repository>:<image.tag>
```

例如：

```yaml
image:
  registry: swr.cn-east-3.myhuaweicloud.com/kweaver-ai
  repository: dip/agent-backend
  tag: 0.5.1
```

则原始地址为：

```text
swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/agent-backend:0.5.1
```

### 7.2 离线包内部镜像名 `localRef`

离线包内部必须完整去掉 `.image.registry` 这整段，只保留：

```text
<image.repository>:<image.tag>
```

所以上例离线包内命名为：

```text
dip/agent-backend:0.5.1
```

这保证了：

- 包内镜像命名不携带源 registry
- 即使 `.image.registry` 自身带 path 前缀，也不会污染离线包内镜像名

## 8. `--override-registry` 的实际语义

`--override-registry` 的行为不是简单“前缀替换”，而是：

- 在导出语义上，将 chart values 中的 `.image.registry` 重写为传入值
- 之后继续使用原始 `.image.repository` 与 `.image.tag`

例如：

原 values：

```yaml
image:
  registry: acr.aishu.cn
  repository: dip/agent-backend
  tag: 0.5.1
```

命令：

```bash
--override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

则实际拉取：

```text
swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/agent-backend:0.5.1
```

## 9. 平台处理

平台处理使用 `oras-go` 完成。

### 9.1 支持的输入

- `linux/amd64`
- `linux/arm64`
- 别名：
  - `amd64`
  - `arm64`
  - `x86_64`
  - `aarch64`

### 9.2 单平台导出

若只指定一个平台：

- 若远端 tag 是单架构 manifest，则要求与目标平台匹配
- 若远端 tag 是多架构 index，则只选中该平台对应 manifest

### 9.3 多平台导出

若指定多个平台：

- 对每个平台分别选择 manifest
- 将这些 manifest 重新组装为新的 OCI image index
- 再用原 tag 关联到这个新 index

因此：

- 单平台包不会意外保留多架构 index
- 双平台包只包含显式选定的平台集合

## 10. 镜像导出实现

镜像导出使用 `oras-go`，不依赖 `skopeo` 二进制。

流程：

1. 解析源镜像引用
2. 根据 `--platform` 选择匹配 manifest
3. 使用 `oras.CopyGraph` 复制内容到本地 OCI Layout
4. 为单平台 manifest 或重新组装后的 index 打本地 tag

OCI Layout 内 tag 命名规则：

```text
<image.repository>:<image.tag>
```

## 11. 导出包结构

导出包结构：

```text
offline-app-package.tar
├── manifest.yaml
├── images/
└── charts/
```

其中：

- `charts/`
  - 平铺保存 chart tgz
- `images/`
  - 单一 OCI Layout
- `manifest.yaml`
  - 原始 `VersionSet` + 导出元数据

## 12. 导出后的 `manifest.yaml`

导出后的 `manifest.yaml` 增补如下元数据：

```yaml
offlinePackage:
  platform: linux/amd64
  platforms:
    - linux/amd64
  overrideRegistry: swr.cn-east-3.myhuaweicloud.com/kweaver-ai
  exportedAt: ...
  images:
    - source: ...
      pullSource: ...
      repository: ...
      localRef: ...
      requestedPlatforms: ...
      exportedPlatforms: ...
      exported: true
  imageErrors:
    - ...
```

字段说明：

- `platform`
  - 单平台导出时保留，便于兼容旧语义
- `platforms`
  - 本次导出的完整平台列表
- `overrideRegistry`
  - 本次导出对 `.image.registry` 使用的覆盖值
- `source`
  - 原始 values 识别出的镜像地址
- `pullSource`
  - 实际用于拉取的镜像地址
  - 可能受 `--override-registry` 影响
- `repository`
  - 离线包内部 repository 名
- `localRef`
  - OCI Layout 内 tag
- `requestedPlatforms`
  - 请求导出的平台
- `exportedPlatforms`
  - 实际导出的平台
- `exported`
  - 是否成功进入离线包
- `imageErrors`
  - 被忽略的镜像错误列表

## 13. 导入实现

导入前会：

1. 解压 tar
2. 校验存在：
   - `manifest.yaml`
   - `charts/`
   - `images/`

### 13.1 镜像导入

镜像导入使用 `oras-go`：

- 从 OCI Layout 中读取所有 tag
- 逐个推送到目标 registry

目标命名规则：

```text
<registry>/<localRef>
```

例如：

- `localRef = dip/agent-backend:0.5.1`
- `registry = internal.example.com:5000`

则推送目标为：

```text
internal.example.com:5000/dip/agent-backend:0.5.1
```

### 13.2 Chart 导入

Chart 导入通过 ChartMuseum 接口逐个上传：

- 遍历 `charts/*.tgz`
- 上传到目标 ChartMuseum

`--force` 用于覆盖已存在的 chart。

## 14. 错误处理策略

### 14.1 直接失败

以下情况直接失败：

- 清单读取失败
- 清单 YAML 解析失败
- 清单字段缺失
- chart 下载失败
- tar 写入失败
- 导入包结构缺失
- 任一镜像导入失败
- 任一 chart 上传失败

### 14.2 可降级为 warning

以下情况可被降级：

- chart 无法识别镜像结构
- 已识别镜像但无法拉取

降级条件：

- 必须显式指定 `--ignore-missing-images`

降级后的行为：

- 记录 warning
- 写入 `manifest.yaml`
- 继续导出剩余镜像

## 15. 当前与 `FEATURE.md` 的对照

### 15.1 已实现

- 本地文件和 HTTP(S) manifest 输入
- `VersionSet` 必填字段校验
- 忽略 `dependencies`
- 按 `source.helmRepoUrl` 下载 chart
- 从 values 提取两类镜像结构
- 镜像去重
- 使用 OCI Layout 保存镜像
- 导入到 OCI registry
- 上传到 ChartMuseum
- `--platform` 平台选择
- 双平台导出
- `--ignore-missing-images`
- `--force`
- `--override-registry`

### 15.2 当前未实现

- `dependencies` 递归展开
- 更复杂的 values 模板求值
- `global.imageRegistry`
- `digest` 形式镜像
- Helm subchart 依赖分析
- 导入前 dry-run
- 并发度配置
- registry 认证参数用于 export 侧拉取私有镜像

## 16. 当前已验证结果

已验证的真实场景：

- `kweaver-core.yaml` 能成功下载全部 chart
- 默认情况下原始 `acr.aishu.cn` 镜像会失败
- 通过：

```bash
--override-registry swr.cn-east-3.myhuaweicloud.com/kweaver-ai
```

可将示例清单中的镜像导出提升到 `33/33`

同时已验证：

- 单架构导出
- 双架构导出
- 导入到本地 registry
- 导入到本地 ChartMuseum

## 17. 维护建议

后续如果继续扩展，建议优先处理：

1. export 侧 registry 认证
2. 更复杂 values 镜像结构识别
3. `dependencies` 递归展开
4. 平台选择结果与导入结果的回归测试
