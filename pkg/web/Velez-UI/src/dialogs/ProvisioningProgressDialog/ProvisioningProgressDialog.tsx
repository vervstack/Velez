import cls from "@/dialogs/ProvisioningProgressDialog/ProvisioningProgressDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {queryClient} from "@/app/queryClient.ts"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

interface Props {
    title: string
    entityId: string
    action: string
    queryKey: readonly unknown[]
}

export default function ProvisioningProgressDialog({title, entityId, action, queryKey}: Props) {
    const {CloseDialog} = useDialog()

    function handleStart(): Promise<{ entityId: string, action: string }> {
        return Promise.resolve({entityId, action})
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey})
        queryClient.invalidateQueries({queryKey: ["services"]})
    }

    return (
        <div className={cls.ProvisioningProgressDialogContainer}>
            <DialogShell title={title} onClose={CloseDialog} isFlush>
                <TaskProgressScreen
                    title={title}
                    metaLine={entityId}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            </DialogShell>
        </div>
    )
}
