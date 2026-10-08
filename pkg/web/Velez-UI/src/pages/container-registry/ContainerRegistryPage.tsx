import {useEffect, useMemo} from "react"

import cls from "@/pages/container-registry/ContainerRegistryPage.module.css"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {isInstanceProvisioning} from "@/processes/mappings/provisioning.ts"
import {
    REGISTRY_INSTANCES_QUERY_KEY,
    useListRegistryInstancesQuery,
} from "@/processes/queries/registry_instances.ts"
import {sortRegistryInstancesByName} from "@/processes/mappings/registry_instances.ts"
import Button from "@/components/base/Button.tsx"
import InstanceCount from "@/components/InstanceCount/InstanceCount.tsx"
import CreateServiceDialog from "@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx"
import RegistryInstanceRow from "@/pages/container-registry/components/RegistryInstanceRow/RegistryInstanceRow.tsx"
import ProvisioningRow from "@/widgets/ProvisioningRow/ProvisioningRow.tsx"
import ContainerRegistryEmptyState
    from "@/pages/container-registry/components/ContainerRegistryEmptyState/ContainerRegistryEmptyState.tsx"

// Mirrors labels.RegistryaasNamePrefix in internal/domain/labels/verv_labels.go: the listed instance
// name is the service name, while the create task's entity id is the bare name the user typed.
const REGISTRY_NAME_PREFIX = "cr_"

const COLUMNS = ["", "Name", "Environment", "Port", "Username", "Storage", "Created", ""]

export default function ContainerRegistryPage() {
    const {OpenDialog} = useDialog()
    const toaster = useToaster()

    const instancesQuery = useListRegistryInstancesQuery()
    useEffect(() => {
        if (instancesQuery.error) toaster.catchGrpc(instancesQuery.error)
    }, [instancesQuery.error])

    const instances = useMemo(
        () => sortRegistryInstancesByName(instancesQuery.data?.instances ?? []),
        [instancesQuery.data]
    )

    const tasks = useMemo(() => instancesQuery.data?.provisioning ?? [], [instancesQuery.data])
    const visibleInstances = useMemo(
        () => instances.filter(
            (instance) => !isInstanceProvisioning(instance.name ?? "", tasks, REGISTRY_NAME_PREFIX)
        ),
        [instances, tasks]
    )

    function handleCreate() {
        OpenDialog(<CreateServiceDialog initialScreen="registry"/>)
    }

    function renderColumnHeader(label: string, i: number) {
        return <span key={i} className={cls.headerCell}>{label}</span>
    }

    function renderProvisioningRow(task: ProvisioningTask) {
        return (
            <ProvisioningRow
                key={task.taskId}
                task={task}
                noun="container registry"
                prefix={REGISTRY_NAME_PREFIX}
                queryKey={REGISTRY_INSTANCES_QUERY_KEY}
            />
        )
    }

    function renderRow(instance: typeof instances[number]) {
        return <RegistryInstanceRow key={instance.name} instance={instance}/>
    }

    let content: React.ReactNode
    if (instancesQuery.isLoading) {
        content = <div className={cls.loading}>Loading…</div>
    } else if (instances.length === 0 && tasks.length === 0) {
        content = <ContainerRegistryEmptyState onCreate={handleCreate}/>
    } else {
        content = (
            <div className={cls.table}>
                <div className={cls.tableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {tasks.map(renderProvisioningRow)}
                {visibleInstances.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.ContainerRegistryPageContainer}>
            <div className={cls.toolbar}>
                <h1 className={cls.pageTitle}>Container Registries</h1>
                <InstanceCount count={instances.length} label="instances" isLoading={instancesQuery.isLoading}/>
                <div className={cls.toolbarRight}>
                    <Button variant="primary" onClick={handleCreate}>
                        Create registry
                    </Button>
                </div>
            </div>
            {content}
        </div>
    )
}
