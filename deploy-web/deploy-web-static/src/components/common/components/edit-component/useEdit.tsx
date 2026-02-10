import React, { useState, useMemo } from 'react';

interface useEditReturn {
    trigger: React.ReactElement<any>
    setEdit: React.Dispatch<React.SetStateAction<boolean>>
}

function useEdit(props: { triggerElement: React.ReactElement<any>, edit?: boolean }, deps: any[] = []): useEditReturn {
    const { triggerElement } = props;

    const [ edit, setEdit ] = useState(props.edit || false);

    const trigger = useMemo(() => {
        return React.cloneElement(triggerElement, {
            onClick: () => {
                setEdit(true);
            }
        });
    }, deps);

    return {
        trigger,
        setEdit,
    };
}

export default useEdit;