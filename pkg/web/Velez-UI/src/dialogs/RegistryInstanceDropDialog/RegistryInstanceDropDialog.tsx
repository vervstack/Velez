import cls from "@/dialogs/RegistryInstanceDropDialog/RegistryInstanceDropDialog.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {DropRegistryInstanceMutation} from "@/processes/queries/registry_instances.ts"
import Button from "@/components/base/Button.tsx"

interface Props {
    name: string
}

export default function RegistryInstanceDropDialog({name}: Props) {
    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const dropRegistryInstance = DropRegistryInstanceMutation()

    function handleConfirm() {
        dropRegistryInstance.mutate(name, {
            onSuccess: () => {
                toaster.bake({title: "Registry dropped", description: name, level: "Info"})
                CloseDialog()
            },
            onError: toaster.catchGrpc,
        })
    }

    return (
        <div className={cls.RegistryInstanceDropDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Drop registry</h2>
            </div>

            <div className={cls.Content}>
                <p className={cls.ConfirmText}>
                    Are you sure you want to drop <strong>{name}</strong>? This action cannot be undone.
                </p>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={dropRegistryInstance.isPending}>
                        Cancel
                    </Button>
                    <Button variant="danger" onClick={handleConfirm} disabled={dropRegistryInstance.isPending}>
                        {dropRegistryInstance.isPending ? "Dropping…" : "Drop"}
                    </Button>
                </div>
            </div>
        </div>
    )
}
