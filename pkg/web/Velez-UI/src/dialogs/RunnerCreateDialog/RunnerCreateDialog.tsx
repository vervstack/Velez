import {useState} from "react"

import cls from "@/dialogs/RunnerCreateDialog/RunnerCreateDialog.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {RunnerScope} from "@/app/api/velez"
import {CreateRunnerMutation} from "@/processes/queries/runners.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import LabeledSelect from "@/dialogs/RunnerCreateDialog/components/LabeledSelect/LabeledSelect.tsx"
import {buildCreateRunnerRequest} from "@/dialogs/RunnerCreateDialog/processes/buildCreateRunnerRequest.ts"

const SCOPE_OPTIONS: Array<{value: RunnerScope, label: string}> = [
    {value: RunnerScope.REPO, label: "Repository"},
    {value: RunnerScope.ORG, label: "Organization"},
]

export default function RunnerCreateDialog() {
    const [name, setName] = useState("")
    const [scope, setScope] = useState<RunnerScope>(RunnerScope.REPO)
    const [target, setTarget] = useState("")
    const [labels, setLabels] = useState("")
    const [environment, setEnvironment] = useState("")
    const [accessToken, setAccessToken] = useState("")
    const [dockerSocketAddress, setDockerSocketAddress] = useState("")
    const [showAdvanced, setShowAdvanced] = useState(false)

    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const createRunner = CreateRunnerMutation()

    const scopeOptions = SCOPE_OPTIONS.map(opt => ({
        value: opt.value,
        label: opt.label,
    }))

    function handleCreate() {
        const req = buildCreateRunnerRequest({
            name,
            scope,
            target,
            labels,
            environment,
            accessToken,
            dockerSocketAddress,
        })
        if (!req) return

        createRunner.mutate(req, {
            onSuccess: () => {
                toaster.bake({title: "Runner created", description: name.trim(), level: "Info"})
                CloseDialog()
            },
            onError: toaster.catchGrpc,
        })
    }

    const isFormValid = name.trim() && scope && target.trim() && accessToken.trim()

    return (
        <div className={cls.RunnerCreateDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Create runner</h2>
            </div>

            <div className={cls.Content}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={createRunner.isPending}
                    />

                    <LabeledSelect
                        label="Scope"
                        value={scope}
                        placeholder="Select scope"
                        options={scopeOptions}
                        onChange={(val) => setScope(val as RunnerScope)}
                        disabled={createRunner.isPending}
                    />

                    <Input
                        label="Target"
                        inputValue={target}
                        onChange={setTarget}
                        disabled={createRunner.isPending}
                    />

                    <Input
                        label="Labels (comma-separated)"
                        inputValue={labels}
                        onChange={setLabels}
                        disabled={createRunner.isPending}
                    />

                    <Input
                        label="Environment (optional)"
                        inputValue={environment}
                        onChange={setEnvironment}
                        disabled={createRunner.isPending}
                    />

                    <Input
                        label="GitHub Access Token"
                        inputValue={accessToken}
                        onChange={setAccessToken}
                        disabled={createRunner.isPending}
                    />

                    <div
                        className={cls.AdvancedToggle}
                        onClick={() => setShowAdvanced(!showAdvanced)}
                    >
                        {showAdvanced ? "▼" : "▶"} Advanced
                    </div>

                    {showAdvanced && (
                        <Input
                            label="Docker Socket Address (optional)"
                            inputValue={dockerSocketAddress}
                            onChange={setDockerSocketAddress}
                            disabled={createRunner.isPending}
                        />
                    )}
                </div>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={createRunner.isPending}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleCreate}
                        disabled={createRunner.isPending || !isFormValid}
                    >
                        {createRunner.isPending ? "Creating…" : "Create"}
                    </Button>
                </div>
            </div>
        </div>
    )
}
