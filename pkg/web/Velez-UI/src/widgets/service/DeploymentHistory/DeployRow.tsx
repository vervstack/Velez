import { DeploymentInfo, DeploymentStatus } from '@/app/api/velez'
import cls from '@/widgets/service/DeploymentHistory/DeployRow.module.css'

interface DeployRowProps {
    deployment: DeploymentInfo
    isCurrent: boolean
    isLast: boolean
    prevDeployId?: string
}

function formatTimestamp(ts: { seconds?: string | number; nanos?: number } | undefined): string {
    const seconds = Number(ts?.seconds || 0)
    if (!seconds) return '—'
    return new Date(seconds * 1000).toLocaleString()
}

function getStatusClass(status: DeploymentStatus | undefined): string {
    switch (status) {
        case DeploymentStatus.RUNNING:
            return cls.dotRunning
        case DeploymentStatus.FAILED:
        case DeploymentStatus.DELETED:
            return cls.dotFailed
        default:
            return cls.dotUnknown
    }
}

function parseImageTag(image: string | undefined): string {
    if (!image) return '—'
    const colonIdx = image.lastIndexOf(':')
    if (colonIdx === -1 || colonIdx === image.length - 1) return image
    return image.slice(colonIdx + 1)
}

export default function DeployRow({ deployment: dep, isCurrent }: DeployRowProps) {
    const shortId = dep.id ? dep.id.slice(0, 8) : '—'
    const version = parseImageTag(dep.image) || (dep.specId ? dep.specId.slice(0, 8) : '—')
    const deployedAt = formatTimestamp(dep.createdAt as { seconds?: string | number; nanos?: number } | undefined)
    const triggeredBy = dep.triggeredBy || '—'
    // TODO: replace with git commit hash when GetServiceGraph adds git field
    const digest = dep.imageDigest ? dep.imageDigest.slice(0, 12) : '—'
    const statusClass = getStatusClass(dep.status)

    return (
        <div className={cls.DeployRowContainer}>
            <div className={cls.IdCell}>
                <span className={`${cls.StatusDot} ${statusClass}`} />
                <span className={cls.IdText}>#{shortId}</span>
            </div>

            <div className={cls.VersionCell}>
                <span className={cls.VersionText}>{version}</span>
                {isCurrent && <span className={cls.LiveBadge}>live</span>}
            </div>

            <div className={cls.Cell}>{deployedAt}</div>

            <div className={cls.Cell}>{triggeredBy}</div>

            <div className={cls.DigestCell}>
                {dep.imageDigest
                    ? <span className={cls.DigestChip}>{digest}</span>
                    : '—'
                }
            </div>

            <div className={cls.Cell}>
                {/* TODO: DeploymentInfo needs env field — see api/grpc/service_api.proto */}
                —
            </div>

            <div className={cls.ActionsCell}>
                <button className={cls.ActionBtn} disabled>⇄ Diff</button>
                <button className={cls.ActionBtn} disabled={isCurrent}>↺ Rollback</button>
            </div>
        </div>
    )
}
