import {useMemo, useState} from "react"
import {Checkbox, Dropdown} from "@vervstack/chures"

import cls from "@/dialogs/SetupPgReplicationDialog/SetupPgReplicationDialog.module.css"

import type {SetupPgReplicationRequest} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import InfoMark from "@/components/base/InfoMark.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import {
    SetupPgReplicationMutation,
    useListPgInstancesQuery,
    useRefreshPgReplication,
} from "@/processes/queries/pg_instances.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

interface Props {
    name: string
}

const SLOT_TOOLTIP = "The master keeps WAL until the replica consumes it. "
    + "A replica survives any outage, but an abandoned replica fills the master disk."

export default function SetupPgReplicationDialog({name}: Props) {
    const {CloseDialog, LockClosing, UnlockClosing} = useDialog()
    const toaster = useToaster()
    const instancesQuery = useListPgInstancesQuery(true)
    const setupReplication = SetupPgReplicationMutation()
    const refreshReplication = useRefreshPgReplication()
    const [source, setSource] = useState(name)
    const [target, setTarget] = useState("")
    const [hasSlot, setHasSlot] = useState(true)
    const [request, setRequest] = useState<SetupPgReplicationRequest | null>(null)

    const sourceOptions = useMemo(() => (instancesQuery.data?.instances ?? []).flatMap((instance) => (
        instance.name ? [{id: instance.name, name: instance.displayName || instance.name}] : []
    )), [instancesQuery.data?.instances])
    const targetOptions = useMemo(() => sourceOptions.filter((option) => option.id !== source), [sourceOptions, source])
    const isSourceAvailable = sourceOptions.some((option) => option.id === source)
    const isTargetAvailable = targetOptions.some((option) => option.id === target)
    const canSubmit = !instancesQuery.isError && isSourceAvailable && isTargetAvailable

    function handleSourceChange(ids: string[]) {
        const nextSource = ids[0] ?? ""
        setSource(nextSource)
        if (nextSource === target) setTarget("")
    }

    function handleTargetChange(ids: string[]) {
        setTarget(ids[0] ?? "")
    }

    function handleConfirm() {
        if (!canSubmit) return
        setRequest({sourceInstance: source, targetInstance: target, useSlot: hasSlot, mode: "physical"})
    }

    function renderEndpoints() {
        if (instancesQuery.isLoading) {
            return <>
                <SkeletonLoader shape="block" height="2.5rem"/>
                <SkeletonLoader shape="block" height="2.5rem"/>
            </>
        }
        if (instancesQuery.isError) {
            return <QueryErrorState message="Could not load PostgreSQL instances." onRetry={instancesQuery.refetch}/>
        }
        return <>
            <Dropdown
                label="Source instance"
                placeholder="Select a source instance"
                options={sourceOptions}
                value={isSourceAvailable ? [source] : []}
                onChange={handleSourceChange}
                portal
            />
            <Dropdown
                label="Target instance"
                placeholder="Select a target instance"
                options={targetOptions}
                value={isTargetAvailable ? [target] : []}
                onChange={handleTargetChange}
                portal
            />
        </>
    }

    if (request) {
        const submittedRequest = request
        function handleStart() {
            LockClosing()
            return setupReplication.mutateAsync(submittedRequest).finally(UnlockClosing)
        }

        function handleSuccess() {
            refreshReplication(source, target)
            toaster.bake({
                title: "Replication set up",
                description: `${target} is a replica of ${source}.`,
                level: "Info",
            })
            CloseDialog()
        }

        return <div className={cls.SetupPgReplicationDialogContainer}>
            <DialogShell title="Setting up replication" onClose={CloseDialog} isFlush>
                <TaskProgressScreen
                    title="Setting up replication"
                    metaLine={`${request.sourceInstance} → ${request.targetInstance}`}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            </DialogShell>
        </div>
    }

    return <div className={cls.SetupPgReplicationDialogContainer}>
        <DialogShell title="Set up replication" onClose={CloseDialog}>
            {renderEndpoints()}
            <div className={cls.ModeRow}>
                <span className={cls.ModeLabel}>Mode</span>
                <span>Physical (streaming)</span>
            </div>
            <div className={cls.SlotRow}>
                <Checkbox label="Use replication slot" checked={hasSlot} onChange={setHasSlot}/>
                <InfoMark tooltip={SLOT_TOOLTIP}/>
            </div>
            <p className={cls.Warning}>
                The target instance will be wiped and become an exact copy of the source, including users and passwords.
            </p>
            <div className={cls.ActionsWrapper}>
                <Button variant="secondary" onClick={CloseDialog}>Cancel</Button>
                <Button variant="danger" disabled={!canSubmit} onClick={handleConfirm}>Set up replication</Button>
            </div>
        </DialogShell>
    </div>
}
