import {AutoUpdateBumpLevel} from "@/app/api/velez/auto_update_api.pb"
import type {ListAutoUpdateServicesResponseService} from "@/app/api/velez/auto_update_api.pb"
import {TaskStatusStatus} from "@/app/api/velez"
import type {TaskStatus} from "@/app/api/velez"

export const AUTO_UPDATE_ACTION = "upgrade_smerd"

export const DEFAULT_BUMP_LEVEL = AutoUpdateBumpLevel.PATCH

export interface BumpLevelOption {
    id: AutoUpdateBumpLevel
    name: string
    hint: string
}

export const BUMP_LEVEL_OPTIONS: BumpLevelOption[] = [
    {id: AutoUpdateBumpLevel.MAJOR, name: "Major only", hint: "Only major releases, e.g. 1.0.2 to 2.0.0"},
    {id: AutoUpdateBumpLevel.MINOR, name: "Major + minor", hint: "Major and minor releases, e.g. 1.0.2 to 1.1.0"},
    {
        id: AutoUpdateBumpLevel.PATCH,
        name: "All versions (major, minor, patch)",
        hint: "Every newer release, e.g. 1.0.2 to 1.0.3",
    },
]

export function initialBumpLevel(level: AutoUpdateBumpLevel | undefined): AutoUpdateBumpLevel {
    if (!level || level === AutoUpdateBumpLevel.AUTO_UPDATE_BUMP_LEVEL_UNSPECIFIED) return DEFAULT_BUMP_LEVEL
    return level
}

export function bumpLevelName(level: AutoUpdateBumpLevel | undefined): string {
    const option = BUMP_LEVEL_OPTIONS.find((o) => o.id === initialBumpLevel(level))
    return option?.name ?? ""
}

type AutoUpdateService = ListAutoUpdateServicesResponseService

export function sortServices(services: AutoUpdateService[]): AutoUpdateService[] {
    return [...services].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""))
}

function pluralize(value: number, unit: string): string {
    return `${value} ${unit}${value === 1 ? "" : "s"}`
}

export function describeCheckPeriod(seconds: string | number | undefined): string {
    const total = Number(seconds)
    if (!Number.isFinite(total) || total <= 0) return "Check period is not configured"
    if (total % 3600 === 0) return `Checks every ${pluralize(total / 3600, "hour")}`
    if (total % 60 === 0) return `Checks every ${pluralize(total / 60, "minute")}`
    return `Checks every ${pluralize(total, "second")}`
}

export function confirmTitle(wasEnabled: boolean, isEnabled: boolean): string {
    if (wasEnabled === isEnabled) return "Change bump level?"
    return isEnabled ? "Enable auto-update?" : "Disable auto-update?"
}

export function restartWarning(name: string, isEnabled: boolean): string {
    if (isEnabled) {
        return `${name} restarts once to apply this change, and again whenever a newer allowed version is found.`
    }
    return `${name} restarts once to apply this change and is no longer updated automatically.`
}

export function isTaskFailed(task: TaskStatus | undefined): boolean {
    return task?.status === TaskStatusStatus.FAILED
}
