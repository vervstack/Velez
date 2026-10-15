import {ConfirmDialog} from "@vervstack/chures"
import cn from "classnames"

import cls from "@/pages/transfer/components/TransferRow/TransferRow.module.css"
import type {TransferInfo} from "@/app/api/velez/transfer_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {formatBytes} from "@/processes/mappings/s3.ts"
import {
    isTransferActive,
    isTransferFailed,
    statusLabel,
    statusTone,
    transferPercent,
} from "@/processes/mappings/transfer.ts"
import {CancelTransferMutation} from "@/processes/queries/transfer.ts"
import Button from "@/components/base/Button.tsx"

interface Props {
    transfer: TransferInfo
}

export default function TransferRow({transfer}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const cancelTransfer = CancelTransferMutation()

    const containerName = transfer.containerName ?? ""
    const isActive = isTransferActive(transfer.status)
    const isCancellable = Boolean(transfer.isSource) && isActive
    const percent = transferPercent(transfer)
    const isBarShown = isActive || percent !== null
    const progressLabel = percent === null ? formatBytes(transfer.doneBytes) : `${percent}%`

    function handleConfirmCancel(): Promise<void> {
        return cancelTransfer.mutateAsync(transfer.id ?? "")
            .then(function handleCancelled() {
                toaster.bake({title: "Transfer cancelled", description: containerName, level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function handleCancel() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Cancel transfer?"
                message={`The transfer of ${containerName} stops. The source container stays stopped.`}
                confirmLabel="Cancel transfer"
                onConfirm={handleConfirmCancel}
                onClose={CloseDialog}
            />
        )
    }

    return (
        <div className={cls.TransferRowContainer}>
            <div className={cls.Row}>
                <span className={cls.Cell}>{transfer.isSource ? "Outgoing" : "Incoming"}</span>
                <span className={cls.Name}>{containerName}</span>
                <span className={cls.Cell}>{transfer.peerAddress || "-"}</span>
                <span className={cn(cls.Chip, cls[statusTone(transfer.status)])}>
                    {statusLabel(transfer.status)}
                </span>
                <div className={cls.Progress}>
                    {isBarShown && <progress className={cls.Bar} value={percent ?? undefined} max={100}/>}
                    {isBarShown && <span className={cls.ProgressLabel}>{progressLabel}</span>}
                </div>
                <div className={cls.Actions}>
                    {isCancellable && (
                        <Button sm variant="danger" onClick={handleCancel} disabled={cancelTransfer.isPending}>
                            Cancel
                        </Button>
                    )}
                </div>
            </div>
            {isTransferFailed(transfer.status) && transfer.error && (
                <div className={cls.ErrorLine}>{transfer.error}</div>
            )}
        </div>
    )
}
