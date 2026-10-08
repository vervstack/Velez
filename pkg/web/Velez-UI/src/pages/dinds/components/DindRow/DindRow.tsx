import {ConfirmDialog} from "@vervstack/chures"

import cls from "@/pages/dinds/components/DindRow/DindRow.module.css"
import type {DindInfo} from "@/app/api/velez/dind_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DropDindMutation} from "@/processes/queries/dinds.ts"
import Button from "@/components/base/Button.tsx"

interface Props {
    dind: DindInfo
}

export default function DindRow({dind}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const dropDind = DropDindMutation()

    const name = dind.name ?? ""

    function handleConfirmDrop(): Promise<void> {
        return dropDind.mutateAsync(name).catch(toaster.catchGrpc)
    }

    function handleDrop() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Drop Docker daemon?"
                message={`${name} will be removed. Runners that use it stop working.`}
                confirmLabel="Drop"
                onConfirm={handleConfirmDrop}
                onClose={CloseDialog}
            />
        )
    }

    return (
        <div className={cls.DindRowContainer}>
            <div className={cls.Row}>
                <span className={cls.Name}>{name}</span>
                <span className={cls.Cell}>{dind.address || "-"}</span>
                <span className={cls.Cell}>{dind.isSysboxEnabled ? "Yes" : "No"}</span>
                <span className={cls.Cell}>
                    {dind.createdAt ? new Date(dind.createdAt as never).toLocaleDateString() : "-"}
                </span>
                <div className={cls.Actions}>
                    <Button sm variant="danger" onClick={handleDrop} disabled={dropDind.isPending}>
                        Drop
                    </Button>
                </div>
            </div>
        </div>
    )
}
