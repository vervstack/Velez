import {useEffect, useMemo} from "react"

import cls from "@/pages/postgres/PostgresPage.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import {sortPgInstancesByName} from "@/processes/mappings/pg_instances.ts"
import Button from "@/components/base/Button.tsx"
import PgInstanceCreateDialog from "@/dialogs/PgInstanceCreateDialog/PgInstanceCreateDialog.tsx"
import PgInstanceRow from "@/pages/postgres/components/PgInstanceRow/PgInstanceRow.tsx"
import PostgresEmptyState from "@/pages/postgres/components/PostgresEmptyState/PostgresEmptyState.tsx"

const COLUMNS = ["", "Name", "Environment", "Database", "Username", "Port", "Created", ""]

export default function PostgresPage() {
    const {OpenDialog} = useDialog()
    const toaster = useToaster()

    const instancesQuery = useListPgInstancesQuery()
    useEffect(() => {
        if (instancesQuery.error) toaster.catchGrpc(instancesQuery.error)
    }, [instancesQuery.error])

    const instances = useMemo(
        () => sortPgInstancesByName(instancesQuery.data?.instances ?? []),
        [instancesQuery.data]
    )

    function handleCreate() {
        OpenDialog(<PgInstanceCreateDialog/>)
    }

    function renderColumnHeader(label: string, i: number) {
        return <span key={i} className={cls.headerCell}>{label}</span>
    }

    function renderRow(instance: typeof instances[number]) {
        return <PgInstanceRow key={instance.name} instance={instance}/>
    }

    let content: React.ReactNode
    if (instancesQuery.isLoading) {
        content = <div className={cls.loading}>Loading…</div>
    } else if (instances.length === 0) {
        content = <PostgresEmptyState onCreate={handleCreate}/>
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
        <div className={cls.PostgresPageContainer}>
            <div className={cls.toolbar}>
                <h1 className={cls.pageTitle}>Postgres</h1>
                <span className={cls.count}>{instances.length} instances</span>
                <div className={cls.toolbarRight}>
                    <Button variant="primary" onClick={handleCreate}>
                        Create database
                    </Button>
                </div>
            </div>
            {content}
        </div>
    )
}
