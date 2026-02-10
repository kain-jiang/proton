export interface Props {
  // 标题
  title: string;

  // 标题提示
  tip?: string;

  // 删除组件
  deleteCallback?: (services) => void;
}
