import {useState} from "react"

import cls from "@/dialogs/PgDatabaseReplicateDialog/PgDatabaseReplicateDialog.module.css"
import {ReplicatePgDatabaseRequest} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {pgaasService} from "@/processes/api/pgaas.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import Button from "@/components/base/Button.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import Input from "@/components/base/Input.tsx"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

type EndpointMode = "instance" | "plain"

interface EndpointFields {
    instanceId: string
    host: string
    port: string
    user: string
    password: string
    dbName: string
}

const EMPTY_ENDPOINT: EndpointFields = {
    instanceId: "",
    host: "",
    port: "5432",
    user: "",
    password: "",
    dbName: "",
}

export default function PgDatabaseReplicateDialog() {
    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const instancesQuery = useListPgInstancesQuery()
    const [sourceMode, setSourceMode] = useState<EndpointMode>("instance")
    const [targetMode, setTargetMode] = useState<EndpointMode>("instance")
    const [source, setSource] = useState<EndpointFields>(EMPTY_ENDPOINT)
    const [target, setTarget] = useState<EndpointFields>(EMPTY_ENDPOINT)
    const [submittedRequest, setSubmittedRequest] = useState<ReplicatePgDatabaseRequest | null>(null)

    function updateSource(field: keyof EndpointFields, value: string) {
        setSource({...source, [field]: value})
    }

    function updateTarget(field: keyof EndpointFields, value: string) {
        setTarget({...target, [field]: value})
    }

    function buildRequest(): ReplicatePgDatabaseRequest | null {
        const sourcePort = Number(source.port)
        const targetPort = Number(target.port)
        if (sourceMode === "instance" && !source.instanceId) return null
        if (targetMode === "instance" && !target.instanceId) return null
        if (sourceMode === "plain" && (
            !source.host || !source.user || !source.password || !source.dbName || !sourcePort
        )) return null
        if (targetMode === "plain" && (!target.host || !target.user || !target.password || !targetPort)) return null

        return {
            sourceInstanceId: sourceMode === "instance" ? source.instanceId : "",
            sourceHost: sourceMode === "plain" ? source.host : "",
            sourcePort: sourceMode === "plain" ? sourcePort : 0,
            sourceUser: sourceMode === "plain" ? source.user : "",
            sourcePassword: sourceMode === "plain" ? source.password : "",
            sourceDbName: sourceMode === "plain" ? source.dbName : "",
            targetInstanceId: targetMode === "instance" ? target.instanceId : "",
            targetHost: targetMode === "plain" ? target.host : "",
            targetPort: targetMode === "plain" ? targetPort : 0,
            targetUser: targetMode === "plain" ? target.user : "",
            targetPassword: targetMode === "plain" ? target.password : "",
        }
    }

    function handleReplicate() {
        const request = buildRequest()
        if (!request) {
            toaster.bake({
                title: "Complete both endpoints",
                description: "Fill in every required endpoint field before replicating.",
                level: "Error",
            })
            return
        }
        setSubmittedRequest(request)
    }

    function renderEndpoint(
        title: string,
        mode: EndpointMode,
        fields: EndpointFields,
        update: (field: keyof EndpointFields, value: string) => void,
        setMode: (mode: EndpointMode) => void,
        isSource: boolean,
    ) {
        const availableInstances = instancesQuery.data?.instances ?? []
        const instanceChoices = availableInstances.map(function renderInstance(instance) {
            const isSelected = fields.instanceId === instance.name
            return <Button
                key={instance.name}
                variant={isSelected ? "primary" : "secondary"}
                onClick={() => update("instanceId", instance.name ?? "")}
            >
                {instance.displayName || instance.name}
            </Button>
        })

        return <section className={cls.EndpointWrapper}>
            <h3>{title}</h3>
            <div className={cls.ModeButtons}>
                <Button variant={mode === "instance" ? "primary" : "secondary"} onClick={() => setMode("instance")}>
                    Instance
                </Button>
                <Button variant={mode === "plain" ? "primary" : "secondary"} onClick={() => setMode("plain")}>
                    Connection
                </Button>
            </div>
            {mode === "instance" ? <div className={cls.InstancesWrapper}>{instanceChoices}</div> : <div className={cls.FieldsWrapper}>
                <Input label="Host" inputValue={fields.host} onChange={(value) => update("host", value)}/>
                <Input label="Port" inputValue={fields.port} onChange={(value) => update("port", value)}/>
                <Input label="Username" inputValue={fields.user} onChange={(value) => update("user", value)}/>
                <Input label="Password" inputValue={fields.password} onChange={(value) => update("password", value)}/>
                {isSource && <Input label="Database" inputValue={fields.dbName} onChange={(value) => update("dbName", value)}/>} 
            </div>}
        </section>
    }

    function handleSuccess() {
        CloseDialog()
    }

    if (submittedRequest) {
        return <div className={cls.PgDatabaseReplicateDialogContainer}>
            <DialogShell title="Replicating database" onClose={CloseDialog} isFlush>
                <TaskProgressScreen
                    title="Replicating database"
                    metaLine={submittedRequest.sourceDbName ?? "database"}
                    start={() => pgaasService.replicatePgDatabase(submittedRequest)}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            </DialogShell>
        </div>
    }

    return <div className={cls.PgDatabaseReplicateDialogContainer}>
        <DialogShell title="Replicate database" onClose={CloseDialog}>
            <p className={cls.Warning}>Репликация — снапшот на момент старта. Записи во время копирования не попадут в реплику.</p>
            {renderEndpoint("Source", sourceMode, source, updateSource, setSourceMode, true)}
            {renderEndpoint("Target", targetMode, target, updateTarget, setTargetMode, false)}
            <div className={cls.ActionsWrapper}>
                <Button variant="secondary" onClick={CloseDialog}>Cancel</Button>
                <Button variant="primary" onClick={handleReplicate}>Replicate</Button>
            </div>
        </DialogShell>
    </div>
}
