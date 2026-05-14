import type { ServiceResource } from '@/model/service_page/ServicePageModel'

import cls from './ResourceCard.module.css'

interface ResourceCardProps {
    resource: ServiceResource
}

export default function ResourceCard({ resource }: ResourceCardProps) {
    const statusClass = resource.status === 'healthy'
        ? cls.healthy
        : resource.status === 'degraded'
            ? cls.degraded
            : cls.unhealthy

    return (
        <div
            className={cls.ResourceCardContainer}
            style={{ '--resource-color': resource.color } as React.CSSProperties}
        >
            <div className={cls.HeaderWrapper}>
                <div className={cls.IconSquare}>{resource.icon}</div>
                <div className={cls.NameGroup}>
                    <span className={cls.ResourceName}>{resource.id}</span>
                    <span className={cls.KindDesc}>{resource.kind} · {resource.desc}</span>
                </div>
                <span className={`${cls.StatusDot} ${statusClass}`} />
            </div>
            <p className={cls.HostLine}>{resource.host}</p>
            <div className={cls.MetaRow}>
                <span className={cls.MetaItem}><span>use</span>{resource.use}</span>
                <span className={cls.MetaItem}><span>hits</span>{resource.hits}</span>
            </div>
        </div>
    )
}
