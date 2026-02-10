# Proton 统一架构设计 — proton-server + proton-agent + proton-web

> **设计目标**：将现有 5 个后端服务 + 2 个前端项目合并为 **proton-server（K8s Deployment）+ proton-agent（DaemonSet）+ proton-web（统一 UI）**，提供覆盖集群管理到数据服务到应用部署的**单一 Web 控制台**。

---

## 一、现状问题分析

### 当前架构（5 后端 + 2 前端）

```
┌─────────────────────────────────────────────────────────────────────────┐
│  用户                                                                    │
│    │                                                                     │
│    ├── proton-cli server :8888 ──→ proton-cli-web (React, 嵌入式)         │
│    │       └── POST /init → apply.Apply()                                │
│    │                                                                     │
│    └── deploy-web-static (React) ──→ deploy-web-service (Node.js BFF)    │
│              │                            │                              │
│              │                   ┌────────┴──────────┐                   │
│              │                   │  OAuth2 (Hydra)   │                   │
│              │                   │  20+ 服务代理       │                  │
│              │                   └────────┬──────────┘                   │
│              │                            │                              │
│              ├── deploy-service (Python/Tornado :9703)                    │
│              │      └── OSS 网关/客户端包/证书/通信配置                      │
│              │                                                           │
│              ├── deployrunner (Go/Gin :9090)                              │
│              │      └── DAG 任务执行/应用包管理/组件管理                      │
│              │      └── proton_component: Kafka/Redis/MariaDB/...         │
│              │                                                           │
│              └── component-manage (Go, K8s 内部)                          │
│                     └── Helm Release 生命周期管理                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 核心问题

| 问题 | 影响 |
|------|------|
| **5 个后端 3 种语言** (Go/Python/Node.js) | 运维复杂、构建慢、依赖多 |
| **Node.js BFF 纯代理** | 90% 代码是 fetch 转发，增加延迟和故障点 |
| **Python deploy-service 功能窄** | 仅 OSS 网关+客户端包+证书，可合并 |
| **proton-cli-web 与 deploy-web-static 割裂** | 安装向导和管理控制台分离，用户体验断裂 |
| **数据服务管理分散** | 初始配置在 proton-cli-web，运行时管理在 deployrunner，信息不一致 |
| **Hydra OAuth2 硬依赖** | 开源环境难以部署，增加启动门槛 |

---

## 二、目标架构（1 后端 + 1 前端）

### 2.1 业界对标：单一控制台模式

| 项目 | 方案 | 做法 |
|------|------|------|
| **Rancher** | 单 UI，多视图 | 一个控制台，上层切换集群，下层管理工作负载。集群管理和应用管理在同一 UI 的不同导航区域 |
| **KubeSphere** | 单 UI，分层导航 | "平台管理"（集群/节点/存储）+ "工作空间"（项目/服务/DevOps），同一 UI 左侧导航分区 |
| **OpenShift** | 单 UI，双视角 | "Administrator 视角"看集群基础设施，"Developer 视角"看应用部署，右上角切换 |
| **Sealos** | 单 UI（云桌面） | K8s 集群管理和应用部署融合在一个"桌面"界面 |
| **Longhorn** | DaemonSet Agent | 每个节点运行 agent 管理本地磁盘/卷，控制平面通过 K8s API 下发指令 |

**结论**：业界全部是**一个 Web 控制台**，不存在两套界面。proton-cli 的 Web 向导应该消失，Web 管理全部归 proton-server。

### 2.2 总体架构（3 个组件）

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          proton-web (React SPA)                          │
│                    唯一 Web 控制台（参照 Rancher/KubeSphere）               │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │ 集群概览       │ 节点管理       │ 数据服务管理     │ 应用管理        │   │
│  │ K8s 版本/资源  │ 磁盘/目录/状态 │ CRUD + 实时状态 │ 安装/升级/任务  │   │
│  │                │ (via agent)   │                 │                │   │
│  ├────────────────┴───────────────┴─────────────────┴────────────────┤   │
│  │ 系统设置 │ 用户管理 │ 首次安装向导（内嵌，非独立 UI）                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                    │ HTTP / WebSocket                    │
└────────────────────────────────────┼────────────────────────────────────┘
                                     │
┌────────────────────────────────────┼────────────────────────────────────┐
│           proton-server (Go/Gin) — K8s Deployment                       │
│                                                                         │
│  ┌────────────┐ ┌──────────────────┐ ┌────────────┐ ┌──────────┐     │
│  │ DataSvcMgr │ │ AppMgr           │ │ NodeMgr    │ │ Auth     │     │
│  │ Helm CRUD  │ │ ┌──────────────┐ │ │ Agent 通信  │ │ JWT/OIDC │     │
│  │ Status     │ │ │ 模块化部署    │ │ │ 节点信息    │ │          │     │
│  │ Watch      │ │ │ DAG Runner   │ │ │ 磁盘/目录   │ │          │     │
│  └────────────┘ │ │ Job Executor │ │ └──────┬─────┘ └──────────┘     │
│                  │ ├──────────────┤ │        │                         │
│                  │ │ 原生 Helm    │ │        │                         │
│                  │ │ Repo 浏览    │ │        │                         │
│                  │ │ Chart 可视化 │ │        │                         │
│                  │ │ Release 管理 │ │        │                         │
│                  │ └──────────────┘ │        │                         │
│                  └──────────────────┘        │                         │
│  ┌────────────────┐ ┌────────────────┐ ┌─────────────────────────┐    │
│  │ SystemMgr     │ │ ClusterInfoMgr │ │ Storage Strategy        │    │
│  │ 证书/配置/地址 │ │ K8s 只读查询   │ │ StorageClass → PVC      │    │
│  └────────────────┘ └────────────────┘ │ hostPath → agent 创建   │    │
│                                        └─────────────────────────┘    │
│  Infrastructure:                                                       │
│  Helm3 (in-cluster) │ K8s client-go │ Chart Repo │ gRPC to agents     │
└───────────────────────────────────────┼─────────────────────────────────┘
                                        │ K8s API (Pod exec / gRPC)
          ┌─────────────────────────────┼─────────────────────────┐
          │                             │                         │
          ▼                             ▼                         ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ proton-agent     │  │ proton-agent     │  │ proton-agent     │
│ (DaemonSet Pod)  │  │ (DaemonSet Pod)  │  │ (DaemonSet Pod)  │
│ Node 1           │  │ Node 2           │  │ Node 3           │
│                  │  │                  │  │                  │
│ ✅ 创建 hostPath │  │ ✅ 创建 hostPath │  │ ✅ 创建 hostPath │
│ ✅ 磁盘信息      │  │ ✅ 磁盘信息      │  │ ✅ 磁盘信息      │
│ ✅ 目录清理      │  │ ✅ 目录清理      │  │ ✅ 目录清理      │
│ ✅ 节点健康      │  │ ✅ 节点健康      │  │ ✅ 节点健康      │
│ (hostPath mount) │  │ (hostPath mount) │  │ (hostPath mount) │
└──────────────────┘  └──────────────────┘  └──────────────────┘
```

### 2.3 proton-agent 的必要性

| 场景 | 需要 agent？ | 存储方式 | 主机操作 |
|------|-------------|---------|---------|
| **离线裸机** | ✅ 必须 | **hostPath**（用户指定目录） | 创建目录、设置权限、磁盘查询、清理 |
| **自建 K8s + 本地盘** | ✅ 需要 | **hostPath / local PV** | 同上 |
| **托管 K8s（EKS/AKS）** | ⚠️ 可选 | **StorageClass + PVC** | 不需要主机操作，但 agent 可采集节点信息 |
| **托管 K8s + 云盘** | ❌ 不需要 | **StorageClass + PVC** | 无 |

**proton-agent 以 DaemonSet 运行，不依赖 SSH**：
- 通过 K8s API 部署，无需 SSH 到任何主机
- 挂载 hostPath（如 `/data`）实现主机目录操作
- 通过 gRPC / K8s API（Pod labels/annotations）与 proton-server 通信
- 离线环境必装，托管 K8s 可选装

### 2.4 两条部署路径（更新）

```
═══════════════════════════════════════════════════════════════════════
路径 A: 裸机 / VM / 离线环境（proton-cli 向导 → proton-server 自动跳转）
═══════════════════════════════════════════════════════════════════════

  管理员浏览器                 管理员主机                    目标集群节点
  ┌─────────────┐            ┌─────────────┐            ┌─────────────────┐
  │             │  打开向导   │ proton-cli   │   SSH      │                 │
  │  proton-cli │ ────────→  │ serve :3000  │ ────────→  │ 裸机 / VM       │
  │  Web 向导   │            │              │            │                 │
  │ (React SPA) │            │ apply chain: │            │ 执行:           │
  │             │            │ 1. firewall  │            │ 防火墙 → 节点   │
  │ Step 1-4:   │            │ 2. nodes     │            │ → CR → K8s     │
  │ 节点/K8s    │  WebSocket │ 3. cr        │            │                 │
  │ 配置        │ ←───────── │ 4. cs        │            │ 最后两步:       │
  │             │  实时进度   │ 5. ingress   │            │ 5. ingress-nginx│
  │             │            │ 6. proton ★  │            │ 6. proton-server│
  │             │            └─────────────┘            │    + agent      │
  │             │                                        │                 │
  │  安装完成！  │            proton-cli 探测到            └─────────────────┘
  │  正在跳转.. │            proton-server Ready:                │
  │             │            GET https://<ingress>/health        │
  │  ┌────────────────────────────────────────────────────┐     │
  │  │ 自动跳转 → https://<ingress-host>/                  │     │
  │  │ proton-cli 向导关闭，打开 proton-server Dashboard   │     │
  │  └────────────────────────────────────────────────────┘     │
  │             │                                                │
  └──────┬──────┘                                                │
         │ 302 Redirect                                          │
         ▼                                                       │
  ┌─────────────┐                                                │
  │ proton-web  │ ←──────── proton-server (Pod) ←────────────────┘
  │ Dashboard   │            :8080 via Ingress
  │             │
  │ 首次登录:   │   proton-server 检测到首次启动:
  │ admin/随机  │   → 显示内嵌向导（数据服务配置）
  │ → 内嵌向导  │   → 数据服务/应用在 proton-server 中管理
  └─────────────┘

═══════════════════════════════════════════════════════════════════════
路径 B: 托管 K8s / 已有 K8s（不需要 proton-cli）
═══════════════════════════════════════════════════════════════════════

  用户本机                            托管 K8s (EKS/AKS/GKE/...)
  ┌────────────────┐                ┌──────────────────────────────────┐
  │ $ helm install  │   K8s API     │                                  │
  │   proton \\      │ ───────────→  │  proton-server (Pod)              │
  │   --set ...     │               │  :8080 Web UI + API               │
  │                 │               │                                  │
  │ agent 可选:     │               │  proton-agent: 不安装或可选安装    │
  │ --set agent.    │               │  存储: StorageClass + PVC         │
  │  enabled=false  │               │                                  │
  └────────────────┘               │  首次访问 Web UI:                 │
                                    │  → 向导引导选择存储策略            │
                                    │  → StorageClass / external        │
                                    │  → 一键安装数据服务               │
                                    └──────────────────────────────────┘
```

### 2.5 两阶段向导衔接设计

proton-cli 的 Web 向导**仅负责 K8s 基础设施初始化**（裸机场景）。完成后自动跳转到 proton-server Dashboard，
后续数据服务/应用管理全部在 proton-server 的统一 UI 中完成。

```
┌─────────────────────────────────────────────────────────────────────┐
│  proton-cli 向导（裸机场景，localhost:3000）                          │
│                                                                     │
│  Step 1: 节点配置（IP/SSH/角色）                                     │
│  Step 2: K8s 配置（Pod CIDR/Svc CIDR/Container Runtime）            │
│  Step 3: 容器仓库（本地 Harbor / 离线镜像包）                         │
│  Step 4: Ingress 配置（ingress-nginx，访问域名/IP）                   │
│  Step 5: 确认 & 开始安装                                             │
│           → firewall → nodes → cr → cs → ingress-nginx              │
│           → helm install proton-server + proton-agent                │
│           → 轮询 GET https://<ingress>/api/v1/health                │
│           → Ready! 自动跳转 ↓                                        │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  ✅ 集群初始化完成！                                         │    │
│  │  proton-server 已就绪，3 秒后自动跳转到管理控制台...          │    │
│  │  https://proton.example.com/                                │    │
│  │                                        [立即跳转 →]          │    │
│  └─────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────┬──────────────────────┘
                                               │ 浏览器 302 跳转
                                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│  proton-server Dashboard（https://proton.example.com/）              │
│                                                                     │
│  首次登录 → admin / <随机密码，proton-cli 安装时打印>                 │
│  检测到首次启动 → 自动弹出内嵌向导:                                   │
│                                                                     │
│  Step 1: 数据服务配置（hostPath 模式，选节点/路径/大小）              │
│  Step 2: 应用安装（可选，从本地 Chart 仓库选择）                      │
│  Step 3: 完成 → 进入 Dashboard                                      │
│                                                                     │
│  此后 proton-cli 可以关闭，所有管理在此 UI 完成                       │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.6 统一 UI 设计（对标 Rancher/KubeSphere）

所有 Web 管理统一在 proton-server 的 Web UI 中，通过导航分区覆盖所有场景。
proton-cli 的 Web 向导仅在裸机初始化时使用一次，之后所有操作在 proton-server 完成。

```
┌──────────────────────────────────────────────────────────────────────┐
│  Proton Console                                       admin ▾  🌐   │
├──────────┬───────────────────────────────────────────────────────────┤
│          │                                                           │
│  📊 概览  │  K8s 版本、节点数、CPU/Memory、数据服务状态、应用/Helm 状态 │
│          │                                                           │
│ ─────── │───────────────────────────────────────────────────────────│
│ 集群     │                                                           │
│  🖥 节点  │  节点列表、CPU/Memory/磁盘使用率                            │
│          │  离线环境: agent 上报磁盘详情、hostPath 目录列表             │
│          │  托管 K8s: 只读展示                                        │
│  💾 存储  │  StorageClass 列表、PVC 列表                               │
│          │  离线环境: hostPath 卷管理（通过 agent）                    │
│          │                                                           │
│ ─────── │───────────────────────────────────────────────────────────│
│ 数据服务 │                                                           │
│  🗄 服务  │  数据服务列表、创建/编辑/删除、实时状态、连接信息             │
│          │                                                           │
│ ─────── │───────────────────────────────────────────────────────────│
│ 应用     │                                                           │
│  📦 模块  │  模块化部署（DAG）：上传应用包 → 依赖图 → 并行安装          │
│  📋 任务  │  安装任务列表、进度跟踪、日志查看                           │
│          │                                                           │
│ ─────── │───────────────────────────────────────────────────────────│
│ Helm ★  │  （新增：原生 Helm 可视化管理，对标 Rancher Apps）           │
│  � 仓库  │  Chart 仓库管理（添加/删除 Helm repo）                     │
│  📜 Chart │  Chart 浏览：搜索、版本列表、README、values.yaml 预览      │
│  🚀 部署  │  可视化安装：填写 values 表单 / YAML 编辑 → helm install   │
│  📋 Release│ Release 列表：状态、版本历史、升级/回滚/卸载              │
│          │                                                           │
│ ─────── │───────────────────────────────────────────────────────────│
│ 系统     │                                                           │
│  ⚙ 设置  │  证书管理、访问地址、通信配置                               │
│  👤 用户  │  用户管理、密码修改                                        │
│          │                                                           │
│ ─────── │───────────────────────────────────────────────────────────│
│  🧙 向导  │  首次安装向导（内嵌在同一 UI 中，非独立界面）               │
│          │  首次访问时自动弹出，完成后变为侧边栏入口                    │
│          │                                                           │
└──────────┴───────────────────────────────────────────────────────────┘
```

### 2.7 核心设计原则

1. **两阶段衔接** — proton-cli 向导负责 K8s 初始化，完成后自动跳转到 proton-server Dashboard
2. **单一控制台** — proton-server Web UI 是日常管理的唯一入口（数据服务 + 模块化部署 + Helm 管理）
3. **容器优先** — proton-server 和 proton-agent 都以 K8s 工作负载运行（Deployment + DaemonSet）
4. **Agent 取代 SSH** — 主机级操作通过 proton-agent DaemonSet 完成，不依赖 SSH
5. **存储策略自适应** — 离线用 hostPath（agent 创建目录），云上用 StorageClass + PVC，用户自选
6. **双模式应用管理** — 模块化部署（DAG 依赖图）与原生 Helm 可视化并存
7. **Go 统一** — 消除 Python/Node.js 依赖，proton-server/agent/cli 全部 Go
8. **本地认证优先** — JWT + bcrypt 本地用户，可选对接外部 OIDC
9. **WebSocket 实时推送** — 安装进度/服务状态/节点状态/Helm 操作进度实时更新

---

## 三、后端设计（proton-server）

### 3.1 模块合并映射

```
┌───────────────────────────────────────────────────────────────┐
│                    合并前 → 合并后                              │
├───────────────────────────────────────────────────────────────┤
│  proton-cli serve.go (apply chain)                            │
│    firewall→nodes→cr→cs                                      │
│    → 保留在 proton-cli（裸机场景独立使用）                      │
│    → proton-server 不包含此功能                                │
│                                                               │
│  proton-cli 安装向导 (POST /init)                              │
│    → proton-server: ClusterInfoMgr (只读集群信息)              │
│      GET  /api/v1/cluster/info        集群概览(节点/版本/资源)  │
│      GET  /api/v1/cluster/nodes       节点列表+状态             │
│      GET  /api/v1/cluster/storage     StorageClass 列表        │
│      GET  /api/v1/cluster/namespaces  Namespace 列表           │
├───────────────────────────────────────────────────────────────┤
│  deployrunner proton_component/*                              │
│    GET/PUT /components/release/{type}/{name}                  │
│    GET/PUT /components/info/{type}                            │
│    → proton-server: DataSvcMgr                                │
│      GET    /api/v1/datasvc/                                  │
│      GET    /api/v1/datasvc/:type/:name                       │
│      PUT    /api/v1/datasvc/:type/:name                       │
│      POST   /api/v1/datasvc/:type (创建)                      │
│      DELETE /api/v1/datasvc/:type/:name (卸载)                 │
│      GET    /api/v1/datasvc/:type/:name/status (实时状态)      │
├───────────────────────────────────────────────────────────────┤
│  deployrunner rest/store.go + rest/job.go                     │
│    /application/*, /job/*, /system/*                          │
│    → proton-server: AppMgr（模块化部署，保留 DAG 逻辑）        │
│      POST /api/v1/apps/upload         上传应用包（含多 Chart） │
│      GET  /api/v1/apps/               应用列表                │
│      POST /api/v1/apps/:name/install  安装（DAG 依赖图执行）   │
│      DELETE /api/v1/apps/:name        卸载                    │
│      GET  /api/v1/jobs/               任务列表                │
│      POST /api/v1/jobs/:id/start      启动任务                │
│      PATCH /api/v1/jobs/:id/cancel    取消任务                │
│                                                               │
│  新增: HelmMgr（原生 Helm 可视化管理） ★                       │
│    → proton-server: HelmMgr                                   │
│      GET    /api/v1/helm/repos           仓库列表              │
│      POST   /api/v1/helm/repos           添加仓库              │
│      DELETE /api/v1/helm/repos/:name     删除仓库              │
│      POST   /api/v1/helm/repos/:name/sync 同步仓库索引         │
│      GET    /api/v1/helm/charts          Chart 搜索/浏览       │
│      GET    /api/v1/helm/charts/:repo/:name  Chart 详情        │
│      GET    /api/v1/helm/charts/:repo/:name/versions  版本列表 │
│      GET    /api/v1/helm/charts/:repo/:name/values    默认values│
│      POST   /api/v1/helm/releases        安装 Release          │
│      GET    /api/v1/helm/releases        Release 列表          │
│      GET    /api/v1/helm/releases/:ns/:name  Release 详情      │
│      PUT    /api/v1/helm/releases/:ns/:name  升级 Release      │
│      POST   /api/v1/helm/releases/:ns/:name/rollback  回滚     │
│      DELETE /api/v1/helm/releases/:ns/:name  卸载 Release      │
│      GET    /api/v1/helm/releases/:ns/:name/history   版本历史 │
├───────────────────────────────────────────────────────────────┤
│  deploy-service                                               │
│    /api/deploy-manager/cert/*                                 │
│    /api/deploy-manager/v1/access-addr/*                       │
│    /api/deploy-manager/v1/communication/*                     │
│    → proton-server: SystemMgr                                 │
│      GET/PUT /api/v1/system/cert                              │
│      GET/PUT /api/v1/system/access-addr                       │
│      GET/PUT /api/v1/system/config                            │
├───────────────────────────────────────────────────────────────┤
│  deploy-web-service (BFF)                                     │
│    OAuth2 proxy + 20+ 服务转发                                 │
│    → 完全删除，由 proton-server 直接提供 API                    │
│      POST /api/v1/auth/login                                  │
│      POST /api/v1/auth/logout                                 │
│      GET  /api/v1/auth/me                                     │
└───────────────────────────────────────────────────────────────┘
```

### 3.2 项目结构

```
proton-server/
├── cmd/
│   └── proton-server/
│       └── main.go                 # 入口：CLI + HTTP Server
├── internal/
│   ├── api/                        # Gin HTTP handlers
│   │   ├── router.go               # 路由注册
│   │   ├── middleware/
│   │   │   ├── auth.go             # JWT 认证中间件
│   │   │   ├── cors.go
│   │   │   └── logger.go
│   │   ├── clusterinfo/             # 集群信息 API（只读）
│   │   │   ├── overview.go          # GET /api/v1/cluster/info
│   │   │   ├── nodes.go             # GET /api/v1/cluster/nodes
│   │   │   ├── storage.go           # GET /api/v1/cluster/storage
│   │   │   └── namespaces.go        # GET /api/v1/cluster/namespaces
│   │   ├── datasvc/                # 数据服务管理 API ★
│   │   │   ├── list.go             # GET  /api/v1/datasvc/
│   │   │   ├── get.go              # GET  /api/v1/datasvc/:type/:name
│   │   │   ├── create.go           # POST /api/v1/datasvc/:type
│   │   │   ├── update.go           # PUT  /api/v1/datasvc/:type/:name
│   │   │   ├── delete.go           # DELETE /api/v1/datasvc/:type/:name
│   │   │   └── status.go           # GET  /api/v1/datasvc/:type/:name/status
│   │   ├── apps/                   # 模块化部署 API（DAG）
│   │   │   ├── upload.go           # POST /api/v1/apps/upload
│   │   │   ├── install.go          # POST /api/v1/apps/:name/install
│   │   │   ├── list.go             # GET  /api/v1/apps/
│   │   │   └── jobs.go             # GET/POST/PATCH /api/v1/jobs/*
│   │   ├── helm/                   # 原生 Helm 可视化 API ★
│   │   │   ├── repos.go            # CRUD /api/v1/helm/repos
│   │   │   ├── charts.go           # GET  /api/v1/helm/charts/*
│   │   │   ├── releases.go         # CRUD /api/v1/helm/releases/*
│   │   │   └── values.go           # GET  values / diff
│   │   ├── system/                 # 系统配置 API
│   │   │   ├── cert.go
│   │   │   ├── access.go
│   │   │   └── config.go
│   │   ├── auth/                   # 认证 API
│   │   │   ├── login.go
│   │   │   └── user.go
│   │   └── ws/                     # WebSocket
│   │       └── events.go           # 实时事件推送
│   │
│   ├── core/                       # 核心业务逻辑
│   │   ├── clusterinfo/             # 集群信息（只读，K8s client-go）
│   │   │   └── manager.go           # 节点/NS/SC/资源概览
│   │   ├── datasvc/                # 数据服务管理 ★
│   │   │   ├── manager.go          # DataServiceManager 核心
│   │   │   ├── helm_operator.go    # Helm Release CRUD
│   │   │   ├── status_watcher.go   # K8s Pod/Service 状态监控
│   │   │   ├── config_sync.go      # ClusterConfig 同步
│   │   │   └── types.go            # 数据服务类型定义
│   │   ├── apps/                   # 模块化部署（复用 deployrunner DAG）
│   │   │   ├── executor.go         # DAG 任务执行器
│   │   │   ├── store.go            # 应用/任务持久化
│   │   │   └── graph.go            # 依赖图算法
│   │   ├── helm/                   # 原生 Helm 管理 ★
│   │   │   ├── repo_manager.go     # Chart 仓库索引/同步
│   │   │   ├── chart_browser.go    # Chart 搜索/详情/values 解析
│   │   │   ├── release_manager.go  # Release 安装/升级/回滚/卸载
│   │   │   └── values_merger.go    # values 合并（用户输入 + 默认值）
│   │   └── auth/                   # 认证
│   │       ├── jwt.go
│   │       └── users.go
│   │
│   ├── infra/                      # 基础设施客户端
│   │   ├── helm/                   # Helm3 客户端（in-cluster 操作）
│   │   ├── k8s/                    # K8s client-go 封装（in-cluster config）
│   │   └── registry/               # 镜像仓库客户端（拉取 Chart）
│   │
│   └── config/                     # 配置管理
│       └── server_config.go        # Server 自身配置（端口/认证模式/Chart仓库等）
│
├── web/                            # //go:embed 前端静态文件（构建时从 proton-web 复制）
│   └── .gitkeep
│
├── go.mod
└── Makefile
```

### 3.3 数据服务管理核心设计 ★

这是本次设计的**核心新增能力**。

#### 3.3.1 数据服务抽象模型

```go
// DataService 统一数据服务抽象
type DataService struct {
    Name        string            `json:"name"`         // 实例名，如 "mariadb", "proton-redis"
    Type        DataServiceType   `json:"type"`         // 服务类型
    Source      SourceType        `json:"source"`       // internal / external
    Status      ServiceStatus     `json:"status"`       // 运行状态
    HelmRelease *HelmReleaseInfo  `json:"helm_release"` // Helm Release 信息（internal）
    ConnectInfo *ConnectInfo      `json:"connect_info"` // 连接信息
    Config      map[string]any    `json:"config"`       // 服务特定配置
    Storage     *StorageSpec      `json:"storage"`      // 存储配置（PVC + StorageClass）
    Resources   *ResourceSpec     `json:"resources"`    // CPU/Memory 配额
    Namespace   string            `json:"namespace"`    // 部署命名空间
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type DataServiceType string
const (
    TypeRDS           DataServiceType = "rds"             // MariaDB/MySQL
    TypeMongoDB       DataServiceType = "mongodb"
    TypeRedis         DataServiceType = "redis"
    TypeOpenSearch    DataServiceType = "opensearch"
    TypeKafka         DataServiceType = "kafka"
    TypeZooKeeper     DataServiceType = "zookeeper"
    TypeETCD          DataServiceType = "etcd"
    TypePolicyEngine  DataServiceType = "policy-engine"
    TypeNSQ           DataServiceType = "nsq"
    TypePrometheus    DataServiceType = "prometheus"
    TypeGrafana       DataServiceType = "grafana"
    TypeNebula        DataServiceType = "nebula"
)

type SourceType string
const (
    SourceInternal SourceType = "internal"  // Helm 部署在 K8s 内
    SourceExternal SourceType = "external"  // 外部已有服务
)

type ServiceStatus struct {
    Phase      string    `json:"phase"`       // Running / Pending / Failed / Unknown
    Ready      string    `json:"ready"`       // "3/3"
    Message    string    `json:"message"`
    Pods       []PodInfo `json:"pods"`        // Pod 详细状态
    CheckedAt  time.Time `json:"checked_at"`
}

type ConnectInfo struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password,omitempty"` // API 返回时脱敏
    Database string `json:"database,omitempty"`
    // 特定类型扩展字段
    Extra    map[string]any `json:"extra,omitempty"`
}

// StorageSpec 存储配置（两种模式，用户自选）
type StorageSpec struct {
    // 模式 1: StorageClass + PVC（托管 K8s / 有 CSI 驱动的环境）
    StorageClass string `json:"storage_class,omitempty"` // StorageClass 名称
    Size         string `json:"size"`                    // 存储大小，如 "50Gi"

    // 模式 2: hostPath（离线裸机环境，通过 proton-agent 创建和管理）
    HostPath     string `json:"host_path,omitempty"`     // 主机目录，如 "/data/mariadb"
    NodeName     string `json:"node_name,omitempty"`     // 指定节点（hostPath 必须绑定节点）
    // proton-server 创建数据服务时:
    // - 如果 HostPath 非空 → 通知 proton-agent 在指定节点创建目录
    //   → 生成 local PV + PVC 绑定到该节点
    // - 如果 HostPath 为空 → 使用 StorageClass 动态分配 PVC
}
```

#### 3.3.2 DataServiceManager 核心接口

```go
// DataServiceManager 数据服务管理器
type DataServiceManager struct {
    helm       helm3.Client
    k8s        kubernetes.Interface
    watcher    *StatusWatcher
    config     *configuration.ClusterConfig
    eventBus   *EventBus  // WebSocket 事件推送
}

// CRUD 接口
func (m *DataServiceManager) List(ctx context.Context) ([]DataService, error)
func (m *DataServiceManager) Get(ctx context.Context, svcType, name string) (*DataService, error)
func (m *DataServiceManager) Create(ctx context.Context, req *CreateDataServiceRequest) (*DataService, error)
func (m *DataServiceManager) Update(ctx context.Context, svcType, name string, req *UpdateDataServiceRequest) (*DataService, error)
func (m *DataServiceManager) Delete(ctx context.Context, svcType, name string) error

// 状态相关
func (m *DataServiceManager) Status(ctx context.Context, svcType, name string) (*ServiceStatus, error)
func (m *DataServiceManager) HealthCheck(ctx context.Context, svcType, name string) (*HealthResult, error)

// 连接信息
func (m *DataServiceManager) ConnectInfo(ctx context.Context, svcType, name string) (*ConnectInfo, error)
func (m *DataServiceManager) TestConnection(ctx context.Context, info *ConnectInfo) error

// 配置同步（更新 ClusterConfig 中的 ResourceConnectInfo）
func (m *DataServiceManager) SyncConfig(ctx context.Context) error
```

#### 3.3.3 Internal 数据服务的生命周期

```
创建（Create）                    状态监控（Watch）
    │                                │
    ▼                                ▼
┌─────────┐   helm install    ┌─────────────┐   K8s Watch   ┌──────────┐
│ 前端表单  │ ──────────────→ │ Helm Release │ ────────────→ │ Pod 状态  │
│ 填写配置  │                 │  (internal)  │               │ Ready/Err│
└─────────┘                   └──────┬──────┘               └────┬─────┘
                                     │                           │
                              写入 K8s Secret              WebSocket 推送
                                     │                           │
                                     ▼                           ▼
                            ┌────────────────┐          ┌──────────────┐
                            │ ClusterConfig  │          │  前端实时更新  │
                            │ ResourceConnect│          │  状态面板     │
                            │ Info 同步      │          └──────────────┘
                            └────────────────┘

更新（Update）                    删除（Delete）
    │                                │
    ▼                                ▼
  helm upgrade                  检查依赖 → helm uninstall
  + 更新 ConnectInfo             + 清理 ConnectInfo
  + 通知依赖服务                  + 通知上层应用
```

#### 3.3.4 External 数据服务管理

外部数据服务不需要 Helm 操作，但仍需要：

```go
// 外部数据服务只管理连接信息
func (m *DataServiceManager) CreateExternal(ctx context.Context, req *CreateExternalRequest) (*DataService, error) {
    // 1. 测试连接可达性
    if err := m.TestConnection(ctx, req.ConnectInfo); err != nil {
        return nil, fmt.Errorf("connection test failed: %w", err)
    }
    // 2. 保存到 ClusterConfig.ResourceConnectInfo
    m.config.ResourceConnectInfo.Set(req.Type, req.ConnectInfo)
    // 3. 上传到 K8s Secret
    configuration.UploadToKubernetes(ctx, m.config, m.k8s)
    // 4. 返回数据服务对象
    return &DataService{Source: SourceExternal, ...}, nil
}
```

### 3.4 API 设计

#### 3.4.1 数据服务 API（核心新增）

```yaml
# 列出所有数据服务
GET /api/v1/datasvc/
Response:
  data:
    - name: mariadb
      type: rds
      source: internal
      status: { phase: Running, ready: "1/1" }
      connect_info: { host: mariadb.resource, port: 3306, username: root }
    - name: proton-redis
      type: redis
      source: internal
      status: { phase: Running, ready: "3/3" }
    - name: external-mysql
      type: rds
      source: external
      status: { phase: Connected }
      connect_info: { host: 10.0.1.100, port: 3306 }
  total: 3

# 获取单个数据服务详情
GET /api/v1/datasvc/rds/mariadb
Response:
  name: mariadb
  type: rds
  source: internal
  status: { phase: Running, ready: "1/1", pods: [...] }
  helm_release: { name: mariadb, chart: mariadb-11.0.0, namespace: resource }
  connect_info: { host: mariadb.resource, port: 3306, username: root }
  config: { replicas: 1 }
  storage: { storage_class: "gp3", size: "50Gi" }
  resources: { cpu: "500m", memory: "1Gi" }
  namespace: proton-data

# 创建内置数据服务
POST /api/v1/datasvc/redis
Body:
  name: proton-redis
  source: internal
  config:
    replicas: 3
    password: "auto-generated"
  storage:
    storage_class: ""            # 空=使用集群默认 StorageClass
    size: "10Gi"
  resources: { cpu: "500m", memory: "1Gi" }
  namespace: proton-data

# 创建外置数据服务
POST /api/v1/datasvc/rds
Body:
  name: external-mysql
  source: external
  connect_info:
    host: 10.0.1.100
    port: 3306
    username: admin
    password: "..."
    database: proton

# 更新数据服务配置
PUT /api/v1/datasvc/redis/proton-redis
Body:
  config:
    replicas: 5
  resources: { cpu: "1", memory: "2Gi" }

# 删除数据服务
DELETE /api/v1/datasvc/redis/proton-redis

# 测试外部连接
POST /api/v1/datasvc/test-connection
Body:
  type: rds
  host: 10.0.1.100
  port: 3306
  username: admin
  password: "..."

# 获取服务健康状态
GET /api/v1/datasvc/rds/mariadb/health
Response:
  healthy: true
  latency_ms: 5
  details: { connections: 10, uptime: "3d 5h" }
```

#### 3.4.2 集群信息 API（只读）

proton-server **不管理 K8s 基础设施**（节点扩缩容、K8s 升级等由 proton-cli 或云厂商负责）。
仅提供只读查询接口，用于 Dashboard 展示和数据服务创建时选择 StorageClass。

```yaml
# 集群概览
GET /api/v1/cluster/info
Response:
  kubernetes_version: "1.28.5"
  platform: "EKS"               # 自动检测: EKS/AKS/GKE/kubeadm/k3s/unknown
  nodes: 3
  total_cpu: "12"
  total_memory: "48Gi"
  data_services: { total: 8, running: 7, failed: 1 }
  applications: { total: 5, running: 5 }

# 节点列表（只读）
GET /api/v1/cluster/nodes
Response:
  - name: ip-10-0-1-5.ec2.internal
    status: Ready
    roles: [worker]
    cpu: { capacity: "4", usage: "1.2" }
    memory: { capacity: "16Gi", usage: "8.5Gi" }
    pods: 15

# StorageClass 列表（创建数据服务时需要选择）
GET /api/v1/cluster/storage
Response:
  - name: gp3
    provisioner: ebs.csi.aws.com
    default: true
    reclaim_policy: Delete
  - name: local-path
    provisioner: rancher.io/local-path
    default: false

# Namespace 列表
GET /api/v1/cluster/namespaces
Response:
  - name: proton-system
    status: Active
    pods: 2
  - name: proton-data
    status: Active
    pods: 12

# WebSocket 实时事件
WS /ws/events
→ { type: "datasvc.status", data: { name: "mariadb", phase: "Running" } }
→ { type: "datasvc.created", data: { name: "redis", type: "redis" } }
```

#### 3.4.3 认证 API

```yaml
# 登录（本地认证）
POST /api/v1/auth/login
Body: { username: "admin", password: "..." }
Response: { token: "eyJ...", expires_in: 86400 }

# 获取当前用户
GET /api/v1/auth/me
Response: { username: "admin", role: "admin" }

# 修改密码
PUT /api/v1/auth/password
Body: { old_password: "...", new_password: "..." }
```

#### 3.4.4 原生 Helm 可视化 API ★

```yaml
# ──── Chart 仓库管理 ────
# 列出已添加的 Chart 仓库
GET /api/v1/helm/repos
Response:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
    status: synced
    chart_count: 142
    last_sync: "2026-02-07T01:00:00Z"
  - name: local
    url: http://chartmuseum.proton-system:8080
    status: synced

# 添加 Chart 仓库
POST /api/v1/helm/repos
Body:
  name: grafana
  url: https://grafana.github.io/helm-charts
  # 离线环境: 使用本地 ChartMuseum / OCI Registry
  # url: oci://registry.local:5000/charts

# 同步仓库索引
POST /api/v1/helm/repos/bitnami/sync

# 删除仓库
DELETE /api/v1/helm/repos/grafana

# ──── Chart 浏览 ────
# 搜索 Chart（跨仓库）
GET /api/v1/helm/charts?q=redis&repo=bitnami
Response:
  - repo: bitnami
    name: redis
    version: "18.6.1"
    app_version: "7.2.4"
    description: "Redis is an open source..."
    icon: "https://..."

# Chart 详情（README + values schema）
GET /api/v1/helm/charts/bitnami/redis
Response:
  name: redis
  versions: ["18.6.1", "18.6.0", "18.5.0", ...]
  readme: "# Redis\n..."
  values_schema: { ... }            # JSON Schema（如果 Chart 提供）

# Chart 版本列表
GET /api/v1/helm/charts/bitnami/redis/versions

# 获取某版本默认 values.yaml
GET /api/v1/helm/charts/bitnami/redis/values?version=18.6.1
Response:
  # 返回原始 values.yaml 内容（YAML 格式）
  architecture: replication
  auth:
    enabled: true
    password: ""
  master:
    persistence:
      size: 8Gi
  ...

# ──── Release 管理 ────
# 安装 Chart（可视化填写 values 或 YAML 编辑）
POST /api/v1/helm/releases
Body:
  name: my-redis
  namespace: default
  chart: bitnami/redis
  version: "18.6.1"
  values:                            # 用户自定义 values（与默认值合并）
    auth:
      password: "my-secret"
    master:
      persistence:
        storageClass: "gp3"
        size: "50Gi"

# Release 列表（当前集群所有 Helm Release）
GET /api/v1/helm/releases?namespace=all
Response:
  - name: my-redis
    namespace: default
    chart: redis-18.6.1
    app_version: "7.2.4"
    status: deployed
    revision: 1
    updated: "2026-02-07T02:00:00Z"
  - name: mariadb          # DataSvcMgr 管理的也会显示，但标记为 managed
    namespace: proton-data
    chart: mariadb-11.0.0
    status: deployed
    managed_by: proton      # 标记：由 proton DataSvcMgr 管理，不建议直接操作

# Release 详情
GET /api/v1/helm/releases/default/my-redis
Response:
  name: my-redis
  namespace: default
  chart: redis-18.6.1
  status: deployed
  revision: 3
  values: { ... }                    # 当前生效的 values
  resources:                          # K8s 资源列表
    - kind: StatefulSet
      name: my-redis-master
      ready: "1/1"
    - kind: Service
      name: my-redis-master
      type: ClusterIP
  notes: "..."                        # Helm NOTES.txt 输出

# 升级 Release
PUT /api/v1/helm/releases/default/my-redis
Body:
  version: "18.7.0"                   # 升级到新版本
  values:
    master:
      persistence:
        size: "100Gi"

# 回滚 Release
POST /api/v1/helm/releases/default/my-redis/rollback
Body: { revision: 2 }

# 查看版本历史
GET /api/v1/helm/releases/default/my-redis/history
Response:
  - revision: 1
    chart: redis-18.6.1
    status: superseded
    updated: "2026-02-07T02:00:00Z"
  - revision: 2
    chart: redis-18.6.1
    status: superseded
  - revision: 3
    chart: redis-18.7.0
    status: deployed

# 卸载 Release
DELETE /api/v1/helm/releases/default/my-redis

# WebSocket 事件
WS /ws/events
→ { type: "helm.install.progress", data: { name: "my-redis", phase: "installing" } }
→ { type: "helm.install.done", data: { name: "my-redis", status: "deployed" } }
```

#### 3.4.5 模块化部署 API（DAG，复用 deployrunner）

```yaml
# 上传应用包（包含多个 HelmComponent + 依赖关系声明）
POST /api/v1/apps/upload
Body: <multipart/form-data: app-package.tar.gz>
Response:
  name: kweaver
  version: "3.0.0"
  components:
    - { name: kweaver-web, type: helm, version: "3.0.0" }
    - { name: kweaver-api, type: helm, version: "3.0.0", depends: [mariadb, redis] }
    - { name: kweaver-engine, type: helm, version: "3.0.0", depends: [kafka, opensearch] }
  graph: { edges: [...] }            # DAG 依赖图

# 应用列表
GET /api/v1/apps/
Response:
  - name: kweaver
    version: "3.0.0"
    status: installed
    components: 3

# 安装应用（DAG 并行执行）
POST /api/v1/apps/kweaver/install
Body:
  config:
    kweaver-api:
      replicas: 2
    kweaver-engine:
      resources: { cpu: "2", memory: "4Gi" }
Response: { job_id: 42 }

# 任务列表
GET /api/v1/jobs/
Response:
  - id: 42
    app: kweaver
    type: install
    status: running
    progress: 60
    components:
      - { name: kweaver-web, status: done }
      - { name: kweaver-api, status: running, progress: 80 }
      - { name: kweaver-engine, status: pending }

# 启动/取消任务
POST  /api/v1/jobs/42/start
PATCH /api/v1/jobs/42/cancel

# WebSocket 任务进度
WS /ws/events
→ { type: "job.progress", data: { job_id: 42, component: "kweaver-api", progress: 90 } }
→ { type: "job.done", data: { job_id: 42, status: "success" } }
```

### 3.5 认证设计

```go
// 两种认证模式，通过配置切换
type AuthMode string
const (
    AuthLocal AuthMode = "local"   // 默认：JWT + 本地用户（存储在 K8s Secret）
    AuthOIDC  AuthMode = "oidc"    // 可选：对接外部 OIDC Provider
)

// 本地认证流程
// 1. 首次启动时，如果没有用户，自动创建 admin 用户，密码随机生成并打印到日志
// 2. 登录后返回 JWT Token（有效期 24h，可配置）
// 3. 所有 /api/* 请求需要 Bearer Token（白名单：/api/v1/auth/login, /health）
// 4. 用户信息存储在 K8s Secret（proton-server-users）中
```

### 3.6 状态监控设计（StatusWatcher）

```go
// StatusWatcher 基于 K8s Informer 实时监控数据服务状态
type StatusWatcher struct {
    k8s       kubernetes.Interface
    informer  cache.SharedInformerFactory
    eventBus  *EventBus
    services  map[string]*DataService  // 缓存
    mu        sync.RWMutex
}

// Watch 启动监控
func (w *StatusWatcher) Watch(ctx context.Context) {
    // 监控 Pod 状态变化
    w.informer.Core().V1().Pods().Informer().AddEventHandler(
        cache.ResourceEventHandlerFuncs{
            UpdateFunc: func(old, new interface{}) {
                pod := new.(*corev1.Pod)
                // 匹配数据服务 label
                if svcType, ok := pod.Labels["proton.io/data-service"]; ok {
                    status := w.computeStatus(svcType, pod.Labels["app.kubernetes.io/instance"])
                    w.eventBus.Publish(Event{
                        Type: "datasvc.status",
                        Data: status,
                    })
                }
            },
        },
    )
    // 监控 Helm Release（通过 Secret with label owner=helm）
    w.informer.Core().V1().Secrets().Informer().AddEventHandler(...)
}
```

---

## 三½、节点代理设计（proton-agent）

### 3.7 proton-agent 概述

proton-agent 是一个 **DaemonSet**，在每个需要主机操作的节点上运行。核心职责：

| 职责 | 说明 |
|------|------|
| **hostPath 目录管理** | 创建/删除/检查数据服务的主机目录（如 `/data/mariadb`） |
| **磁盘信息采集** | 上报各挂载点的容量/使用率/IO 状态 |
| **目录权限设置** | 设置 UID/GID 确保 Pod 能正确读写 |
| **数据清理** | 数据服务卸载后清理残留目录（需用户确认） |
| **节点健康探测** | 磁盘健康、目录可写性检查 |

### 3.8 proton-agent 架构

```
┌─────────────────────────────────────────────────────────────────┐
│  proton-agent (DaemonSet Pod on each node)                       │
│                                                                  │
│  Deployment:                                                     │
│    hostPID: false                                                │
│    volumes:                                                      │
│      - hostPath: /data (或用户自定义根目录，通过 Helm values 配置) │
│        mountPath: /host-data                                     │
│    securityContext:                                               │
│      privileged: false                                           │
│      capabilities: [SYS_ADMIN]  # 仅需创建目录/查询磁盘           │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  gRPC Server (:50051)                                      │  │
│  │                                                            │  │
│  │  rpc PrepareHostPath(req) → 创建目录 + 设置权限             │  │
│  │  rpc CleanupHostPath(req) → 删除目录（需确认标记）          │  │
│  │  rpc GetDiskInfo(req)     → 返回磁盘容量/使用率             │  │
│  │  rpc ListDirectories(req) → 列出已有数据目录                │  │
│  │  rpc HealthCheck(req)     → 节点和磁盘健康状态              │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Periodic Reporter                                         │  │
│  │  每 30s 将节点信息写入 Node Annotations:                    │  │
│  │    proton.io/disk-info: '{"total":"500G","used":"120G"}'   │  │
│  │    proton.io/agent-status: "healthy"                       │  │
│  │    proton.io/agent-version: "1.0.0"                        │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.9 proton-server ↔ proton-agent 通信

```
proton-server                              proton-agent (on target node)
     │                                          │
     │  方案: gRPC via K8s Service (headless)    │
     │                                          │
     │  1. proton-server 发现 agent:             │
     │     kubectl get pods -l app=proton-agent  │
     │     → 获取每个节点上 agent 的 Pod IP       │
     │                                          │
     │  2. 直接 gRPC 调用:                       │
     │     agent-pod-ip:50051/PrepareHostPath    │
     │     ──────────────────────────────────→   │
     │                                          │  创建 /data/mariadb
     │     ←──────────────────────────────────   │  设置 chown 999:999
     │     { success: true, path: "/data/mariadb" }
     │                                          │
     │  3. 或通过 Node Annotation 读取信息:      │
     │     kubectl get node <name> -o json       │
     │     → annotations["proton.io/disk-info"]  │
     │                                          │
```

### 3.10 hostPath 数据服务的创建流程

```
用户在 Web UI 创建数据服务（hostPath 模式）

  1. 用户选择:
     - 服务类型: MariaDB
     - 存储模式: hostPath
     - 节点: node-1 (从节点列表选择，agent 上报了磁盘信息)
     - 路径: /data/mariadb (可自定义)
     - 大小: 50Gi

  2. proton-server:
     ├── gRPC → proton-agent@node-1: PrepareHostPath("/data/mariadb", uid=999)
     │   └── agent 创建目录、设置权限、返回成功
     ├── 创建 PersistentVolume (local type):
     │   spec:
     │     capacity: { storage: 50Gi }
     │     local: { path: /data/mariadb }
     │     nodeAffinity: { node-1 }
     ├── 创建 PersistentVolumeClaim 绑定该 PV
     └── helm install mariadb --set persistence.existingClaim=<pvc-name>

  3. Helm Chart 的 StatefulSet:
     └── Pod 调度到 node-1（nodeAffinity）
     └── 挂载 PVC → 实际使用 /data/mariadb
```

### 3.11 proton-agent 项目结构

```
proton-agent/
├── cmd/
│   └── proton-agent/
│       └── main.go                  # DaemonSet 入口
├── internal/
│   ├── server/
│   │   └── grpc.go                  # gRPC 服务实现
│   ├── hostpath/
│   │   ├── manager.go               # 目录创建/删除/权限管理
│   │   └── validator.go             # 路径安全校验（防止 ../../ 等）
│   ├── disk/
│   │   ├── info.go                  # 磁盘信息采集（df、lsblk）
│   │   └── health.go                # 磁盘健康检查
│   └── reporter/
│       └── node_annotation.go       # 定期上报节点信息到 K8s Annotation
├── api/
│   └── proto/
│       └── agent.proto              # gRPC 接口定义
├── go.mod
└── Dockerfile
```

### 3.12 agent gRPC 接口定义

```protobuf
syntax = "proto3";
package proton.agent.v1;

service AgentService {
    // 目录管理
    rpc PrepareHostPath(PrepareHostPathRequest) returns (PrepareHostPathResponse);
    rpc CleanupHostPath(CleanupHostPathRequest) returns (CleanupHostPathResponse);
    rpc ListHostPaths(ListHostPathsRequest) returns (ListHostPathsResponse);

    // 磁盘信息
    rpc GetDiskInfo(GetDiskInfoRequest) returns (GetDiskInfoResponse);

    // 健康检查
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

message PrepareHostPathRequest {
    string path = 1;        // 主机目录路径，如 "/data/mariadb"
    int32  uid  = 2;        // 目录所有者 UID（不同数据库的运行用户不同）
    int32  gid  = 3;        // 目录所有者 GID
    string mode = 4;        // 权限模式，如 "0755"
}

message PrepareHostPathResponse {
    bool   success = 1;
    string path    = 2;     // 实际创建的路径
    string error   = 3;
}

message GetDiskInfoResponse {
    repeated DiskMount mounts = 1;
}

message DiskMount {
    string mount_point = 1;  // 如 "/data"
    string filesystem  = 2;  // 如 "ext4"
    uint64 total_bytes = 3;
    uint64 used_bytes  = 4;
    uint64 avail_bytes = 5;
}
```

### 3.13 存储策略总览

```
┌────────────────────────────────────────────────────────────────┐
│  proton-server 创建数据服务时的存储决策                          │
│                                                                │
│  用户选择存储模式:                                              │
│                                                                │
│  ┌─ hostPath 模式（离线/裸机）────────────────────────────┐    │
│  │                                                        │    │
│  │  前端: 选择节点 → 输入路径 → 选择大小                    │    │
│  │                                                        │    │
│  │  后端:                                                  │    │
│  │  1. gRPC → agent@node: PrepareHostPath(path, uid, gid) │    │
│  │  2. 创建 local PersistentVolume (nodeAffinity)          │    │
│  │  3. 创建 PersistentVolumeClaim                          │    │
│  │  4. helm install --set persistence.existingClaim=...    │    │
│  │                                                        │    │
│  │  卸载时:                                                │    │
│  │  1. helm uninstall                                      │    │
│  │  2. 删除 PVC + PV                                       │    │
│  │  3. 可选: gRPC → agent: CleanupHostPath(path)           │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌─ StorageClass 模式（托管 K8s / 有 CSI）───────────────┐    │
│  │                                                        │    │
│  │  前端: 选择 StorageClass → 选择大小                     │    │
│  │                                                        │    │
│  │  后端:                                                  │    │
│  │  1. helm install --set persistence.storageClass=...     │    │
│  │     --set persistence.size=50Gi                         │    │
│  │  2. Helm Chart 自动创建 PVC，CSI 驱动动态分配存储        │    │
│  │                                                        │    │
│  │  卸载时:                                                │    │
│  │  1. helm uninstall                                      │    │
│  │  2. PVC reclaimPolicy 决定是否自动清理                   │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌─ 外部服务模式 ─────────────────────────────────────────┐    │
│  │                                                        │    │
│  │  前端: 填写连接信息 (host/port/user/pass)               │    │
│  │  后端: 仅保存 ConnectInfo，无 Helm/存储操作              │    │
│  └────────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────────┘
```

---

## 四、前端设计（proton-web）

### 4.1 技术栈

| 选型 | 方案 | 理由 |
|------|------|------|
| **框架** | React 18 + TypeScript | 沿用现有技术栈，降低迁移成本 |
| **构建** | Vite | 替代 webpack，开发体验更好 |
| **UI 库** | Ant Design 5 | 替代 `@aishutech/ui`（API 兼容），公共 npm 包 |
| **状态管理** | Zustand | 轻量级，替代 Class Component state |
| **路由** | React Router v6 | 沿用 |
| **i18n** | react-i18next | 替代 `@anyshare/i18nfactory`，支持中英文 |
| **图表** | ECharts / Recharts | 数据服务监控面板 |
| **实时通信** | 原生 WebSocket | 安装进度/服务状态实时更新 |

### 4.2 页面结构

```
proton-web/
├── src/
│   ├── pages/
│   │   ├── wizard/                    # 安装向导（来自 proton-cli-web）
│   │   │   ├── WizardLayout.tsx       # 向导布局（Steps）
│   │   │   ├── ChooseTemplate.tsx     # Step 0: 部署模式
│   │   │   ├── NodeConfig.tsx         # Step 1: 节点配置
│   │   │   ├── NetworkConfig.tsx      # Step 2: K8s 配置
│   │   │   ├── CRConfig.tsx           # Step 3: 仓库配置
│   │   │   ├── DataServiceConfig.tsx  # Step 4: 数据服务（重设计 ★）
│   │   │   ├── ReviewAndApply.tsx     # Step 5: 配置预览 + 确认（新增 ★）
│   │   │   └── Progress.tsx           # 安装进度（WebSocket 实时 ★）
│   │   │
│   │   ├── dashboard/                 # 集群概览
│   │   │   └── Dashboard.tsx          # 节点/服务/应用概览
│   │   │
│   │   ├── datasvc/                   # 数据服务管理 ★
│   │   │   ├── DataServiceList.tsx    # 数据服务列表（状态总览）
│   │   │   ├── DataServiceDetail.tsx  # 数据服务详情
│   │   │   ├── CreateDataService.tsx  # 创建数据服务表单
│   │   │   ├── EditDataService.tsx    # 编辑配置
│   │   │   └── components/
│   │   │       ├── StatusBadge.tsx    # 状态徽标
│   │   │       ├── ConnectInfoCard.tsx# 连接信息卡片
│   │   │       ├── PodStatusTable.tsx # Pod 状态表格
│   │   │       ├── ResourceChart.tsx  # CPU/Memory 图表
│   │   │       └── ServiceTypeIcon.tsx# 服务类型图标
│   │   │
│   │   ├── apps/                      # 应用管理
│   │   │   ├── AppStore.tsx           # 应用商店
│   │   │   ├── AppDetail.tsx          # 应用详情
│   │   │   └── JobList.tsx            # 安装任务列表
│   │   │
│   │   ├── system/                    # 系统设置
│   │   │   ├── CertManagement.tsx     # 证书管理
│   │   │   ├── AccessConfig.tsx       # 访问地址配置
│   │   │   └── UserManagement.tsx     # 用户管理
│   │   │
│   │   └── login/
│   │       └── Login.tsx
│   │
│   ├── stores/                        # Zustand 状态
│   │   ├── cluster.ts
│   │   ├── datasvc.ts
│   │   ├── apps.ts
│   │   └── auth.ts
│   │
│   ├── hooks/
│   │   ├── useWebSocket.ts           # WebSocket 连接管理
│   │   ├── useDataService.ts         # 数据服务 CRUD hooks
│   │   └── useClusterStatus.ts       # 集群状态 hook
│   │
│   ├── api/                           # API 客户端
│   │   ├── client.ts                  # Axios + Token 拦截器
│   │   ├── cluster.ts
│   │   ├── datasvc.ts
│   │   └── auth.ts
│   │
│   └── i18n/
│       ├── zh-CN.json
│       └── en-US.json
```

### 4.3 数据服务管理页面设计 ★

#### 数据服务列表页

```
┌─────────────────────────────────────────────────────────────────┐
│  Proton Console                  admin ▾                         │
├──────┬──────────────────────────────────────────────────────────┤
│      │  数据服务                                     [+ 添加服务] │
│  📊  │                                                          │
│ 概览  │  ┌─ 筛选 ──────────────────────────────────────────────┐ │
│      │  │ 类型: [全部▾]  来源: [全部▾]  状态: [全部▾]  🔍 搜索  │ │
│  🗄️  │  └──────────────────────────────────────────────────────┘ │
│ 数据  │                                                          │
│ 服务  │  ┌──────┬──────────┬────────┬────────┬─────────┬──────┐ │
│      │  │ 名称  │ 类型      │ 来源   │ 状态    │ 连接地址  │ 操作 │ │
│  📦  │  ├──────┼──────────┼────────┼────────┼─────────┼──────┤ │
│ 应用  │  │🐬 mariadb │ RDS    │ 内置 │ ● 运行中│ :3306   │ ⚙ 🗑│ │
│      │  │🍃 mongodb  │MongoDB │ 内置 │ ● 运行中│ :27017  │ ⚙ 🗑│ │
│  ⚙️  │  │🔴 redis    │ Redis  │ 内置 │ ● 运行中│ :6379   │ ⚙ 🗑│ │
│ 系统  │  │🔍 opensrch │OpenSrch│ 内置 │ ● 运行中│ :9200   │ ⚙ 🗑│ │
│      │  │📨 kafka    │ Kafka  │ 内置 │ ⚠ 异常  │ :9092   │ ⚙ 🗑│ │
│      │  │🐘 ext-pg   │ RDS    │ 外置 │ ● 已连接│10.0.1.5 │ ⚙ 🗑│ │
│      │  └──────┴──────────┴────────┴────────┴─────────┴──────┘ │
│      │                                                          │
│      │  总计 6 个服务 | 5 运行中 | 1 异常                         │
└──────┴──────────────────────────────────────────────────────────┘
```

#### 数据服务详情页

```
┌─────────────────────────────────────────────────────────────────┐
│  ← 返回    mariadb (RDS - MariaDB)           [编辑] [重启] [删除]│
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ 状态 ──────────────────────────────────────────────────────┐│
│  │  ● 运行中    Ready: 1/1    运行时间: 15d 3h                  ││
│  │  Helm Release: mariadb (mariadb-11.0.0)  Namespace: resource ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─ 连接信息 ──────────────────────────────────────────────────┐│
│  │  Host:     mariadb.resource.svc.cluster.local     [📋 复制] ││
│  │  Port:     3306                                             ││
│  │  Username: root                                             ││
│  │  Password: ●●●●●●●●●●                     [👁 显示] [📋]    ││
│  │                                                             ││
│  │  连接命令: mysql -h mariadb.resource -u root -p     [📋]    ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─ 配置 ──────────────────────────────────────────────────────┐│
│  │  部署节点: node1                                             ││
│  │  数据目录: /data/mariadb                                     ││
│  │  副本数:   1                                                 ││
│  │  CPU:     500m / Memory: 1Gi                                 ││
│  │  存储:    50Gi (PVC: mariadb-data)                           ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─ Pod 状态 ──────────────────────────────────────────────────┐│
│  │  Pod 名称           │ 节点   │ 状态  │ 重启  │ 创建时间       ││
│  │  mariadb-0          │ node1 │ ● Run │  0   │ 15d ago       ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─ 依赖关系 ──────────────────────────────────────────────────┐│
│  │  被以下应用依赖: deploy-service, kweaver-studio              ││
│  └──────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

#### 创建数据服务

```
┌─────────────────────────────────────────────────────────────────┐
│  添加数据服务                                                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  来源选择:  ◉ 内置部署    ○ 外置连接                               │
│                                                                  │
│  服务类型:                                                        │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐        │
│  │ 🐬     │ │ 🍃     │ │ 🔴     │ │ 🔍     │ │ 📨     │        │
│  │ MariaDB│ │MongoDB │ │ Redis  │ │OpenSrch│ │ Kafka  │        │
│  │  ✓     │ │        │ │        │ │        │ │        │        │
│  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘        │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐                   │
│  │ 🔐     │ │ 📊     │ │ 📈     │ │ 🌐     │                   │
│  │ ETCD   │ │Promthus│ │Grafana │ │ Nebula │                   │
│  └────────┘ └────────┘ └────────┘ └────────┘                   │
│                                                                  │
│  ── MariaDB 配置 ──────────────────────────────────────         │
│                                                                  │
│  实例名称:  [mariadb            ]                                │
│  部署节点:  [node1 ▾] [+ 添加节点]                                │
│  数据目录:  [/data/mariadb      ]                                │
│  Root 密码: [●●●●●●●] [🔄 随机生成]                              │
│  存储大小:  [50  ] Gi                                            │
│  资源配额:  CPU [500 ]m  Memory [1024]Mi                         │
│                                                                  │
│                              [取消]  [创建]                       │
└─────────────────────────────────────────────────────────────────┘
```

### 4.4 安装向导增强（Step 4 重设计 ★）

原来 proton-cli-web 的 Step 4 (DataBaseConfig) 一次展示 15+ 服务，信息过载。新设计**分层展示**：

```
┌─────────────────────────────────────────────────────────────────┐
│  Step 4: 数据服务配置                                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ── 核心服务（必选）──────────────────────────────────────────── │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ☑ MariaDB (RDS)    ◉ 内置  ○ 外置   存储:[50Gi]         │   │
│  │ ☑ Redis            ◉ 内置  ○ 外置   存储:[10Gi]         │   │
│  │ ☑ ETCD             ◉ 内置  ○ 外置   存储:[20Gi]         │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ── 可选服务 ────────────────────────────── [展开高级配置 ▾]  ── │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ☐ MongoDB           ◉ 内置  ○ 外置                       │   │
│  │ ☐ OpenSearch         ◉ 内置  ○ 外置                       │   │
│  │ ☐ Kafka + ZooKeeper  ◉ 内置  ○ 外置                       │   │
│  │ ☐ Prometheus + Grafana                                    │   │
│  │ ☐ Nebula Graph                                            │   │
│  │ ☐ PolicyEngine                                            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ── 已选服务预览 ─────────────────────────────────────────────  │
│  MariaDB(内置,50Gi) + Redis(内置,10Gi) + ETCD(内置,20Gi)        │
│  预计资源占用: CPU 1.5核 / Memory 3Gi / Storage 80Gi          │
│                                                                  │
│                          [上一步]  [下一步: 配置预览]              │
└─────────────────────────────────────────────────────────────────┘
```

### 4.5 新增 Step 5: 配置预览 ★

```
┌─────────────────────────────────────────────────────────────────┐
│  Step 5: 配置预览                                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  请确认以下配置，点击"开始安装"后将不可中断                          │
│                                                                  │
│  ┌─ 部署模式 ──────────┐  ┌─ 存储 ───────────────────────┐       │
│  │ 标准模式 (3节点)     │  │ StorageClass: gp3 (默认)    │       │
│  └─────────────────────┘  │ Namespace: proton-data      │       │
│                            └─────────────────────────┘       │
│  ┌─ Kubernetes ────────┐  ┌─ 数据服务 ──────────────────┐    │
│  │ Pod CIDR: 10.244/16 │  │ MariaDB  (内置, 50Gi)         │    │
│  │ Svc CIDR: 10.96/12  │  │ Redis    (内置, 10Gi)         │    │
│  └─────────────────────┘  │ ETCD     (内置, 20Gi)         │    │
│                            └──────────────────────────────┘    │
│                                                                  │
│  ┌─ 生成的 ClusterConfig (YAML) ──────────── [📋 复制] [💾 下载]│
│  │ nodes:                                                      │ │
│  │   - name: node1                                             │ │
│  │     ipv4: 10.0.0.1                                          │ │
│  │ cs:                                                         │ │
│  │   master: [node1]                                           │ │
│  │   ...                                                       │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│                   [上一步]  [开始安装 →]                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## 五、从现有代码复用的策略

### 5.1 Go 代码复用

| 来源 | 复用内容 | 复用方式 |
|------|---------|---------|
| **proton-cli** `pkg/core/apply/` | 完整 apply 链（firewall→nodes→cr→cs→modules） | 直接 import 为 Go package |
| **proton-cli** `pkg/configuration/` | ClusterConfig 结构体 + 校验 + K8s 读写 | 直接 import |
| **proton-cli** `pkg/client/helm3/` | Helm3 客户端 | 直接 import |
| **proton-cli** `pkg/client/registry/` | Registry 客户端 | 直接 import |
| **proton-cli** `pkg/proton/componentmanage/` | 数据组件安装逻辑 | 直接 import |
| **deployrunner** `pkg/component/` | Helm Component 抽象 | 适配后 import |
| **deployrunner** `api/rest/proton_component/` | 数据服务 CRUD 逻辑 | 核心逻辑提取为 service 层 |
| **deployrunner** `pkg/app/executor/` | DAG 任务执行器 | 适配后 import |
| **deployrunner** `pkg/graph/` | DAG 图算法 | 直接 import |
| **deploy-service** `helm.py` | Helm 操作逻辑 | 已有 Go 等价实现 |
| **component-manage** | 组件生命周期 | 直接 import |

### 5.2 前端代码复用

| 来源 | 复用内容 | 复用方式 |
|------|---------|---------|
| **proton-cli-web** | 安装向导 5 步表单 + 校验逻辑 | 重构为 React Hooks + 函数组件 |
| **proton-cli-web** `helper.ts` | 数据转换 `exChangeData()`、IP 校验函数 | 提取为 utils 库 |
| **proton-cli-web** `index.d.ts` | TypeScript 类型定义 | 直接复用 |
| **deploy-web-static** | 证书管理/系统配置页面 | 适配后复用 |
| **全新开发** | 数据服务管理页面 ★ | 新写 |
| **全新开发** | Dashboard 概览页 ★ | 新写 |

### 5.3 可以直接删除的

| 组件 | 理由 |
|------|------|
| **deploy-web-service** (Node.js BFF) | 完全被 proton-server 替代 |
| **deploy-service** (Python/Tornado) | 功能合并到 proton-server |
| **proton-cli serve.go** | 被 proton-server 替代 |
| **cmd/download-proton-cli-web** | 前端已内嵌，不需要 FTP 下载 |

---

## 六、构建与部署

### 6.1 构建产物：Docker 镜像（非单二进制）

```dockerfile
# Dockerfile (multi-stage)

# Stage 1: 构建前端
FROM node:20-alpine AS web-builder
WORKDIR /app
COPY proton-web/package*.json ./
RUN npm ci
COPY proton-web/ ./
RUN npm run build

# Stage 2: 构建后端
FROM golang:1.22-alpine AS server-builder
WORKDIR /app
COPY proton-server/ ./
COPY --from=web-builder /app/dist ./web/
RUN CGO_ENABLED=0 go build -o proton-server ./cmd/proton-server/

# Stage 3: 最终镜像（包含 Helm CLI 用于 in-cluster 操作）
FROM alpine:3.19
RUN apk add --no-cache helm kubectl
COPY --from=server-builder /app/proton-server /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["proton-server"]
```

```bash
# Makefile
.PHONY: image

# 构建容器镜像
image:
	docker build -t ghcr.io/kweaver-ai/proton-server:latest .

# 推送到公共仓库
push:
	docker push ghcr.io/kweaver-ai/proton-server:latest

# 本地开发（可选：Go 二进制直连远程 K8s，用于开发调试）
dev:
	cd proton-web && npm run build && cp -r dist/* ../proton-server/web/
	cd proton-server && go run ./cmd/proton-server/ --kubeconfig=~/.kube/config
```

### 6.2 部署方式：Helm Chart（唯一方式）

```yaml
# proton-server Helm Chart values.yaml
replicaCount: 1

image:
  repository: ghcr.io/kweaver-ai/proton-server
  tag: "latest"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP          # 托管 K8s 可改为 LoadBalancer
  port: 8080

serviceAccount:
  create: true
  name: proton-admin
  # RBAC: 需要 Helm/Pod/Secret/ConfigMap/Namespace 等权限

config:
  authMode: "local"         # local | oidc
  logLevel: "info"
  chartRepo: "oci://ghcr.io/kweaver-ai/charts"
  dataServiceNamespace: "proton-data"

# proton-agent DaemonSet 配置
agent:
  enabled: true             # 离线/裸机: true、托管 K8s: false
  image:
    repository: ghcr.io/kweaver-ai/proton-agent
    tag: "latest"
  hostDataRoot: "/data"     # agent 挂载的主机根目录
  resources:
    cpu: "100m"
    memory: "128Mi"

ingress:
  enabled: false
  className: "nginx"
  hosts:
    - host: proton.example.com
      paths: [{ path: "/" }]

persistence:
  enabled: true
  storageClass: ""
  size: 1Gi
```

```bash
# 路径 A: 裸机 / 离线 — proton-cli 自动安装（agent 默认开启）
# proton-cli apply 链的最后一步自动执行:
#   helm install proton proton/proton-server \
#     --set agent.enabled=true \
#     --set agent.hostDataRoot=/data

# 路径 B: 托管 K8s — 用户手动安装（agent 关闭）
helm repo add proton https://kweaver-ai.github.io/charts
helm install proton proton/proton-server \
  --namespace proton-system --create-namespace \
  --set service.type=LoadBalancer \
  --set agent.enabled=false

# 路径 C: 自建 K8s + 本地盘 — 手动安装（agent 开启）
helm install proton proton/proton-server \
  --namespace proton-system --create-namespace \
  --set agent.enabled=true \
  --set agent.hostDataRoot=/mnt/data

# 安装完成后
kubectl get svc -n proton-system proton-server
# → EXTERNAL-IP:8080 即可访问 Web UI
```

### 6.3 两条路径的完整流程

```
═══════════════════════════════════════════════════════════════════════
路径 A: 裸机 / VM / 离线（proton-cli 向导 → 自动跳转 proton-server）
═══════════════════════════════════════════════════════════════════════

  1. curl -sfL https://get.proton.kweaver.ai | sh          # 安装 proton-cli
  2. proton-cli serve                                       # 启动 Web 向导 :3000
  3. 浏览器打开 http://管理机:3000                            # Web 向导配置
     → Step 1-4: 节点/K8s/CR/Ingress 配置
     → Step 5: 确认 → 开始安装
  4. proton-cli apply chain:                                # 后台执行
     → firewall → nodes → cr → cs                          # K8s 基础设施
     → ingress-nginx                                        # Ingress 控制器
     → proton-server + proton-agent                         # 管理平台
  5. proton-cli 轮询 GET https://<ingress>/api/v1/health    # 等待 proton-server Ready
  6. 向导页面自动跳转 → https://<ingress-host>/              # 浏览器重定向
  7. proton-server 首次登录（admin/<随机密码>）               # 进入 Dashboard
     → 内嵌向导: 配置数据服务 → 安装应用                     # 数据服务在此管理
  8. proton-cli 可关闭                                      # 后续全在 Web UI

  ⚠️ 如需扩缩容节点/升级 K8s → 仍需 proton-cli（主机级操作）

═══════════════════════════════════════════════════════════════════════
路径 B: 托管 K8s（EKS/AKS/GKE）/ 已有 K8s
═══════════════════════════════════════════════════════════════════════

  1. helm install proton-server proton/proton-server \       # 一条命令
       --set service.type=LoadBalancer
  2. 打开 http://<LB-IP>:8080                                # 自动创建 admin
  3. Web UI 向导引导:                                         # 无需 proton-cli
     → 选择数据服务 (MariaDB/Redis/...) → internal/external
     → 选择 StorageClass → 配置资源配额
     → 一键安装（Helm in-cluster）
  4. 数据服务运行在同一 K8s 集群中

  ✅ 完全不需要 proton-cli、SSH、主机访问

═══════════════════════════════════════════════════════════════════════
```

### 6.4 RBAC 设计

```yaml
# proton-server ServiceAccount 需要的 K8s 权限
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: proton-server
rules:
  # 读取集群信息
  - apiGroups: [""]
    resources: ["nodes", "namespaces", "pods", "services", "events",
                "persistentvolumeclaims", "configmaps", "secrets"]
    verbs: ["get", "list", "watch"]
  # 管理数据服务相关资源（在 proton-data namespace）
  - apiGroups: [""]
    resources: ["namespaces", "secrets", "configmaps",
                "services", "persistentvolumeclaims"]
    verbs: ["create", "update", "patch", "delete"]
  # StorageClass（只读）
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list"]
  # Helm 操作需要的 Secret 权限（Helm 在 Secret 中存储 release 信息）
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # Deployment/StatefulSet（数据服务通常是 StatefulSet）
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "daemonsets"]
    verbs: ["get", "list", "watch"]
  # PersistentVolume（hostPath 模式需要创建 local PV）
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # Node annotations（读取 agent 上报的磁盘信息）
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
```

---

## 七、实施计划

### Phase 1: proton-server + proton-agent 骨架（3 周）

| # | 任务 | 天 | 说明 |
|---|------|-----|------|
| 1 | proton-server 项目初始化 | 2 | Gin 框架 + 项目结构 + embed 静态文件 |
| 2 | JWT 本地认证 | 2 | login/logout/me + K8s Secret 存储用户 |
| 3 | **proton-agent** | 3 | DaemonSet + gRPC 服务 + hostPath 管理 + 磁盘信息采集 |
| 4 | NodeMgr + ClusterInfoMgr | 2 | agent 通信 + K8s 只读查询 + 存储策略分发 |
| 5 | **DataServiceManager 核心** | 5 | Helm CRUD + hostPath/StorageClass 双模式 + StatusWatcher |
| 6 | 数据服务 REST API | 1 | 全部 CRUD + health + test-connection |

### Phase 2: proton-web 统一前端（3 周）

| # | 任务 | 天 | 说明 |
|---|------|-----|------|
| 7 | 前端项目初始化 | 1 | Vite + React 18 + Ant Design 5 + i18n |
| 8 | 统一导航 + Dashboard | 2 | Rancher 风格左侧导航 + 集群概览 |
| 9 | 节点管理页面 | 2 | 节点列表、磁盘信息（agent）、hostPath 目录 |
| 10 | 存储管理页面 | 1 | StorageClass / PVC / hostPath 卷列表 |
| 11 | **数据服务管理页面** | 5 | 列表/详情/创建（hostPath+SC 双模式）/编辑/状态 |
| 12 | 首次安装向导（内嵌） | 3 | 从 proton-cli-web 迁移，内嵌在统一 UI 中，非独立界面 |
| 13 | WebSocket + 系统设置 | 2 | 实时状态 + 证书/用户管理 |

### Phase 3: 应用管理 + Helm 可视化（3 周）

| # | 任务 | 天 | 说明 |
|---|------|-----|------|
| 14 | AppMgr（DAG 执行器） | 3 | 复用 deployrunner 逻辑（executor/graph/store） |
| 15 | 模块化部署前端 | 2 | 应用包上传 + DAG 依赖图可视化 + 任务进度 |
| 16 | **HelmMgr 后端** | 4 | Chart 仓库索引/同步 + Chart 浏览 + Release CRUD/升级/回滚 |
| 17 | **Helm 可视化前端** | 4 | 仓库管理 + Chart 浏览(README/values) + Release 安装表单 + 历史 |

### Phase 4: 集成打磨（2 周）

| # | 任务 | 天 | 说明 |
|---|------|-----|------|
| 18 | proton-cli 向导衔接 | 2 | apply 链加 ingress-nginx + proton-server，健康探测 + 自动跳转 |
| 19 | proton-cli-web 精简 | 2 | 仅保留 K8s 初始化向导（节点/K8s/CR/Ingress），移除数据服务配置 |
| 20 | Helm Chart 打包 | 2 | proton-server + proton-agent 统一 Chart |
| 21 | E2E 测试 + 文档 | 4 | 离线场景全流程 + 托管 K8s + Helm 管理场景测试 |

---

## 八、对比总结

### 改造前 vs 改造后

| 维度 | 改造前 | 改造后 |
|------|--------|--------|
| **后端服务** | 5 个（Go+Python+Node.js） | **proton-server**（Go）+ **proton-agent**（Go DaemonSet） |
| **前端项目** | 2 个（proton-cli-web + deploy-web-static） | **1 个**（proton-web，统一控制台） |
| **Web 界面** | 2 套割裂（安装向导 + 管理控制台） | **两阶段衔接**: proton-cli 向导→自动跳转 proton-server 控制台 |
| **编程语言** | Go + Python + Node.js + TypeScript | **Go + TypeScript** |
| **主机操作** | SSH 连接执行 | **DaemonSet Agent**（无 SSH 依赖） |
| **存储管理** | 硬编码 hostPath 在 apply YAML | **双模式**: hostPath（via agent）/ StorageClass，用户自选 |
| **认证** | Hydra OAuth2（重依赖） | **JWT 本地认证**（零依赖） |
| **数据服务管理** | 分散（初始化在向导，运行时在 deployrunner） | **统一管理界面** |
| **Helm 管理** | 无可视化，仅 CLI | **原生 Helm 可视化**（仓库/Chart/Release CRUD，对标 Rancher Apps） |
| **模块化部署** | deployrunner DAG（无前端） | **DAG + 可视化依赖图 + 任务进度** |
| **实时状态** | 轮询 `/alpha/result` | **WebSocket 实时推送** |
| **部署方式** | 5 个 Helm Chart | **1 个 Helm Chart**（包含 server + agent） |
| **离线支持** | 仅 proton-cli apply | **proton-server Web UI 直接管理**（agent 处理 hostPath） |

### 数据服务管理能力对比

| 能力 | 改造前 | 改造后 |
|------|--------|--------|
| 初始化配置 | ✅ proton-cli-web 向导（独立 UI） | ✅ 向导（内嵌在统一控制台中） |
| 运行时查看状态 | ⚠️ deployrunner API（无前端） | ✅ **Web UI 实时状态面板** |
| 动态添加服务 | ❌ 需要重新 apply | ✅ **Web UI 一键创建** |
| 动态删除服务 | ❌ 需要 helm uninstall | ✅ **Web UI 一键删除**（依赖检查） |
| 修改配置 | ⚠️ deployrunner API（无前端） | ✅ **Web UI 表单编辑** |
| 连接信息查看 | ⚠️ 需要 kubectl 查看 Secret | ✅ **Web UI 一键复制** |
| **hostPath 存储** | ⚠️ 仅 proton-cli YAML 配置 | ✅ **Web UI 选节点+路径，agent 自动创建** |
| **StorageClass 存储** | ❌ 不支持 | ✅ **Web UI 选 SC + 大小，动态分配** |
| **磁盘信息** | ❌ 需要 SSH 登录查看 | ✅ **Web UI 实时展示（agent 采集）** |
| 健康检查 | ❌ | ✅ **连接测试 + 健康探针** |
| 外部服务管理 | ⚠️ 仅 ConnectInfo | ✅ **连接测试 + 状态监控** |
