import { ComponentType } from "../../../api/service-management/service-deploy/declare";
import { ServiceJSONSchemaItem } from "../../../api/service-management/service-deploy/declare";

export interface ServiceTableType {
  cid: number;
  name: string;
  version: string;
  status: number;
}

/**
 * @description 将组件对象变为table所要的数组格式
 * @param components 格式化前的数据
 * @return 表格所需要的数据格式
 */
export const formatTable = (
  components: { [key: string]: ComponentType } | ComponentType[]
): ServiceTableType[] => {
  return Object.values(components).map((component) => {
    return {
      cid: component.cid,
      name: component.component.name,
      version: component.component.version,
      status: component.trait.status,
      componentDefineType: component.component.componentDefineType,
    };
  });
};

/**
 * @description 将微服务信息转换为title所要的格式
 * @param formData 格式化前的数据
 * @return 服务标题所需要的数据格式
 */
export const formatServiceTitle = (
  formData: ComponentType
): ServiceJSONSchemaItem => {
  if (!formData.component) {
    return { status: 0 } as ServiceJSONSchemaItem;
  }
  return {
    ...formData,
    ...formData.component,
    ...formData.trait,
  };
};
