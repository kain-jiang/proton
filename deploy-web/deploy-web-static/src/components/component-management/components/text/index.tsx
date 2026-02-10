import React, { FC, useState, useEffect } from "react";
import styles from "./styles.module.less";

interface IProps {
    children: React.ReactNode;
}

export const Text: FC<IProps> = ({ children }) => {
    return (
        <span className={styles["text"]} title={children as string}>
            {children}
        </span>
    );
};
