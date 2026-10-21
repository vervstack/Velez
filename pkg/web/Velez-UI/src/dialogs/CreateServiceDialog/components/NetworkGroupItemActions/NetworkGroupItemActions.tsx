import cls from "@/dialogs/CreateServiceDialog/components/NetworkGroupItemActions/NetworkGroupItemActions.module.css"
import Button from "@/components/base/Button.tsx"

export interface ItemActionsProps {
    canMakeService: boolean
    canMakeRoot: boolean
    onMakeService(): void
    onMakeRoot(): void
}

export default function NetworkGroupItemActions(
    {canMakeService, canMakeRoot, onMakeService, onMakeRoot}: ItemActionsProps,
) {
    if (!canMakeService && !canMakeRoot) return null

    return (
        <div className={cls.NetworkGroupItemActionsContainer}>
            {canMakeService && <Button variant="secondary" sm onClick={onMakeService}>Make service</Button>}
            {canMakeRoot && <Button variant="secondary" sm onClick={onMakeRoot}>Make network root</Button>}
        </div>
    )
}
