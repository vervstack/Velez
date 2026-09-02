import cls from '@/components/service/ServiceLabelBadge.module.css'
import Badge from '@/components/base/Badge'

interface ServiceLabelBadgeProps {
    labels?: string[]
}

interface BadgeSpec {
    label: string
    color: string
}

const RESOURCE_PREFIX = 'resource-'

function toBadgeSpec(label: string): BadgeSpec | null {
    if (label.startsWith(RESOURCE_PREFIX)) {
        return {label: label.slice(RESOURCE_PREFIX.length) || 'resource', color: 'var(--violet)'}
    }
    if (label === 'service-core') {
        return {label: 'core', color: 'var(--amber)'}
    }
    // service-app and anything unknown render nothing — keeps app cards clean.
    return null
}

export default function ServiceLabelBadge({labels}: ServiceLabelBadgeProps) {
    const specs: BadgeSpec[] = []
    for (const label of labels ?? []) {
        const spec = toBadgeSpec(label)
        if (spec) specs.push(spec)
    }

    if (specs.length === 0) return null

    return (
        <div className={cls.ServiceLabelBadgeContainer}>
            {specs.map((spec) => <Badge key={spec.label} label={spec.label} color={spec.color}/>)}
        </div>
    )
}
