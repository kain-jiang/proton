import React, { createRef } from "react";
import { RJSFSchema } from "@rjsf/utils";
import { Props, State } from "./declare";
import WebComponent from "../../webcomponent";
import { FormInstance } from "@kweaver-ai/ui";

export class ServiceFrameworkBase extends WebComponent<Props, State> {
  // state: Readonly<State> = {
  //     formData: FORM_DATA,
  // };
  formRef = createRef<any>();

  componentDidMount(): void {
    this.props.setCallback &&
      this.props.setCallback(() => this.formRef?.current?.submit());
  }
}
