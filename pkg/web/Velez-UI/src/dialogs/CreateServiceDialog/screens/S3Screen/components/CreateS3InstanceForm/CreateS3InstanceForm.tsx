import {useState} from "react"
import {Checkbox, Dropdown, DropdownOption, parseGrpcError} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/screens/S3Screen/components/CreateS3InstanceForm/CreateS3InstanceForm.module.css"
import type {CreateS3InstanceRequest} from "@/app/api/velez/s3_api.pb"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import {ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts"
import Button from "@/components/base/Button.tsx"
import Choice from "@/components/base/Choice.tsx"
import Input from "@/components/base/Input.tsx"
import {
    buildCreateS3InstanceRequest,
    LOCKED_REPLICATION_FACTOR,
} from "@/dialogs/CreateServiceDialog/screens/S3Screen/processes/buildCreateS3InstanceRequest.ts"

const BOX_OPTIONS = ["small", "medium", "large"] as const

const REPLICATION_HINT = "Locked to 1 — multi-node Garage clusters come later."

interface Props {
    isHidden: boolean

    onSubmit(req: CreateS3InstanceRequest): void
    onCancel(): void
}

export default function CreateS3InstanceForm({isHidden, onSubmit, onCancel}: Props) {
    const [name, setName] = useState("")
    const [environment, setEnvironment] = useState("")
    const [box, setBox] = useState<string>("small")
    const [isPortExposed, setIsPortExposed] = useState(false)
    const [port, setPort] = useState("")
    const [region, setRegion] = useState("")
    const [isWebUiEnabled, setIsWebUiEnabled] = useState(false)

    const toaster = useToaster()
    const environmentsQuery = ListEnvironmentsQuery()

    const environmentOptions: DropdownOption[] = (environmentsQuery.data?.environments ?? []).map((env) => ({
        id: env.name ?? "",
        name: env.name ?? "",
    }))

    const nameError = validateInstanceName(name)
    const req = buildCreateS3InstanceRequest({name, environment, box, isPortExposed, port, region, isWebUiEnabled})

    function handleError(err: unknown) {
        toaster.catchGrpc(parseGrpcError(err))
    }

    function handleEnvironmentChange(ids: string[]) {
        setEnvironment(ids[0] ?? "")
    }

    function handleSubmit() {
        if (req) onSubmit(req)
    }

    function renderBoxChoice(option: string) {
        function handleClick() {
            setBox(option)
        }

        return <Choice key={option} title={option} active={box === option} onClick={handleClick}/>
    }

    return (
        <div className={cls.CreateS3InstanceFormContainer} hidden={isHidden}>
            <div className={cls.FieldsWrapper}>
                <Input label="Name" inputValue={name} onChange={setName} error={nameError}/>

                <Dropdown
                    label="Environment"
                    placeholder="Default"
                    options={environmentOptions}
                    value={environment ? [environment] : []}
                    onChange={handleEnvironmentChange}
                    isLoading={environmentsQuery.isLoading}
                    onError={handleError}
                    portal
                />

                <span className={cls.FieldLabel}>Box</span>
                <div className={cls.ChoiceRow}>
                    {BOX_OPTIONS.map(renderBoxChoice)}
                </div>

                <Checkbox label="Expose port" checked={isPortExposed} onChange={setIsPortExposed}/>
                {isPortExposed && <Input label="Port" inputValue={port} onChange={setPort}/>}

                <Input label="Region" inputValue={region} onChange={setRegion} placeholder="garage"/>

                <Input
                    label="Replication factor"
                    inputValue={String(LOCKED_REPLICATION_FACTOR)}
                    disabled
                />
                <p className={cls.Description}>{REPLICATION_HINT}</p>

                <Checkbox label="Enable web UI" checked={isWebUiEnabled} onChange={setIsWebUiEnabled}/>
                <p className={cls.Description}>Deploys the garage-webui sidecar next to the instance.</p>
            </div>

            <div className={cls.ActionsRow}>
                <Button variant="secondary" onClick={onCancel}>Cancel</Button>
                <Button variant="primary" onClick={handleSubmit} disabled={!req}>Create</Button>
            </div>
        </div>
    )
}
