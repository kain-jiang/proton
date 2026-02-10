/*
 * @File: 封装常见的文字使用场景 1. 提示文字 2. 输入框下的文字 3. 带颜色的文字 4. 带圆点的文字 5. 省略号
 * @Author: Zeng.xu
 * @Date: 2022-04-26 13:58:39
 */

import React from "react";
import { Tooltip } from "@kweaver-ai/ui";
import { TextProps } from "./declare";
import classnames from "classnames";
import "./index.less";

const Text: React.FC<TextProps> = ({
  fontSize = "14px",
  fontWeight = "normal",
  textClassName,
  textColor,
  ellipsis,
  dot,
  dotClassName,
  dotColor,
  children,
  tooltipProps,
}) => {
  const childElement = React.useMemo(() => {
    let style: React.CSSProperties = {
      fontSize,
      fontWeight: fontWeight as string,
    };

    if (textColor) {
      style.color = textColor;
    }

    return (
      <span style={{ ...style }} className={classnames(textClassName, "text")}>
        {children}
      </span>
    );
  }, [fontSize, textClassName, children]);

  const dotStyle = React.useMemo(() => {
    if (dotColor) {
      return {
        color: dotColor,
      };
    }
    return {};
  }, [dotColor]);

  const enable = ellipsis !== undefined && ellipsis !== false;

  const width =
    typeof ellipsis === "string" || typeof ellipsis === "number"
      ? ellipsis
      : undefined;

  const style = React.useMemo(() => {
    if (width !== undefined) {
      const widthString = typeof width === "string" ? width : width + "px";
      return { maxWidth: widthString };
    }
    return {};
  }, [width]);

  const EllipsisText = React.useMemo(() => {
    if (enable) {
      return (
        <div
          className={classnames("es-text-ellipsis", {
            "es-text-max-width": width === undefined,
          })}
          style={style}
        >
          {childElement}
        </div>
      );
    }
    return <>{childElement}</>;
  }, [enable, childElement, style, width]);

  return (
    <div className="es-text">
      {dot && (
        <span
          style={dotStyle}
          className={classnames("blue-point", dotClassName)}
        ></span>
      )}
      {tooltipProps ? (
        <Tooltip {...tooltipProps}>{EllipsisText}</Tooltip>
      ) : (
        <>{EllipsisText}</>
      )}
    </div>
  );
};

export default Text;
