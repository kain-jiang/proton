import React, { FC, useEffect, useState } from "react";
import { Space, Input, Button } from "@kweaver-ai/ui";
import { FormOutlined } from "@kweaver-ai/ui/icons";
import { Dot } from "../../../../common/components/text/dot";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";
import { serviceConfigStatus } from "../../type.d";
import { ServiceModeType } from "../type";
import { ServiceJSONSchemaItem } from "../../../../../api/service-management/service-deploy/declare";
import { safetyTime } from "../../../utils/timer";
import { safetyStr } from "../../../../common/utils/string";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";

interface IProps {
  // 服务信息
  serviceDetailInfo: ServiceJSONSchemaItem;
  // 服务类型
  serviceModeType: ServiceModeType;
}
export const ServiceTitle: FC<IProps> = ({
  serviceDetailInfo,
  serviceModeType,
}) => {
  // 是否编辑备注
  const [isEditing, setIsEditing] = useState<boolean>(false);
  // 备注内容
  const [comment, setComment] = useState<string>("");
  // 获取备注初始信息
  useEffect(() => {
    setComment(serviceDetailInfo.comment);
  }, [serviceDetailInfo.comment]);
  /**
   * @description 显示输入框
   */
  const showInput = () => {
    setIsEditing(true);
  };
  /**
   * @description 修改输入框内容
   * @param e 输入框change事件
   */
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setComment(e.target.value);
  };
  /**
   * @description 保存备注修改内容
   */
  const handleInputBlur = () => {
    setIsEditing(false);
    // 发送请求（暂不支持）
  };
  return (
    <Space direction="vertical">
      <Space size="large">
        <span className={styles["service-name"]}>
          {serviceModeType === ServiceMode.Service
            ? serviceDetailInfo.title
            : serviceDetailInfo.name}
        </span>
        <Dot color={serviceConfigStatus[serviceDetailInfo.status].color}>
          {serviceConfigStatus[serviceDetailInfo.status].text}
        </Dot>
      </Space>
      <Space size={16} className={styles["title-info"]}>
        {serviceModeType === ServiceMode.Service ? (
          <div>
            <span>{__("服务标识：")}</span>
            <span
              className={styles["text-item"]}
              title={safetyStr(serviceDetailInfo.name)}
            >
              {safetyStr(serviceDetailInfo.name)}
            </span>
          </div>
        ) : null}
        <div>
          <span>{__("版本：")}</span>
          <span
            className={styles["text-item"]}
            title={safetyStr(serviceDetailInfo.version)}
          >
            {safetyStr(serviceDetailInfo.version)}
          </span>
        </div>
        <div>
          <span>{__("更新时间：")}</span>
          <span
            className={styles["text-item"]}
            title={safetyTime(serviceDetailInfo.startTime)}
          >
            {safetyTime(serviceDetailInfo.startTime)}
          </span>
        </div>
        {serviceModeType === ServiceMode.Service && (
          <div>
            {isEditing ? (
              <span>
                {__("备注：")}
                <Input
                  autoFocus
                  value={comment}
                  className={styles["input-title"]}
                  onBlur={handleInputBlur}
                  onChange={handleInputChange}
                />
              </span>
            ) : (
              <>
                <span>{__("备注：")}</span>
                <span
                  className={styles["text-item"]}
                  style={{ lineHeight: "32px", maxWidth: "100px" }}
                  title={safetyStr(comment)}
                >
                  {safetyStr(comment)}
                </span>
              </>
            )}
            <Button
              type="link"
              icon={
                <FormOutlined
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                />
              }
              className={styles["icon"]}
              onClick={showInput}
              // disabled={isEditing}
              disabled={true}
            />
          </div>
        )}
      </Space>
    </Space>
  );
};
