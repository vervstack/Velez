import cls from "@/dialogs/CreateServiceDialog/components/AdoptActions/AdoptActions.module.css"
import Button from "@/components/base/Button.tsx"

interface Props {
    isClusterMode: boolean
    isRegisterDisabled: boolean
    onCancel(): void
    onRegister(): void
}

export default function AdoptActions({isClusterMode, isRegisterDisabled, onCancel, onRegister}: Props) {
    return (
        <div className={cls.AdoptActionsContainer}>
            <Button variant="secondary" onClick={onCancel}>Cancel</Button>
            <Button variant={isClusterMode ? "primary" : "danger"} onClick={onRegister} disabled={isRegisterDisabled}>
                Onboard
            </Button>
        </div>
    )
}
