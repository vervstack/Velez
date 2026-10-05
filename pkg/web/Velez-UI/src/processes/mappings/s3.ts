import type {S3BucketAccess, S3Instance} from "@/app/api/velez/s3_api.pb"
import {getLinkToPort} from "@/model/services/VervPlugins.tsx"

export type S3InstanceStatus = "running" | "degraded" | "stopped"

// Mirrors labels.S3NamePrefix in internal/domain/labels/verv_labels.go -
// the single source of truth this string must match.
const S3_NAME_PREFIX = "s3_"

const BYTE_UNITS = ["B", "KiB", "MiB", "GiB", "TiB"]
const BYTES_PER_UNIT = 1024
const ROUNDING_THRESHOLD = 10

export function mapS3InstanceStatus(status?: string): S3InstanceStatus {
    const s = (status ?? "").toLowerCase()
    if (s.includes("run")) return "running"
    if (s.includes("degrad") || s.includes("restart")) return "degraded"
    return "stopped"
}

export function s3ServiceName(instanceName: string): string {
    return S3_NAME_PREFIX + instanceName
}

export function sortS3InstancesByName(instances: S3Instance[]): S3Instance[] {
    return [...instances].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""))
}

export function formatBytes(bytes?: string | number): string {
    let value = Number(bytes ?? 0)
    if (!Number.isFinite(value) || value <= 0) return "0 B"

    let unit = 0
    while (value >= BYTES_PER_UNIT && unit < BYTE_UNITS.length - 1) {
        value /= BYTES_PER_UNIT
        unit++
    }

    const rounded = unit === 0 || value >= ROUNDING_THRESHOLD ? Math.round(value) : Math.round(value * 10) / 10
    return `${rounded} ${BYTE_UNITS[unit]}`
}

export function formatS3Date(ts?: { seconds?: string | number }): string {
    const seconds = Number(ts?.seconds ?? 0)
    if (!seconds) return "-"
    return new Date(seconds * 1000).toLocaleDateString()
}

export function formatAccessFlags(access: S3BucketAccess): string {
    const flags = [
        access.isRead ? "read" : "",
        access.isWrite ? "write" : "",
        access.isOwner ? "owner" : "",
    ].filter(Boolean)
    return flags.length > 0 ? flags.join("+") : "none"
}

export function formatKeyBuckets(access: S3BucketAccess[]): string {
    if (access.length === 0) return "-"
    return access.map((entry) => `${entry.bucketName} (${formatAccessFlags(entry)})`).join(", ")
}

export function s3WebUiLink(instance: S3Instance): string | undefined {
    if (!instance.webUiPort) return undefined
    return getLinkToPort(instance.webUiPort)
}
