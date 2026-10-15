import {useMemo} from "react"

import cls from "@/pages/transfer/components/TransfersList/TransfersList.module.css"
import type {TransferInfo} from "@/app/api/velez/transfer_api.pb"
import {sortTransfersNewestFirst} from "@/processes/mappings/transfer.ts"
import {useListTransfersQuery} from "@/processes/queries/transfer.ts"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import TransferRow from "@/pages/transfer/components/TransferRow/TransferRow.tsx"
import TransfersListSkeleton from "@/pages/transfer/components/TransfersListSkeleton/TransfersListSkeleton.tsx"

const COLUMNS = ["Direction", "Container", "Peer", "Status", "Progress", ""]

function renderColumnHeader(label: string, i: number) {
    return <span key={i} className={cls.HeaderCell}>{label}</span>
}

function renderRow(transfer: TransferInfo) {
    return <TransferRow key={transfer.id} transfer={transfer}/>
}

export default function TransfersList() {
    const transfersQuery = useListTransfersQuery()

    const transfers = useMemo(
        () => sortTransfersNewestFirst(transfersQuery.data?.transfers ?? []),
        [transfersQuery.data]
    )

    function handleRetry() {
        transfersQuery.refetch()
    }

    function renderContent() {
        if (transfersQuery.isLoading) return <TransfersListSkeleton/>
        if (transfersQuery.isError) {
            return <QueryErrorState message="Failed to load transfers." onRetry={handleRetry}/>
        }
        if (transfers.length === 0) {
            return <div className={cls.EmptyMessage}>No transfers yet.</div>
        }
        return (
            <div className={cls.Table}>
                <div className={cls.TableHeader}>{COLUMNS.map(renderColumnHeader)}</div>
                {transfers.map(renderRow)}
            </div>
        )
    }

    return (
        <section className={cls.TransfersListContainer}>
            <h2 className={cls.SectionTitle}>Transfers</h2>
            {renderContent()}
        </section>
    )
}
