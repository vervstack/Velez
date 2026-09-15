import {useState} from "react"

import cls from "@/dialogs/RegistryInstanceCreateDialog/RegistryInstanceCreateDialog.module.css"
import {CreateRegistryInstanceRequest, CreateRegistryInstanceResponse} from "@/app/api/velez"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {queryClient} from "@/app/queryClient.ts"
import {ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts"
import {useListServicesQuery} from "@/processes/queries/services.ts"
import {
    CreateRegistryInstanceMutation,
    REGISTRY_INSTANCES_QUERY_KEY,
} from "@/processes/queries/registry_instances.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import Choice from "@/components/base/Choice.tsx"
import Checkbox from "@/components/base/Checkbox.tsx"
import LabeledSelect from "@/dialogs/RegistryInstanceCreateDialog/components/LabeledSelect/LabeledSelect.tsx"
import {
    buildCreateRegistryInstanceRequest,
} from "@/dialogs/RegistryInstanceCreateDialog/processes/buildCreateRegistryInstanceRequest.ts"
import RegistryDeployProgressScreen from "@/dialogs/RegistryInstanceCreateDialog/screens/RegistryDeployProgressScreen.tsx"

const BOX_OPTIONS = ["small", "medium", "large"] as const

type Box = typeof BOX_OPTIONS[number]

export default function RegistryInstanceCreateDialog() {
    const [name, setName] = useState("")
    const [environment, setEnvironment] = useState("")
    const [box, setBox] = useState<Box>("small")
    const [exposePort, setExposePort] = useState(false)
    const [port, setPort] = useState("")
    const [ownerService, setOwnerService] = useState("")
    const [enableUi, setEnableUi] = useState(false)
    const [submittedReq, setSubmittedReq] = useState<CreateRegistryInstanceRequest | null>(null)

    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const environmentsQuery = ListEnvironmentsQuery()
    const servicesQuery = useListServicesQuery()
    const createRegistryInstance = CreateRegistryInstanceMutation()

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

    function handleToggleEnableUi() {
        setEnableUi(!enableUi)
    }

    function handleCreate() {
        const req = buildCreateRegistryInstanceRequest({
            name, environment, box, exposePort, port, ownerService, enableUi,
        })
        if (!req) return

        setSubmittedReq(req)
    }

    function handleStart(): Promise<CreateRegistryInstanceResponse> {
        if (!submittedReq) {
            return Promise.reject(new Error("no pending create request"))
        }
        return createRegistryInstance.mutateAsync(submittedReq)
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: REGISTRY_INSTANCES_QUERY_KEY})
        toaster.bake({title: "Registry created", description: "", level: "Info"})
    }

    if (submittedReq) {
        return (
            <div className={cls.RegistryInstanceCreateDialogContainer}>
                <RegistryDeployProgressScreen
                    name={submittedReq.name ?? ""}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            </div>
        )
    }

    return (
        <div className={cls.RegistryInstanceCreateDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Create registry</h2>
            </div>

            <div className={cls.Content}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={createRegistryInstance.isPending}
                    />

                    <LabeledSelect
                        label="Environment"
                        value={environment}
                        placeholder="Default"
                        options={environmentOptions}
                        onChange={setEnvironment}
                        disabled={createRegistryInstance.isPending}
                    />

                    <span className={cls.FieldLabel}>Box</span>
                    <div className={cls.ChoiceRow}>
                        {BOX_OPTIONS.map(renderBoxChoice)}
                    </div>

                    <Checkbox label="Expose port" checked={exposePort} onChange={handleToggleExposePort}/>

                    {exposePort && (
                        <Input label="Port" inputValue={port} onChange={setPort}/>
                    )}

                    <Checkbox label="Enable UI" checked={enableUi} onChange={handleToggleEnableUi}/>

                    <LabeledSelect
                        label="Owner service (optional)"
                        value={ownerService}
                        placeholder="None"
                        options={serviceOptions}
                        onChange={setOwnerService}
                        disabled={createRegistryInstance.isPending}
                    />
                </div>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={createRegistryInstance.isPending}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleCreate}
                        disabled={createRegistryInstance.isPending || !name.trim()}
                    >
                        Create
                    </Button>
                </div>
            </div>
        </div>
    )
}
