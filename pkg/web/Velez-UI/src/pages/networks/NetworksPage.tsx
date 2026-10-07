import {useMemo, useState} from "react"
import {ConfirmDialog, Toggle} from "@vervstack/chures"
import {useNavigate} from "react-router-dom"

import cls from "@/pages/networks/NetworksPage.module.css"
import type {Network} from "@/app/api/velez/network_api.pb"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {Routes} from "@/app/router/Routes.ts"
import {shouldShowClusterBanner, sortNetworks} from "@/processes/mappings/networks.ts"
import {
    DeleteNetworkMutation,
    DisconnectContainerMutation,
    useListNetworksQuery,
    useNetworkStatusQuery,
} from "@/processes/queries/networks.ts"
import Button from "@/components/base/Button.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import NetworkClusterBanner from "@/components/NetworkClusterBanner/NetworkClusterBanner.tsx"
import AttachContainerDialog from "@/dialogs/AttachContainerDialog/AttachContainerDialog.tsx"
import CreateNetworkDialog from "@/dialogs/CreateNetworkDialog/CreateNetworkDialog.tsx"
import NetworkRow from "@/pages/networks/components/NetworkRow/NetworkRow.tsx"
import NetworksListSkeleton from "@/pages/networks/components/NetworksListSkeleton/NetworksListSkeleton.tsx"

export default function NetworksPage() {
    const [isForeignIncluded, setIsForeignIncluded] = useState(false)

    const navigate = useNavigate()
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const environment = useEnvironmentStore((state) => state.selectedEnvironment)

    const statusQuery = useNetworkStatusQuery()
    const networksQuery = useListNetworksQuery(environment, isForeignIncluded)
    const deleteNetwork = DeleteNetworkMutation()
    const disconnectContainer = DisconnectContainerMutation()

    const networks = useMemo(() => sortNetworks(networksQuery.data?.networks ?? []), [networksQuery.data])

    function handleOpenVcn() {
        navigate(Routes.VCN)
    }

    function handleCreate() {
        OpenDialog(<CreateNetworkDialog/>)
    }

    function handleAttach(network: Network) {
        OpenDialog(<AttachContainerDialog network={network}/>)
    }

    function handleDelete(network: Network) {
        function handleConfirm() {
            deleteNetwork.mutateAsync({id: network.id, environment})
                .then(function handleDeleted() {
                    toaster.bake({title: "Network deleted", description: network.name ?? "", level: "Info"})
                    CloseDialog()
                })
                .catch(toaster.catchGrpc)
        }

        OpenDialog(
            <ConfirmDialog
                danger
                title="Delete network?"
                message={`${network.name} will be removed.`}
                confirmLabel="Delete"
                onConfirm={handleConfirm}
                onClose={CloseDialog}
            />
        )
    }

    function handleDetach(network: Network, containerName: string) {
        function handleConfirm() {
            disconnectContainer.mutateAsync({networkId: network.id, containerName, environment})
                .then(function handleDetached() {
                    toaster.bake({title: "Container detached", description: containerName, level: "Info"})
                    CloseDialog()
                })
                .catch(toaster.catchGrpc)
        }

        OpenDialog(
            <ConfirmDialog
                danger
                title="Detach container?"
                message={`${containerName} will be disconnected from ${network.name}.`}
                confirmLabel="Detach"
                onConfirm={handleConfirm}
                onClose={CloseDialog}
            />
        )
    }

    function handleRetry() {
        networksQuery.refetch()
    }

    function renderNetwork(network: Network) {
        return (
            <NetworkRow
                key={network.id} network={network}
                onAttach={handleAttach} onDelete={handleDelete} onDetach={handleDetach}
            />
        )
    }

    function renderContent() {
        if (networksQuery.isLoading) return <NetworksListSkeleton/>
        if (networksQuery.isError) {
            return <QueryErrorState message="Failed to load networks." onRetry={handleRetry}/>
        }
        if (networks.length === 0) {
            return <div className={cls.EmptyMessage}>No networks in this environment.</div>
        }
        return <div className={cls.List}>{networks.map(renderNetwork)}</div>
    }

    return (
        <div className={cls.NetworksPageContainer}>
            <div className={cls.Toolbar}>
                <h1 className={cls.PageTitle}>Networks</h1>
                <div className={cls.ToolbarRight}>
                    <Toggle
                        label="Show all Docker networks"
                        checked={isForeignIncluded}
                        onChange={setIsForeignIncluded}
                    />
                    <Button variant="primary" onClick={handleCreate}>Create network</Button>
                </div>
            </div>
            {shouldShowClusterBanner(statusQuery.data) && <NetworkClusterBanner onOpenVcn={handleOpenVcn}/>}
            {renderContent()}
        </div>
    )
}
