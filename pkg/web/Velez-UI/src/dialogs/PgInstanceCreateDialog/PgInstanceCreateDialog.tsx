import {useState} from "react"

import cls from "@/dialogs/PgInstanceCreateDialog/PgInstanceCreateDialog.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts"
import {useListServicesQuery} from "@/processes/queries/services.ts"
import {CreatePgInstanceMutation} from "@/processes/queries/pg_instances.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import Choice from "@/components/base/Choice.tsx"
import Checkbox from "@/components/base/Checkbox.tsx"
import LabeledSelect from "@/dialogs/PgInstanceCreateDialog/components/LabeledSelect/LabeledSelect.tsx"
import {buildCreatePgInstanceRequest} from "@/dialogs/PgInstanceCreateDialog/processes/buildCreatePgInstanceRequest.ts"

const BOX_OPTIONS = ["small", "medium", "large"] as const

type Box = typeof BOX_OPTIONS[number]

const ISOLATION_OPTIONS: Array<{ id: string, title: string, sub?: string, disabled?: boolean }> = [
    {id: "separate_instance", title: "Separate instance"},
    {id: "shared_pool", title: "Shared pool", sub: "Coming soon", disabled: true},
]

function handleIsolationNoop() {
}

function renderIsolationChoice(opt: typeof ISOLATION_OPTIONS[number]) {
    return (
        <Choice
            key={opt.id}
            title={opt.title}
            sub={opt.sub}
            active={opt.id === "separate_instance"}
            disabled={opt.disabled}
            onClick={handleIsolationNoop}
        />
    )
}

export default function PgInstanceCreateDialog() {
    const [name, setName] = useState("")
    const [environment, setEnvironment] = useState("")
    const [box, setBox] = useState<Box>("small")
    const [exposePort, setExposePort] = useState(false)
    const [port, setPort] = useState("")
    const [ownerService, setOwnerService] = useState("")

    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const environmentsQuery = ListEnvironmentsQuery()
    const servicesQuery = useListServicesQuery()
    const createPgInstance = CreatePgInstanceMutation()

    const environmentOptions = (environmentsQuery.data?.environments ?? []).map((env) => ({
        value: env.name ?? "",
        label: env.name ?? "",
    }))
    const serviceOptions = (servicesQuery.data?.services ?? []).map((s) => ({
        value: s.name ?? "",
        label: s.name ?? "",
    }))

    function renderBoxChoice(option: Box) {
        function handleClick() {
            setBox(option)
        }

        return <Choice key={option} title={option} active={box === option} onClick={handleClick}/>
    }

    function handleToggleExposePort() {
        setExposePort(!exposePort)
    }

    function handleCreate() {
        const req = buildCreatePgInstanceRequest({name, environment, box, exposePort, port, ownerService})
        if (!req) return

        createPgInstance.mutate(req, {
            onSuccess: () => {
                toaster.bake({title: "Database created", description: name.trim(), level: "Info"})
                CloseDialog()
            },
            onError: toaster.catchGrpc,
        })
    }

    return (
        <div className={cls.PgInstanceCreateDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Create database</h2>
            </div>

            <div className={cls.Content}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={createPgInstance.isPending}
                    />

                    <LabeledSelect
                        label="Environment"
                        value={environment}
                        placeholder="Default"
                        options={environmentOptions}
                        onChange={setEnvironment}
                        disabled={createPgInstance.isPending}
                    />

                    <span className={cls.FieldLabel}>Box</span>
                    <div className={cls.ChoiceRow}>
                        {BOX_OPTIONS.map(renderBoxChoice)}
                    </div>

                    <span className={cls.FieldLabel}>Isolation</span>
                    <div className={cls.ChoiceRow}>
                        {ISOLATION_OPTIONS.map(renderIsolationChoice)}
                    </div>

                    <Checkbox label="Expose port" checked={exposePort} onChange={handleToggleExposePort}/>

                    {exposePort && (
                        <Input label="Port" inputValue={port} onChange={setPort}/>
                    )}

                    <LabeledSelect
                        label="Owner service (optional)"
                        value={ownerService}
                        placeholder="None"
                        options={serviceOptions}
                        onChange={setOwnerService}
                        disabled={createPgInstance.isPending}
                    />
                </div>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={createPgInstance.isPending}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleCreate}
                        disabled={createPgInstance.isPending || !name.trim()}
                    >
                        {createPgInstance.isPending ? "Creating…" : "Create"}
                    </Button>
                </div>
            </div>
        </div>
    )
}
