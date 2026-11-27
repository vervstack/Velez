import {Button, ConfirmDialog} from "@vervstack/chures"

import type {VcnNetwork} from "@/app/api/velez/verv_closed_network.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import {useDropVcnNetworkMutation, useVcnNetworkQuery} from "@/processes/queries/vcn.ts"
import cls from "@/pages/vcn/components/VcnNetworkRow/VcnNetworkRow.module.css"

interface Props {
    network: VcnNetwork
    onOpen: (networkId: string) => void
}

export default function VcnNetworkRow({network, onOpen}: Props) {
    const {OpenDialog, CloseDialog, LockClosing, UnlockClosing} = useDialog()
    const toaster = useToaster()
    const networkId = network.id ?? ""
    const name = network.name || networkId || "Unnamed network"
    const detailsQuery = useVcnNetworkQuery(networkId)
    const dropNetwork = useDropVcnNetworkMutation()
    const isUnavailable = !networkId || detailsQuery.isError
    const isChecking = !isUnavailable && detailsQuery.isPending

    function handleOpen() {
        onOpen(networkId)
    }

    function handleRetry() {
        detailsQuery.refetch()
    }

    function handleConfirmDrop() {
        LockClosing()
        return dropNetwork.mutateAsync(networkId)
            .then(function handleDropped() {
                UnlockClosing()
                CloseDialog()
                toaster.bake({title: "Network dropped", description: name, level: "Info"})
            })
            .catch(toaster.catchGrpc)
            .finally(UnlockClosing)
    }

    function handleDrop() {
        const message = network.source === "deployed"
            ? `Dropping ${name} deletes its Headscale data, peers, and API keys.`
            : `Dropping ${name} removes its registration from Velez.`
        OpenDialog(
            <ConfirmDialog
                danger
                title="Drop network?"
                message={message}
                confirmLabel="Drop network"
                onConfirm={handleConfirmDrop}
                onClose={CloseDialog}
            />
        )
    }

    function renderStatus() {
        if (isChecking) return <span className={cls.Status}>Checking…</span>
        if (isUnavailable) return <span className={cls.Unavailable}>Unavailable</span>
        return <span className={cls.Reachable}>Reachable</span>
    }

    function renderCount() {
        if (isChecking) return <SkeletonLoader shape="line" width="2rem"/>
        if (isUnavailable) return <span className={cls.Status}>Unknown</span>
        return detailsQuery.data?.peers?.length ?? 0
    }

    return (
        <tr className={cls.VcnNetworkRowContainer}>
            <th scope="row">{name}</th>
            <td className={cls.Source}>{network.source || "Unknown"}</td>
            <td>{renderStatus()}</td>
            <td>{renderCount()}</td>
            <td>{renderActions()}</td>
        </tr>
    )

    function renderActions() {
        return (
            <div className={cls.ActionsWrapper}>
                {isUnavailable && networkId && (
                    <Button onClick={handleRetry} disabled={detailsQuery.isFetching}>Retry</Button>
                )}
                <Button onClick={handleOpen} disabled={!networkId} aria-label={`View details for ${name}`}>
                    Details
                </Button>
                <Button
                    variant="danger"
                    onClick={handleDrop}
                    disabled={!networkId || dropNetwork.isPending}
                    aria-label={`Drop ${name}`}
                >
                    Drop
                </Button>
            </div>
        )
    }
}
