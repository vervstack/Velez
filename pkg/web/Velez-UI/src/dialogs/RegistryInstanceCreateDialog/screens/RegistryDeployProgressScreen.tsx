import {useEffect, useState} from "react"

import cls from "@/dialogs/RegistryInstanceCreateDialog/screens/RegistryDeployProgressScreen.module.css"
import {TaskStatus, TaskStatusStatus, WatchTaskRequest} from "@/app/api/velez"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import Button from "@/components/base/Button.tsx"
import {WatchTaskStream} from "@/processes/api/tasks.ts"
import {REGISTRY_INSTANCES_QUERY_KEY} from "@/processes/queries/registry_instances.ts"

interface WatchableTaskStart {
    entityId?: string
    action?: string
}

interface Props {
    name: string

    start(): Promise<WatchableTaskStart>

    onSuccess?(): void

    onClose(): void
}

type Phase = "starting" | "running" | "done" | "failed"

function humanizeJobName(name: string): string {
    return name
        .split("_")
        .filter(Boolean)
        .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
        .join(" ")
}

function currentJobLabel(taskStatus: TaskStatus | undefined): string {
    const runningJob = taskStatus?.jobs?.find((job) => job.status === TaskStatusStatus.RUNNING)
    if (runningJob?.name) {
        return humanizeJobName(runningJob.name)
    }
    return "Starting"
}

function statusText(phase: Phase, taskStatus: TaskStatus | undefined): string {
    if (phase === "done") {
        return "Registry deployed"
    }
    if (phase === "failed") {
        return "Deployment failed"
    }
    return `${currentJobLabel(taskStatus)}…`
}

export default function RegistryDeployProgressScreen({name, start, onSuccess, onClose}: Props) {
    const [phase, setPhase] = useState<Phase>("starting")
    const [error, setError] = useState<string | undefined>(undefined)
    const [taskStatus, setTaskStatus] = useState<TaskStatus | undefined>(undefined)

    useEffect(() => {
        let cancelled = false

        function onStatus(status: TaskStatus) {
            if (cancelled) {
                return
            }
            setTaskStatus(status)

            if (status.status === TaskStatusStatus.RUNNING) {
                setPhase("running")
            }
            if (status.status === TaskStatusStatus.DONE) {
                queryClient.invalidateQueries({queryKey: REGISTRY_INSTANCES_QUERY_KEY})
                onSuccess?.()
                setPhase("done")
            }
            if (status.status === TaskStatusStatus.FAILED) {
                setPhase("failed")
                setError(status.error || "Task failed")
            }
        }

        start()
            .then((res) => {
                const watchReq: WatchTaskRequest = {entityId: res.entityId, action: res.action}
                return WatchTaskStream(watchReq, onStatus)
            })
            .catch((err: Error) => {
                if (cancelled) {
                    return
                }
                setPhase("failed")
                setError(err.message)
                useToaster.getState().catchGrpc(err)
            })

        return () => {
            cancelled = true
        }
    }, [])

    const isActive = phase === "starting" || phase === "running"

    return (
        <div className={cls.RegistryDeployProgressScreenContainer}>
            {isActive && <span className={cls.Spinner}/>}
            {phase === "done" && <span className={cls.SuccessMark}>✓</span>}
            {phase === "failed" && <span className={cls.FailureMark}>!</span>}

            <span className={cls.StatusText}>{statusText(phase, taskStatus)}</span>
            {isActive && <span className={cls.NameText}>{name}</span>}
            {phase === "failed" && error && <span className={cls.ErrorText}>{error}</span>}

            {phase === "done" && <Button variant="primary" onClick={onClose}>Done</Button>}
            {phase === "failed" && <Button variant="secondary" onClick={onClose}>Close</Button>}
        </div>
    )
}
