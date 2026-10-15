import cls from "@/pages/transfer/components/IncomingTransferCard/IncomingTransferCard.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateTransferTokenMutation} from "@/processes/queries/transfer.ts"
import Button from "@/components/base/Button.tsx"
import TransferTokenDialog from "@/dialogs/TransferTokenDialog/TransferTokenDialog.tsx"

export default function IncomingTransferCard() {
    const {OpenDialog} = useDialog()
    const toaster = useToaster()
    const createToken = CreateTransferTokenMutation()

    function handleAllow() {
        createToken.mutateAsync()
            .then(function handleCreated(response) {
                OpenDialog(<TransferTokenDialog token={response.token ?? ""} expiresAt={response.expiresAt}/>)
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <section className={cls.IncomingTransferCardContainer}>
            <div className={cls.Text}>
                <h2 className={cls.Title}>Incoming transfer</h2>
                <div className={cls.Hint}>
                    Issue a one-time token and give it, with this node's address, to the sender.
                </div>
            </div>
            <Button variant="primary" onClick={handleAllow} disabled={createToken.isPending}>
                Allow incoming transfer
            </Button>
        </section>
    )
}
