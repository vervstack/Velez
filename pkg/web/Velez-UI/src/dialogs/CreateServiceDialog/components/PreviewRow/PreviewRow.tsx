import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/PreviewRow/PreviewRow.module.css"
import type {PreviewValue} from "@/app/api/velez"
import {isAdded} from "@/dialogs/CreateServiceDialog/processes/registerPreview.ts"

interface Props {
    value: PreviewValue
}

export default function PreviewRow({value}: Props) {
    const isValueAdded = isAdded(value)

    return (
        <li className={cn(cls.PreviewRowContainer, isValueAdded && cls.Added)}>
            <span className={cls.Key}>{value.key}</span>
            <span className={cls.Value}>{value.value}</span>
            {isValueAdded && <span className={cls.Tag}>added</span>}
        </li>
    )
}
