import cn from "classnames"

import cls from "@/components/complex/SlidingLabel/SlidingLabel.module.css"

interface Props {
    label: string
    swapLabel: string
    isSwapped: boolean
}

export default function SlidingLabel({label, swapLabel, isSwapped}: Props) {
    return (
        <span className={cn(cls.SlidingLabelContainer, {[cls.IsSwapped]: isSwapped})}>
            <span className={cn(cls.Face, cls.Primary)} aria-hidden={isSwapped}>{label}</span>
            <span className={cn(cls.Face, cls.Secondary)} aria-hidden={!isSwapped}>{swapLabel}</span>
        </span>
    )
}
