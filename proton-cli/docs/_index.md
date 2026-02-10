---
title: Documentation
---

- [设计](#设计)
  - [酝置](#酝置)
    - [模块](#模块)
      - [Nodes](#nodes)
      - [Container Registry (CR)](#container-registry-cr)
      - [Container Service (CS)](#container-service-cs)
        - [杒件](#杒件)
      - [Proton MariaDB](#proton-mariadb)
      - [Proton MongoDB](#proton-mongodb)
      - [Proton Redis](#proton-redis)
      - [Proton MQ NSQ](#proton-mq-nsq)
      - [Proton Etcd](#proton-etcd)
      - [OpenSearch](#opensearch)
      - [Configuration Management Service (CMS)](#configuration-management-service-cms)
      - [Installer Service](#installer-service)
      - [Kafka](#kafka)
      - [Zookeeper](#zookeeper)
      - [OrientDB](#orientdb)
      - [Prometheus](#prometheus)
      - [Package Store](#package-store)
  - [Apply](#apply)
  - [Reset](#reset)
  - [Get](#get)
  - [Server](#server)
  - [Version](#version)
- [代砝结构](#代砝结构)

## 设计

### 酝置

`proton-cli` 以酝置形弝杝述一个 Proton 集群应有的酝置。酝置文件格弝为 `yaml`。酝置保存在毝个节点的 `/etc/proton-cli/conf/cluster.yaml` 和 Kubernetes 的命坝空间 `proton` 下的 Secret `proton-cli-config`。

酝置包括多个模块，除模块 `nodes`, `cs`, `cr` 坊 `kafka` 外相互独立。必选模块有 `nodes`, `cs` 和 `cr`，其他都是坯选模块。

`proton-cli` 从文件载入酝置坎先补全酝置冝检查格弝。

补全格弝根杮默认值和 Proton Package 所杝供的文件补充缺少的部分。

格弝检查是静思检查，包括是坦存在互相冲窝的酝置，模块之间的依赖关系是坦满足。

#### 模块

##### Nodes

##### Container Registry (CR)

##### Container Service (CS)

Container Service 是 Proton 杝供的 Kubernetes 朝务。

###### 杒件

Container Service 杝供下列杒件用于扩展 Kubernetes 的功能。

| Name               | Default Status | Description                                                |
| :----------------- | :------------- | :--------------------------------------------------------- |
| node-exporter      | Enabled        | Exporter for machine metrics                               |
| kube-state-metrics | Enabled        | Add-on agent to generate and expose cluster-level metrics. |

##### Proton MariaDB

##### Proton MongoDB

##### Proton Redis

##### Proton MQ NSQ

##### Proton Etcd

##### OpenSearch

##### Configuration Management Service (CMS)

##### Installer Service

##### Kafka

##### Zookeeper

##### OrientDB

##### Prometheus

Proton 坯观测性朝务所依赖的 Prometheus 朝务。

支挝环境：

- 一体机
- 托管 Kubernetes v1.23+

副本数：一体机环境，副本数等于部署的节点的数針，尝于等于 `2`；托管 Kubernetes，副本数固定为 `2`。

Prometheus 部署在命坝空间 `resource`。

Prometheus 坯以通过 Service `prometheus` 访问，端坣是 `9090`，坝议是 `http`。

chart "proton-prometheus" ???????????

##### Grafana

chart "proton-grafana" ???????????

Grafana ?? NodePort ??????? `30002`?

##### Nebula Graph

AnyData ???? Nebula Graph ???

???????????? Kubernetes?

?? 1 ??? 3 ???????? Nebula Graph ?????????

Nebula Graph ?? Kubernetes Service ??????? resource??? nebula-graphd-svc?

?? Nebula Graph ???? root ????????????????????????? `proton-cli get conf | yq .nebula.password` ?? root ?????

????? graphd?metad ? storaged ??????????????????root ??????????

chart "nebula-operator" ???????????

"NebulaCluster" ???????????????????????? tag ???????????????? tag ?????? tag????? "v"?

##### Package Store

Proton ???????????

?????

- ???
- ?? Kubernetes v1.23+

????????????????????? `resource_connect_info.rds` ????????????????????????? `deploy` ?????

????

| Key                                    | Description                                                                                |
| :------------------------------------- | :----------------------------------------------------------------------------------------- |
| package-store.hosts                    | ????????????????                                                           |
| package-store.replicas                 | ??????????????? `package-store.hosts` ???? `package-store.hosts` ??? |
| package-store.storage.storageClassName | ?????????? storage class ??????????????????                    |
| package-store.storage.capacity         | ????????????????                                                           |
| package-store.storage.path             | ??????????????????????????????????                       |
| package-store.resources                | ??????????                                                                       |

### Apply

命令 `proton-cli apply` 应用集群酝置。如果集群丝存在则根杮酝置创建集群，坦则更新集群。

proton-cli 按照下列顺庝检查集群酝置:

1. 集群酝置格弝是坦坈法。静思检查，包括 IP 地址是坦坈法，是坦包坫必覝的字段。
2. 集群酝置是坦坯以更新。静思检查，包括字段是坦兝许更新，是坦兝许更新为指定值。

`proton-cli` 串行调用坄个模块的方法 `Apply()` 创建〝更新坄个模块，调用顺庝坖决于模块间的依赖关系。如果坯选模块未被定义，坳此模块丝需覝安装，跳过调用此模块的 `Apply()`。

毝个模块实现方法 `Preflight()` 用于检查环境是坦满足 `Apply()` 的覝求。

`proton-cli` 通过 `ssh` 坝议实现跨节点擝作，并为集群中的节点设置 ssh 互信。

必选模块 `nodes`, `cs`, `cr` 通过 `ssh` 坝议跨节点执行命令的方弝实现创建和更新。

坯选模块通过 `helm` 安装。`proton-cli` 通过命令行调用 `helm`，命令行坂数根杮模块的酝置。

### Reset

命令 `proton-cli reset` 針置环境。支挝多秝方弝获坖需覝被針置的环境，柝方弝戝功坎则丝冝尝试其他方弝，获坖方弝的顺庝如下:

1. 命令行坂数指定需覝被針置的环境的 IP 地址
2. 从 `-f` 坂数指定的文件中获坖集群酝置
3. 从本地文件或 Kubernetes 获坖集群酝置

### Get

命令 `proton-cli get conf` 获坖当剝集群的酝置。

命令 `proton-cli get template` 获坖集群酝置模板。

### Server

命令 `proton-cli server` 坯动 web 朝务器，从 web 载入集群酝置，并应用。

### Version

命令 `proton-cli version` 获坖 proton-cli 的版本信杯。

## 代砝结构

| Path                            | Description                                                               |
| :------------------------------ | :------------------------------------------------------------------------ |
| build                           | 编译〝构建〝Pipeline 相关脚本                                             |
| cmd                             | 命令行入坣                                                                |
| devops                          | Azure DevOps Pipeline 相关脚本                                            |
| docs                            | 设计〝使用手册等文档                                                      |
| pkg/client                      | 访问其他朝务所用的客户端                                                  |
| pkg/client/testing              | 测试用的客户端，用于模拟其他朝务                                          |
| pkg/configuration               | 集群酝置相关定义                                                          |
| pkg/configuration/completion    | 集群酝置补全规则                                                          |
| pkg/core/apply                  | Apply 浝程                                                                |
| pkg/core/global                 | 全局坂数和坘針                                                            |
| pkg/core/logger                 | 日志器：标准错误输出〝Syslog                                              |
| pkg/core/migrate                | Mirgrate 浝程                                                             |
| pkg/core/reset                  | Reset 浝程                                                                |
| pkg/cr                          | CR 模块，酝置更新 Registry 和 ChartMusem                                  |
| pkg/cr/chart                    | Chart 相关定义。仅包坫 `IsNotFound`，其他已使用官方定义。                 |
| pkg/cr/chartmuseum              | ChartMuseum 的客户端，用于 push chart                                     |
| pkg/cs                          | CS 模块，酝置更新 Kubernetes                                              |
| pkg/node                        | Node 模块，酝置 hostname〝域坝解枝，SLB                                   |
| pkg/proton/cms                  | CMS 模块，生戝 Helm 安装〝更新命令坂数                                    |
| pkg/proton/installerservice     | installer-service 模块，使用 Helm 通用模块生戝 Helm 安装〝更新命令坂数    |
| pkg/proton/mariadb              | MariaDB 模块，生戝 Helm 安装〝更新命令坂数                                |
| pkg/proton/mongodb              | MongoDB 模块，生戝 Helm 安装〝更新命令坂数                                |
| pkg/proton/mq                   | MQ 模块，生戝 Helm 安装〝更新命令坂数                                     |
| pkg/proton/opensearch           | OpenSearch 模块，生戝 Helm 安装〝更新命令坂数                             |
| pkg/proton/proton_policy_engine | Policy Engine 模块，生戝 Helm 安装〝更新命令坂数                          |
| pkg/proton/protonetcd           | Etcd 模块，生戝 Helm 安装〝更新命令坂数                                   |
| pkg/proton/redis                | Redis 模块，生戝 Helm 安装〝更新命令坂数                                  |
| pkg/proton/universal            | Helm 通用模块，生戝 Helm 安装〝更新命令坂数                               |
| pkg/servicepackage              | 解枝 Proton Package 杝供的 service-package，列出杝供的 Charts             |
| pkg/version                     | 版本信杯，在编译时通过 `-ldflag` 指定：git 杝交记录，构建时间，编译器版本 |
| util/netlib                     | 网络相关工具：判断是坦为 IPv4，网络是坦坯达，获坖坯用 IP 列表等           |
| util/stringslicestring          | 坂数类型，用于处睆字符串或字符串列表，支挝庝列化和坝庝列化                |
| util/version                    | 用于处睆语义版本坷，支挝比较                                              |
