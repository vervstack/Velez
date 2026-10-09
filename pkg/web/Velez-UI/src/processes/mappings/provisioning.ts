import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"

const ACTIVE_REFETCH_INTERVAL_MS = 2000

export function isProvisioningActive(task: ProvisioningTask): boolean {
    return task.status === ProvisioningTaskStatus.PENDING || task.status === ProvisioningTaskStatus.RUNNING
}

export function isProvisioningFailed(task: ProvisioningTask): boolean {
    return task.status === ProvisioningTaskStatus.FAILED
}

export function provisioningRefetchInterval(tasks?: ProvisioningTask[]): number | false {
    return tasks?.some(isProvisioningActive) ? ACTIVE_REFETCH_INTERVAL_MS : false
}

export function currentStepName(task: ProvisioningTask): string {
    const jobs = task.jobs ?? []
    const current = jobs.find((job) => job.status === ProvisioningTaskStatus.RUNNING)
        ?? jobs.find((job) => job.status === ProvisioningTaskStatus.PENDING)
    return current?.name ?? ""
}

export function stepProgress(task: ProvisioningTask): { done: number, total: number } {
    const jobs = task.jobs ?? []
    return {
        done: jobs.filter((job) => job.status === ProvisioningTaskStatus.DONE).length,
        total: jobs.length,
    }
}

export function isInstanceProvisioning(instanceName: string, tasks: ProvisioningTask[], prefix = ""): boolean {
    return tasks.some((task) => {
        const entityId = task.entityId ?? ""
        return isProvisioningActive(task) && (entityId === instanceName || prefix + entityId === instanceName)
    })
}

const DROP_ACTION_PREFIX = "drop_"

export function isDropAction(action?: string): boolean {
    return (action ?? "").startsWith(DROP_ACTION_PREFIX)
}

export function provisioningVerb(task: ProvisioningTask): "Creating" | "Removing" {
    return isDropAction(task.action) ? "Removing" : "Creating"
}

export function provisioningTitle(task: ProvisioningTask, noun: string): string {
    return `${provisioningVerb(task)} ${noun}`
}

export function stripInstancePrefix(name: string, prefix: string): string {
    return prefix !== "" && name.startsWith(prefix) ? name.slice(prefix.length) : name
}

const REGISTER_CONTAINER_ACTION = "register_container"

export function provisioningName(task: ProvisioningTask, prefix = ""): string {
    const entityId = task.entityId ?? ""
    if (task.action === REGISTER_CONTAINER_ACTION) {
        const separator = entityId.lastIndexOf("/")
        return separator === -1 ? entityId : entityId.slice(0, separator)
    }
    return stripInstancePrefix(entityId, prefix)
}
