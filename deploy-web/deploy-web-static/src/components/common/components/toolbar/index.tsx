import React from "react";
import className from "classnames";
import { Row, Col, Space } from "@kweaver-ai/ui";
import { AB_PREFIX } from "../config";
import "./index.less";

export interface IToolbar {
  cols?: Array<React.ComponentProps<typeof Col>>;
  custom?: React.ReactNode;
  left?: React.ReactNode;
  right?: React.ReactNode;
  search?: React.ReactNode;
  leftSize?: number;
  rightSize?: number;
  wrapperClass?: string;
  noMarginTop?: boolean;
  moduleName?: string;
}

const Toolbar: React.FC<IToolbar> = (props) => {
  const {
    cols,
    custom,
    left,
    right,
    search,
    wrapperClass,
    leftSize,
    rightSize,
    noMarginTop,
  } = props;
  const moduleName = props.moduleName || AB_PREFIX;
  const render = () => {
    const leftColProps = Array.isArray(cols) ? cols[0] : {};
    const rightColProps = Array.isArray(cols) ? cols[1] : {};
    // 右边工具栏，靠右展示
    const rightCol = {
      ...{ span: rightSize || 12 },
      ...rightColProps,
    };
    // 左边工具栏
    const leftCol = {
      ...{ span: leftSize || 12, offset: left ? 0 : leftSize || 12 },
      ...leftColProps,
    };
    return (
      <React.Fragment>
        <Row>
          {left && (
            <Col
              {...leftCol}
              className={`${moduleName}-custom-toolbar-left toolbar-left--pop-container`}
            >
              <Space size={10}>{left}</Space>
            </Col>
          )}
          {right && (
            <Col {...rightCol} className={`${moduleName}-custom-toolbar-right`}>
              <Space size={10}>{right}</Space>
            </Col>
          )}
        </Row>
        {search && <Row>{search}</Row>}
      </React.Fragment>
    );
  };

  return (
    <div
      className={className(`${moduleName}-custom-toolbar`, wrapperClass, {
        "without-margin": noMarginTop,
      })}
    >
      {custom || render()}
    </div>
  );
};
export default Toolbar;
