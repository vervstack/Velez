import {useMemo} from "react"

import cls from "@/pages/dinds/DindsPage.module.css"
import type {DindInfo} from "@/app/api/velez/dind_api.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {isInstanceProvisioning} from "@/processes/mappings/provisioning.ts"
import {DINDS_QUERY_KEY, useListDindsQuery} from "@/processes/queries/dinds.ts"
import Button from "@/components/base/Button.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import CreateDindDialog from "@/dialogs/CreateDindDialog/CreateDindDialog.tsx"
import DindRow from "@/pages/dinds/components/DindRow/DindRow.tsx"
import DindsTableSkeleton from "@/pages/dinds/components/DindsTableSkeleton/DindsTableSkeleton.tsx"
import ProvisioningRow from "@/widgets/ProvisioningRow/ProvisioningRow.tsx"

const COLUMNS = ["Name", "Address", "Sysbox", "Created", ""]

function renderColumnHeader(label: string, i: number) {
    return <span key={i} className={cls.HeaderCell}>{label}</span>
}

function renderRow(dind: DindInfo) {
    return <DindRow key={dind.name} dind={dind}/>
}

function renderProvisioningRow(task: ProvisioningTask) {
    return (
        <ProvisioningRow key={task.taskId} task={task} title="Creating Docker daemon" queryKey={DINDS_QUERY_KEY}/>
    )
}

export default function DindsPage() {
    const {OpenDialog} = useDialog()
    const dindsQuery = useListDindsQuery()

    const dinds = useMemo(
        () => [...(dindsQuery.data?.dinds ?? [])].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? "")),
        [dindsQuery.data]
    )

    const tasks = useMemo(() => dindsQuery.data?.provisioning ?? [], [dindsQuery.data])
    const visibleDinds = useMemo(
        () => dinds.filter((dind) => !isInstanceProvisioning(dind.name ?? "", tasks)),
        [dinds, tasks]
    )

    function handleCreate() {
        OpenDialog(<CreateDindDialog/>)
    }

    function handleRetry() {
        dindsQuery.refetch()
    }

    function renderContent() {
        if (dindsQuery.isLoading) return <DindsTableSkeleton/>
        if (dindsQuery.isError) {
            return <QueryErrorState message="Failed to load Docker daemons." onRetry={handleRetry}/>
        }
        if (dinds.length === 0 && tasks.length === 0) {
            return <div className={cls.EmptyMessage}>No Docker daemons on this node.</div>
        }
        return (
            <div className={cls.Table}>
                <div className={cls.TableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {tasks.map(renderProvisioningRow)}
                {visibleDinds.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.DindsPageContainer}>
            <div className={cls.Toolbar}>
                <h1 className={cls.PageTitle}>Docker daemons</h1>
                <span className={cls.Count}>{dinds.length} daemons</span>
                <div className={cls.ToolbarRight}>
                    <Button variant="primary" onClick={handleCreate}>
                        Create Docker daemon
                    </Button>
                </div>
            </div>
            {renderContent()}
        </div>
    )
}
