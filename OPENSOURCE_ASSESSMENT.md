# Proton 开源部署系统 — 从零到运行的架构设计

> **设计目标**：以 `proton-cli` 为入口，对标 K3s/KubeKey/KubeSphere 等业界顶级开源项目的一键启动体验，设计 Proton 部署系统从裸机到可运行的完整流程，支持在线/离线可选。

---

## 一、业界对标分析

### 顶级开源项目的 Bootstrap 模式

| 项目 | 入口 | 一键命令 | 离线支持 | Web UI |
|------|------|---------|---------|--------|
| **K3s** | 单二进制 `k3s` | `curl -sfL \| sh` → 30s 启动 | ✅ airgap images tarball | Rancher 单独安装 |
| **KubeKey** | 单二进制 `kk` | `kk create cluster` | ✅ `kk artifact` 打离线包 | KubeSphere 随后安装 |
| **KubeSphere** | KubeKey + Web | 安装完自带 Web Console | ✅ manifest + artifact | ✅ 内置 |
| **Rancher** | Helm Chart | `helm install rancher` | ✅ 镜像同步 + 私有 registry | ✅ 内置 |
| **Sealos** | 单二进制 `sealos` | `sealos run kubernetes:v1.28` | ✅ OCI 集群镜像 | ✅ 内置 |

### 共同特征

1. **单一入口**：一个二进制 / 一条命令启动全部
2. **渐进式**：最小安装 → 按需扩展
3. **离线即一等公民**：离线包格式标准化（OCI/tar），不是事后补丁
4. **配置即声明**：YAML 配置文件描述期望状态，工具执行收敛
5. **内置 Web UI**：安装完成后立即有管理界面

---

## 二、Proton 现有架构的能力盘点

### proton-cli 已具备的能力（核心优势）

经过深入分析 `proton-cli` 源码，它已经是一个**成熟的集群生命周期管理工具**：

```
proton-cli
├── server       # 🌟 内置 Web 向导 (serve.go, //go:embed web)
│   ├── /        # 静态 Web UI（向导式配置界面）
│   ├── /init    # POST JSON → apply.Apply(conf) 执行完整初始化
│   └── /alpha/result  # GET 查询初始化结果
├── apply -f     # 声明式配置应用（conf.yaml → 集群收敛）
├── reset        # 集群重置
├── push-images  # 离线镜像推送（OCI 格式 → skopeo copy）
├── push-charts  # 离线 chart 推送（Chartmuseum/OCI）
├── component    # 数据组件管理
├── get conf     # 查看当前集群配置
├── kubernetes   # K8s 管理（calico 升级等）
├── backup       # 集群备份
└── recover      # 集群恢复
```

### Apply 执行链（必选 → 可选模块顺序）

```
必选模块（按依赖顺序）：
  1. firewall    — 防火墙规则配置
  2. nodes       — 节点准备（SSH/ECMS agent）
  3. cr          — Container Registry（本地/外部）
  4. cs          — Container Service（K8s 集群）

可选模块（K8s 就绪后）：
  5. component_manage  — 基础组件管理服务
  6. nvidia_device_plugin
  7. proton_mq_nsq     — 消息队列
  8. prometheus        — 监控
  9. grafana           — 可视化
  10. package_store    — 包存储服务
  11. eceph            — 分布式存储（最后执行）
```

### 两种部署模式（已支持）

| 模式 | 配置模板 | CR | CS | 数据服务 |
|------|---------|----|----|---------|
| **Internal（一体机）** | `templateInternal.yaml` | 本地 Registry + Chartmuseum | 本地 K8s (kubeadm) | 本地 MariaDB/Redis/MongoDB/... |
| **External（托管云）** | `templateExternal.yaml` | 外部 Registry + OCI/Chartmuseum | 外部 K8s | 外部数据服务（可全部外置） |

### 关键差距

| 能力 | 现状 | 差距 |
|------|------|------|
| Web 向导 | ✅ `proton-cli server` 已有架构 | ❌ Web 前端在内部 FTP，未入开源仓库 |
| 单二进制分发 | ✅ Go 编译单二进制 | ❌ 依赖内部 CI 构建 |
| 在线安装 | ⚠️ 框架存在，但镜像仓库全部内部 | ❌ 公共仓库未配置 |
| 离线安装 | ✅ push-images/push-charts 完整 | ⚠️ 离线包构建在内部流水线 |
| proton-cli → 上层平台 | ⚠️ CMS/component-manage 已有 | ❌ deploy-service/deploy-web 安装未自动化 |
| Go module 路径 | `devops.aishu.cn/...` 内部路径 | ❌ 无法公网 go get |
| 开源产品 Helm Chart 支持 | ✅ helm-repo 已有 100+ chart | ⚠️ 与 proton 的 chart 仓库未打通 |

---

## 三、从零到运行的完整架构设计

### 总体流程图

```
┌──────────────────────────────────────────────────────────────────────┐
│                    用户的裸金属服务器 / 云服务器                         │
│                                                                      │
│  ① 获取 proton-cli                                                   │
│     在线: curl -sfL https://get.proton.kweaver.ai | sh               │
│     离线: 解压 proton-offline-bundle.tar.gz                           │
│                                                                      │
│  ② 启动 Web 向导                                                      │
│     $ proton-cli server --port 8888                                   │
│     浏览器访问 http://<IP>:8888 → 图形化配置向导                        │
│                                                                      │
│  ② (备选) CLI 方式                                                    │
│     $ proton-cli apply -f cluster.yaml                                │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │              proton-cli apply 执行链                            │  │
│  │                                                                │  │
│  │  [Stage 1: 基础设施]                                           │  │
│  │  firewall → nodes → cr(registry) → cs(kubernetes)              │  │
│  │                                                                │  │
│  │  [Stage 2: 数据服务]  (内置 or 外置可选)                        │  │
│  │  component-manage → MariaDB → Redis → MongoDB → ...            │  │
│  │                                                                │  │
│  │  [Stage 3: Proton 平台]  ← 🆕 需要新增                         │  │
│  │  deploy-service → deploy-web → deployrunner                    │  │
│  │                                                                │  │
│  │  [Stage 4: 开源产品]  (通过 Proton Web UI 部署)                 │  │
│  │  KWeaver / 其他 Helm Chart 产品                                 │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ③ 访问 Proton Web UI                                                │
│     https://<IP> → 系统工作台 → 管理/部署/升级上层产品                  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### 3.1 Stage 0：获取 proton-cli（入口点）

#### 在线模式

对标 K3s 的 `curl | sh` 体验：

```bash
# 一条命令下载并安装 proton-cli
curl -sfL https://get.proton.kweaver.ai | sh

# 或者从 GitHub Releases 直接下载
curl -LO https://github.com/kweaver-ai/proton/releases/latest/download/proton-cli-linux-amd64.tar.gz
tar xzf proton-cli-linux-amd64.tar.gz
sudo mv proton-cli /usr/local/bin/
```

**需要交付的产物**：
- `proton-cli` 单二进制文件（Go 编译，内嵌 Web 向导前端 + 配置模板）
- `install.sh` 安装脚本（检测架构，下载对应二进制）
- GitHub Releases 多架构发布（linux/amd64, linux/arm64）

#### 离线模式

对标 KubeKey 的 artifact 模式：

```bash
# 在有网环境打离线包
proton-cli artifact create --config manifest.yaml -o proton-offline-bundle.tar.gz

# 传输到离线环境，解压即用
tar xzf proton-offline-bundle.tar.gz
cd proton-offline-bundle/
./proton-cli server --port 8888 --service-package ./service-package
```

**离线包内容结构**（标准化）：

```
proton-offline-bundle/
├── proton-cli                    # 二进制
├── service-package/
│   ├── images/                   # OCI 格式镜像（skopeo 推送）
│   │   ├── index.json
│   │   └── blobs/
│   └── charts/                   # Helm chart tgz 文件
│       ├── component-manage-*.tgz
│       ├── deploy-service-*.tgz
│       ├── deploy-web-*.tgz
│       ├── mariadb-*.tgz
│       ├── redis-*.tgz
│       └── ...
├── templates/
│   ├── cluster-internal.yaml     # 一体机模板
│   └── cluster-external.yaml     # 托管云模板
└── README.md
```

### 3.2 Stage 1：Web 向导初始化（proton-cli server + proton-cli-web）

#### 后端骨架（proton-cli serve.go — 已完成）

- `//go:embed web` 嵌入静态前端到 Go 二进制
- `POST /init?accout=&password=` 接收 JSON ClusterConfig，调用 `apply.Apply(conf)`
- `GET /alpha/result` 轮询初始化状态（running / success / fail）
- 初始化成功后 30s 自动关闭 HTTP server

#### 前端项目（proton-cli-web — ✅ 已入库）

`proton-cli-web/` 是一个**功能完整的安装向导前端**，已经具备生产级能力：

| 维度 | 详情 |
|------|------|
| **技术栈** | React 18 + TypeScript + webpack 5 + SASS |
| **UI 库** | `@aishutech/ui`（Ant Design 封装）+ `@anyshare/i18nfactory` |
| **源码规模** | 110 个文件，核心逻辑约 4500 行 |
| **路由** | `/` → 部署向导，`/success` → 成功页 |
| **构建** | `npm run build` → `dist/` → 嵌入 Go 二进制 |
| **开发** | `npm run serve` → express mock server (port 3000) |
| **CI** | Azure Pipelines → Docker build → FTP 上传 |

#### 向导流程（已实现的 5 步）

```
┌─────────────────────────────────────────────────────────────────┐
│  Proton 部署工具  (http://<IP>:8888)                             │
│                                                                   │
│  Step 0: 选择部署模式 (ChooseTemplate)                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │
│  │ 标准模式部署    │  │ 云主机部署    │  │托管K8s部署    │            │
│  │ (一体机/物理机) │  │ (ECS云主机)  │  │ (已有K8s)    │            │
│  └──────────────┘  └──────────────┘  └──────────────┘            │
│  + 产品型号选择（当前为 AnyShare 型号，需开源化）                      │
│                                                                   │
│  Step 1: 节点配置 (NodeConfig)                                     │
│  - 节点列表（name/IPv4/IPv6/内部IP）                                │
│  - SSH 账户/密码                                                   │
│  - Chrony 时间同步（系统默认/内置NTP/外部NTP）                       │
│  - 防火墙模式（firewalld/用户自管）                                  │
│  - IPv4/IPv6/双栈选择                                              │
│                                                                   │
│  Step 2: Kubernetes 配置 (NetworkConfig)                           │
│  - Master 节点选择                                                 │
│  - Pod/Service CIDR、Docker BIP                                    │
│  - etcd/docker 数据目录                                            │
│  - K8s 插件（node-exporter, kube-state-metrics）                   │
│  - 部署命名空间/服务账号                                             │
│                                                                   │
│  Step 3: 仓库配置 (CRConfig)                                      │
│  - 本地 CR：节点选择/端口/HA端口/存储路径                             │
│  - 外部 CR：Registry/Chartmuseum/OCI 三选                          │
│                                                                   │
│  Step 4: 基础服务配置 (DataBaseConfig)                              │
│  - 15+ 可配服务：MariaDB/MongoDB/Redis/NSQ/OpenSearch/             │
│    Kafka/ZooKeeper/PolicyEngine/ETCD/Prometheus/Grafana/           │
│    Nebula/ProtonMonitor/PackageStore/ECeph                         │
│  - 每个服务：节点选择/数据路径/资源配额/副本数                         │
│  - 服务可动态添加/删除                                               │
│                                                                   │
│  Step 5: 连接配置 (ConnectInfo)                                    │
│  - 每个数据服务：内置(internal) vs 外置(external)                    │
│  - 外置：填写 host/port/username/password                          │
│  - 支持多种外置类型：RDS(MySQL/DM8/GoldenDB/TiDB/...)              │
│    Redis(哨兵/主从/单机/集群) MQ(NSQ/Kafka/TongLink/...)           │
│                                                                   │
│  [完成] → POST /init → apply.Apply() → /success 页面               │
└─────────────────────────────────────────────────────────────────┘
```

#### 数据流

```
前端 ConfigData → exChangeData() 转换 → ClusterConfig JSON
    → POST /init?accout=xxx&password=xxx
    → serve.go initialHandler
    → configuration.LoadFromBytes(body)
    → apply.Apply(conf)
    → 轮询 GET /alpha/result → 成功跳转 /success
```

#### proton-cli-web 的内部依赖问题

| 问题 | 位置 | 影响 | 改造方案 |
|------|------|------|---------|
| **`@aishutech/ui`** npm 包 | `package.json` + 全部组件 | 🔴 构建阻断 | 发布到公共 npm 或替换为 `antd` |
| **`@anyshare/i18nfactory`** | `core/i18n/index.ts` | 🔴 构建阻断 | 替换为 `react-i18next` 或类似 |
| **AnyShare 产品型号** | `helper.ts` DEVICESPECS | 🟡 用户困惑 | 移除或改为通用 "Standard/Enterprise" |
| **默认用户名 "anyshare"** | `helper.ts` DEFAULT_INTERNAL_RDS/MONGODB | 🟡 品牌残留 | 改为 "proton" 或 "admin" |
| **OpenSearch AnyShare URL** | `helper.ts` hanlpRemoteextDict | 🟡 无效默认值 | 清空或移除 |
| **`/anyshare/opensearch`** 路径 | `helper.ts` DefaultConfigData | 🟡 品牌残留 | 改为 `/sysvol/opensearch` |
| **"部署AnyShare场景"** UI 文字 | `ChooseTemplate` | 🟡 品牌残留 | 改为通用提示 |
| **中文 only** | 全部 UI 文字 | 🟢 体验 | 添加英文 i18n |
| **Azure Pipelines CI** | `azure-pipelines.yml` | 🟡 构建 | 迁移到 GitHub Actions |

#### proton-cli-web 的架构优势

1. **功能完整** — 覆盖 proton-cli 全部配置项，3 种部署模式全支持
2. **前后端解耦** — 纯静态 SPA，只通过 `/init` 和 `/alpha/result` 两个 API 与后端通信
3. **数据转换层清晰** — `exChangeData()` 将 UI 数据模型转为 ClusterConfig JSON
4. **内置 mock 服务器** — `express.js` + `payload.js` 开发调试方便
5. **表单校验完善** — IP 格式、CIDR 范围、节点名称正则、重复检测等
6. **服务可动态增删** — 非必须服务可按需添加（Prometheus/Grafana/Nebula 等）

#### proton-cli-web 需要改进的点

1. **Class 组件** — 使用 React Class Component（已不推荐），核心逻辑集中在 `component.base.tsx`（1631 行）和 `helper.ts`（1953 行），后续可渐进式迁移到 Hooks
2. **缺少安装进度反馈** — 当前只有 Spin "初始化中, 请耐心等待..."，无法看到执行到哪个阶段（firewall/cr/cs），需要扩展 `/alpha/result` 返回详细进度
3. **服务过多** — 默认展示 15+ 服务配置，对首次使用者信息过载。建议分为"核心必选"和"高级可选"两组
4. **未集成 Proton Platform 安装** — 当前向导完成后 K8s + 数据服务就绪，但 deploy-service/deploy-web 的安装需要手动。应新增 Stage 3

#### 构建集成方案

proton-cli-web 已在仓库内，与 proton-cli 的集成路径清晰：

```bash
# 1. 构建前端
cd proton-cli-web && npm install && npm run build

# 2. 将产物复制到 Go embed 目录
cp -r dist/* ../proton-cli/cmd/proton-cli/cmd/web/

# 3. 构建 Go 二进制（自动嵌入前端）
cd ../proton-cli && go build -o proton-cli ./cmd/proton-cli/

# proton-cli 单二进制内含完整 Web 向导
```

在 CI 中（GitHub Actions）自动化此流程，确保每次 Release 的 `proton-cli` 二进制都内嵌最新 Web 向导。

### 3.3 Stage 2-3：Apply 执行链的开源化改造

当前 apply 链已完善，但有几个必须调整的点：

#### 3.3.1 新增 "Proton 平台" 安装阶段

当前 apply 链在安装完 K8s + 数据服务后就结束了。需要在 `appendOptionalModules` 中新增 **Proton 平台自身的安装**：

```go
// 新增 Proton Platform 模块（在 component_manage 之后）
if clusterConf.ProtonPlatform != nil {
    modules = append(modules, module{
        name: "proton_platform",
        applier: protonplatform.NewManager(
            helm3, clusterConf, registry.Address(),
            global.ServicePackage, charts,
        ),
    })
}
```

这个模块负责安装：
1. `deploy-service` — Helm install
2. `deploy-web` — Helm install（含 BFF + 前端静态资源）
3. `deployrunner` — Helm install
4. Ingress 规则配置
5. 数据库初始化（`init.sql`）

#### 3.3.2 配置模板新增 Proton Platform 段

在 `templateInternal.yaml` 和 `templateExternal.yaml` 中新增：

```yaml
# Proton 部署管理平台（新增）
proton_platform:
  # 安装后即可通过 Web UI 管理上层产品
  enabled: true
  # Helm chart 仓库（在线模式使用公共仓库）
  chart_repo: https://kweaver-ai.github.io/helm-repo/
  # 访问端口
  ingress:
    host: ""  # 留空则使用 IP 访问
    tls: true
  # 认证模式
  auth:
    mode: local  # local | oauth2
    # local 模式使用内置用户
    admin_user: admin
    admin_password: ""  # 留空自动生成
```

#### 3.3.3 resource_connect_info 精简

当前 `resource_connect_info` 包含大量只有 AnyShare 才需要的组件。开源版精简为：

```yaml
resource_connect_info:
  rds:
    source_type: internal    # MariaDB（deploy-service/deployrunner 必须）
  redis:
    source_type: internal    # Redis（deploy-web session 必须）
```

MongoDB、OpenSearch、Kafka、ZooKeeper、PolicyEngine、etcd 等改为**按需安装**，由上层产品通过 Proton Web UI 自行配置。

### 3.4 Stage 4：通过 Proton Web UI 部署上层产品

安装完成后，用户访问 Proton Web UI，可以：

1. **添加 Helm Chart 仓库**（如 `https://kweaver-ai.github.io/helm-repo/`）
2. **浏览可用产品**（KWeaver Studio、Ontology、Agent Operator 等）
3. **一键安装**（基于 deployrunner 的 DAG 任务编排）
4. **管理升级/回滚**

这与 Rancher 管理应用的体验一致。

---

## 四、必须完成的开源化改造清单

### 🔴 P0 — 阻断性（不改无法运行）

#### 4.1 Go Module 路径迁移

当前所有 Go 代码的 import 路径为内部 Azure DevOps：

```
devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/...
```

**必须迁移为**：
```
github.com/kweaver-ai/proton/proton-cli/v3/...
```

涉及文件：
- `proton-cli/go.mod` — module 声明
- `proton-cli/` 下所有 `.go` 文件的 import 路径（约 200+ 文件）
- `deployrunner/go.mod` — `component-manage v0.0.0` replace 指令
- `component-manage/go.mod` — module 声明

**执行方式**：`sed` 全局替换 + `go mod tidy` 验证

#### 4.2 内部域名/仓库替换

| 替换项 | 当前值 | 目标值 |
|--------|--------|--------|
| 镜像仓库 | `acr.aishu.cn` | `ghcr.io/kweaver-ai` 或 `docker.io/kweaver` |
| npm registry | `registry.npm.aishu.cn` | `registry.npmjs.org` |
| pip 源 | `mirrors.huaweicloud.com` | 删除（使用默认 PyPI） |
| Go proxy | `goproxy.aishu.cn` | `https://proxy.golang.org` |
| Chartmuseum | `chartmuseum.aishu.cn:15001` | 配置化 |
| CI 镜像 | `acr.aishu.cn/library/golang` | `golang:1.25` (Docker Hub) |
| proton-cli-web FTP 下载 | `cmd/download-proton-cli-web` 引用 `ftp-ict.aishu.cn` | 删除此工具，改为本地构建集成 |

#### 4.3 proton-cli-web 内部 npm 包替换（✅ 源码已入库）

`proton-cli-web/` 源码已在仓库内，不再需要从 FTP 迁出。但有 **2 个内部 npm 包** 阻断构建：

**🔴 必须解决（构建阻断）**：

| npm 包 | 引用位置 | 方案 |
|--------|---------|------|
| `@aishutech/ui` ^0.0.2 | `package.json` + 全部 25+ 组件文件 | **方案 A**（推荐）：发布到公共 npmjs.com；**方案 B**：替换为 `antd`（API 兼容） |
| `@anyshare/i18nfactory` ^1.0.0 | `core/i18n/index.ts` | 替换为 `react-i18next` 或内联简单 i18n 函数 |

**🟡 需要清理（AnyShare 残留）**：

| 位置 | 内容 | 改造 |
|------|------|------|
| `helper.ts` DEVICESPECS | 30+ AnyShare 产品型号（AS10000/AS9000/smas...） | 移除整个产品型号选择器，开源不需要 |
| `helper.ts` DEFAULT_INTERNAL_RDS | `username: "anyshare"` | 改为 `"proton"` |
| `helper.ts` DEFAULT_INTERNAL_MONGODB | `username: "anyshare"` | 改为 `"proton"` |
| `helper.ts` DefaultConfigData opensearch | `hanlpRemoteextDict: "http://ecoconfig-private.anyshare:32128/..."` | 清空为 `""` |
| `helper.ts` DefaultConfigData opensearch | `data_path: "/anyshare/opensearch"` | 改为 `"/sysvol/opensearch"` |
| `ChooseTemplate` | `"部署AnyShare场景选择"` | 改为 `"请选择产品型号（可选）"` |
| `payload.js` mock 数据 | `ecoconfig-private.anyshare` + `/anyshare/opensearch` | 同步更新 |
| `azure-pipelines.yml` | Azure Pipelines + FTP 上传 | 迁移到 GitHub Actions |

**构建集成**：`npm run build` → `dist/` → 复制到 `proton-cli/cmd/proton-cli/cmd/web/` → `go build` 自动嵌入。

#### 4.4 AnyShare 产品引用清除（含 proton-cli-web）

全项目（含 proton-cli-web）的 AnyShare 引用汇总：

| 位置 | 操作 |
|------|------|
| `deploy-service/main.py` → `init_anyshare_cms()` | 改为通用 `init_cms()` 或条件禁用 |
| `deploy-web/deploy-web-service/conf/anyshare.js` | 重命名为 `default.js`，移除 AnyShare 品牌 |
| `migrations/init.sql` 中 iOS App Store 链接 | 删除该 INSERT 语句 |
| `proton-cli-web/helper.ts` DEVICESPECS/DEVICESPECSMAP | 移除 AnyShare 产品型号 |
| `proton-cli-web/helper.ts` opensearch 配置中 `ecoconfig-private.anyshare` | 清空 URL |
| `proton-cli-web/helper.ts` 默认用户名 `"anyshare"` | 改为 `"proton"` |
| `proton-cli-web/ChooseTemplate` UI 文字 | 移除 "部署AnyShare场景选择" |
| 前端 OEM 配置中的品牌文字 | 替换为 "Proton System Console" |

#### 4.5 认证模式降级

deploy-web 的 auth.ts 深度绑定 Hydra OAuth2。必须增加 **local auth** 模式：

```typescript
// auth.ts 中新增
if (authMode === 'local') {
    // 使用简单的 session + 本地用户表验证
    // 不依赖 Hydra/user-management/eacp
    return localAuth(req, res);
}
```

BFF 层 `handlers/tools/index.js` 中 20+ 硬编码服务需要**按实际部署的服务动态加载**。

#### 4.6 数据库自动初始化

deploy-service 和 deployrunner 的 SQL 需要在 Helm install 时自动执行：
- 在 Helm chart 中添加 `initContainers` 或 Job 执行 SQL
- 精简 `init.sql`，移除 AnyShare 特有数据

---

### 🟡 P1 — 快速启动必须

#### 4.7 CI/CD 迁移到 GitHub Actions

需要新建：
- `.github/workflows/build-proton-cli.yml` — 构建 proton-cli 多架构二进制
- `.github/workflows/build-images.yml` — 构建 Docker 镜像推送到 GHCR
- `.github/workflows/build-charts.yml` — 打包 Helm chart 到 helm-repo
- `.github/workflows/release.yml` — 发布离线包到 GitHub Releases

#### 4.8 Helm Chart 标准化

- 所有 `Chart.yaml` 升级到 `apiVersion: v2`
- 移除 namespace 硬编码（`anyshare`/`proton`/`resource`）
- 默认 `image.registry` 改为公共仓库
- 添加 `values.schema.json` + `NOTES.txt`

#### 4.9 前端依赖开源化

- 确认 `@kweaver-ai/ui`、`@kweaver-ai/template`、`@kweaver-ai/workshop-framework-system` 已发布到公共 npmjs.com
- 重新 `npm install` 生成干净的 `package-lock.json`（当前含 4500+ 处内部 registry 引用）

#### 4.10 ecms 模块处理

`ecms/` 是 AnyShare ECMS Agent 代码，开源不需要。同时 proton-cli 中多处通过 ECMS agent 做远程执行（SSH），开源场景应改为直接 SSH：

| 当前方式 | 开源方案 |
|---------|---------|
| `ecms/v1alpha1.NewForHost(ip)` → ECMS HTTP API | 直接 SSH 执行 |
| ECMS agent 部署在每个节点 | 不需要 agent |

短期可保留 ECMS 兼容层，长期应迁移到纯 SSH。

---

### 🟢 P2 — 生产就绪

#### 4.11 proton-cli 新增快速启动命令

对标 K3s 的极简体验：

```bash
# 最简单的一体机部署（自动检测本机IP，全部使用默认配置）
proton-cli quickstart

# 等同于：
# 1. 生成默认 cluster.yaml（单节点，内置全部）
# 2. proton-cli apply -f cluster.yaml
# 3. 输出访问地址
```

#### 4.12 统一包管理与 helm-repo

当前 `kweaver/helm-repo/packages/` 已有 100+ 个 chart（studio-web、agent-backend、hydra 等）。Proton 部署上层产品时，应默认对接这个仓库：

```yaml
# proton_platform.chart_repos 默认值
chart_repos:
  - name: kweaver
    url: https://kweaver-ai.github.io/helm-repo/
    type: helm
```

#### 4.13 文档体系

```
docs/
├── quickstart.md              # 5 分钟快速开始
├── architecture.md            # 架构说明 + 组件关系图
├── installation/
│   ├── online.md              # 在线安装指南
│   ├── offline.md             # 离线安装指南
│   └── cloud.md               # 托管云安装指南
├── configuration/
│   ├── cluster-config.md      # 集群配置参考
│   └── templates.md           # 配置模板说明
├── user-guide/
│   ├── deploy-product.md      # 通过 Web UI 部署产品
│   ├── upgrade.md             # 升级指南
│   └── backup-restore.md      # 备份恢复
└── contributing.md             # 贡献指南
```

---

## 五、开源化改造的优先级路线图

### Phase 1：可构建可运行（2 周）

目标：`proton-cli apply -f cluster.yaml` 能在裸机上从零部署出完整 Proton 平台，Web 向导可用。

| # | 任务 | 天 | 说明 |
|---|------|-----|------|
| 1 | Go module 路径迁移 | 2 | 全局 sed + go mod tidy |
| 2 | 内部域名/仓库替换 | 3 | 镜像/npm/pip/GoProxy/FTP下载工具删除 |
| 3 | proton-cli-web npm 包替换 | 2 | `@aishutech/ui` → antd, `@anyshare/i18nfactory` → react-i18next |
| 4 | proton-cli-web AnyShare 清理 | 1 | DEVICESPECS 移除 + 默认用户名/URL/路径替换 |
| 5 | 其他模块 AnyShare 引用清除 | 2 | deploy-service/deploy-web 共 73+ 文件 |
| 6 | 新增 `proton_platform` apply 模块 | 3 | 自动安装 deploy-service/web/runner |
| 7 | 本地认证模式 | 2 | 绕过 Hydra 依赖 |

### Phase 2：开源标准化（2 周）

目标：完整的在线/离线安装体验，GitHub 可构建，CI 自动化。

| # | 任务 | 天 | 说明 |
|---|------|-----|------|
| 8 | GitHub Actions CI/CD | 3 | 含 proton-cli-web 构建 → Go embed → 多架构二进制发布 |
| 9 | Helm Chart 标准化 | 2 | apiVersion v2 + 公共仓库 |
| 10 | deploy-web npm 包开源 | 3 | @kweaver-ai/* 发布到 npmjs + 重生成 lock |
| 11 | BFF 服务依赖配置化 | 2 | 20+ 硬编码服务地址 |
| 12 | 数据库自动初始化 | 1 | Helm Job + 精简 init.sql |
| 13 | install.sh 安装脚本 | 1 | curl \| sh 体验 |

### Phase 3：生产就绪（2 周）

目标：文档完善、离线包 GitHub Releases 发布、quickstart 命令、安装进度优化。

| # | 任务 | 天 | 说明 |
|---|------|-----|------|
| 14 | `proton-cli quickstart` 命令 | 2 | 一键单节点部署 |
| 15 | 离线包构建 + Releases 发布 | 3 | artifact create 流水线 |
| 16 | 对接 helm-repo 开源 chart | 2 | 默认仓库配置 |
| 17 | proton-cli-web 安装进度增强 | 2 | `/alpha/result` 返回分阶段进度，前端展示各模块状态 |
| 18 | proton-cli-web 英文 i18n | 2 | 中英双语支持 |
| 19 | 完整文档体系 | 3 | quickstart + 架构 + 配置参考 |
| 20 | ecms → SSH 迁移（可选） | 3 | 去除 agent 依赖 |

---

## 六、最终目标用户体验

### 在线一键部署（5 分钟）

```bash
# 1. 安装 proton-cli（10 秒）
curl -sfL https://get.proton.kweaver.ai | sh

# 2. 一键部署（单节点，自动检测环境）
sudo proton-cli quickstart

# 输出:
# ✅ Firewall configured
# ✅ Container Registry started (localhost:5000)
# ✅ Kubernetes cluster initialized (v1.28)
# ✅ MariaDB installed
# ✅ Redis installed
# ✅ Proton Platform deployed
#
# 🎉 Proton is ready!
# Access Web Console: https://10.0.0.1
# Admin User: admin
# Admin Password: <auto-generated>
#
# To deploy KWeaver products:
#   1. Login to Web Console
#   2. Add Helm Repo: https://kweaver-ai.github.io/helm-repo/
#   3. Browse and install products
```

### 离线部署

```bash
# 在有网机器上打包
proton-cli artifact create -o proton-bundle.tar.gz

# 传输到离线环境
scp proton-bundle.tar.gz user@offline-server:~/

# 在离线环境安装
ssh user@offline-server
tar xzf proton-bundle.tar.gz && cd proton-bundle
sudo ./proton-cli quickstart --service-package ./service-package
```

### Web 向导部署（企业用户）

```bash
# 启动 Web 向导
sudo proton-cli server --port 8888

# 浏览器打开 http://<IP>:8888
# 按向导配置：多节点、外部数据库、高可用等
# 点击 "开始安装" → 实时查看进度
```

---

## 七、前端项目专项评估

### deploy-web（系统工作台）

| 子项目 | 技术栈 | 核心功能 |
|--------|--------|---------|
| `deploy-web-static` | React 18 + Antd + webpack | 前端 UI：服务管理/套件管理/组件管理/站点配置 |
| `deploy-web-service` | Node.js Express + TypeScript | BFF：OAuth 认证代理/API 代理/Redis Session |

### 关键阻断

1. **@kweaver-ai/* npm 包**：3 个核心包必须在公共 npm 可用
2. **package-lock.json**：4500+ 处内部 npm registry 引用
3. **BFF 20+ 服务硬编码**：`ossgateway`, `license`, `audit-log`, `eacp`, `ShareMgnt` 等开源不存在
4. **OAuth2 全链路**：login → Hydra → user-management → 角色判断，每步都是硬依赖
5. **微前端 qiankun**：作为子应用注册，独立运行需验证

### 改造建议

**短期（Phase 1）**：
- 增加 `AUTH_MODE=local` 环境变量，BFF 用简单 session + 本地用户替代 OAuth2
- BFF Config 类中不存在的服务设为 `null`，调用时跳过
- 前端 API 调用加 try-catch 降级

**中期（Phase 2）**：
- 确保 npm 包公开 + 重新生成 lock 文件
- 精简前端，移除仅 AnyShare 需要的功能模块

### proton-cli-web（安装向导 — ✅ 源码已入库）

独立前端项目，通过 `//go:embed` 嵌入 proton-cli 二进制，是用户首次接触 Proton 的**第一印象界面**。

| 维度 | 详情 |
|------|------|
| **路径** | `proton-cli-web/` |
| **技术栈** | React 18 + TypeScript + webpack 5 + SASS |
| **UI 库** | `@aishutech/ui`（Ant Design 封装）+ `@anyshare/i18nfactory` |
| **源码规模** | 110 文件，核心逻辑 ~4500 行 |
| **组件架构** | Class Components (base + view 分离) |
| **核心文件** | `component.base.tsx` (1631行) — 业务逻辑；`helper.ts` (1953行) — 常量/转换/校验；`component.view.tsx` (379行) — 渲染 |

#### 向导步骤详解

| 步骤 | 组件 | 功能 |
|------|------|------|
| Step 0 | `ChooseTemplate` | 选择部署模式：标准(一体机)/云主机/托管K8s + 产品型号选择 |
| Step 1 | `NodeConfig` | 节点 IP(v4/v6/双栈)/名称/SSH/Chrony/防火墙 |
| Step 2 | `NetworkConfig` | K8s Master/Pod CIDR/Service CIDR/etcd 数据目录/插件 |
| Step 3 | `CRConfig` | 本地 Registry(端口/HA/存储) 或 外部(Registry/Chartmuseum/OCI) |
| Step 4 | `DataBaseConfig` | 15+ 服务配置（9 子组件：MariDB/MongoDB/Redis/Opensearch/Nebula/Monitor/ECeph/PackageStore/ServiceConfig） |
| Step 5 | `ConnectInfo` | 7 种连接配置(RDS/MongoDB/Redis/MQ/OpenSearch/PolicyEngine/ETCD)，每个支持 internal/external |

#### 数据流：前端 → 后端

```
UI ConfigData → exChangeData(data, crType, storageType)
  → 转换为 ClusterConfig JSON（nodes/cs/cr/chrony/firewall/各服务/resource_connect_info）
  → POST /init?accout=&password= → proton-cli serve.go
  → apply.Apply(conf) → 异步执行
  → 轮询 GET /alpha/result → 成功跳转 /success
```

#### 开源阻断项

| 问题 | 影响 | 工作量 |
|------|------|--------|
| `@aishutech/ui` 内部 npm 包 | 🔴 无法 `npm install` | 发布到 npmjs 或全局替换为 `antd`（API 兼容，约 25 文件） |
| `@anyshare/i18nfactory` 内部包 | 🔴 无法 `npm install` | 替换为 `react-i18next`（仅 1 文件） |
| DEVICESPECS 产品型号 (AS10000/AS9000/smas...) | 🟡 AnyShare 业务残留 | 移除整个选择器或改为通用型号 |
| 默认用户名 `"anyshare"` | 🟡 品牌残留 | 改为 `"proton"` |
| OpenSearch `ecoconfig-private.anyshare` URL | 🟡 无效引用 | 清空 |
| 全部中文 UI、无 i18n | 🟢 体验 | Phase 3 添加英文 |

#### 优势评估

- **功能完备**：完整覆盖 proton-cli 全部配置项和 3 种部署模式
- **校验严格**：IPv4/IPv6 格式、CIDR 范围、节点名正则、重复检测
- **服务可选**：默认服务 + 可动态添加（Prometheus/Grafana/Nebula）
- **Mock 开发**：`express.js` + `payload.js` 可独立调试
- **构建路径清晰**：`npm build` → `dist/` → copy 到 Go embed 目录

#### 改进建议

1. **安装进度** — 当前只有 Spin 动画，无法看到执行到哪个阶段。需扩展 `/alpha/result` 返回 `{stage: "cr", progress: 60, modules: [{name:"firewall",status:"done"}, ...]}`
2. **服务分组** — 15+ 服务一次性展示太多，建议分"核心必选"(MariaDB/Redis) 和"高级可选"(Kafka/OpenSearch/Nebula...)
3. **配置预览** — 完成前增加 YAML 预览页，让用户确认生成的 ClusterConfig
4. **Class → Hooks** — 长期可渐进迁移，短期不影响功能

---

## 八、总结

Proton 的核心架构已经是**企业级部署管理平台**的完整形态，且 **Web 安装向导前端已入库**：

| 能力 | 状态 | 对标 |
|------|------|------|
| 声明式集群配置 | ✅ 成熟 | KubeKey |
| 内置 K8s 安装 | ✅ 完整（kubeadm + calico） | K3s/KubeKey |
| 本地 Registry | ✅ 完整 | K3s embedded registry |
| 数据服务编排 | ✅ 完整（MariaDB/Redis/MongoDB/ES/...） | 超越对标 |
| **Web 安装向导** | **✅ 前后端完整（proton-cli-web 已入库）** | KubeSphere |
| 离线支持 | ✅ OCI 镜像 + Chart 推送 | KubeKey artifact |
| 上层应用管理 | ✅ DAG 任务编排 + Helm 管理 | Rancher Apps |
| 备份恢复 | ✅ 已有 | — |

**前端项目总览**：

| 前端项目 | 定位 | 源码状态 | 阻断级别 |
|---------|------|---------|---------|
| `proton-cli-web` | 安装向导（嵌入 proton-cli） | ✅ 已入库 | 🔴 2 个内部 npm 包需替换 |
| `deploy-web-static` | 系统管理工作台 | ✅ 已入库 | 🔴 3 个内部 npm 包 + OAuth2 依赖 |
| `deploy-web-service` | BFF 层 | ✅ 已入库 | 🟡 20+ 硬编码服务 + Hydra 绑定 |

**主要工作不是能力开发，而是开源化改造**：

1. **清除内部绑定**（Go module 路径 + 域名/仓库 + AnyShare 引用 + 内部 npm 包）— 约 8 天
2. **补全最后一公里**（proton_platform 安装模块 + 本地认证）— 约 5 天
3. **标准化发布**（GitHub Actions 含 proton-cli-web 构建 + Helm Chart 标准化 + install.sh）— 约 6 天

**Phase 1 完成后**（约 2 周），用户可以：
1. 通过 `proton-cli server` 启动 Web 向导，图形化配置集群
2. 或通过 `proton-cli apply -f cluster.yaml` CLI 方式部署
3. 从裸机一键部署出完整的 Proton 平台（K8s + 数据服务 + 管理平台）
4. 通过 Proton Web UI 管理上层开源 Helm Chart 产品的全生命周期

这将是一个可与 KubeSphere/Rancher 同级的企业级开源部署管理平台。
