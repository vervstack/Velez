import {useState} from "react"
import {Dropdown, DropdownOption} from "@vervstack/chures"
import cn from "classnames"

import cls from "@/dialogs/RunnerCreateDialog/RunnerCreateDialog.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {CreateRunnerRequest, CreateRunnerResponse, RunnerProvider, RunnerScope} from "@/app/api/velez"
import {queryClient} from "@/app/queryClient.ts"
import {CreateRunnerMutation, RUNNERS_QUERY_KEY} from "@/processes/queries/runners.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"
import {buildCreateRunnerRequest} from "@/dialogs/RunnerCreateDialog/processes/buildCreateRunnerRequest.ts"

const PROVIDER_OPTIONS: DropdownOption[] = [
    {id: RunnerProvider.GITHUB, name: "GitHub"},
    {id: RunnerProvider.GITLAB, name: "GitLab"},
]

const SCOPE_OPTIONS: DropdownOption[] = [
    {id: RunnerScope.REPO, name: "Repository"},
    {id: RunnerScope.ORG, name: "Organization"},
]

export default function RunnerCreateDialog() {
    const [name, setName] = useState("")
    const [provider, setProvider] = useState<RunnerProvider>(RunnerProvider.GITHUB)
    const [scope, setScope] = useState<RunnerScope>(RunnerScope.REPO)
    const [target, setTarget] = useState("")
    const [labels, setLabels] = useState("")
    const [environment, setEnvironment] = useState("")
    const [accessToken, setAccessToken] = useState("")
    const [baseUrl, setBaseUrl] = useState("")
    const [dockerSocketAddress, setDockerSocketAddress] = useState("")
    const [showAdvanced, setShowAdvanced] = useState(false)
    const [targetTouched, setTargetTouched] = useState(false)
    const [submittedReq, setSubmittedReq] = useState<CreateRunnerRequest | null>(null)

    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const createRunner = CreateRunnerMutation()

    const isGitlab = provider === RunnerProvider.GITLAB

    function handleProviderChange(ids: string[]) {
        setProvider((ids[0] as RunnerProvider) ?? RunnerProvider.GITHUB)
    }

    function handleScopeChange(ids: string[]) {
        setScope((ids[0] as RunnerScope) ?? RunnerScope.REPO)
    }

    function handleTargetChange(v: string) {
        setTargetTouched(true)
        setTarget(v)
    }

    function handleToggleAdvanced() {
        setShowAdvanced(!showAdvanced)
    }

    function handleCreate() {
        const req = buildCreateRunnerRequest({
            name,
            provider,
            scope,
            target,
            labels,
            environment,
            accessToken,
            baseUrl,
            dockerSocketAddress,
        })
        if (!req) return

        setSubmittedReq(req)
    }

    function handleStart(): Promise<CreateRunnerResponse> {
        if (!submittedReq) {
            return Promise.reject(new Error("no pending create request"))
        }
        return createRunner.mutateAsync(submittedReq)
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: RUNNERS_QUERY_KEY})
        toaster.bake({title: "Runner created", description: name.trim(), level: "Info"})
    }

    const isFormValid = name.trim() && scope && target.trim() && accessToken.trim()
    const showTargetError = targetTouched && !target.trim()

    if (submittedReq) {
        return (
            <div className={cls.RunnerCreateDialogContainer}>
                <TaskProgressScreen
                    title="Creating runner"
                    metaLine={name.trim()}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            </div>
        )
    }

    return (
        <div className={cls.RunnerCreateDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Create runner</h2>
            </div>

            <div className={cls.Content}>
                <div className={cls.FieldsWrapper}>
                    <span className={cls.FieldLabel}>Required</span>

                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={createRunner.isPending}
                    />

                    <Dropdown
                        label="Provider"
                        placeholder="Select provider"
                        options={PROVIDER_OPTIONS}
                        value={[provider]}
                        onChange={handleProviderChange}
                        portal
                    />

                    <Dropdown
                        label="Scope"
                        placeholder="Select scope"
                        options={SCOPE_OPTIONS}
                        value={[scope]}
                        onChange={handleScopeChange}
                        portal
                    />

                    <div className={cls.TargetFieldWrapper}>
                        <Input
                            label="Target"
                            inputValue={target}
                            onChange={handleTargetChange}
                            disabled={createRunner.isPending}
                        />
                        {showTargetError && (
                            <span className={cls.FieldError}>Target is required</span>
                        )}
                    </div>

                    <Input
                        label={isGitlab ? "GitLab Runner Token" : "GitHub Access Token"}
                        inputValue={accessToken}
                        onChange={setAccessToken}
                        disabled={createRunner.isPending}
                    />

                    {isGitlab && (
                        <Input
                            label="GitLab Base URL (optional, defaults to gitlab.com)"
                            inputValue={baseUrl}
                            onChange={setBaseUrl}
                            disabled={createRunner.isPending}
                        />
                    )}

                    <div
                        className={cls.AdvancedToggle}
                        onClick={handleToggleAdvanced}
                    >
                        {showAdvanced ? "▼" : "▶"} Optional
                    </div>

                    <div className={cn(cls.AdvancedSection, showAdvanced && cls.AdvancedSectionOpen)}>
                        <Input
                            label="Docker Socket Address (optional)"
                            inputValue={dockerSocketAddress}
                            onChange={setDockerSocketAddress}
                            disabled={createRunner.isPending}
                        />

                        <div className={cls.RiskNotice}>
                            Empty uses the default host Docker socket — full root access to the host.
                            Only override this if you understand the risk.
                        </div>

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
                    </div>
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
