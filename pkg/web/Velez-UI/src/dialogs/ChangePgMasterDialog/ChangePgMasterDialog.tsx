import {useMemo, useState} from "react"
import {useQueryClient} from "@tanstack/react-query"
import {Dropdown, SegmentedControl} from "@vervstack/chures"

import cls from "@/dialogs/ChangePgMasterDialog/ChangePgMasterDialog.module.css"

import type {ChangePgMasterRequest} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import {HAPROXY_INSTANCES_QUERY_KEY, useListHaproxyInstancesQuery} from "@/processes/queries/haproxy.ts"
import {useChangePgMasterMutation, useRefreshPgReplication} from "@/processes/queries/pg_instances.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

interface Props {
    newMaster: string
    upstream: string
}

type OldMasterStrategy = "rewind" | "recopy"

const REWIND_TOOLTIP = "Reuses the old master's data and copies only changed pages. "
    + "Requires replication readiness (wal_log_hints) to have been enabled before the switch."
const RECOPY_TOOLTIP = "Erases the old master and copies everything over the network from scratch. "
    + "Slow on large databases, but always works."

const STRATEGY_OPTIONS = [
    {
        value: "rewind",
        content: <span data-tooltip-id="root-tooltip" data-tooltip-content={REWIND_TOOLTIP}>
            Fast catch-up (pg_rewind)
        </span>,
    },
    {
        value: "recopy",
        content: <span data-tooltip-id="root-tooltip" data-tooltip-content={RECOPY_TOOLTIP}>
            Wipe and re-copy (pg_basebackup)
        </span>,
    },
]

export default function ChangePgMasterDialog({newMaster, upstream}: Props) {
    const {CloseDialog, LockClosing, UnlockClosing} = useDialog()
    const toaster = useToaster()
    const queryClient = useQueryClient()
    const balancersQuery = useListHaproxyInstancesQuery()
    const changeMaster = useChangePgMasterMutation()
    const refreshReplication = useRefreshPgReplication()
    const [balancer, setBalancer] = useState("")
    const [strategy, setStrategy] = useState<OldMasterStrategy>("rewind")
    const [request, setRequest] = useState<ChangePgMasterRequest | null>(null)

    const balancerOptions = useMemo(() => (balancersQuery.data?.instances ?? []).flatMap((instance) => (
        instance.name && instance.masterPgInstance === upstream && instance.status === "running"
            ? [{id: instance.name, name: instance.name}]
            : []
    )), [balancersQuery.data?.instances, upstream])
    const isBalancerAvailable = balancerOptions.some((option) => option.id === balancer)
    const canSubmit = !balancersQuery.isError && isBalancerAvailable

    function handleBalancerChange(ids: string[]) {
        setBalancer(ids[0] ?? "")
    }

    function handleStrategyChange(value: string) {
        if (value === "rewind" || value === "recopy") setStrategy(value)
    }

    function handleConfirm() {
        if (!canSubmit) return
        setRequest({balancerInstance: balancer, newMasterInstance: newMaster, oldMasterStrategy: strategy})
    }

    function renderBalancer() {
        if (balancersQuery.isLoading) return <SkeletonLoader shape="block" height="2.5rem"/>
        if (balancersQuery.isError) {
            return <QueryErrorState message="Could not load balancers." onRetry={balancersQuery.refetch}/>
        }
        return <Dropdown
            label="Balancer"
            placeholder="Select a balancer"
            emptyHint={`No running balancers point to ${upstream}.`}
            options={balancerOptions}
            value={isBalancerAvailable ? [balancer] : []}
            onChange={handleBalancerChange}
            portal
        />
    }

    if (request) {
        const submittedRequest = request

        function handleStart() {
            LockClosing()
            return changeMaster.mutateAsync(submittedRequest).finally(UnlockClosing)
        }

        function handleSuccess() {
            refreshReplication(newMaster, upstream)
            queryClient.invalidateQueries({queryKey: HAPROXY_INSTANCES_QUERY_KEY})
            toaster.bake({
                title: "Master changed",
                description: `${newMaster} is now the master. ${upstream} is a replica.`,
                level: "Info",
            })
            CloseDialog()
        }

        return <div className={cls.ChangePgMasterDialogContainer}>
            <DialogShell title="Changing master" onClose={CloseDialog} isFlush>
                <TaskProgressScreen
                    title="Changing PostgreSQL master"
                    metaLine={submittedRequest.balancerInstance ?? ""}
                    metaLineSecondary={`${upstream} → ${newMaster}`}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            </DialogShell>
        </div>
    }

    return <div className={cls.ChangePgMasterDialogContainer}>
        <DialogShell title="Change master (failover)" onClose={CloseDialog}>
            {renderBalancer()}
            <fieldset className={cls.StrategyWrapper}>
                <legend>Old master recovery strategy</legend>
                <SegmentedControl options={STRATEGY_OPTIONS} value={strategy} onChange={handleStrategyChange}/>
            </fieldset>
            <p className={cls.Warning}>
                Stop your services before switching. They keep their balancer DSN and reconnect to the new master.
            </p>
            <div className={cls.ActionsWrapper}>
                <Button variant="secondary" onClick={CloseDialog}>Cancel</Button>
                <Button variant="danger" disabled={!canSubmit} onClick={handleConfirm}>Promote to master</Button>
            </div>
        </DialogShell>
    </div>
}
