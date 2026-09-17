import {useEffect, useMemo} from "react"

import cls from "@/pages/runners/RunnersPage.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useListRunnersQuery} from "@/processes/queries/runners.ts"
import Button from "@/components/base/Button.tsx"
import RunnerCreateDialog from "@/dialogs/RunnerCreateDialog/RunnerCreateDialog.tsx"
import RunnerRow from "@/pages/runners/components/RunnerRow/RunnerRow.tsx"
import RunnersEmptyState from "@/pages/runners/components/RunnersEmptyState/RunnersEmptyState.tsx"

const COLUMNS = ["", "Name", "Provider", "Scope", "Target", "Status", "Created", ""]

export default function RunnersPage() {
    const {OpenDialog} = useDialog()
    const toaster = useToaster()

    const runnersQuery = useListRunnersQuery()
    useEffect(() => {
        if (runnersQuery.error) toaster.catchGrpc(runnersQuery.error)
    }, [runnersQuery.error])

    const runners = useMemo(
        () => (runnersQuery.data?.runners ?? []).sort((a, b) => (a.name ?? "").localeCompare(b.name ?? "")),
        [runnersQuery.data]
    )

    function handleCreate() {
        OpenDialog(<RunnerCreateDialog/>)
    }

    function renderColumnHeader(label: string, i: number) {
        return <span key={i} className={cls.headerCell}>{label}</span>
    }

    function renderRow(runner: typeof runners[number]) {
        return <RunnerRow key={runner.name} runner={runner}/>
    }

    let content: React.ReactNode
    if (runnersQuery.isLoading) {
        content = <div className={cls.loading}>Loading…</div>
    } else if (runners.length === 0) {
        content = <RunnersEmptyState onCreate={handleCreate}/>
    } else {
        content = (
            <div className={cls.table}>
                <div className={cls.tableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {runners.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.RunnersPageContainer}>
            <div className={cls.toolbar}>
                <h1 className={cls.pageTitle}>Runners</h1>
                <span className={cls.count}>{runners.length} runners</span>
                <div className={cls.toolbarRight}>
                    <Button variant="primary" onClick={handleCreate}>
                        Create runner
                    </Button>
                </div>
            </div>
            {content}
        </div>
    )
}
