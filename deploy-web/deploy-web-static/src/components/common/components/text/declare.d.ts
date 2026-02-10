import React from 'react'
import { Color } from '../color'
import { TooltipProps } from 'antd/lib/tooltip'

/**
 * @interface IconProps
 * @param useSvg 是否svg图标
 * @param name 图标后缀名
 * @param size 图标尺寸大小
 * @param fontSize 字体图标大小
 * @param iClass 字体图标的类名
 * @param tooltip 提示信息
 * @param tooltipProps
 * @param flex 是否使用flex布局
 * @param opacity 透明度，小数值（主要用于黑色调节透明度达到灰色效果）
 * */
export interface IconProps {
    useSvg?: boolean
    className?: string
    name: string
    size?: { width?: string; height?: string }
    fontSize?: string
    iClass?: string
    color?: Color
    tooltip?: React.ReactNode
    tooltipProps?: Omit<TooltipProps, 'title'>
    flex?: boolean
    opacity?: number
    children?: React.ReactNode
}

export interface DerivedIconProps extends Omit<IconProps, 'name'> {
    coverClass?: string
}

interface TextProps {
    fontSize?: string
    fontWeight?: 'normal' | 'bold' | 'bolder' | 'lighter' | number
    textClassName?: string
    textColor?: string
    ellipsis?: boolean | number | string
    dot?: boolean
    dotClassName?: string
    dotColor?: string
    tooltipProps?: TooltipProps
    children: React.ReactNode
}

export interface TipsProps {
    iconProps?: DerivedIconProps
    textProps?: TextProps
    type?: 'info' | 'warning'
    children: React.ReactNode
}

export interface DotProps {
    color?: Color
    children: React.ReactNode
    textProps?: TextProps
}

export interface TextInterface extends React.FC<TextProps> {
    Tips: React.FC<TipsProps>
    Dot: React.FC<DotProps>
}
