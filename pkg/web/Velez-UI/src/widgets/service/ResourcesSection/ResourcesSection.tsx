import {useMemo} from 'react'

import type {ServiceResource} from '@/model/service_page/ServicePageModel'
import {useGetServiceResourcesQuery, useGetVervonomiconQuery} from '@/processes/queries/services'
import {mergeResourcesWithReconciliation} from '@/processes/vervonomicon.ts'
import {getResourceMeta} from '@/processes/api/service.ts'
import {useEnvironmentStore} from '@/app/hooks/environment/Environment.ts'
import cls from '@/widgets/service/ResourcesSection/ResourcesSection.module.css'

import ResourceCard from './ResourceCard'

interface ResourcesSectionProps {
    serviceName: string
}

export default function ResourcesSection({serviceName}: ResourcesSectionProps) {
    const selectedEnvironment = useEnvironmentStore(state => state.selectedEnvironment)
    const {data: resources = []} = useGetServiceResourcesQuery(serviceName)
    const {data: docs} = useGetVervonomiconQuery(serviceName, selectedEnvironment)

    const mergedResources = useMemo(
        () => mergeResourcesWithReconciliation(resources, docs?.resourceStatuses ?? [], getResourceMeta),
        [resources, docs],
    )

    return (
        <div className={cls.ResourcesSectionContainer}>
            <ResourcesHeader/>
            {
                mergedResources.length === 0 ?
                    <EmptyState/>
                    :
                    <ResourcesRow resources={mergedResources}/>
            }
        </div>
    )
}

function ResourcesHeader() {
    return (
        <div className={cls.HeaderWrapper}>
            <h3 className={cls.Title}>Resources</h3>
            <p className={cls.Subtitle}>External dependencies attached to this service</p>
        </div>
    )
}

function EmptyState() {
    return (
        <p className={cls.Empty}>No resources configured</p>
    )
}


function ResourcesRow({resources}: { resources: ServiceResource[] }) {
    return (
        <div className={cls.Grid}>
            {resources
                .map((resource) =>
                    <ResourceCard
                        key={resource.name}
                        resource={resource}/>
                )}
        </div>
    )
}
