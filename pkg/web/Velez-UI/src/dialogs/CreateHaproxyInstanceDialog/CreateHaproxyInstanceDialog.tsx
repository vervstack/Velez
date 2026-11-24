import {useMemo, useState} from "react"
import {Dropdown} from "@vervstack/chures"
import {useQueryClient} from "@tanstack/react-query"

import type {CreateHaproxyInstanceRequest} from "@/app/api/velez/haproxy_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import {HAPROXY_INSTANCES_QUERY_KEY, useCreateHaproxyInstanceMutation} from "@/processes/queries/haproxy.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

const MAX_PORT = 65535

export default function CreateHaproxyInstanceDialog() {
    const {CloseDialog} = useDialog()
    const queryClient = useQueryClient()
    const createInstance = useCreateHaproxyInstanceMutation()
    const pgQuery = useListPgInstancesQuery(true)
    const [name, setName] = useState("")
    const [port, setPort] = useState("")
    const [box, setBox] = useState("")
    const [master, setMaster] = useState("")
    const [request, setRequest] = useState<CreateHaproxyInstanceRequest>()
    const pgInstances = useMemo(() => pgQuery.data?.instances ?? [], [pgQuery.data])
    const options = useMemo(
        () => pgInstances.filter((pg) => pg.name).map((pg) => ({
            id: pg.name ?? "",
            name: pg.environment ? `${pg.displayName || pg.name} (${pg.environment})` : pg.displayName || pg.name || "",
        })),
        [pgInstances],
    )
    const pg = pgInstances.find((instance) => instance.name === master)
    const portNumber = Number(port)
    const nameError = validateInstanceName(name.trim())
    const portError = port && (!/^\d+$/.test(port) || portNumber < 1 || portNumber > MAX_PORT)
        ? "Enter a port between 1 and 65535."
        : undefined

    function handleCreate() {
        if (!pg || !name.trim() || nameError || portError) return

        setRequest({
            name: name.trim(),
            masterPgInstance: master,
            environment: pg.environment || undefined,
            box: box.trim() || undefined,
            port: port ? portNumber : undefined,
        })
    }

    function handleMasterChange(ids: string[]) {
        setMaster(ids[0] ?? "")
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: HAPROXY_INSTANCES_QUERY_KEY})
        queryClient.invalidateQueries({queryKey: ["services"]})
    }

    function renderMasterField() {
        if (pgQuery.isLoading) return <SkeletonLoader shape="line"/>
        if (pgQuery.isError) {
            return <QueryErrorState message="Failed to load PostgreSQL instances." onRetry={pgQuery.refetch}/>
        }

        return <Dropdown label="Master" placeholder="Select a PostgreSQL instance" options={options}
                         value={master ? [master] : []} onChange={handleMasterChange} portal/>
    }

    if (request) {
        const submittedRequest = request
        function start() {
            return createInstance.mutateAsync(submittedRequest)
        }

        return (
            <DialogShell title="Create HAProxy instance" onClose={CloseDialog} isFlush>
                <TaskProgressScreen title="Creating HAProxy instance" metaLine={request.name ?? ""}
                                    start={start} onSuccess={handleSuccess} onClose={CloseDialog}/>
            </DialogShell>
        )
    }

    return (
        <DialogShell title="Create HAProxy instance" onClose={CloseDialog}>
            <div className={cls.Fields}>
                <Input label="Name" inputValue={name} onChange={setName} error={nameError}/>
                {renderMasterField()}
                <Input label="Port (default 5432)" inputValue={port} onChange={setPort} error={portError}/>
                <Input label="Box (optional)" inputValue={box} onChange={setBox}/>
                <div className={cls.Actions}>
                    <Button variant="secondary" onClick={CloseDialog}>Cancel</Button>
                    <Button variant="primary" onClick={handleCreate}
                            disabled={!pg || !name.trim() || !!nameError || !!portError || pgQuery.isError}>
                        Create
                    </Button>
                </div>
            </div>
        </DialogShell>
    )
}
