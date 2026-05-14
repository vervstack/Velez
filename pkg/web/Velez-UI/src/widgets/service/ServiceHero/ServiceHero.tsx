import { useQuery } from '@tanstack/react-query'

import cls from '@/widgets/service/ServiceHero/ServiceHero.module.css'

import { FetchServiceMetrics } from '@/processes/api/service'
import type { ServiceMetrics } from '@/model/service_page/ServicePageModel'

interface ServiceHeroProps {
    serviceId: string
    serviceName: string
    serviceStatus?: string
    imageFromSmerd?: string
}

function statusDotClass(status?: string): string {
    if (!status || status === 'stopped' || status === 'failed') return cls.dotDisabled
    if (status === 'degraded') return cls.dotDegraded
    return cls.dotRunning
}

function progressBarFillClass(pct: number): string {
    if (pct > 80) return cls.fillDanger
    if (pct > 60) return cls.fillWarning
    return cls.fillNormal
}

interface MetricTileProps {
    label: string
    value: string
    sub: string
    barPct?: number
}

function MetricTile({ label, value, sub, barPct }: MetricTileProps) {
    return (
        <div className={cls.MetricTile}>
            <span className={cls.TileLabel}>{label}</span>
            <span className={cls.TileValue}>{value}</span>
            {barPct !== undefined && (
                <div className={cls.ProgressBar}>
                    <div
                        className={`${cls.ProgressFill} ${progressBarFillClass(barPct)}`}
                        style={{ width: `${Math.min(barPct, 100)}%` }}
                    />
                </div>
            )}
            <span className={cls.TileSub}>{sub}</span>
        </div>
    )
}

export default function ServiceHero({
    serviceId,
    serviceName,
    serviceStatus,
    imageFromSmerd,
}: ServiceHeroProps) {
    const { data: metrics } = useQuery<ServiceMetrics>({
        queryKey: ['service-metrics', serviceId],
        queryFn: function fetchMetrics() {
            return FetchServiceMetrics(serviceId)
        },
    })

    const replicas = metrics?.replicas ?? '—'
    const uptime = metrics?.uptime ?? '—'
    const cpu = metrics?.cpu ?? 0
    const mem = metrics?.mem ?? 0
    const memMax = metrics?.memMax ?? 0

    const cpuLabel = cpu > 0 ? `${cpu.toFixed(1)}%` : '—'
    const memLabel = mem > 0 ? `${mem} MiB` : '—'
    const memSub = memMax > 0 ? `of ${memMax} MiB` : 'no limit'
    const cpuPct = cpu
    const memPct = memMax > 0 ? (mem / memMax) * 100 : 0

    const description = metrics?.description || 'No description available.'
    const serviceType = metrics?.type || '—'
    const team = metrics?.team || '—'
    const image = metrics?.image || imageFromSmerd || ''
    const port = metrics?.port || ''
    const repo = metrics?.repo || ''

    return (
        <div className={cls.ServiceHeroContainer}>
            <div className={cls.LeftWrapper}>
                <div className={cls.NameRow}>
                    <span className={`${cls.StatusDot} ${statusDotClass(serviceStatus)}`} />
                    <h1 className={cls.ServiceName}>{serviceName}</h1>
                    <span className={cls.TagCyan}>{serviceType}</span>
                    <span className={cls.Tag}>team · {team}</span>
                </div>

                <p className={cls.Description}>{description}</p>

                <div className={cls.MetaTagsRow}>
                    {image && <><span className={cls.MetaLabel}>image</span><code className={cls.CodeTag}>{image}</code></>}
                    {port && <><span className={cls.MetaLabel}>port</span><code className={cls.CodeTag}>{port}</code></>}
                    {repo && <><span className={cls.MetaLabel}>repo</span><code className={cls.CodeTag}>{repo}</code></>}
                </div>
            </div>

            <div className={cls.RightWrapper}>
                <div className={cls.MetricsGrid}>
                    <MetricTile label="Replicas" value={replicas} sub="desired / running" />
                    <MetricTile label="Uptime" value={uptime} sub="since last deploy" />
                    <MetricTile label="CPU" value={cpuLabel} sub="avg across pods" barPct={cpuPct} />
                    <MetricTile label="Memory" value={memLabel} sub={memSub} barPct={memPct} />
                </div>
            </div>
        </div>
    )
}
