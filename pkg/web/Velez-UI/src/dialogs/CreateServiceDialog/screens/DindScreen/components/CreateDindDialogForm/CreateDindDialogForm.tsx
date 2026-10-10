import {useState} from "react"
import {Checkbox} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/screens/DindScreen/components/CreateDindDialogForm/CreateDindDialogForm.module.css"
import type {CreateDindRequest} from "@/app/api/velez/dind_api.pb"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import {buildCreateDindRequest} from "@/widgets/CreateDindForm/processes/buildCreateDindRequest.ts"

const SYSBOX_DESCRIPTION =
    "Recommended for public repositories. Runs the Docker daemon without --privileged. " +
    "Turn off only for fully trusted private repositories — the daemon then runs privileged."

interface Props {
    onSubmit(req: CreateDindRequest): void
    onCancel(): void
}

export default function CreateDindDialogForm({onSubmit, onCancel}: Props) {
    const [name, setName] = useState("")
    const [environment, setEnvironment] = useState("")
    const [isSysboxEnabled, setIsSysboxEnabled] = useState(true)

    const nameError = validateInstanceName(name)
    const req = buildCreateDindRequest({name, environment, isSysboxEnabled})

    function handleCreate() {
        if (!req) return

        onSubmit(req)
    }

    return (
        <div className={cls.CreateDindDialogFormContainer}>
            <div className={cls.FieldsWrapper}>
                <Input label="Name" inputValue={name} onChange={setName} error={nameError}/>
                <Input label="Environment (optional)" inputValue={environment} onChange={setEnvironment}/>
                <Checkbox label="Sysbox isolation" checked={isSysboxEnabled} onChange={setIsSysboxEnabled}/>
                <p className={cls.Description}>{SYSBOX_DESCRIPTION}</p>
            </div>

            <div className={cls.ActionsRow}>
                <Button variant="secondary" onClick={onCancel}>Cancel</Button>
                <Button variant="primary" onClick={handleCreate} disabled={!req}>Create</Button>
            </div>
        </div>
    )
}
