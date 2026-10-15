import cls from "@/dialogs/TransferTokenDialog/TransferTokenDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {formatTransferTime} from "@/processes/mappings/transfer.ts"
import Button from "@/components/base/Button.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"

interface Props {
    token: string
    expiresAt?: {seconds?: string | number}
}

export default function TransferTokenDialog({token, expiresAt}: Props) {
    const {CloseDialog} = useDialog()
    const toaster = useToaster()

    function handleCopy() {
        navigator.clipboard.writeText(token)
            .then(function onCopied() {
                toaster.bake({title: "Token copied", description: "Transfer token", level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.TransferTokenDialogContainer}>
            <DialogShell title="Incoming transfer token" onClose={CloseDialog}>
                <div className={cls.TokenWrapper}>
                    <code className={cls.Token}>{token}</code>
                    <Button sm onClick={handleCopy}>Copy</Button>
                </div>
                <div className={cls.Note}>
                    Shown once and works for a single transfer. Expires at {formatTransferTime(expiresAt)}.
                    Give it to the sender together with this node's address.
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="primary" onClick={CloseDialog}>Done</Button>
                </div>
            </DialogShell>
        </div>
    )
}
