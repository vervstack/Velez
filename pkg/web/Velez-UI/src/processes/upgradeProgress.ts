import {TaskStatus, TaskStatusJobStatus, TaskStatusStatus} from "@/app/api/velez"

export interface UpgradeProgress {
    isInFlight: boolean
    percent: number
    stepLabel: string
    isFailed: boolean
}

const STEP_LABELS: Record<string, string> = {
    check_self_upgrade: "Checking upgrade",
    capture_old_container: "Capturing current container",
    discover_sidecars: "Discovering sidecars",
    prepare_image: "Pulling image",
    pause_old_container: "Pausing current container",
    create_config_fetcher_container: "Starting config reader",
    get_config_from_container: "Reading config",
    drop_config_fetcher_container: "Removing config reader",
    fetch_config: "Fetching config",
    prepare_verv_config: "Preparing config",
    create_final_container: "Creating new container",
    start_final_container: "Starting new container",
    healthcheck: "Waiting for healthcheck",
    sync_runner_proxy: "Syncing runner proxy",
    rename_old_container: "Retiring old container",
    drop_old_container: "Removing old container",
    rename_new_container: "Promoting new container",
    recreate_sidecars: "Recreating sidecars",
    sync_addresses: "Syncing addresses",
}

const QUEUED_LABEL = "Queued"
const APPLIED_LABEL = "Applied"

const NOT_STARTED: UpgradeProgress = {isInFlight: false, percent: 0, stepLabel: "", isFailed: false}

export function jobLabel(name: string): string {
    const known = STEP_LABELS[name]
    if (known) {
        return known
    }

    const spaced = name.replace(/_/g, " ").trim()
    return spaced.charAt(0).toUpperCase() + spaced.slice(1)
}

function currentJob(jobs: TaskStatusJobStatus[]): TaskStatusJobStatus {
    const running = jobs.find((j) => j.status === TaskStatusStatus.RUNNING)
    if (running) {
        return running
    }

    const pending = jobs.find((j) => j.status === TaskStatusStatus.PENDING)
    return pending ?? jobs[jobs.length - 1]
}

function percentDone(jobs: TaskStatusJobStatus[]): number {
    if (jobs.length === 0) {
        return 0
    }

    const done = jobs.filter((j) => j.status === TaskStatusStatus.DONE).length
    return Math.round((100 * done) / jobs.length)
}

export function upgradeProgress(task: TaskStatus | undefined): UpgradeProgress {
    if (!task) {
        return NOT_STARTED
    }

    if (task.status === TaskStatusStatus.DONE) {
        return {isInFlight: false, percent: 100, stepLabel: APPLIED_LABEL, isFailed: false}
    }

    const jobs = task.jobs ?? []

    if (task.status === TaskStatusStatus.FAILED) {
        const stepLabel = task.error ? `Failed: ${task.error}` : "Failed"
        return {isInFlight: false, percent: percentDone(jobs), stepLabel, isFailed: true}
    }

    const isInFlight = task.status === TaskStatusStatus.PENDING || task.status === TaskStatusStatus.RUNNING
    if (!isInFlight) {
        return NOT_STARTED
    }

    if (jobs.length === 0) {
        return {isInFlight, percent: 0, stepLabel: QUEUED_LABEL, isFailed: false}
    }

    const job = currentJob(jobs)
    return {isInFlight, percent: percentDone(jobs), stepLabel: jobLabel(job.name ?? ""), isFailed: false}
}
