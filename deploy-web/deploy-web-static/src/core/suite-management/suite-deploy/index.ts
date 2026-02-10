/**
 * @description jsonschema配置项操作类型
 */
export const enum SchemaOperateType {
  // 安装
  Install = "Install",
  // 更新
  Update = "Update",
  // 回退版本
  Revert = "Revert",
  // 更改配置
  ChangeConfig = "ChangeConfig",
}

// 任务类型
export const enum JobType {
  // 所有任务
  All,
  // 批量任务
  Batch,
  // 套件任务
  Suite,
}
