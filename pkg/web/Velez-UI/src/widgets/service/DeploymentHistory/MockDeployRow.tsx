import type { MockDeploy } from '@/widgets/service/DeploymentHistory/mockDeploys'
import cls from '@/widgets/service/DeploymentHistory/DeployRow.module.css'

interface MockDeployRowProps {
    deploy: MockDeploy
    isLast: boolean
    prevDeployId?: number
}

function statusDotClass(status: MockDeploy['status']): string {
    if (status === 'failed') return cls.dotFailed
    if (status === 'rolled') return cls.dotUnknown
    return cls.dotRunning
}

function initials(who: string): string {
    return who.split('.').map(function firstChar(p) { return p[0] }).join('').toUpperCase()
}

export default function MockDeployRow({ deploy: d, isLast, prevDeployId }: MockDeployRowProps) {
    return (
        <div className={`${cls.DeployRowContainer} ${isLast ? cls.lastRow : ''}`}>
            <div className={cls.IdCell}>
                <span className={`${cls.StatusDot} ${statusDotClass(d.status)}`} />
                <span className={cls.IdText}>#{d.id}</span>
            </div>

            <div className={cls.VersionCell}>
                <span className={cls.VersionText}>{d.version}</span>
                {d.current && <span className={cls.LiveBadge}>live · {d.env}</span>}
            </div>

            <div className={cls.Cell}>{d.date}</div>

            <div className={cls.WhoCell}>
                <span className={cls.Avatar}>{initials(d.who)}</span>
                <span>{d.who}</span>
            </div>

            <div className={cls.GitCell}>
                <span className={cls.DigestChip}>{d.git}</span>
                <span className={cls.BranchLabel}>⎇ {d.branch}</span>
                {d.tag && <span className={cls.TagChip}>{d.tag}</span>}
            </div>

            <div className={cls.Cell}>
                <span className={cls.EnvPill}>{d.env}</span>
            </div>

            <div className={cls.ActionsCell}>
                <button
                    className={cls.ActionBtn}
                    disabled={!prevDeployId}
                    title={prevDeployId ? `compare with #${prevDeployId}` : 'no previous version'}
                >
                    ⇄ Diff
                </button>
                <button
                    className={`${cls.ActionBtn} ${cls.rollbackBtn}`}
                    disabled={!!d.current}
                    title={d.current ? 'already live' : `rollback to ${d.version}`}
                >
                    ↺ Rollback
                </button>
            </div>
        </div>
    )
}
