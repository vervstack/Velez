import cls from "@/dialogs/PgInstanceDropDialog/PgInstanceDropDialog.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {DropPgInstanceMutation} from "@/processes/queries/pg_instances.ts"
import Button from "@/components/base/Button.tsx"

interface Props {
    name: string
}

export default function PgInstanceDropDialog({name}: Props) {
    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const dropPgInstance = DropPgInstanceMutation()

    function handleConfirm() {
        dropPgInstance.mutate(name, {
            onSuccess: () => {
                toaster.bake({title: "Database dropped", description: name, level: "Info"})
                CloseDialog()
            },
            onError: toaster.catchGrpc,
        })
    }

    return (
        <div className={cls.PgInstanceDropDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Drop database</h2>
            </div>

            <div className={cls.Content}>
                <p className={cls.ConfirmText}>
                    Are you sure you want to drop <strong>{name}</strong>? This action cannot be undone.
                </p>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={dropPgInstance.isPending}>
                        Cancel
                    </Button>
                    <Button variant="danger" onClick={handleConfirm} disabled={dropPgInstance.isPending}>
                        {dropPgInstance.isPending ? "Dropping…" : "Drop"}
                    </Button>
                </div>
            </div>
        </div>
    )
}
