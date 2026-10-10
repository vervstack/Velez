import { Smerd, SmerdStatus, ServiceBaseInfo } from '@/app/api/velez';
import { ServiceCardData } from '@/components/service/ServiceCard';

export interface AppData {
    name: string;
    image: string;
    status: 'running' | 'degraded' | 'stopped';
    env: string;
    incident: boolean;
    releaseFrozen: boolean;
    node: { id: string; host: string; status: 'online' | 'degraded' | 'offline' };
    deployments: number;
    lastDeployed: string;
    configSource: string;
    version: string;
}

export type ServiceIconKind = 'gitlab' | 'postgres' | 's3' | 'dind';

export interface ServiceListItem {
    name: string;
    displayName: string;
    subtitle?: string;
    icon?: ServiceIconKind;
    image: string;
    status: 'running' | 'degraded' | 'stopped';
    labels: string[];
    env: string;
    incident: boolean;
    releaseFrozen: boolean;
    node: { id: string; host: string; status: 'online' | 'degraded' | 'offline' };
    deployments: number;
    lastDeployed: string;
    configSource: string;
    version: string;
}

export const LOCAL_NODE = { id: 'local', host: 'localhost', status: 'online' as const };

// Label keys written by Velez on the Docker container — see
// internal/domain/labels/verv_labels.go, the single source of truth these
// strings must match.
const DISPLAY_NAME_LABEL = 'velez.display_name';
const VERV_SERVICE_LABEL = 'VERV_SERVICE';

export function deriveServiceDisplayName(s: Smerd): string {
    return s.labels?.[DISPLAY_NAME_LABEL] || s.labels?.[VERV_SERVICE_LABEL] || s.name || '';
}

export function mapSmerdToServiceCard(s: Smerd): ServiceCardData {
    return {
        name: s.name ?? '',
        displayName: deriveServiceDisplayName(s),
        image: s.imageName ?? '',
        status: mapSmerdStatus(s.status),
        cpu: 0,
        mem: 0,
        uptime: '-',
        restarts: 0,
        env: s.labels?.['env'] ?? 'prod',
        incident: false,
        releaseFrozen: false,
        node: LOCAL_NODE,
    };
}

export function mapSmerdToAppData(s: Smerd): AppData {
    return {
        name: s.name ?? '',
        image: s.imageName ?? '',
        status: mapSmerdStatus(s.status),
        env: s.labels?.['env'] ?? 'prod',
        incident: false,
        releaseFrozen: false,
        node: LOCAL_NODE,
        deployments: 0,
        lastDeployed: '-',
        configSource: 'none',
        version: s.imageName?.split(':')[1] ?? 'latest',
    };
}

function mapServiceStatus(status?: string): 'running' | 'degraded' | 'stopped' {
    const s = (status ?? '').toLowerCase();
    if (s.includes('run')) return 'running';
    if (s.includes('degrad') || s.includes('restart')) return 'degraded';
    return 'stopped';
}

function formatDeployedAt(ts?: { seconds?: string | number }): string {
    const seconds = Number(ts?.seconds ?? 0);
    if (!seconds) return '-';
    const diffDays = Math.floor((Date.now() - seconds * 1000) / 86400000);
    if (diffDays <= 0) return 'today';
    if (diffDays === 1) return 'yesterday';
    if (diffDays < 30) return diffDays + 'd ago';
    return new Date(seconds * 1000).toLocaleDateString();
}

// Derived labels the API attaches to GetService/ListServices.labels — see
// internal/domain/service_labels.go's ClassifyService, the single source of
// truth these strings must match.
const RUNNER_GITLAB_LABEL = 'service-runner-gitlab';
const PGAAS_LABEL = 'service-pgaas';
const S3_LABEL = 'service-s3';
const DIND_LABEL = 'service-dind';

function deriveServicePresentation(labels: string[]): { subtitle?: string; icon?: ServiceIconKind } {
    if (labels.includes(RUNNER_GITLAB_LABEL)) {
        return {subtitle: 'Gitlab Runner', icon: 'gitlab'};
    }
    if (labels.includes(PGAAS_LABEL)) {
        return {subtitle: 'Postgres', icon: 'postgres'};
    }
    if (labels.includes(S3_LABEL)) {
        return {subtitle: 'S3', icon: 's3'};
    }
    if (labels.includes(DIND_LABEL)) {
        return {subtitle: 'Docker in Docker', icon: 'dind'};
    }
    return {};
}

export function mapServiceToListItem(service: ServiceBaseInfo, smerds: Smerd[]): ServiceListItem {
    const name = service.name ?? '';
    const displayName = service.displayName || name;
    const relatedSmerd = smerds.find(
        (s) => s.name === name || (!!s.name && !!name && s.name.startsWith(name + '-'))
    );
    const image = relatedSmerd?.imageName ?? service.imageName ?? '';
    const status = relatedSmerd
        ? mapSmerdStatus(relatedSmerd.status)
        : mapServiceStatus(service.status);
    const labels = service.labels ?? [];
    const {subtitle, icon} = deriveServicePresentation(labels);

    return {
        name,
        displayName,
        subtitle,
        icon,
        image,
        status,
        labels: service.labels ?? [],
        env: service.env || 'prod',
        incident: false,
        releaseFrozen: false,
        node: LOCAL_NODE,
        deployments: 0,
        lastDeployed: formatDeployedAt(service.lastDeployedAt),
        configSource: 'none',
        version: image.split(':')[1] ?? 'latest',
    };
}

const TAG_LABEL_PREFIX = 'verv.tag.';

export interface SmerdTag {
    key: string;
    value: string;
}

export function getSmerdTags(smerd?: Smerd): SmerdTag[] {
    const labels = smerd?.labels ?? {};
    return Object.entries(labels)
        .filter(([key]) => key.startsWith(TAG_LABEL_PREFIX))
        .map(([key, value]) => ({key: key.slice(TAG_LABEL_PREFIX.length), value}));
}

function mapSmerdStatus(s?: SmerdStatus): 'running' | 'degraded' | 'stopped' {
    switch (s) {
        case SmerdStatus.running:
            return 'running';
        case SmerdStatus.restarting:
            return 'degraded';
        case SmerdStatus.dead:
        case SmerdStatus.exited:
        case SmerdStatus.paused:
        case SmerdStatus.removing:
            return 'stopped';
        default:
            return 'stopped';
    }
}
