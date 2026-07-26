import Badge from '@/components/base/Badge.tsx';
import {DeploymentStatus} from '@/app/api/velez';

interface DeploymentStatusBadgeProps {
    status?: DeploymentStatus;
}

const STATUS_COLOR: Record<DeploymentStatus, string> = {
    [DeploymentStatus.DEPLOYMENT_STATUS_UNKNOWN]: 'var(--fg-dim)',
    [DeploymentStatus.SCHEDULED_DEPLOYMENT]: 'var(--amber)',
    [DeploymentStatus.SCHEDULED_DELETION]: 'var(--amber)',
    [DeploymentStatus.SCHEDULED_UPGRADE]: 'var(--amber)',
    [DeploymentStatus.RUNNING]: 'var(--green)',
    [DeploymentStatus.FAILED]: 'var(--red)',
    [DeploymentStatus.DELETED]: 'var(--fg-dim)',
    [DeploymentStatus.STOPPED]: 'var(--fg-dim)',
};

const STATUS_DIM: Record<DeploymentStatus, string> = {
    [DeploymentStatus.DEPLOYMENT_STATUS_UNKNOWN]: 'transparent',
    [DeploymentStatus.SCHEDULED_DEPLOYMENT]: 'var(--amber-dim)',
    [DeploymentStatus.SCHEDULED_DELETION]: 'var(--amber-dim)',
    [DeploymentStatus.SCHEDULED_UPGRADE]: 'var(--amber-dim)',
    [DeploymentStatus.RUNNING]: 'var(--green-dim)',
    [DeploymentStatus.FAILED]: 'var(--red-dim)',
    [DeploymentStatus.DELETED]: 'transparent',
    [DeploymentStatus.STOPPED]: 'transparent',
};

export default function DeploymentStatusBadge({status}: DeploymentStatusBadgeProps) {
    const resolved = status ?? DeploymentStatus.DEPLOYMENT_STATUS_UNKNOWN;

    return (
        <Badge
            label={resolved}
            color={STATUS_COLOR[resolved]}
            dim={STATUS_DIM[resolved]}
        />
    );
}
