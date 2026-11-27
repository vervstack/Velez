import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import VcnNetworkRow from "@/pages/vcn/components/VcnNetworkRow/VcnNetworkRow.tsx"
import {useListVcnNetworksQuery} from "@/processes/queries/vcn.ts"
import cls from "@/pages/vcn/components/VcnNetworkList/VcnNetworkList.module.css"

interface Props {
    onOpen: (networkId: string) => void
}

export default function VcnNetworkList({onOpen}: Props) {
    const networksQuery = useListVcnNetworksQuery()
    const networks = networksQuery.data ?? []

    if (networksQuery.isLoading) return <SkeletonLoader shape="block" height="8rem"/>
    if (networksQuery.isError) {
        return <QueryErrorState message="Failed to load Verv Closed Networks." onRetry={networksQuery.refetch}/>
    }
    function renderTableHeading() {
        return (
            <thead>
                <tr>
                    <th scope="col">Network</th>
                    <th scope="col">Source</th>
                    <th scope="col">Status</th>
                    <th scope="col">Peers</th>
                    <th scope="col">Actions</th>
                </tr>
            </thead>
        )
    }

    function renderTable() {
        if (networks.length === 0) {
            return <p className={cls.EmptyMessage}>No Verv Closed Networks yet.</p>
        }

        return (
            <table className={cls.Table} aria-label="Verv Closed Networks">
                {renderTableHeading()}
                <tbody>
                    {networks.map(network => <VcnNetworkRow key={network.id} network={network} onOpen={onOpen}/>)}
                </tbody>
            </table>
        )
    }

    return <div className={cls.VcnNetworkListContainer}>{renderTable()}</div>
}
