import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/PreviewRow/PreviewRow.module.css"
import type {PreviewValue} from "@/app/api/velez"
import {isAdded} from "@/dialogs/CreateServiceDialog/processes/registerPreview.ts"

const MASK = "••••••"

interface Props {
    value: PreviewValue
}

export default function PreviewRow({value}: Props) {
    const isSecret = value.isSecret === true
    const isValueAdded = isAdded(value)

    return (
        <li className={cn(cls.PreviewRowContainer, isValueAdded && cls.Added)}>
            <span className={cls.Key}>{value.key}</span>
            {isSecret
                ? <span className={cn(cls.Value, cls.Masked)} aria-label="hidden value">{MASK}</span>
                : <span className={cls.Value}>{value.value}</span>}
            {isSecret && <span className={cn(cls.Tag, cls.Hidden)}>hidden</span>}
            {!isSecret && isValueAdded && <span className={cls.Tag}>added</span>}
        </li>
    )
}
