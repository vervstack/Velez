import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import VcnApiKeys from "@/pages/vcn/components/VcnApiKeys/VcnApiKeys.tsx"
import VcnPeers from "@/pages/vcn/components/VcnPeers/VcnPeers.tsx"
import {formatContainerTimestamp} from "@/processes/mappings/containers.ts"
import {useVcnNetworkQuery} from "@/processes/queries/vcn.ts"
import cls from "@/pages/vcn/components/VcnNetworkDetails/VcnNetworkDetails.module.css"

interface Props {
    networkId: string
}

export default function VcnNetworkDetails({networkId}: Props) {
    const detailsQuery = useVcnNetworkQuery(networkId)
    const network = detailsQuery.data?.network

    if (detailsQuery.isLoading) return <SkeletonLoader shape="block" height="12rem"/>
    if (!network) {
        return <QueryErrorState message="Failed to load network details." onRetry={detailsQuery.refetch}/>
    }

    return (
        <div className={cls.VcnNetworkDetailsContainer}>
            <h1 className={cls.Title}>{network.name || networkId}</h1>
            {renderMetadata()}
            {detailsQuery.isError && (
                <QueryErrorState
                    message="Network unavailable. Showing the last loaded details."
                    onRetry={detailsQuery.refetch}
                />
            )}
            <VcnPeers peers={detailsQuery.data?.peers}/>
            <VcnApiKeys networkId={networkId} apiKeys={detailsQuery.data?.apiKeys}/>
        </div>
    )

    function renderMetadata() {
        return (
            <dl className={cls.Metadata}>
                <dt>Source</dt>
                <dd className={cls.Source}>{network?.source || "Unknown"}</dd>
                <dt>Status</dt>
                <dd>{detailsQuery.isError ? "Unavailable" : "Reachable"}</dd>
                <dt>Provider</dt>
                <dd>{network?.provider || "—"}</dd>
                <dt>Server URL</dt>
                <dd>{network?.url || "—"}</dd>
                <dt>Headscale service</dt>
                <dd>{network?.serviceName || "—"}</dd>
                <dt>Created</dt>
                <dd>{formatContainerTimestamp(network?.createdAt)}</dd>
            </dl>
        )
    }
}
