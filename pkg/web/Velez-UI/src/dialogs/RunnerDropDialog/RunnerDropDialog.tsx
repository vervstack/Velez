import cls from "@/dialogs/RunnerDropDialog/RunnerDropDialog.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {DropRunnerMutation} from "@/processes/queries/runners.ts"
import Button from "@/components/base/Button.tsx"

interface Props {
    name: string
}

export default function RunnerDropDialog({name}: Props) {
    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const dropRunner = DropRunnerMutation()

    function handleConfirm() {
        dropRunner.mutate(name, {
            onSuccess: () => {
                toaster.bake({title: "Runner dropped", description: name, level: "Info"})
                CloseDialog()
            },
            onError: toaster.catchGrpc,
        })
    }

    return (
        <div className={cls.RunnerDropDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Drop runner</h2>
            </div>

            <div className={cls.Content}>
                <p className={cls.ConfirmText}>
                    Are you sure you want to drop <strong>{name}</strong>? This action cannot be undone.
                </p>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={dropRunner.isPending}>
                        Cancel
                    </Button>
                    <Button variant="danger" onClick={handleConfirm} disabled={dropRunner.isPending}>
                        {dropRunner.isPending ? "Dropping…" : "Drop"}
                    </Button>
                </div>
            </div>
        </div>
    )
}
