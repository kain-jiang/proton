

# 一、schema index

[toc]

# 二、component schema

## Component

| 成员                | 类型                             | 说明                                                         |
| ------------------- | -------------------------------- | ------------------------------------------------------------ |
| name                | string                           | 组件名称，要求符合[RFC 1035](https://tools.ietf.org/html/rfc1035) |
| version             | string                           | 组件版本号，要求符合[语义化版本2.0](https://semver.org/lang/zh-CN/) |
| componentDefineType | string                           | 组件类型, 固定为fool                                         |
| configSchema        | object                           | 组件配置文档，要求使用[jsonschema](https://json-schema.org/) |
| AttributeSchema     | object                           | 组件属性文档，要求使用[jsonschema](https://json-schema.org/) |
| Attribute           | object                           | 组件属性默认值                                               |
| Dendence            | \[\][CompoentNode](#ComponentNode) | 依赖组件声明数组 |
| timeout | int | 组件安装/升级/更新执行超时时间默认值 |
| deploys | \[\][Deployment](#Deployment) | 无状态服务定义数组 |
| statefulset | \[\][Statefulset](#Statefulset) | 有状态服务定义数组 |

## ComponentNode
| 成员    | 类型   | 说明                                                         |
| ------- | ------ | ------------------------------------------------------------ |
| name    | string | 组件名称，要求符合[RFC 1035](https://tools.ietf.org/html/rfc1035) |
| version | string | 组件版本号，要求符合[语义化版本2.0](https://semver.org/lang/zh-CN/)。可选 |

## Deployment

| 成员          | 类型                    | 说明                                                         |
| ------------- | ----------------------- | ------------------------------------------------------------ |
| name          | string                  | 组件名称，要求符合[RFC 1035](https://tools.ietf.org/html/rfc1035) |
| replica       | [Replica](#Replica)     | 副本数定义                                                   |
| configSchema  | object                  | deployment配置文档，要求使用[jsonschema](https://json-schema.org/) |
| initContainer | [Container](#Container) | initcontainer定义                                            |
| containers    | \[\]Container           | 服务内container定义                                          |
| services      | \[][Service](#Service)  | 服务通信service定义                                          |

## Replica

| 成员           | 类型 | 说明                                                    |
| -------------- | ---- | ------------------------------------------------------- |
| custom         | bool | 副本数是否特殊输入,非特殊输入且为配置时受应用层配置影响 |
| DefaultReplica | int  | 默认副本数,正整数                                       |

## Service

| 成员 | 类型             | 说明                                                         |
| ---- | ---------------- | ------------------------------------------------------------ |
| name | string           | 组件名称，要求符合[RFC 1035](https://tools.ietf.org/html/rfc1035) |
| port | \[][Port](#Port) | 服务端口定义数组                                             |

## Port

| 成员       | 说明                                                         | 类型   |
| ---------- | ------------------------------------------------------------ | ------ |
| name       | 组件名称，要求符合[RFC 1035](https://tools.ietf.org/html/rfc1035) | string |
| port       | service服务端口值, 值区间[1,65536]                           | int    |
| targetPort | service后端服务端口值, 值区间[1,65536]                       | int    |
| protocol   | service协议, 仅"TCP", "UDP"                                  | string |

## Container

| 成员           | 类型                    | 说明                                                         |
| -------------- | ----------------------- | ------------------------------------------------------------ |
| name           | string                  | 容器名称，要求符合[RFC 1035](https://tools.ietf.org/html/rfc1035) |
| command        | []string                | 容器入口命令                                                 |
| args           | []string                | 容器入口命令参数                                             |
| image          | [Image](#Image)         | 容器运行所需镜像                                             |
| resources      | [Resources](#Resources) | 容器资源配置                                                 |
| livenessProbe  | [Probe](#Probe)         | 存活探针,当探针满足失败条件时会重启容器,但不会重启POD        |
| readinessProbe | [Probe](#Probe)         | 就绪探针,当探针满足失败条件时会将POD从service中移除进行熔断,直到满足探针成功条件 |
| startupProbe   | [Probe](#Probe)         | 启动探针,在容器启动时,只有探针满足成功条件时,才会认为容器启动成功,随后才会进行readinessProbe探测 |

## Probe

| 成员 | 类型   | 说明                                                         |
| ---- | ------ | ------------------------------------------------------------ |
| name | string | 容器名称，要求符合[RFC 1035](https://tools.ietf.org/html/rfc1035) |

## Resources

| 成员                | 类型                    | 说明                                                         |
| ------------------- | ----------------------- | ------------------------------------------------------------ |
| failureThreshold    | int                     | 最低连续失败次数,当达到该次数后,认为探针满足失败条件         |
| InitialDelaySeconds | int                     | 容器启动之前,第一次执行探针的延迟时间,单位: 秒               |
| periodSeconds       | int                     | 探针执行频率,单位: 秒                                        |
| successThreshold    | int                     | 连续成功次数,当探针失败后连续成功该次数后,认为探针满足成功条件 |
| timeoutSeconds      | int                     | 探针执行超时时间,单位: 秒                                    |
| httpGet             | [HTTPGet](#HTTPGet)     | 探针以http get方式运行, 几种探针有且只有一种                 |
| exec                | [Exec](#Exec)           | 探针以命令行方式运行,几种探针有且只有一种                    |
| tcpSocket           | [TCPSocket](#TCPScoket) | 探针以tcp连接方式运行,几种探针有且只有一种                   |

## HTTPGet

| 成员   | 类型   | 说明                                    |
| ------ | ------ | --------------------------------------- |
| path   | string | http get请求URL的path部分               |
| port   | int    | http get 请求的容器端口,值区间[1,65536] |
| schema | string | http get请求协议,仅为 "HTTP"和"HTTPS"   |

## Exec

| 成员   | 类型   | 说明                                   |
| ------ | ------ | -------------------------------------- |
| comand | string | 命令行探针执行命令行,返回码非0视为失败 |

## TCPSocket

| 成员 | 类型 | 说明                                   |
| ---- | ---- | -------------------------------------- |
| port | int  | tcp 连接探测的容器端口,值区间[1,65536] |

## ResourceObj

| 类型   | 说明                                                         | 成员   |
| ------ | ------------------------------------------------------------ | ------ |
| string | 内存配置,[数值计算与格式](https://kubernetes.io/zh-cn/docs/reference/kubernetes-api/common-definitions/quantity/) | memory |
| string | cpu配置,[数值计算与格式](https://kubernetes.io/zh-cn/docs/reference/kubernetes-api/common-definitions/quantity/) | cpu    |

## Image

| 成员     | 类型   | 说明         |
| -------- | ------ | ------------ |
| registry | string | 镜像仓库地址 |
| image    | string | 镜像名称URI  |
| tag      | string | 镜像tag      |

## Statefulset

see [Deployment](#Deployment)