import {useMemo, useState} from "react"
import {Dropdown} from "@vervstack/chures"
import {useQueryClient} from "@tanstack/react-query"

import type {HaproxyInstance} from "@/app/api/velez/haproxy_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import cls from "@/pages/haproxy/HaproxyPage.module.css"
import {HAPROXY_INSTANCES_QUERY_KEY, useSelectHaproxyMasterMutation} from "@/processes/queries/haproxy.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

interface Props {
    instance: HaproxyInstance
}

const DEFAULT_ENVIRONMENT = "PROD"

export default function HaproxyInstanceDetail({instance}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const queryClient = useQueryClient()
    const pgQuery = useListPgInstancesQuery(true)
    const selectMaster = useSelectHaproxyMasterMutation()
    const [masterOverride, setMasterOverride] = useState<string>()
    const currentMaster = instance.masterPgInstance ?? ""
    const master = masterOverride ?? currentMaster
    const environment = instance.environment || DEFAULT_ENVIRONMENT
    const options = useMemo(() => {
        const choices = (pgQuery.data?.instances ?? [])
            .filter((pg) => pg.name && (pg.environment || DEFAULT_ENVIRONMENT) === environment)
            .map((pg) => ({id: pg.name ?? "", name: pg.displayName || pg.name || ""}))
        if (currentMaster && !choices.some((option) => option.id === currentMaster)) {
            choices.push({id: currentMaster, name: currentMaster})
        }
        return choices
    }, [pgQuery.data, environment, currentMaster])

    function handleMasterChange(ids: string[]) {
        setMasterOverride(ids[0])
    }

    function handleConfirm() {
        const req = {name: instance.name, masterPgInstance: master}

        function start() {
            return selectMaster.mutateAsync(req)
        }

        function handleSuccess() {
            queryClient.invalidateQueries({queryKey: HAPROXY_INSTANCES_QUERY_KEY})
        }

        function handleClose() {
            setMasterOverride(undefined)
            CloseDialog()
        }

        OpenDialog(
            <DialogShell title="Select master" onClose={handleClose} isFlush>
                <TaskProgressScreen title="Selecting HAProxy master" metaLine={instance.name ?? ""}
                                    metaLineSecondary={master} start={start} onSuccess={handleSuccess}
                                    onClose={handleClose}/>
            </DialogShell>,
        )
    }

    function renderMasterSelection() {
        if (pgQuery.isLoading) return <SkeletonLoader shape="line" width="16rem"/>
        if (pgQuery.isError) {
            return <QueryErrorState message="Failed to load PostgreSQL instances." onRetry={pgQuery.refetch}/>
        }

        return (
            <div className={cls.MasterSelection}>
                <Dropdown label="Select master" options={options} value={master ? [master] : []}
                          onChange={handleMasterChange} portal/>
                <Button variant="primary" onClick={handleConfirm}
                        disabled={instance.status !== "running" || !master || master === currentMaster
                            || !options.some((option) => option.id === master)}>
                    Confirm
                </Button>
            </div>
        )
    }

    return (
        <section className={cls.Detail}>
            <h2>{instance.name}</h2>
            {renderMasterSelection()}
        </section>
    )
}
