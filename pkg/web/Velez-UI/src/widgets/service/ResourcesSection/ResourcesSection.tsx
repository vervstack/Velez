import { useQuery } from '@tanstack/react-query'

import type { ServiceResource } from '@/model/service_page/ServicePageModel'
import { getServiceResourcesQuery } from '@/processes/queries/services'

import ResourceCard from './ResourceCard'
import cls from './ResourcesSection.module.css'

interface ResourcesSectionProps {
    serviceId: string
}

export default function ResourcesSection({ serviceId }: ResourcesSectionProps) {
    const { data: resources = [] } = useQuery(getServiceResourcesQuery(serviceId))

    return (
        <div className={cls.ResourcesSectionContainer}>
            <div className={cls.HeaderWrapper}>
                <h3 className={cls.Title}>Resources</h3>
                <p className={cls.Subtitle}>External dependencies attached to this service</p>
            </div>
            {resources.length === 0 ? (
                <p className={cls.Empty}>No resources configured</p>
            ) : (
                <div className={cls.Grid}>
                    {resources.map(function renderCard(resource: ServiceResource) {
                        return <ResourceCard key={resource.id} resource={resource} />
                    })}
                </div>
            )}
        </div>
    )
}
