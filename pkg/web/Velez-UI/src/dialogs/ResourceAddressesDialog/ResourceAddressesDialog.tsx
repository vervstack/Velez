import cls from "@/dialogs/ResourceAddressesDialog/ResourceAddressesDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import type {ResourceAddress} from "@/model/service_page/ServicePageModel"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import AddressRow from "@/dialogs/ResourceAddressesDialog/components/AddressRow/AddressRow.tsx"

interface Props {
    resourceName: string
    addresses: ResourceAddress[]
}

export default function ResourceAddressesDialog({resourceName, addresses}: Props) {
    const {CloseDialog} = useDialog()

    return (
        <div className={cls.ResourceAddressesDialogContainer}>
            <DialogShell title={`${resourceName} — addresses`} onClose={CloseDialog}>
                <div className={cls.RowsWrapper}>
                    {addresses.map((address) => (
                        <AddressRow key={`${address.scope}-${address.host}-${address.port}`} address={address}/>
                    ))}
                </div>
            </DialogShell>
        </div>
    )
}
