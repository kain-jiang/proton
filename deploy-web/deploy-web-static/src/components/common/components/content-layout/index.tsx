/*
 * @Author: your name
 * @Date: 2021-06-17 09:50:15
 * @LastEditTime: 2023-08-07 09:42:42
 * @LastEditors: Please set LastEditors
 * @Description: 主内容区包裹组件
 */

import React, { ReactNode } from 'react'
import className from 'classnames'
import { IToolbar } from '../toolbar'
import Toolbar from '../toolbar'
import { AB_PREFIX } from '../config'
import './index.less'

interface IContentLayout {
    children: ReactNode
    toolbar?: IToolbar
    header?: React.ReactNode
    wrapperClass?: string
    breadcrumb?: React.ReactNode
    contentClassName?: string
    footer?: React.ReactNode
    contentOverflow?: React.CSSProperties['overflow']
    moduleName?: string
}

const ContentLayout: React.FC<IContentLayout> = (props) => {
    const moduleName = props.moduleName || AB_PREFIX
    const renderHeader = () => {
        // 展示工具栏
        if (props.toolbar) {
            return <Toolbar {...props.toolbar}></Toolbar>
        }
        // 展示header头部
        if (props.header) {
            return <React.Fragment>{props.header}</React.Fragment>
        }
        return null
    }

    const footer = React.useMemo(
        () => (
            <React.Fragment>
                {props.footer && <div className={`${moduleName}-content-footer`}>{props.footer}</div>}
            </React.Fragment>
        ),
        [props.footer]
    )

    const contentStyle = React.useMemo(() => {
        if (props.contentOverflow) {
            return {
                overflow: props.contentOverflow,
            }
        }
        return {}
    }, [props.contentOverflow])

    return (
        <div className={className(`${moduleName}-content-layout`, props.wrapperClass)}>
            {props.breadcrumb && <div className={`${moduleName}-content-layout-breadcrumb`}>{props.breadcrumb}</div>}
            {renderHeader()}
            <div className={className(`${moduleName}-content-main`, props.contentClassName)} style={contentStyle}>
                {props.children}
            </div>
            {footer}
        </div>
    )
}

export default React.memo(ContentLayout)
