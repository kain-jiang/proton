/*
 * @File: 可操作的组件类型的文本
 * @Author: Ximena.zhou
 * @Date: 2023-05-17 12:09:47
 */
import React, { ReactElement } from 'react'
import styles from './style.module.less'

/**
 * @interface IProps
 * @param text 文本
 * @param insert 插入可操作的组件类型的文本
 */
interface IProps {
    text: string
    insert: ReactElement
}
const ElementText: React.FC<IProps> = ({ text, insert }) => {
    let fontArr = text.split('-')

    return (
        <div className={styles["text-container"]}>
            {fontArr[0]}
            {insert}
            {fontArr[1]}
        </div>
    )
}

export default ElementText
