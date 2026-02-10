import * as React from "react";
import { isEqual } from "lodash";
import { Props, State } from "./index";
import { FormInstance } from "@aishutech/ui";

export class RDSConnectInfoBase extends React.Component<Props, State> {
  state = {
    rds: null,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(configData?.resource_connect_info?.rds);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.rds,
        configData?.resource_connect_info?.rds,
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.rds);
    }
  }

  /**
   * 初始化配置
   * @param rds 连接信息类型
   */
  private initConfig(rds) {
    this.setState(
      {
        rds,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.rds,
        });
      },
    );
  }

  /**
   * changeRDSConnectInfo
   */
  public changeRDSConnectInfo(key, val, rds) {
    let cur;
    if (key === "auto_create_database" && !val) {
      cur = {
        ...rds,
        auto_create_database: false,
        admin_user: "",
        admin_passwd: "",
      };
    } else {
      cur = {
        ...rds,
        [key]: val,
      };
    }
    this.setState(
      {
        rds: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      },
    );
  }

  // 清楚校验状态
  protected handleClearValidation() {
    this.form?.current
      ?.validateFields()
      .then(() => {
        // 校验通过，无需操作
      })
      .catch(() => {
        // 校验失败，将错误信息清空
        const fields = this.form?.current?.getFieldsValue();
        const fieldNames = Object.keys(fields);
        this.form?.current?.setFields(
          fieldNames.map((name) => ({
            name,
            errors: [],
          })),
        );
      });
  }

  // 校验联动表单行状态
  protected checkLinkItem(key, rds) {
    if (rds[key]) {
      this.form.current.validateFields([key]);
    }
  }
}
