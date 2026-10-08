import cls from "@/widgets/ProvisioningRow/ProvisioningRow.module.css"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {currentStepName, isProvisioningFailed, stepProgress} from "@/processes/mappings/provisioning.ts"
import {DismissTaskMutation} from "@/processes/queries/provisioning.ts"
import Button from "@/components/base/Button.tsx"
import ProvisioningProgressDialog from "@/dialogs/ProvisioningProgressDialog/ProvisioningProgressDialog.tsx"

interface Props {
    task: ProvisioningTask
    title: string
    queryKey: readonly unknown[]
}

export default function ProvisioningRow({task, title, queryKey}: Props) {
    const {OpenDialog} = useDialog()
    const toaster = useToaster()
    const dismissTask = DismissTaskMutation(queryKey)

    const isFailed = isProvisioningFailed(task)
    const entityId = task.entityId ?? ""
    const {done, total} = stepProgress(task)

    function handleDetails() {
        OpenDialog(
            <ProvisioningProgressDialog
                title={title}
                entityId={entityId}
                action={task.action ?? ""}
                queryKey={queryKey}
            />
        )
    }

    function handleDismiss() {
        dismissTask.mutateAsync(task).catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.ProvisioningRowContainer}>
            {isFailed ? <span className={cls.FailureMark}>!</span> : <span className={cls.Spinner}/>}
            <span className={cls.Name}>{entityId}</span>
            {isFailed
                ? <span className={cls.Error}>{task.error}</span>
                : <span className={cls.Step}>{currentStepName(task)}</span>}
            {!isFailed && total > 0 && <span className={cls.Progress}>{done}/{total}</span>}
            <div className={cls.Actions}>
                {isFailed && (
                    <Button sm variant="secondary" onClick={handleDismiss} disabled={dismissTask.isPending}>
                        Dismiss
                    </Button>
                )}
                <Button sm variant="secondary" onClick={handleDetails}>Details</Button>
            </div>
        </div>
    )
}
