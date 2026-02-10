// 系统空间配置
export interface SystemConfig {
  // 系统空间id
  sid?: number;
  // 命名空间
  namespace: string;
  // 系统空间名称
  systemName: string;
  // 备注
  description?: string;
  // 系统配置项
  config?: string;
}

/**
 * @interface IGetSystemParams
 * @param offset 索引
 * @param limit 页数
 */
export interface IGetSystemParams {
  offset: number;
  limit: number;
  // 单实例模式下且mode设为true，会返回单个系统空间
  mode?: boolean;
}
