import React, { FC } from "react";
import styles from "./styles.module.less";

export const FormLabel: FC<{ text: string }> = ({ text }) => {
  return (
    <span title={text} className={styles["form-label"]}>
      {text}
    </span>
  );
};
