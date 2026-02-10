import { ServiceJSONSchemaItem } from "../../../api/service-management/service-deploy/declare";
import { RJSFSchema } from "@rjsf/utils";

// 合并formData
export const mergeMap = (
  a: RJSFSchema,
  b: RJSFSchema,
  ignoreNil = false
): RJSFSchema => {
  let out = {};
  // 遍历映射a
  for (let [k, v] of Object.entries(a)) {
    out[k] = v;
  }
  // 遍历映射b
  for (let [k, v] of Object.entries(b)) {
    // 检查v是否是另一个映射
    if (typeof v === "object" && v !== null && !Array.isArray(v)) {
      // 检查out[k]是否已经存在，并且也是一个映射
      if (
        typeof out[k] === "object" &&
        out[k] !== null &&
        !Array.isArray(out[k])
      ) {
        // 递归合并映射
        out[k] = mergeMap(out[k], v, ignoreNil);
        continue;
      }
    }
    // 如果v不是null，或者ignoreNil为false，则更新out[k]
    if (v !== null || !ignoreNil) {
      out[k] = v;
    }
  }
  return out;
};

// 合并套件清单配置和系统配置
export const schemaCompose = (
  currentServiceConfig: ServiceJSONSchemaItem,
  suiteSchemaConfig: ServiceJSONSchemaItem,
  systemSchemaConfig: ServiceJSONSchemaItem | null,
  isSynchronousUpdate: boolean
): ServiceJSONSchemaItem => {
  const newFormData = mergeMap(
    currentServiceConfig.formData!,
    systemSchemaConfig
      ? systemSchemaConfig.formData!
      : suiteSchemaConfig.formData!
  );
  if (isSynchronousUpdate) {
    return {
      ...suiteSchemaConfig,
      formData: newFormData,
      schema: systemSchemaConfig
        ? systemSchemaConfig.schema
        : suiteSchemaConfig.schema,
      uiSchema: systemSchemaConfig
        ? systemSchemaConfig.uiSchema
        : suiteSchemaConfig.uiSchema,
      version: systemSchemaConfig
        ? systemSchemaConfig.version
        : suiteSchemaConfig.version,
    };
  } else {
    return {
      ...suiteSchemaConfig,
      formData: newFormData,
      schema: suiteSchemaConfig.schema,
      uiSchema: suiteSchemaConfig.uiSchema,
    };
  }
};
