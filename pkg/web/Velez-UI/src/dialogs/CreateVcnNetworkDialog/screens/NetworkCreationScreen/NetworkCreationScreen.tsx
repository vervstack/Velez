import {useEffect, useRef} from "react"
import {Button, Loader} from "@vervstack/chures"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {
    createRequest,
    type VcnNetworkWizardContext,
} from "@/dialogs/CreateVcnNetworkDialog/processes/vcnNetworkWizard.ts"
import type {FinalStepProps} from "@/dialogs/StepsDialog/Step.ts"
import {useCreateVcnNetworkMutation} from "@/processes/queries/vcn.ts"
import cls from "@/dialogs/CreateVcnNetworkDialog/screens/NetworkCreationScreen/NetworkCreationScreen.module.css"

type Props = FinalStepProps<VcnNetworkWizardContext>

export default function NetworkCreationScreen({context, onClose}: Props) {
    const createNetwork = useCreateVcnNetworkMutation()
    const hasStarted = useRef(false)

    useEffect(() => {
        if (hasStarted.current) return
        hasStarted.current = true

        const {LockClosing, UnlockClosing} = useDialog.getState()
        LockClosing()
        createNetwork.mutateAsync(createRequest(context))
            .catch(useToaster.getState().catchGrpc)
            .finally(UnlockClosing)
    }, [context, createNetwork.mutateAsync])

    return (
        <div className={cls.NetworkCreationScreenContainer}>
            {(createNetwork.isIdle || createNetwork.isPending) && (
                <div className={cls.StatusWrapper} role="status">
                    <Loader size="sm"/>
                    <span>Creating {context.name.trim()}…</span>
                </div>
            )}
            {createNetwork.isSuccess && (
                <>
                    <p role="status">Network {createNetwork.data.network?.name ?? context.name.trim()} created.</p>
                    <Button type="button" variant="primary" onClick={onClose}>Done</Button>
                </>
            )}
            {createNetwork.isError && (
                <>
                    <p className={cls.Error} role="alert">{createNetwork.error.message}</p>
                    <Button type="button" variant="secondary" onClick={onClose}>Close</Button>
                </>
            )}
        </div>
    )
}
