import cn from 'classnames'

import type {ServiceResource} from '@/model/service_page/ServicePageModel'

import cls from './ResourceCard.module.css'

interface ResourceCardProps {
    resource: ServiceResource
}

export default function ResourceCard({resource}: ResourceCardProps) {
    const reconciliationLabel = reconciliationBadgeLabel(resource.reconciliation)
    const reconciliationClass = reconciliationBadgeClass(resource.reconciliation)

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
                    {reconciliationLabel && (
                        <span className={cn(cls.ReconciliationBadge, reconciliationClass)}>
                            {reconciliationLabel}
                        </span>
                    )}
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

function reconciliationBadgeLabel(status: ServiceResource['reconciliation']): string {
    switch (status) {
        case 'already_connected':
            return 'Existing connection'
        case 'must_provision':
            return 'Not yet created'
        default:
            return ''
    }
}

function reconciliationBadgeClass(status: ServiceResource['reconciliation']): string {
    switch (status) {
        case 'already_connected':
            return cls.alreadyConnected
        case 'must_provision':
            return cls.mustProvision
        default:
            return ''
    }
}
