import React, { FC, useEffect } from "react";
import { Spin } from "@kweaver-ai/ui";
import styles from "./styles.module.less";

interface IProps {
  text: string;
}

export const CustomSpin: FC<IProps> = ({ text }) => {
  return (
    <div className={styles["spin-container"]}>
      <Spin size="large" />
      <div>{text}</div>
    </div>
  );
};
