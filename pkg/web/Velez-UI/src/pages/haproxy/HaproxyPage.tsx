import {useMemo, useState} from "react"

import type {HaproxyInstance} from "@/app/api/velez/haproxy_api.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import InstanceCount from "@/components/InstanceCount/InstanceCount.tsx"
import CreateHaproxyInstanceDialog from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.tsx"
import HaproxyInstanceDetail from "@/pages/haproxy/components/HaproxyInstanceDetail/HaproxyInstanceDetail.tsx"
import HaproxyInstanceRow from "@/pages/haproxy/components/HaproxyInstanceRow/HaproxyInstanceRow.tsx"
import cls from "@/pages/haproxy/HaproxyPage.module.css"
import {isInstanceProvisioning} from "@/processes/mappings/provisioning.ts"
import {HAPROXY_INSTANCES_QUERY_KEY, useListHaproxyInstancesQuery} from "@/processes/queries/haproxy.ts"
import ProvisioningRow from "@/widgets/ProvisioningRow/ProvisioningRow.tsx"

const COLUMNS = ["Status", "Name", "Environment", "Port", "Master", "Actions"]

export default function HaproxyPage() {
    const {OpenDialog} = useDialog()
    const [selectedName, setSelectedName] = useState("")
    const instancesQuery = useListHaproxyInstancesQuery()
    const instances = useMemo(() => instancesQuery.data?.instances ?? [], [instancesQuery.data])
    const tasks = useMemo(() => instancesQuery.data?.provisioning ?? [], [instancesQuery.data])
    const visibleInstances = useMemo(
        () => instances.filter((instance) => !isInstanceProvisioning(instance.name ?? "", tasks)),
        [instances, tasks],
    )
    const selectedInstance = useMemo(
        () => visibleInstances.find((instance) => instance.name === selectedName),
        [visibleInstances, selectedName],
    )

    function handleCreate() {
        OpenDialog(<CreateHaproxyInstanceDialog/>)
    }

    function renderProvisioningRow(task: ProvisioningTask) {
        return <ProvisioningRow key={task.taskId} task={task} noun="HAProxy instance"
                                queryKey={HAPROXY_INSTANCES_QUERY_KEY}/>
    }

    function renderInstance(instance: HaproxyInstance) {
        return <HaproxyInstanceRow key={instance.name} instance={instance} onSelect={setSelectedName}
                                   isSelected={instance.name === selectedName}/>
    }

    function renderContent() {
        if (instancesQuery.isLoading) return <SkeletonLoader shape="block" height="8rem"/>
        if (instancesQuery.isError) {
            return <QueryErrorState message="Failed to load HAProxy instances." onRetry={instancesQuery.refetch}/>
        }
        if (instances.length === 0 && tasks.length === 0) {
            return <div className={cls.EmptyMessage}>No HAProxy instances on this node.</div>
        }

        return (
            <div className={cls.Table}>
                <div className={cls.TableHeader}>
                    {COLUMNS.map((column) => <span key={column}>{column}</span>)}
                </div>
                {tasks.map(renderProvisioningRow)}
                {visibleInstances.map(renderInstance)}
            </div>
        )
    }

    return (
        <div className={cls.HaproxyPageContainer}>
            <div className={cls.Toolbar}>
                <h1 className={cls.PageTitle}>HAProxy</h1>
                <InstanceCount count={instances.length} label="instances" isLoading={instancesQuery.isLoading}/>
                <div className={cls.ToolbarRight}>
                    <Button variant="primary" onClick={handleCreate}>Create HAProxy instance</Button>
                </div>
            </div>
            {renderContent()}
            {selectedInstance && <HaproxyInstanceDetail key={selectedInstance.name} instance={selectedInstance}/>}
        </div>
    )
}
