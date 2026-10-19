import {ChangeEvent, ReactNode, useEffect, useState} from "react"

import cls from "@/pages/service/widgets/GitlabRunnerSettings.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import {useListDindsQuery} from "@/processes/queries/dinds.ts"
import {
    GetRunnerConfigQuery,
    GetRunnerCredentialsQuery,
    RedeployRunnerMutation,
    ReregisterRunnerMutation,
    UpdateRunnerConfigMutation,
    useListRunnersQuery,
} from "@/processes/queries/runners.ts"
import AllowedPullPoliciesField from "@/pages/service/widgets/AllowedPullPoliciesField.tsx"
import DindField from "@/pages/service/widgets/DindField.tsx"
import LogLevelField from "@/pages/service/widgets/LogLevelField.tsx"
import PullPolicyField from "@/pages/service/widgets/PullPolicyField.tsx"
import {RunnerLogLevel, RunnerPullPolicy, UpdateRunnerConfigRequest} from "@/app/api/velez"
import {parseConcurrent} from "@/processes/parseConcurrent.ts"
import {arePolicyListsEqual, parseNonNegativeInt} from "@/processes/runnerSettings.ts"

interface Props {
    serviceName: string
}

interface ConfigValues {
    baseUrl: string
    dockerImage: string
    dindName: string
    dockerSocketAddress: string
    concurrent: string
    pullPolicy: RunnerPullPolicy[]
    allowedPullPolicies: RunnerPullPolicy[]
    checkInterval: string
    logLevel: RunnerLogLevel
    shutdownTimeout: string
}

const MASKED_TOKEN = "••••••••••••"
const EMPTY_CONFIG: ConfigValues = {
    baseUrl: "",
    dockerImage: "",
    dindName: "",
    dockerSocketAddress: "",
    concurrent: "1",
    pullPolicy: [],
    allowedPullPolicies: [],
    checkInterval: "0",
    logLevel: RunnerLogLevel.RUNNER_LOG_LEVEL_UNSPECIFIED,
    shutdownTimeout: "0",
}

interface SettingsFieldProps {
    label: string
    value: string
    placeholder: string
    isDisabled?: boolean
    onChange: (value: string) => void
}

function SettingsField({label, value, placeholder, isDisabled, onChange}: SettingsFieldProps) {
    function handleChange(e: ChangeEvent<HTMLInputElement>) {
        onChange(e.target.value)
    }

    return (
        <div className={cls.FieldRow}>
            <label className={cls.FieldLabel}>{label}</label>
            <input
                className={cls.FieldInput}
                value={value}
                placeholder={placeholder}
                disabled={isDisabled}
                onChange={handleChange}
            />
        </div>
    )
}

interface ApplyBannerProps {
    children: ReactNode
    actionLabel: string
    onApply: () => void
    isApplying: boolean
}

function ApplyBanner({children, actionLabel, onApply, isApplying}: ApplyBannerProps) {
    return (
        <div className={cls.ApplyBanner}>
            <span className={cls.ApplyBannerText}>{children}</span>
            <Button sm onClick={onApply} disabled={isApplying}>
                {isApplying ? "Applying…" : actionLabel}
            </Button>
        </div>
    )
}

export default function GitlabRunnerSettings({serviceName}: Props) {
    const toaster = useToaster()

    const configQuery = GetRunnerConfigQuery(serviceName)
    const credentialsQuery = GetRunnerCredentialsQuery(serviceName)
    const updateConfig = UpdateRunnerConfigMutation()
    const reregisterRunner = ReregisterRunnerMutation()
    const redeployRunner = RedeployRunnerMutation()
    const dindsQuery = useListDindsQuery()
    const runnersQuery = useListRunnersQuery()

    const [values, setValues] = useState<ConfigValues>(EMPTY_CONFIG)
    const [savedValues, setSavedValues] = useState<ConfigValues | null>(null)
    const [pendingReregister, setPendingReregister] = useState(false)
    const [pendingRedeploy, setPendingRedeploy] = useState(false)
    const [revealed, setRevealed] = useState(false)

    useEffect(function syncLoadedConfig() {
        if (!configQuery.data) return

        const loaded: ConfigValues = {
            baseUrl: configQuery.data.baseUrl ?? "",
            dockerImage: configQuery.data.dockerImage ?? "",
            dindName: configQuery.data.dindName ?? "",
            dockerSocketAddress: configQuery.data.dockerSocketAddress ?? "",
            concurrent: String(configQuery.data.concurrent || 1),
            pullPolicy: configQuery.data.pullPolicy ?? [],
            allowedPullPolicies: configQuery.data.allowedPullPolicies ?? [],
            checkInterval: String(configQuery.data.checkInterval ?? 0),
            logLevel: configQuery.data.logLevel ?? RunnerLogLevel.RUNNER_LOG_LEVEL_UNSPECIFIED,
            shutdownTimeout: String(configQuery.data.shutdownTimeout ?? 0),
        }

        setValues(loaded)
        setSavedValues(loaded)
    }, [configQuery.data])

    function handleBaseUrlChange(value: string) {
        setValues((prev) => ({...prev, baseUrl: value}))
    }

    function handleDockerImageChange(value: string) {
        setValues((prev) => ({...prev, dockerImage: value}))
    }

    function handleDindNameChange(value: string) {
        setValues((prev) => ({
            ...prev,
            dindName: value,
            dockerSocketAddress: value ? "" : prev.dockerSocketAddress,
        }))
    }

    function handleDockerSocketAddressChange(value: string) {
        setValues((prev) => ({...prev, dockerSocketAddress: value}))
    }

    function handleConcurrentChange(value: string) {
        setValues((prev) => ({...prev, concurrent: value}))
    }

    function handlePullPolicyChange(value: RunnerPullPolicy[]) {
        setValues((prev) => ({...prev, pullPolicy: value}))
    }

    function handleAllowedPullPoliciesChange(value: RunnerPullPolicy[]) {
        setValues((prev) => ({...prev, allowedPullPolicies: value}))
    }

    function handleCheckIntervalChange(value: string) {
        setValues((prev) => ({...prev, checkInterval: value}))
    }

    function handleLogLevelChange(value: RunnerLogLevel) {
        setValues((prev) => ({...prev, logLevel: value}))
    }

    function handleShutdownTimeoutChange(value: string) {
        setValues((prev) => ({...prev, shutdownTimeout: value}))
    }

    function handleSave() {
        if (!savedValues) return

        const parsedConcurrent = parseConcurrent(values.concurrent)
        const parsedCheckInterval = parseNonNegativeInt(values.checkInterval)
        const parsedShutdownTimeout = parseNonNegativeInt(values.shutdownTimeout)
        if (parsedConcurrent === undefined
            || parsedCheckInterval === undefined
            || parsedShutdownTimeout === undefined) return

        const req: UpdateRunnerConfigRequest = {name: serviceName}
        let changed = false

        if (values.baseUrl !== savedValues.baseUrl) {
            req.baseUrl = values.baseUrl
            changed = true
        }
        if (values.dockerImage !== savedValues.dockerImage) {
            req.dockerImage = values.dockerImage
            changed = true
        }
        if (values.dindName !== savedValues.dindName) {
            req.dindName = values.dindName
            changed = true
        }
        if (!values.dindName && values.dockerSocketAddress !== savedValues.dockerSocketAddress) {
            req.dockerSocketAddress = values.dockerSocketAddress
            changed = true
        }
        if (values.concurrent !== savedValues.concurrent) {
            req.concurrent = parsedConcurrent
            changed = true
        }
        if (!arePolicyListsEqual(values.pullPolicy, savedValues.pullPolicy)) {
            req.pullPolicy = {values: values.pullPolicy}
            changed = true
        }
        if (!arePolicyListsEqual(values.allowedPullPolicies, savedValues.allowedPullPolicies)) {
            req.allowedPullPolicies = {values: values.allowedPullPolicies}
            changed = true
        }
        if (values.checkInterval !== savedValues.checkInterval) {
            req.checkInterval = parsedCheckInterval
            changed = true
        }
        if (values.logLevel !== savedValues.logLevel) {
            req.logLevel = values.logLevel
            changed = true
        }
        if (values.shutdownTimeout !== savedValues.shutdownTimeout) {
            req.shutdownTimeout = parsedShutdownTimeout
            changed = true
        }

        if (!changed) return

        updateConfig.mutate(req, {
            onSuccess: (resp) => {
                setSavedValues(values)
                if (resp.requiresReregister) setPendingReregister(true)
                if (resp.requiresRedeploy) setPendingRedeploy(true)
                toaster.bake({title: "Settings saved", description: serviceName, level: "Info"})
            },
            onError: toaster.catchGrpc,
        })
    }

    function handleReregister() {
        reregisterRunner.mutate(serviceName, {
            onSuccess: () => {
                setPendingReregister(false)
                toaster.bake({title: "Runner reregistration started", description: serviceName, level: "Info"})
            },
            onError: toaster.catchGrpc,
        })
    }

    function handleRedeploy() {
        redeployRunner.mutate(serviceName, {
            onSuccess: () => {
                setPendingRedeploy(false)
                toaster.bake({title: "Runner redeploy started", description: serviceName, level: "Info"})
            },
            onError: toaster.catchGrpc,
        })
    }

    function handleCredentialsUnavailable() {
        toaster.bake({
            title: "Token unavailable",
            description: `${serviceName}'s credentials were lost, likely after a server restart. `
                + "Recreate the runner to get a new token.",
            level: "Error",
        })
    }

    function handleToggleReveal() {
        if (revealed) {
            setRevealed(false)
            return
        }
        setRevealed(true)
        if (!credentialsQuery.data) {
            credentialsQuery.refetch().then(function onRefetched(result) {
                if (result.isError) {
                    handleCredentialsUnavailable()
                }
            })
        }
    }

    function copyToClipboard(text: string) {
        navigator.clipboard.writeText(text)
            .then(function onCopied() {
                toaster.bake({title: "Token copied", description: serviceName, level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function handleCopyToken() {
        if (credentialsQuery.data?.token) {
            copyToClipboard(credentialsQuery.data.token)
            return
        }

        credentialsQuery.refetch().then(function onFetched(result) {
            if (result.data?.token) {
                copyToClipboard(result.data.token)
                return
            }
            handleCredentialsUnavailable()
        })
    }

    if (configQuery.isLoading) {
        return (
            <div className={cls.GitlabRunnerSettingsContainer}>
                <SkeletonLoader shape="block" height="10rem"/>
            </div>
        )
    }

    if (configQuery.isError) {
        return (
            <div className={cls.GitlabRunnerSettingsContainer}>
                <QueryErrorState message="Failed to load GitLab runner settings." onRetry={configQuery.refetch}/>
            </div>
        )
    }

    const tokenValue = revealed && credentialsQuery.data?.token
        ? credentialsQuery.data.token
        : MASKED_TOKEN
    const dindNames = (dindsQuery.data?.dinds ?? []).map((dind) => dind.name ?? "")
    const isBuildkitEnabled = runnersQuery.data?.runners?.find((r) => r.name === serviceName)
        ?.isBuildkitEnabled ?? false
    const concurrentParsed = parseConcurrent(values.concurrent)
    const isDirty = savedValues !== null && (
        values.baseUrl !== savedValues.baseUrl
        || values.dockerImage !== savedValues.dockerImage
        || values.dindName !== savedValues.dindName
        || values.dockerSocketAddress !== savedValues.dockerSocketAddress
        || values.concurrent !== savedValues.concurrent
        || !arePolicyListsEqual(values.pullPolicy, savedValues.pullPolicy)
        || !arePolicyListsEqual(values.allowedPullPolicies, savedValues.allowedPullPolicies)
        || values.checkInterval !== savedValues.checkInterval
        || values.logLevel !== savedValues.logLevel
        || values.shutdownTimeout !== savedValues.shutdownTimeout
    )
    const isInvalid = concurrentParsed === undefined
        || parseNonNegativeInt(values.checkInterval) === undefined
        || parseNonNegativeInt(values.shutdownTimeout) === undefined

    return (
        <div className={cls.GitlabRunnerSettingsContainer}>
            <div className={cls.HeaderRow}>
                <span className={cls.Title}>GitLab Runner Settings</span>
            </div>

            {pendingReregister && (
                <ApplyBanner
                    actionLabel="Rerun registration"
                    onApply={handleReregister}
                    isApplying={reregisterRunner.isPending}
                >
                    Base URL / docker image changed — rerun registration to apply.
                </ApplyBanner>
            )}

            {pendingRedeploy && (
                <ApplyBanner
                    actionLabel="Redeploy"
                    onApply={handleRedeploy}
                    isApplying={redeployRunner.isPending}
                >
                    Docker socket address or Docker daemon changed — redeploy the container to apply.
                </ApplyBanner>
            )}

            <div className={cls.FieldsWrapper}>
                <SettingsField
                    label="Base URL"
                    value={values.baseUrl}
                    placeholder="https://gitlab.com"
                    onChange={handleBaseUrlChange}
                />
                <SettingsField
                    label="Docker image"
                    value={values.dockerImage}
                    placeholder="alpine:3.24.2"
                    onChange={handleDockerImageChange}
                />
                <DindField
                    value={values.dindName}
                    dindNames={dindNames}
                    isLoading={dindsQuery.isLoading}
                    isBuildkitEnabled={isBuildkitEnabled}
                    onChange={handleDindNameChange}
                />
                <SettingsField
                    label="Docker socket"
                    value={values.dockerSocketAddress}
                    placeholder="tcp://host:2375 (default: host socket)"
                    isDisabled={Boolean(values.dindName)}
                    onChange={handleDockerSocketAddressChange}
                />
                <SettingsField
                    label="Concurrent jobs"
                    value={values.concurrent}
                    placeholder="1"
                    onChange={handleConcurrentChange}
                />
                <PullPolicyField value={values.pullPolicy} onChange={handlePullPolicyChange}/>
                <AllowedPullPoliciesField
                    value={values.allowedPullPolicies}
                    onChange={handleAllowedPullPoliciesChange}
                />
                <SettingsField
                    label="Check interval (s)"
                    value={values.checkInterval}
                    placeholder="0 (default)"
                    onChange={handleCheckIntervalChange}
                />
                <LogLevelField value={values.logLevel} onChange={handleLogLevelChange}/>
                <SettingsField
                    label="Shutdown (s)"
                    value={values.shutdownTimeout}
                    placeholder="0 (default)"
                    onChange={handleShutdownTimeoutChange}
                />

                <div className={cls.TokenRow}>
                    <label className={cls.FieldLabel}>Token</label>
                    <span className={cls.TokenValue}>
                        {credentialsQuery.isFetching ? "loading…" : tokenValue}
                    </span>
                    <Button sm onClick={handleToggleReveal}>
                        {revealed ? "Hide" : "Reveal"}
                    </Button>
                    <Button sm onClick={handleCopyToken}>
                        Copy
                    </Button>
                </div>
            </div>

            <div className={cls.ActionsRow}>
                <Button
                    onClick={handleSave}
                    disabled={!isDirty || isInvalid || updateConfig.isPending}
                >
                    {updateConfig.isPending ? "Saving…" : "Save"}
                </Button>
            </div>
        </div>
    )
}
