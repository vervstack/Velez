import {useEffect, useMemo} from "react"

import cls from "@/pages/runners/RunnersPage.module.css"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {RUNNERS_QUERY_KEY, useListRunnersQuery} from "@/processes/queries/runners.ts"
import {isRunnerProvisioning} from "@/processes/mappings/runnerDisplay.ts"
import Button from "@/components/base/Button.tsx"
import InstanceCount from "@/components/InstanceCount/InstanceCount.tsx"
import CreateServiceDialog from "@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx"
import RunnerRow from "@/pages/runners/components/RunnerRow/RunnerRow.tsx"
import RunnersEmptyState from "@/pages/runners/components/RunnersEmptyState/RunnersEmptyState.tsx"
import ProvisioningRow from "@/widgets/ProvisioningRow/ProvisioningRow.tsx"

const COLUMNS = ["", "", "Name", "Provider", "Scope", "Target", "Status", "Created", ""]

export default function RunnersPage() {
    const {OpenDialog} = useDialog()
    const toaster = useToaster()

    const runnersQuery = useListRunnersQuery()
    useEffect(() => {
        if (runnersQuery.error) toaster.catchGrpc(runnersQuery.error)
    }, [runnersQuery.error])

    const tasks = runnersQuery.data?.provisioning ?? []
    const runners = useMemo(
        () => (runnersQuery.data?.runners ?? []).sort((a, b) => (a.name ?? "").localeCompare(b.name ?? "")),
        [runnersQuery.data]
    )
    const visibleRunners = useMemo(
        () => runners.filter((runner) => !isRunnerProvisioning(runner.name ?? "", tasks)),
        [runners, runnersQuery.data]
    )

    function handleCreate() {
        OpenDialog(<CreateServiceDialog initialScreen="githubRunner"/>)
    }

    function renderColumnHeader(label: string, i: number) {
        return <span key={i} className={cls.headerCell}>{label}</span>
    }

    function renderRow(runner: typeof runners[number]) {
        return <RunnerRow key={runner.name} runner={runner}/>
    }

    function renderProvisioningRow(task: ProvisioningTask) {
        return (
            <ProvisioningRow
                key={task.taskId}
                task={task}
                noun="runner"
                queryKey={RUNNERS_QUERY_KEY}
            />
        )
    }

    let content: React.ReactNode
    if (runnersQuery.isLoading) {
        content = <div className={cls.loading}>Loading…</div>
    } else if (runners.length === 0 && tasks.length === 0) {
        content = <RunnersEmptyState onCreate={handleCreate}/>
    } else {
        content = (
            <div className={cls.table}>
                <div className={cls.tableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {tasks.map(renderProvisioningRow)}
                {visibleRunners.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.RunnersPageContainer}>
            <div className={cls.toolbar}>
                <h1 className={cls.pageTitle}>Runners</h1>
                <InstanceCount count={runners.length} label="runners" isLoading={runnersQuery.isLoading}/>
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
