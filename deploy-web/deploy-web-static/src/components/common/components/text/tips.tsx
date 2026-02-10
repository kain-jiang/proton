import React from 'react'
import Text from './text'
import { Color } from '../index'
import { TipsProps } from './declare'

export const Tips: React.FC<TipsProps> = React.memo(({ textProps = {}, iconProps = {}, type = 'info', children }) => {
    switch (type) {
        case 'info':
            return (
                <Text {...textProps}>{children}</Text>
            )
        case 'warning':
            return (
                <Text textColor={Color.GRAY_LIGHTER} {...textProps}>
                    {children}
                </Text>
            )
    }
})
