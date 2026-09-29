import { DockerContainer, ServicePattern, SmerdStatus } from '@/app/api/velez';
import { DotStatus } from '@/components/base/StatusDot';

// Label keys written by Velez on the Docker container — see
// internal/domain/labels/verv_labels.go, the single source of truth these
// strings must match.
const DISPLAY_NAME_LABEL = 'velez.display_name';
const VERV_SERVICE_LABEL = 'VERV_SERVICE';
const RUNNER_INSTANCE_LABEL = 'velez.runner';
const RUNNER_PROVIDER_LABEL = 'velez.runner.provider';
const PGAAS_INSTANCE_LABEL = 'velez.pgaas';
const REGISTRYAAS_INSTANCE_LABEL = 'velez.registryaas';

export interface ContainerServicePresentation {
    displayName: string;
    icon?: 'gitlab' | 'service';
}

export function deriveContainerServicePresentation(container: DockerContainer): ContainerServicePresentation {
    const labels = container.labels ?? {};
    const displayName = labels[DISPLAY_NAME_LABEL] || labels[VERV_SERVICE_LABEL] || container.linkedServiceName || '';

    if (labels[RUNNER_INSTANCE_LABEL] !== undefined) {
        const isGitlab = (labels[RUNNER_PROVIDER_LABEL] ?? '').toLowerCase() === 'gitlab';
        return {displayName, icon: isGitlab ? 'gitlab' : 'service'};
    }
    if (labels[PGAAS_INSTANCE_LABEL] !== undefined || labels[REGISTRYAAS_INSTANCE_LABEL] !== undefined) {
        return {displayName, icon: 'service'};
    }
    return {displayName};
}

const STATUS_TO_DOT: Record<SmerdStatus, DotStatus> = {
    [SmerdStatus.unknown]: 'offline',
    [SmerdStatus.created]: 'pending',
    [SmerdStatus.restarting]: 'degraded',
    [SmerdStatus.running]: 'running',
    [SmerdStatus.removing]: 'degraded',
    [SmerdStatus.paused]: 'disabled',
    [SmerdStatus.exited]: 'stopped',
    [SmerdStatus.dead]: 'error',
};

export function mapContainerStatusToDot(status?: SmerdStatus): DotStatus {
    if (!status) return 'offline';
    return STATUS_TO_DOT[status] ?? 'offline';
}

export function containerStatusLabel(status?: SmerdStatus): string {
    return status ?? 'unknown';
}

const SUGGESTED_PATTERN_LABELS: Partial<Record<ServicePattern, string>> = {
    [ServicePattern.SERVICE_PATTERN_POSTGRES]: 'PostgreSQL',
    [ServicePattern.SERVICE_PATTERN_REGISTRY]: 'a container registry',
    [ServicePattern.SERVICE_PATTERN_GITHUB_RUNNER]: 'a GitHub runner',
    [ServicePattern.SERVICE_PATTERN_GITLAB_RUNNER]: 'a GitLab runner',
};

export function suggestedPatternHint(pattern?: ServicePattern): string | undefined {
    const label = pattern ? SUGGESTED_PATTERN_LABELS[pattern] : undefined;
    return label ? `Looks like ${label}` : undefined;
}
