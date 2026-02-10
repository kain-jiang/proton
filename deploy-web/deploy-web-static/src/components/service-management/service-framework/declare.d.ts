import { RJSFSchema, UiSchema } from "@rjsf/utils";

export interface Props extends React.ClassAttributes<any> {
  // 配置项 formData
  formData: RJSFSchema | undefined;
  // 配置项 schema
  schema: RJSFSchema | undefined;
  // 配置项 UIschema
  uiSchema?: UiSchema;
  onChangeFormData?: (formData: RJSFSchema) => void;
  isReadOnly: boolean;
  changeIsFormValidator?: (isFormValidator: boolean) => void;
  setCallback?: (callback: any) => void;
}

export interface State {
  formData: RJSFSchema;
}
