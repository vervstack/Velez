import {useEffect, useMemo} from "react"

import cls from "@/pages/container-registry/ContainerRegistryPage.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useListRegistryInstancesQuery} from "@/processes/queries/registry_instances.ts"
import {sortRegistryInstancesByName} from "@/processes/mappings/registry_instances.ts"
import Button from "@/components/base/Button.tsx"
import RegistryInstanceCreateDialog from "@/dialogs/RegistryInstanceCreateDialog/RegistryInstanceCreateDialog.tsx"
import RegistryInstanceRow from "@/pages/container-registry/components/RegistryInstanceRow/RegistryInstanceRow.tsx"
import ContainerRegistryEmptyState
    from "@/pages/container-registry/components/ContainerRegistryEmptyState/ContainerRegistryEmptyState.tsx"

const COLUMNS = ["", "Name", "Environment", "Port", "Username", "Created", ""]

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

    function handleCreate() {
        OpenDialog(<RegistryInstanceCreateDialog/>)
    }

    function renderColumnHeader(label: string, i: number) {
        return <span key={i} className={cls.headerCell}>{label}</span>
    }

    function renderRow(instance: typeof instances[number]) {
        return <RegistryInstanceRow key={instance.name} instance={instance}/>
    }

    let content: React.ReactNode
    if (instancesQuery.isLoading) {
        content = <div className={cls.loading}>Loading…</div>
    } else if (instances.length === 0) {
        content = <ContainerRegistryEmptyState onCreate={handleCreate}/>
    } else {
        content = (
            <div className={cls.table}>
                <div className={cls.tableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {instances.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.ContainerRegistryPageContainer}>
            <div className={cls.toolbar}>
                <h1 className={cls.pageTitle}>Container Registries</h1>
                <span className={cls.count}>{instances.length} instances</span>
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
