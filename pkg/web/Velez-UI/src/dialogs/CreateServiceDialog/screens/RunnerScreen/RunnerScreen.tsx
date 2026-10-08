import {useEffect, useState} from "react"
import {Checkbox, Dropdown, DropdownOption} from "@vervstack/chures"
import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/RunnerScreen.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {CreateRunnerRequest, CreateRunnerResponse, RunnerProvider, RunnerScope} from "@/app/api/velez"
import {queryClient} from "@/app/queryClient.ts"
import {CreateRunnerMutation, RUNNERS_QUERY_KEY} from "@/processes/queries/runners.ts"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import {parseConcurrent} from "@/processes/parseConcurrent.ts"
import {BUILDKIT_CREATE_TOOLTIP, BUILDKIT_UNSUPPORTED_TOOLTIP} from "@/processes/buildkitControl.ts"
import Button from "@/components/base/Button.tsx"
import InfoMark from "@/components/base/InfoMark.tsx"
import Input from "@/components/base/Input.tsx"
import CreateDindForm from "@/widgets/CreateDindForm/CreateDindForm.tsx"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"
import DockerDaemonPicker
    from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/components/DockerDaemonPicker/DockerDaemonPicker.tsx"
import {isBuildkitSupported, resolveDockerTarget} from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/processes/dockerTarget.ts"
import {
    buildCreateRunnerRequest,
} from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/processes/buildCreateRunnerRequest.ts"

const PROVIDER_OPTIONS: DropdownOption[] = [
    {id: RunnerProvider.GITHUB, name: "GitHub"},
    {id: RunnerProvider.GITLAB, name: "GitLab"},
]

const SCOPE_OPTIONS: DropdownOption[] = [
    {id: RunnerScope.REPO, name: "Repository"},
    {id: RunnerScope.ORG, name: "Organization"},
]

interface Props {
    initialProvider?: RunnerProvider
    onBusyChange(isBusy: boolean): void
}

export default function RunnerScreen({initialProvider = RunnerProvider.GITHUB, onBusyChange}: Props) {
    const [name, setName] = useState("")
    const [provider, setProvider] = useState<RunnerProvider>(initialProvider)
    const [scope, setScope] = useState<RunnerScope>(RunnerScope.REPO)
    const [target, setTarget] = useState("")
    const [labels, setLabels] = useState("")
    const [environment, setEnvironment] = useState("")
    const [accessToken, setAccessToken] = useState("")
    const [baseUrl, setBaseUrl] = useState("")
    const [dockerImage, setDockerImage] = useState("")
    const [dockerChoice, setDockerChoice] = useState("")
    const [externalDockerAddress, setExternalDockerAddress] = useState("")
    const [isBuildkitChecked, setIsBuildkitChecked] = useState(false)
    const [isCreatingDind, setIsCreatingDind] = useState(false)
    const [concurrent, setConcurrent] = useState("")
    const [showAdvanced, setShowAdvanced] = useState(false)
    const [targetTouched, setTargetTouched] = useState(false)
    const [submittedReq, setSubmittedReq] = useState<CreateRunnerRequest | null>(null)

    const toaster = useToaster()
    const {CloseDialog} = useDialog()

    const createRunner = CreateRunnerMutation()

    useEffect(() => {
        onBusyChange(submittedReq !== null || createRunner.isPending)
    }, [submittedReq, createRunner.isPending])

    const isGitlab = provider === RunnerProvider.GITLAB
    const nameError = validateInstanceName(name)
    const concurrentParsed = concurrent ? parseConcurrent(concurrent) : undefined
    const isConcurrentInvalid = concurrent !== "" && concurrentParsed === undefined
    const dockerTarget = resolveDockerTarget(dockerChoice, externalDockerAddress)
    const isBuildkitAvailable = isBuildkitSupported(dockerChoice)
    const isDockerTargetChosen = Boolean(dockerTarget.dindName || dockerTarget.dockerSocketAddress)

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

    function handleStartCreatingDind() {
        setIsCreatingDind(true)
    }

    function handleCancelCreatingDind() {
        setIsCreatingDind(false)
    }

    function handleDindCreated(dindName: string) {
        setDockerChoice(dindName)
        setIsCreatingDind(false)
    }

    function handleToggleAdvanced() {
        setShowAdvanced(!showAdvanced)
    }

    function handleCreate() {
        const req = buildCreateRunnerRequest({
            ...dockerTarget,
            name,
            provider,
            scope,
            target,
            labels,
            environment,
            accessToken,
            baseUrl,
            dockerImage,
            concurrent,
            isBuildkitEnabled: isBuildkitChecked && isBuildkitAvailable,
        })
        if (!req) return

        setSubmittedReq(req)
    }

    function handleBack() {
        setSubmittedReq(null)
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

    const isFormValid = name.trim() && !nameError && scope && target.trim() && accessToken.trim()
        && isDockerTargetChosen && !isConcurrentInvalid
    const showTargetError = targetTouched && !target.trim()

    function renderTargetField() {
        return (
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
        )
    }

    function renderOptionalFields() {
        return (
            <div className={cn(cls.AdvancedSection, showAdvanced && cls.AdvancedSectionOpen)}>
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
        )
    }

    if (isCreatingDind) {
        return <CreateDindForm onCreated={handleDindCreated} onCancel={handleCancelCreatingDind}/>
    }

    return (
        <>
            <div className={cls.RunnerScreenContainer} hidden={submittedReq !== null}>
                <div className={cls.FieldsWrapper}>
                    <span className={cls.FieldLabel}>Required</span>

                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={createRunner.isPending}
                        error={nameError}
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

                    {renderTargetField()}

                    <Input
                        label={isGitlab ? "GitLab Personal Access Token (create_runner scope)" :"GitHub Access Token"}
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

                    {isGitlab && (
                        <Input
                            label="Docker Image (optional, defaults to alpine:3.24.2)"
                            inputValue={dockerImage}
                            onChange={setDockerImage}
                            disabled={createRunner.isPending}
                        />
                    )}

                    {isGitlab && (
                        <Input label="Executor" inputValue="docker"/>
                    )}

                    {isGitlab && (
                        <Input
                            label="Concurrent jobs (optional, defaults to 1)"
                            inputValue={concurrent}
                            onChange={setConcurrent}
                            disabled={createRunner.isPending}
                        />
                    )}

                    <DockerDaemonPicker
                        choice={dockerChoice}
                        externalAddress={externalDockerAddress}
                        isDisabled={createRunner.isPending}
                        onChoiceChange={setDockerChoice}
                        onExternalAddressChange={setExternalDockerAddress}
                        onCreateNew={handleStartCreatingDind}
                    />

                    <div className={cls.BuildkitWrapper}>
                        <Checkbox
                            label="Provision with BuildKit"
                            checked={isBuildkitChecked && isBuildkitAvailable}
                            onChange={setIsBuildkitChecked}
                            disabled={!isBuildkitAvailable || createRunner.isPending}
                        />
                        <InfoMark
                            tooltip={isBuildkitAvailable ? BUILDKIT_CREATE_TOOLTIP : BUILDKIT_UNSUPPORTED_TOOLTIP}
                        />
                    </div>

                    <div
                        className={cls.AdvancedToggle}
                        onClick={handleToggleAdvanced}
                    >
                        {showAdvanced ? "▼" : "▶"} Optional
                    </div>

                    {renderOptionalFields()}
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
            {submittedReq && (
                <TaskProgressScreen
                    title="Creating runner"
                    metaLine={name.trim()}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                    onBack={handleBack}
                />
            )}
        </>
    )
}
