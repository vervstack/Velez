import type {ServiceResource} from '@/model/service_page/ServicePageModel'
import {useGetServiceResourcesQuery} from '@/processes/queries/services'

import ResourceCard from './ResourceCard'
import cls from './ResourcesSection.module.css'

interface ResourcesSectionProps {
    serviceName: string
}

export default function ResourcesSection({serviceName}: ResourcesSectionProps) {
    const {data: resources = []} = useGetServiceResourcesQuery(serviceName)

    return (
        <div className={cls.ResourcesSectionContainer}>
            <ResourcesHeader/>
            {
                resources.length === 0 ?
                    <EmptyState/>
                    :
                    <ResourcesRow resources={resources}/>
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
                        key={resource.id}
                        resource={resource}/>
                )}
        </div>
    )
}
