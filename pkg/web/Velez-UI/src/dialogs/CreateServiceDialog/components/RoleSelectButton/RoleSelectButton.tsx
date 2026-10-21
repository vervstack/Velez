import cls from "@/dialogs/CreateServiceDialog/components/RoleSelectButton/RoleSelectButton.module.css"
import Button from "@/components/base/Button.tsx"
import InfoMark from "@/components/base/InfoMark.tsx"

interface Props {
    label: "Network root" | "Service"
    isSelected: boolean
    isDisabled: boolean
    tooltip: string
    onSelect(): void
}

export default function RoleSelectButton({label, isSelected, isDisabled, tooltip, onSelect}: Props) {
    return (
        <div className={cls.RoleSelectButtonContainer}>
            <Button variant={isSelected ? "primary" : "ghost"} sm disabled={isDisabled} onClick={onSelect}>
                {label}
            </Button>
            <div className={cls.InfoMarkWrapper}>
                <InfoMark tooltip={tooltip}/>
            </div>
        </div>
    )
}
