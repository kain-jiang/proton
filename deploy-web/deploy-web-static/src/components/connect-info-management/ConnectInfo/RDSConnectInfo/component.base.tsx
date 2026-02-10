import * as React from "react";
import { isEqual } from "lodash";
import { Props, State } from "./index";
import { FormInstance } from "@kweaver-ai/ui";
import { SOURCE_TYPE } from "../../../component-management/helper";
import { RDSConnectInfo } from "../../index.d";

export class RDSConnectInfoBase extends React.Component<Props, State> {
  state = {
    rds: null as any,
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
        configData?.resource_connect_info?.rds
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.rds);
    }
  }

  /**
   * 初始化配置
   * @param rds 连接信息类型
   */
  private initConfig(rds: RDSConnectInfo) {
    this.setState(
      {
        rds,
      },
      () => {
        this.form.current?.setFieldsValue({
          ...this.state.rds,
        });
      }
    );
  }

  /**
   * changeRDSConnectInfo
   */
  public changeRDSConnectInfo(
    key: string,
    val: string | boolean,
    rds: RDSConnectInfo
  ) {
    let cur: RDSConnectInfo;
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
      }
    );
  }

  // 是否禁用不可修改信息
  protected getIsDisabled(rds: RDSConnectInfo) {
    const { originConnectInfoType, originSourceType } = this.props;
    if (rds?.source_type === SOURCE_TYPE.INTERNAL) {
      return originSourceType === SOURCE_TYPE.INTERNAL;
    } else {
      if (originSourceType !== SOURCE_TYPE.EXTERNAL) {
        return false;
      } else {
        return originConnectInfoType === rds?.rds_type;
      }
    }
  }

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
          }))
        );
      });
  }

  // 校验联动表单行状态
  protected checkLinkItem(key: string, rds: RDSConnectInfo) {
    if (rds[key]) {
      this.form.current?.validateFields([key]);
    }
  }
}
