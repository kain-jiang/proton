import React from 'react'
import Text from '../text'
import { DotProps } from '../declare'
import { Color } from '../../index'
import styles from './styles.module.less'

const getColor = (color: Color): string => {
    const map = {
        [Color.green]: styles['green-point'],
        [Color.blue]: styles['blue-point'],
        [Color.grey]: styles['gray-point'],
        [Color.red]: styles['red-point'],
        [Color.orange]: styles['orange-point'],
        [Color.yellow]: styles['yellow-point'],
        [Color.GRAY_BLACK]: styles['grayblack-point'],
        [Color.SERVICE_GREEN]: styles['service-green-point'],
        [Color.SERVICE_RED]: styles['service-red-point'],
        [Color.Success]: styles['success-point'],
        [Color.Failure]: styles['failure-point'],
    }
    return map[color]
}

export const Dot: React.FC<DotProps> = (props) => {
    return (
        <Text dot dotClassName={props.color && getColor(props.color)} fontSize="14px" {...props.textProps}>
            {props.children}
        </Text>
    )
}
