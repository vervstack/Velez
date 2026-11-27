import {Button, ConfirmDialog} from "@vervstack/chures"

import type {NetworkApiKey} from "@/app/api/velez/verv_closed_network.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {formatContainerTimestamp} from "@/processes/mappings/containers.ts"
import {useExpireNetworkApiKeyMutation} from "@/processes/queries/vcn.ts"

interface Props {
    networkId: string
    apiKey: NetworkApiKey
}

export default function VcnApiKeyRow({networkId, apiKey}: Props) {
    const {OpenDialog, CloseDialog, LockClosing, UnlockClosing} = useDialog()
    const toaster = useToaster()
    const expireKey = useExpireNetworkApiKeyMutation(networkId)
    const prefix = apiKey.prefix ?? ""

    function handleConfirmExpiry() {
        LockClosing()
        return expireKey.mutateAsync(prefix)
            .then(function handleExpired() {
                UnlockClosing()
                CloseDialog()
                toaster.bake({title: "API key expired", description: prefix, level: "Info"})
            })
            .catch(toaster.catchGrpc)
            .finally(UnlockClosing)
    }

    function handleExpire() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Expire API key?"
                message={`Requests using the key starting with ${prefix} will lose API access.`}
                confirmLabel="Expire API key"
                onConfirm={handleConfirmExpiry}
                onClose={CloseDialog}
            />
        )
    }

    function renderAction() {
        return (
            <Button
                variant="danger"
                onClick={handleExpire}
                disabled={!prefix || expireKey.isPending}
                aria-label={`Expire API key ${prefix}`}
            >
                Expire
            </Button>
        )
    }

    return (
        <tr>
            <th scope="row">{prefix || "—"}</th>
            <td>{formatContainerTimestamp(apiKey.createdAt)}</td>
            <td>{formatContainerTimestamp(apiKey.expiration)}</td>
            <td>{formatContainerTimestamp(apiKey.lastSeen)}</td>
            <td>{renderAction()}</td>
        </tr>
    )
}
