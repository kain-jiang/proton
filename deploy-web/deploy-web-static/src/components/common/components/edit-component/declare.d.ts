import React from 'react'

export interface DescriptionEditProps<T = any> {
    edit?: boolean
    onSave?: (text: T) => void | Promise<void>
    onCancel?: () => void
    value: T
    render?: (value: T) => React.ReactNode
    Component: React.ReactElement<any>
}

export interface useEditParams {
    editComponent: React.ReactElement
}