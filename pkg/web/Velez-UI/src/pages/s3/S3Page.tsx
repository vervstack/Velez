import {useMemo, useState} from "react"

import cls from "@/pages/s3/S3Page.module.css"
import type {S3Instance} from "@/app/api/velez/s3_api.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {isInstanceProvisioning} from "@/processes/mappings/provisioning.ts"
import {sortS3InstancesByName} from "@/processes/mappings/s3.ts"
import {S3_INSTANCES_QUERY_KEY, useListS3InstancesQuery} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import InstanceCount from "@/components/InstanceCount/InstanceCount.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import CreateS3InstanceDialog from "@/dialogs/CreateS3InstanceDialog/CreateS3InstanceDialog.tsx"
import S3InstanceDetail from "@/pages/s3/components/S3InstanceDetail/S3InstanceDetail.tsx"
import S3InstanceRow from "@/pages/s3/components/S3InstanceRow/S3InstanceRow.tsx"
import S3ListSkeleton from "@/pages/s3/components/S3ListSkeleton/S3ListSkeleton.tsx"
import ProvisioningRow from "@/widgets/ProvisioningRow/ProvisioningRow.tsx"

const COLUMNS = ["", "Name", "Environment", "S3 port", "Web UI", "Region", ""]

function renderColumnHeader(label: string, i: number) {
    return <span key={i} className={cls.HeaderCell}>{label}</span>
}

export default function S3Page() {
    const [selectedName, setSelectedName] = useState("")
    const {OpenDialog} = useDialog()
    const instancesQuery = useListS3InstancesQuery()

    const instances = useMemo(
        () => sortS3InstancesByName(instancesQuery.data?.instances ?? []),
        [instancesQuery.data]
    )
    const tasks = useMemo(() => instancesQuery.data?.provisioning ?? [], [instancesQuery.data])
    const visibleInstances = useMemo(
        () => instances.filter((instance) => !isInstanceProvisioning(instance.name ?? "", tasks)),
        [instances, tasks]
    )
    const selectedInstance = useMemo(
        () => instances.find((instance) => instance.name === selectedName),
        [instances, selectedName]
    )

    function handleCreate() {
        OpenDialog(<CreateS3InstanceDialog/>)
    }

    function handleRetry() {
        instancesQuery.refetch()
    }

    function renderProvisioningRow(task: ProvisioningTask) {
        return (
            <ProvisioningRow
                key={task.taskId}
                task={task}
                noun="S3 instance"
                queryKey={S3_INSTANCES_QUERY_KEY}
            />
        )
    }

    function renderRow(instance: S3Instance) {
        return (
            <S3InstanceRow
                key={instance.name}
                instance={instance}
                isSelected={instance.name === selectedName}
                onSelect={setSelectedName}
            />
        )
    }

    function renderContent() {
        if (instancesQuery.isLoading) return <S3ListSkeleton/>
        if (instancesQuery.isError) {
            return <QueryErrorState message="Failed to load S3 instances." onRetry={handleRetry}/>
        }
        if (instances.length === 0 && tasks.length === 0) {
            return <div className={cls.EmptyMessage}>No S3 instances on this node.</div>
        }
        return (
            <div className={cls.Table}>
                <div className={cls.TableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {tasks.map(renderProvisioningRow)}
                {visibleInstances.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.S3PageContainer}>
            <div className={cls.Toolbar}>
                <h1 className={cls.PageTitle}>S3 storage</h1>
                <InstanceCount count={instances.length} label="instances" isLoading={instancesQuery.isLoading}/>
                <div className={cls.ToolbarRight}>
                    <Button variant="primary" onClick={handleCreate}>
                        Create S3 instance
                    </Button>
                </div>
            </div>
            {renderContent()}
            {selectedInstance && <S3InstanceDetail instanceName={selectedInstance.name ?? ""}/>}
        </div>
    )
}
