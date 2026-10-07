import cn from 'classnames'

import {useDialog} from '@/app/hooks/dialog/Dialog.tsx'
import type {ServiceResource} from '@/model/service_page/ServicePageModel'
import {resourceWebUiUrl} from '@/processes/resourceWebUi'
import ResourceAddressesDialog from '@/dialogs/ResourceAddressesDialog/ResourceAddressesDialog.tsx'

import cls from './ResourceCard.module.css'

interface ResourceCardProps {
    resource: ServiceResource
}

export default function ResourceCard({resource}: ResourceCardProps) {
    const reconciliationLabel = reconciliationBadgeLabel(resource.reconciliation)
    const reconciliationClass = reconciliationBadgeClass(resource.reconciliation)
    const {OpenDialog} = useDialog()
    const hasAddresses = resource.addresses.length > 0
    const webUiUrl = hasAddresses ? undefined : resourceWebUiUrl(resource, window.location)
    const isClickable = hasAddresses || Boolean(webUiUrl)

    function handleOpen() {
        if (hasAddresses) {
            OpenDialog(<ResourceAddressesDialog resourceName={resource.name} addresses={resource.addresses}/>)
            return
        }
        if (webUiUrl) {
            window.open(webUiUrl, "_blank", "noopener,noreferrer")
        }
    }

    function handleKeyDown(e: React.KeyboardEvent) {
        if (isClickable && e.key === "Enter") {
            handleOpen()
        }
    }

    return (
        <div
            className={cn(cls.ResourceCardContainer, isClickable && cls.Clickable)}
            style={{'--resource-color': resource.color} as React.CSSProperties}
            onClick={handleOpen}
            onKeyDown={handleKeyDown}
            role={isClickable ? "link" : undefined}
            tabIndex={isClickable ? 0 : undefined}
            title={isClickable ? "Open web UI" : undefined}
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
