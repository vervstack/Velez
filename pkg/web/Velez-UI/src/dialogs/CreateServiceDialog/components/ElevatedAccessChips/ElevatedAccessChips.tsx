import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/ElevatedAccessChips/ElevatedAccessChips.module.css"
import type {ElevatedAccess} from "@/app/api/velez"
import {elevatedAccessItemsOf} from "@/dialogs/CreateServiceDialog/processes/elevatedAccess.ts"

const CARRIED_OVER_TEXT = "Carried over to the new container."
const NO_ACCESS_TEXT = "No elevated access"

interface Props {
    access?: ElevatedAccess
    isRecreating: boolean
}

export default function ElevatedAccessChips({access, isRecreating}: Props) {
    if (!isRecreating) return null

    const items = access ? elevatedAccessItemsOf(access) : []

    function renderChip(item: string) {
        return <span key={item} className={cn(cls.Chip, cls.Warning)}>{item}</span>
    }

    if (items.length === 0) {
        return (
            <div className={cls.ElevatedAccessChipsContainer}>
                <div className={cls.ChipsWrapper}>
                    <span className={cn(cls.Chip, cls.Neutral)}>{NO_ACCESS_TEXT}</span>
                </div>
            </div>
        )
    }

    return (
        <div className={cls.ElevatedAccessChipsContainer}>
            <div className={cls.ChipsWrapper}>{items.map(renderChip)}</div>
            <span className={cls.Caption}>{CARRIED_OVER_TEXT}</span>
        </div>
    )
}
