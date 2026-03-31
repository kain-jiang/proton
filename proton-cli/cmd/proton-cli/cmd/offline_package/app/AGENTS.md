# AGENTS.md

本目录用于 `proton-cli offline-package app` 需求实现与维护。

## 目标

- 实现应用离线包的导出与导入
- 输入为 `VersionSet`
- 输出包含 `manifest.yaml`、`charts/`、`images/`

## 约束

- 改动范围尽量限制在 `proton-cli/cmd/proton-cli/cmd/offline_package`
- 不依赖其他目录已有实现来完成核心逻辑
- 镜像处理必须使用 `oras-go`，不要依赖 `skopeo` 二进制
- `dependencies` 当前只解析保留，不参与导出
- chart 只按主清单 `releases` 处理

## 导出规则

- `--platform` 必须真实参与镜像选择，不能只写入元数据
- 支持：
  - `linux/amd64`
  - `linux/arm64`
  - 双架构组合
- 单架构导出时只保留目标平台 manifest
- 多平台导出时只保留指定平台集合
- `--ignore-missing-images` 仅允许把镜像拉取失败降级为 warning
- 被忽略的失败镜像必须写入 `manifest.yaml`

## 镜像命名

- `.image.registry` 视为完整仓库地址
- `.image.repository` 才是离线包内要保留的 repository
- 离线包内镜像名必须完整去掉 `.image.registry`

示例：

- 原始：`swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/agent-backend:0.5.1`
- 包内：`dip/agent-backend:0.5.1`

## `--override-registry`

- 语义是重写 values 中的 `.image.registry`
- 保留原始 `.image.repository` 与 `.image.tag`
- 不能把路径前缀重复拼接
- `manifest.yaml` 中应同时记录：
  - 原始 `source`
  - 实际 `pullSource`
  - `overrideRegistry`

## 导入规则

- 镜像推送目标为 `<registry>/<localRef>`
- chart 上传到 ChartMuseum
- `--force` 用于覆盖已存在 chart

## 文档要求

- `FEATURE.md` 是需求设计
- `IMPLEMENTATION.md` 是实现方案
- 行为变更后，这两份文档要同步更新

## 测试要求

- 至少覆盖：
  - 平台归一化
  - 镜像提取
  - override 逻辑
  - 导出 manifest 元数据
- 本目录代码变更后，优先执行：

```bash
CGO_ENABLED=0 go test ./cmd/proton-cli/cmd/offline_package/...
```
