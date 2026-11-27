import {Button, Input} from "@vervstack/chures"
import {useState} from "react"

import {TaskStatusStatus, type TaskStatus} from "@/app/api/velez/tasks.pb"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {haproxyService} from "@/processes/api/haproxy.ts"
import {WatchTaskStream} from "@/processes/api/tasks.ts"
import ProgressStepChip from "@/widgets/TaskProgressScreen/components/ProgressStepChip/ProgressStepChip.tsx"
import {type HaproxyWizardContext, parsePort} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

const JOBS = [
    {name: "prepare_image", label: "Pull image"},
    {name: "create_container", label: "Create container"},
    {name: "connect_tailnet", label: "Connect to tailnet"},
    {name: "wait_tailnet_healthy", label: "Healthy"},
]

export default function VcnSetupScreen(props: StepScreenProps<HaproxyWizardContext>) {
    const {sidecarContainerId, isBootstrapping, updateContext, goNext} = props
    const [loginServerUrl, setLoginServerUrl] = useState("")
    const [authKey, setAuthKey] = useState("")
    const [hostname, setHostname] = useState("")
    const [status, setStatus] = useState<TaskStatus>()
    const [error, setError] = useState("")
    const isReady = Boolean(sidecarContainerId)
    const isFormDisabled = isBootstrapping || isReady
    const canSetup = Boolean(loginServerUrl.trim() && authKey.trim())
    const firstPending = status?.jobs?.find((job) => job.status === TaskStatusStatus.PENDING)
    const hasRunningJob = status?.jobs?.some((job) => job.status === TaskStatusStatus.RUNNING)

    async function setupSidecar() {
        if (isReady) {
            goNext?.()
            return
        }

        updateContext({isBootstrapping: true})
        setError("")
        setStatus(undefined)

        try {
            const publishedPort = props.isPortExposed ? parsePort(props.port ?? "") : undefined
            const {entityId, action} = await haproxyService.bootstrapVcnSidecar({
                name: props.name.trim(),
                loginServerUrl: loginServerUrl.trim(),
                authKey: authKey.trim(),
                hostname: hostname.trim(),
                isAsync: true,
            }, publishedPort)
            if (!entityId || !action) throw new Error("The VCN bootstrap task could not be found.")

            let finalStatus: TaskStatus | undefined
            await WatchTaskStream({entityId, action}, (update) => {
                finalStatus = update
                setStatus(update)
            })

            if (finalStatus?.status !== TaskStatusStatus.DONE || !finalStatus.containerId) {
                throw new Error(finalStatus?.error || "VCN setup ended without a running sidecar.")
            }

            updateContext({sidecarContainerId: finalStatus.containerId})
            setAuthKey("")
        } catch (failure) {
            setError(failure instanceof Error ? failure.message : String(failure))
        } finally {
            updateContext({isBootstrapping: false})
        }
    }

    function chipStatus(name: string) {
        if (isReady) return TaskStatusStatus.DONE

        const job = status?.jobs?.find((item) => item.name === name)
        if (isBootstrapping && !hasRunningJob && firstPending?.name === name &&
            status?.status === TaskStatusStatus.RUNNING) {
            return TaskStatusStatus.RUNNING
        }
        return job?.status
    }

    return (
        <div className={cls.Fields}>
            {!isReady && (
                <>
                    <Input label="Login server URL" value={loginServerUrl} setValue={setLoginServerUrl}
                           disabled={isFormDisabled}/>
                    <Input label="Auth key" value={authKey} setValue={setAuthKey} type="password"
                           autoComplete="off" disabled={isFormDisabled}/>
                    <Input label="Hostname (optional)" value={hostname} setValue={setHostname}
                           disabled={isFormDisabled}/>
                </>
            )}
            <div className={cls.Actions}>
                <Button variant="primary" type="button" disabled={isBootstrapping || (!isReady && !canSetup)}
                        onClick={setupSidecar}>{isReady ? "Next" : "Setup VCN sidecar"}</Button>
                <Button variant="unstyled" type="button" className={cls.Link}
                        disabled={isFormDisabled} onClick={goNext}>
                    Proceed without VCN settings
                </Button>
            </div>
            {(isBootstrapping || status || isReady) && (
                <div className={cls.Progress} aria-label="VCN setup progress">
                    {JOBS.map((job, index) => (
                        <ProgressStepChip key={job.name} name={job.name} label={job.label}
                                          index={index + 1} status={chipStatus(job.name)}/>
                    ))}
                </div>
            )}
            {error && <div role="alert" className={cls.Unreachable}>{error}</div>}
        </div>
    )
}
