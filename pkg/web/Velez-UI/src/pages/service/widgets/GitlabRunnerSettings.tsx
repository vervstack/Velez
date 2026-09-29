import {ChangeEvent, ReactNode, useEffect, useState} from "react"

import cls from "@/pages/service/widgets/GitlabRunnerSettings.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import {
    GetRunnerConfigQuery,
    GetRunnerCredentialsQuery,
    RedeployRunnerMutation,
    ReregisterRunnerMutation,
    UpdateRunnerConfigMutation,
} from "@/processes/queries/runners.ts"
import {UpdateRunnerConfigRequest} from "@/app/api/velez"

interface Props {
    serviceName: string
}

interface ConfigValues {
    baseUrl: string
    dockerImage: string
    dockerSocketAddress: string
}

const MASKED_TOKEN = "••••••••••••"
const EMPTY_CONFIG: ConfigValues = {baseUrl: "", dockerImage: "", dockerSocketAddress: ""}

interface SettingsFieldProps {
    label: string
    value: string
    placeholder: string
    onChange: (value: string) => void
}

function SettingsField({label, value, placeholder, onChange}: SettingsFieldProps) {
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
            dockerSocketAddress: configQuery.data.dockerSocketAddress ?? "",
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

    function handleDockerSocketAddressChange(value: string) {
        setValues((prev) => ({...prev, dockerSocketAddress: value}))
    }

    function handleSave() {
        if (!savedValues) return

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
        if (values.dockerSocketAddress !== savedValues.dockerSocketAddress) {
            req.dockerSocketAddress = values.dockerSocketAddress
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
    const isDirty = savedValues !== null && (
        values.baseUrl !== savedValues.baseUrl
        || values.dockerImage !== savedValues.dockerImage
        || values.dockerSocketAddress !== savedValues.dockerSocketAddress
    )

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
                    Docker socket address changed — redeploy the container to apply.
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
                    placeholder="alpine:latest"
                    onChange={handleDockerImageChange}
                />
                <SettingsField
                    label="Docker socket"
                    value={values.dockerSocketAddress}
                    placeholder="tcp://host:2375 (default: host socket)"
                    onChange={handleDockerSocketAddressChange}
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
                <Button onClick={handleSave} disabled={!isDirty || updateConfig.isPending}>
                    {updateConfig.isPending ? "Saving…" : "Save"}
                </Button>
            </div>
        </div>
    )
}
