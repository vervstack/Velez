import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/RoleSlot/RoleSlot.module.css"
import Badge from "@/components/base/Badge.tsx"
import Button from "@/components/base/Button.tsx"
import InfoMark from "@/components/base/InfoMark.tsx"

interface Props {
    chipLabel: "Root" | "Service"
    chipColor: string
    chipDim: string
    buttonLabel: "Make root" | "Make service"
    isHeld: boolean
    isDisabled: boolean
    tooltip: string
    onSelect(): void
}

export default function RoleSlot(
    {chipLabel, chipColor, chipDim, buttonLabel, isHeld, isDisabled, tooltip, onSelect}: Props,
) {
    return (
        <div className={cn(cls.RoleSlotContainer, isHeld && cls.Held)}>
            <div className={cls.SwapWrapper}>
                <span className={cls.Chip} aria-hidden={!isHeld}>
                    <Badge label={chipLabel} color={chipColor} dim={chipDim}/>
                </span>
                <span className={cls.Action} aria-hidden={isHeld}>
                    <Button variant="ghost" sm disabled={isHeld || isDisabled} onClick={onSelect}>
                        {buttonLabel}
                    </Button>
                </span>
            </div>
            <div className={cls.InfoMarkWrapper}>
                <InfoMark tooltip={tooltip}/>
            </div>
        </div>
    )
}
