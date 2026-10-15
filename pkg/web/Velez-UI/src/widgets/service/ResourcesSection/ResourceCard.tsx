import cn from 'classnames'
import {useNavigate} from 'react-router-dom'

import ResourceTypeIcon from '@/components/ResourceTypeIcon/ResourceTypeIcon.tsx'
import type {ServiceResource} from '@/model/service_page/ServicePageModel'
import {resourcePath} from '@/processes/resourceLink'
import ResourceCardActions from '@/widgets/service/ResourcesSection/ResourceCardActions.tsx'
import ResourceCardTitle from '@/widgets/service/ResourcesSection/ResourceCardTitle.tsx'

import cls from './ResourceCard.module.css'

interface ResourceCardProps {
    resource: ServiceResource
}

export default function ResourceCard({resource}: ResourceCardProps) {
    const navigate = useNavigate()
    const path = resourcePath(resource)
    const isClickable = path !== undefined

    function handleClick(e: React.MouseEvent) {
        if (path === undefined) {
            return
        }
        if (e.metaKey || e.ctrlKey) {
            window.open(path, "_blank", "noopener,noreferrer")
            return
        }
        navigate(path)
    }

    function handleAuxClick(e: React.MouseEvent) {
        if (path !== undefined && e.button === 1) {
            window.open(path, "_blank", "noopener,noreferrer")
        }
    }

    function handleKeyDown(e: React.KeyboardEvent) {
        if (path !== undefined && e.key === "Enter" && e.target === e.currentTarget) {
            navigate(path)
        }
    }

    return (
        <div
            className={cn(cls.ResourceCardContainer, isClickable && cls.Clickable)}
            style={{'--resource-color': resource.color} as React.CSSProperties}
            onClick={isClickable ? handleClick : undefined}
            onAuxClick={isClickable ? handleAuxClick : undefined}
            onKeyDown={isClickable ? handleKeyDown : undefined}
            role={isClickable ? "link" : undefined}
            tabIndex={isClickable ? 0 : undefined}
            title={isClickable ? `Open ${resource.name}` : undefined}
        >
            <div className={cls.HeaderWrapper}>
                <div className={cls.IconSquare}>
                    <ResourceTypeIcon type={resource.type} fallback={resource.icon}/>
                </div>
                <ResourceCardTitle resource={resource}/>
                <span className={cn(cls.StatusDot, statusDotClass(resource.status))}/>
            </div>
            <ResourceCardActions resource={resource}/>
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
