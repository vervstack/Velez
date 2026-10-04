import {useState} from "react"
import {Toggle} from "@vervstack/chures"

import cls from "@/widgets/CreateDindForm/CreateDindForm.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateDindMutation} from "@/processes/queries/dinds.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import {buildCreateDindRequest} from "@/widgets/CreateDindForm/processes/buildCreateDindRequest.ts"

const SYSBOX_DESCRIPTION =
    "Recommended for public repositories. Runs the Docker daemon without --privileged. " +
    "Turn off only for fully trusted private repositories — the daemon then runs privileged."

interface Props {
    onCreated(name: string): void
    onCancel(): void
}

export default function CreateDindForm({onCreated, onCancel}: Props) {
    const [name, setName] = useState("")
    const [environment, setEnvironment] = useState("")
    const [isSysboxEnabled, setIsSysboxEnabled] = useState(true)

    const toaster = useToaster()
    const createDind = CreateDindMutation()

    const req = buildCreateDindRequest({name, environment, isSysboxEnabled})

    function handleCreate() {
        if (!req) return

        createDind.mutateAsync(req)
            .then(function handleCreated() {
                toaster.bake({title: "Docker daemon created", description: req.name ?? "", level: "Info"})
                onCreated(req.name ?? "")
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.CreateDindFormContainer}>
            <div className={cls.FieldsWrapper}>
                <Input label="Name" inputValue={name} onChange={setName} disabled={createDind.isPending}/>
                <Input
                    label="Environment (optional)"
                    inputValue={environment}
                    onChange={setEnvironment}
                    disabled={createDind.isPending}
                />
                <Toggle
                    label="Sysbox isolation"
                    checked={isSysboxEnabled}
                    onChange={setIsSysboxEnabled}
                    disabled={createDind.isPending}
                />
                <p className={cls.Description}>{SYSBOX_DESCRIPTION}</p>
            </div>

            <div className={cls.ActionsRow}>
                <Button variant="secondary" onClick={onCancel} disabled={createDind.isPending}>
                    Cancel
                </Button>
                <Button variant="primary" onClick={handleCreate} disabled={createDind.isPending || !req}>
                    {createDind.isPending ? "Creating…" : "Create"}
                </Button>
            </div>
        </div>
    )
}
