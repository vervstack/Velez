import cn from 'classnames'

import type {ServiceResource} from '@/model/service_page/ServicePageModel'

import cls from './ResourceCard.module.css'

interface ResourceCardProps {
    resource: ServiceResource
}

export default function ResourceCard({resource}: ResourceCardProps) {
    return (
        <div
            className={cls.ResourceCardContainer}
            style={{'--resource-color': resource.color} as React.CSSProperties}
        >
            <div className={cls.HeaderWrapper}>
                <div className={cls.IconSquare}>{resource.icon}</div>
                <div className={cls.NameGroup}>
                    <span className={cls.ResourceName}>{resource.name}</span>
                    <span className={cls.KindDesc}>{resource.type}</span>
                </div>
                <span className={cn(cls.StatusDot, statusDotClass(resource.status))}/>
            </div>
        </div>
    )
}

function statusDotClass(status: ServiceResource['status']): string {
    switch (status) {
        case 'healthy':
            return cls.healthy
        case 'degraded':
            return cls.degraded
        case 'unhealthy':
            return cls.unhealthy
        default:
            return cls.unknown
    }
}
